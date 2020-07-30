package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/qlik-oss/corectl/pkg/log"
	"github.com/qlik-oss/enigma-go"
)

// Object is a struct describing the generic object
type Object struct {
	Info       *enigma.NxInfo                  `json:"qInfo,omitempty"`
	Properties *enigma.GenericObjectProperties `json:"qProperty,omitempty"`
}

type Layout struct {
	Meta *ObjectMeta `json:"qMeta"`
}
type ObjectMeta struct {
	Published bool `json:"published"`
}

func (o Object) validate() error {
	if o.Info != nil {
		if o.Info.Id == "" {
			return errors.New("missing qInfo qId attribute")
		}
		if o.Info.Type == "" {
			return errors.New("missing qInfo qType attribute")
		}
	} else if o.Properties != nil {
		if o.Properties.Info == nil {
			return errors.New("missing qInfo attribute inside the qProperty")
		}
		if o.Properties.Info.Id == "" {
			return errors.New("missing qInfo qId attribute inside qProperty")
		}
		if o.Properties.Info.Type == "" {
			return errors.New("missing qInfo qType attribute inside qProperty")
		}
	} else {
		return errors.New("need to supply atleast one of qInfo or qProperty")
	}
	return nil
}

// ListObjects fetches all generic objects and returns them sorted in an array
func ListObjects(ctx context.Context, doc *enigma.Doc) []NamedItemWithType {
	allInfos, _ := doc.GetAllInfos(ctx)
	unsortedResult := make(map[string]*NamedItemWithType)
	keys := []string{}

	waitChannel := make(chan *NamedItemWithType)
	defer close(waitChannel)

	for _, item := range allInfos {
		go func(item *enigma.NxInfo) {
			object, _ := doc.GetObject(ctx, item.Id)
			if object != nil && object.Type != "" {
				rawProps, _ := object.GetPropertiesRaw(ctx)
				propsWithTitle := &PropsWithTitle{}
				json.Unmarshal(rawProps, propsWithTitle)
				waitChannel <- &NamedItemWithType{Title: propsWithTitle.Meta.Title, ID: item.Id, Type: item.Type}
			} else {
				waitChannel <- nil
			}
		}(item)
	}
	//Put all responses into a map by their Id
	for range allInfos {
		item := <-waitChannel
		if item != nil {
			keys = append(keys, item.ID)
			unsortedResult[item.ID] = item
		}
	}
	//Loop over the keys that are sorted on qId and fetch the result for each object
	sort.Strings(keys)
	resultInSortedOrder := make([]NamedItemWithType, len(keys))
	for i, key := range keys {
		resultInSortedOrder[i] = *unsortedResult[key]
	}
	return resultInSortedOrder
}

// SetObjects creates or updates all objects on given glob patterns
func SetObjects(ctx context.Context, doc *enigma.Doc, paths []string) {
	for _, path := range paths {
		rawEntities, err := parseEntityFile(path)
		if err != nil {
			log.Fatalf("could not parse file %s: %s\n", path, err)
		}

		// Run in parallel
		ch := make(chan error)

		for _, raw := range rawEntities {
			go func(raw json.RawMessage) {
				var object Object
				err = json.Unmarshal(raw, &object)
				if err != nil {
					ch <- fmt.Errorf("could not parse data in file %s: %s", path, err)
					return
				}
				err = object.validate()
				if err != nil {
					ch <- fmt.Errorf("validation error in file %s: %s", path, err)
					return
				}
				ch <- setObject(ctx, doc, object.Info, object.Properties, raw)
			}(raw)
		}

		// Loop through the responses and see if there are any failures, if so exit with a fatal
		success := true
		for range rawEntities {
			err := <-ch
			if err != nil {
				log.Errorln(err)
				success = false
			}
		}

		if !success {
			log.Fatalln("One or more objects failed to be created or updated")
		}
	}
}

func setObject(ctx context.Context, doc *enigma.Doc, info *enigma.NxInfo, props *enigma.GenericObjectProperties, raw json.RawMessage) error {
	var objectID string
	isGenericObjectEntry := false
	if info != nil {
		objectID = info.Id
	} else {
		objectID = props.Info.Id
		isGenericObjectEntry = true
	}
	object, err := doc.GetObject(ctx, objectID)
	if err != nil {
		return err
	}
	if object.Handle != 0 {
		if isGenericObjectEntry {
			log.Verboseln("Updating object " + objectID + " using SetFullPropertyTree")
			err = object.SetFullPropertyTreeRaw(ctx, raw)
		} else {
			log.Verboseln("Updating object " + objectID + " using SetProperties")
			err = object.SetPropertiesRaw(ctx, raw)
		}
		if err != nil {
			return fmt.Errorf("failed to update %s %s: %s", "object", objectID, err)
		}
	} else {
		log.Verboseln("Creating object " + objectID)
		if isGenericObjectEntry {
			var createdObject *enigma.GenericObject
			objectType := props.Info.Type
			createdObject, err = doc.CreateObject(ctx, &enigma.GenericObjectProperties{Info: &enigma.NxInfo{Id: objectID, Type: objectType}})
			log.Verboseln("Setting object  " + objectID + " using SetFullPropertyTree")
			err = createdObject.SetFullPropertyTreeRaw(ctx, raw)
		} else {
			_, err = doc.CreateObjectRaw(ctx, raw)
		}
		if err != nil {
			return fmt.Errorf("failed to create %s %s: %s", "object", objectID, err)
		}
	}
	return nil
}

func Publish(ctx context.Context, doc *enigma.Doc, objectID string) error {
	object, err := doc.GetObject(ctx, objectID)
	if err != nil {
		return err
	}
	if object.Handle != 0 {
		var layout Layout
		var raw json.RawMessage
		raw, err = object.GetLayoutRaw(ctx)
		err = json.Unmarshal(raw, &layout)
		if layout.Meta.Published {
			log.Infoln(objectID + " is published.")
			return fmt.Errorf("Cannot publish a published object: " + objectID)
		} else {
			log.Infoln("Publishing object " + objectID)
			err = object.Publish(ctx)
			if err != nil {
				return fmt.Errorf("Unable to publish %s with %s: %s", "object", objectID, err)
			}
		}
	}
	return nil
}

func UnPublish(ctx context.Context, doc *enigma.Doc, objectID string) error {
	object, err := doc.GetObject(ctx, objectID)
	if err != nil {
		return err
	}
	if object.Handle != 0 {
		var layout Layout
		var raw json.RawMessage
		raw, err = object.GetLayoutRaw(ctx)
		err = json.Unmarshal(raw, &layout)
		if !layout.Meta.Published {
			log.Infoln(objectID + " is not published.")
			return fmt.Errorf("Cannot unpublish an unpublished object: " + objectID)
		} else {
			log.Infoln("Unpublishing object " + objectID)
			err = object.UnPublish(ctx)
			if err != nil {
				return fmt.Errorf("Unable to unpublish %s with %s: %s", "object", objectID, err)
			}
		}
	}
	return nil
}
