package types

import (
	"github.com/lawrencegripper/ion/common"
)

//EventContext is a single entry in a document
type EventContext struct {
	*common.Context
	ParentEventID string               `bson:"parentEventId" json:"parentEventId"`
	Files         []string             `bson:"files" json:"files"`
	Data          common.KeyValuePairs `bson:"data" json:"data"`
}

//EventBundle Wraps the event context and event into a single object
type EventBundle struct {
	Event        *common.Event
	EventContext *EventContext
	DataFiles    []string
}
