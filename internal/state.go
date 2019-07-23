package internal

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/qlik-oss/corectl/internal/log"
	"github.com/qlik-oss/enigma-go"
	"github.com/spf13/viper"
)

// State contains all needed info about the current app including a go context to use when communicating with the engine.
type State struct {
	Doc     *enigma.Doc
	Ctx     context.Context
	Global  *enigma.Global
	AppName string
	AppID   string
	Verbose bool
}

func logConnectError(err error, engine string) {
	msg := fmt.Sprintf("could not connect to engine on %s\nDetails: %s\n", engine, err)
	if strings.Contains(err.Error(), "401") {
		msg += fmt.Sprintln("This probably means that you have provided either incorrect or no authorization credentials.")
		msg += fmt.Sprintln("Check that the headers specified are correct.")
	} else {
		msg += fmt.Sprintln("This probably means that there is no engine running on the specified url.")
		msg += fmt.Sprintln("Check that the engine is up and that the url specified is correct.")
	}
	log.Fatalln(msg)
}

func connectToEngine(ctx context.Context, appName, ttl string, headers http.Header, certificates *tls.Config) *enigma.Global {
	engineURL := buildWebSocketURL(ttl)
	log.Debugln("Engine: " + engineURL)

	if headers.Get("X-Qlik-Session") == "" {
		sessionID := getSessionID(appName)
		log.Debugln("SessionId: " + sessionID)
		headers.Set("X-Qlik-Session", sessionID)
	}
	log.Debugln("SessionId " + headers.Get("X-Qlik-Session"))

	var dialer = enigma.Dialer{}

	if log.Traffic {
		dialer.TrafficLogger = log.TrafficLogger{}
	}

	if certificates != nil {
		dialer.TLSClientConfig = certificates
	}

	global, err := dialer.Dial(ctx, engineURL, headers)
	if err != nil {
		logConnectError(err, engineURL)
	}
	return global
}

//AppExists returns wether or not an app exists along with any eventual error from engine
func AppExists(ctx context.Context, engine string, appName string, headers http.Header, certificates *tls.Config) (bool, error) {
	global := PrepareEngineStateWithoutApp(ctx, headers, certificates).Global
	appID, _ := applyNameToIDTransformation(appName)
	_, err := global.GetAppEntry(ctx, appID)
	if err != nil {
		return false, fmt.Errorf("could not find any app by ID '%s': %s", appID, err)
	}
	return true, nil
}

//DeleteApp removes the specified app from the engine.
func DeleteApp(ctx context.Context, engine string, appName string, headers http.Header, certificates *tls.Config) {
	global := PrepareEngineStateWithoutApp(ctx, headers, certificates).Global
	appID, _ := applyNameToIDTransformation(appName)
	succ, err := global.DeleteApp(ctx, appID)
	if err != nil {
		log.Fatalf("could not delete app with name '%s' and ID '%s': %s", appName, appID, err)
	} else if !succ {
		log.Fatalf("could not delete app with name '%s' and ID '%s'", appName, appID)
	}
	SetAppIDToKnownApps(appName, appID, true)
}

// PrepareEngineState makes sure that the app idenfied by the supplied parameters is created or opened or reconnected to
// depending on the state. The TTL feature is used to keep the app session loaded to improve performance.
func PrepareEngineState(ctx context.Context, headers http.Header, certificates *tls.Config, createAppIfMissing bool) *State {
	engine := viper.GetString("engine")
	appName := viper.GetString("app")
	ttl := viper.GetString("ttl")
	noData := viper.GetBool("no-data")

	if appName == "" {
		// No app name provided, lets check if one exists in the url
		appName = TryParseAppFromURL(engine)
		if appName == "" {
			log.Fatalln("no app specified")
		}
	}

	appID, _ := applyNameToIDTransformation(appName)

	log.Debugln("---------- Connecting to app ----------")
	global := connectToEngine(ctx, appName, ttl, headers, certificates)
	sessionMessages := global.SessionMessageChannel()
	err := waitForOnConnectedMessage(sessionMessages)
	if err != nil {
		log.Fatalln("could not connect to engine: ", err)
	}
	go printSessionMessagesIfInVerboseMode(sessionMessages)
	doc, err := global.GetActiveDoc(ctx)
	if doc != nil {
		// There is an already opened doc!
		log.Debugln("App with name: " + appName + " and id: " + appID + "(reconnected)")
	} else {
		doc, err = global.OpenDoc(ctx, appID, "", "", "", noData)
		if doc != nil {
			if noData {
				log.Debugln("Opened app with name: " + appName + " and id: " + appID + " without data")
			} else {
				log.Debugln("Opened app with name: " + appName + " and id: " + appID)
			}
		} else if createAppIfMissing {
			var success bool
			success, appID, err = global.CreateApp(ctx, appName, "")
			if err != nil {
				log.Fatalf("could not create app with name '%s': %s", appName, err)
			}
			if !success {
				log.Fatalf("could not create app with name '%s'", appName)
			}
			// Write app id to config
			SetAppIDToKnownApps(appName, appID, false)
			doc, err = global.OpenDoc(ctx, appID, "", "", "", noData)
			if err != nil {
				log.Fatalf("could not do open app with ID '%s': %s", appID, err)
			}
			if doc != nil {
				log.Debugln("App with name: " + appName + " and id: " + appID + "(new)")
			}
		} else {
			log.Fatalf("could not open app with ID '%s': %s", appID, err)
		}
	}

	return &State{
		Doc:     doc,
		Global:  global,
		AppName: appName,
		AppID:   appID,
		Ctx:     ctx,
	}
}

func waitForOnConnectedMessage(sessionMessages chan enigma.SessionMessage) error {
	for sessionEvent := range sessionMessages {
		log.Debugln(sessionEvent.Topic + " " + string(sessionEvent.Content))
		if sessionEvent.Topic == "OnConnected" {
			var parsedEvent map[string]string
			err := json.Unmarshal(sessionEvent.Content, &parsedEvent)
			if err != nil {
				log.Fatalln("could not parse response from engine: ", err)
			}
			if parsedEvent["qSessionState"] == "SESSION_CREATED" || parsedEvent["qSessionState"] == "SESSION_ATTACHED" {
				return nil
			}
			return errors.New(parsedEvent["qSessionState"])
		}
	}
	return errors.New("session closed before reciving OnConnected message")
}

func printSessionMessagesIfInVerboseMode(sessionMessages chan enigma.SessionMessage) {
	for sessionEvent := range sessionMessages {
		log.Debugln(sessionEvent.Topic + " " + string(sessionEvent.Content))
	}
}

// PrepareEngineStateWithoutApp creates a connection to the engine with no dependency to any app.
func PrepareEngineStateWithoutApp(ctx context.Context, headers http.Header, certificates *tls.Config) *State {
	ttl := viper.GetString("ttl")

	log.Debugln("---------- Connecting to engine ----------")

	engineURL := buildWebSocketURL(ttl)

	log.Debugln("Engine: " + engineURL)

	var dialer = enigma.Dialer{}

	if log.Traffic {
		dialer.TrafficLogger = log.TrafficLogger{}
	}

	if certificates != nil {
		log.Debugln("Using certificates in: " + viper.GetString("certificates"))
		dialer.TLSClientConfig = certificates
	}

	global, err := dialer.Dial(ctx, engineURL, headers)

	if err != nil {
		logConnectError(err, engineURL)
	}
	sessionMessages := global.SessionMessageChannel()
	err = waitForOnConnectedMessage(sessionMessages)
	if err != nil {
		log.Fatalln("could not connect to engine: ", err)
	}
	go printSessionMessagesIfInVerboseMode(sessionMessages)

	return &State{
		Doc:     nil,
		Global:  global,
		AppName: "",
		AppID:   "",
		Ctx:     ctx,
	}
}

func getSessionID(appID string) string {
	// If no-data or ttl flag is used the user should not get the default session id
	noData := viper.GetBool("no-data")
	ttl := viper.GetString("ttl")

	currentUser, err := user.Current()
	if err != nil {
		log.Fatalln("unexpected error when retrieving current user: ", err)
	}
	hostName, err := os.Hostname()
	if err != nil {
		log.Fatalln("unexpected error when retrieving hostname: ", err)
	}
	sessionID := base64.StdEncoding.EncodeToString([]byte("corectl-" + currentUser.Username + "-" + hostName + "-" + appID + "-" + ttl + "-" + strconv.FormatBool(noData)))
	return sessionID
}
