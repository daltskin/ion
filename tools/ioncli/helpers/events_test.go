package helpers

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/lawrencegripper/ion/common"
	"github.com/lawrencegripper/ion/tools/ioncli/types"
)

const event0ID = "d55a6cf9-665b-4f4e-9d64-c18d6c97fb65"

func TestGetEvents_ReturnsExpectedEventCount(t *testing.T) {
	events, err := GetEventsFromDev("./testdata/.dev")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if len(events) != 5 {
		t.Errorf("Expected: 5 events actual: %v", len(events))
	}
}

func TestGetEvents_SpecificEventHasCorrectValues(t *testing.T) {
	events, err := GetEventsFromDev("./testdata/.dev")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	event := events[event0ID]

	if event.Event == nil || event.EventContext == nil {
		t.Errorf("Event object invalid: %+v", event)
	}

	if event.Event.EventID != event0ID && event.EventContext.EventID != event0ID {
		t.Error("Event ID not set correctly")
	}
}

func TestGetEvents_HasCorrentFullFilePaths(t *testing.T) {
	events, err := GetEventsFromDev("./testdata/.dev")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	event := events[event0ID]

	fullPath := event.DataFiles[0]
	t.Log(fullPath)
	file, err := os.Open(fullPath)
	if err != nil {
		t.Error(err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		t.Error(err)
	}
	if string(bytes) != "face!" {
		t.Errorf("File content not as expected got: %v", string(bytes))
	}
}

func TestSaveFolderNames_CreateAndExtract(t *testing.T) {
	eventBundle := types.EventBundle{
		Event: &common.Event{
			Context: &common.Context{
				Name:    "modulename",
				EventID: "eventid",
			},
			Type: "eventtype",
		},
	}

	foldername := getSaveEventFolderName(eventBundle)
	moduleName, eventType, eventid := getDetailsFromSaveFolderName(foldername)

	if moduleName != eventBundle.Event.Name || eventType != eventBundle.Event.Type || eventBundle.Event.EventID != eventid {
		t.Log(foldername)
		t.Error("Failed to pull out correct details")
	}
}

func TestSaveEvent_CreatesArgs_CopysBlobs_CreatesMeta(t *testing.T) {
	//Read an existing event from the test data
	events, err := GetEventsFromDev("./testdata/.dev")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	event := events[event0ID]

	// Test the method for save
	tmpdir := os.TempDir()
	err = SaveEvent("testmodule", "./testdata", tmpdir, event)
	if err != nil {
		t.Error(err)
	}

	tmpdir = filepath.Join(tmpdir, getSaveEventFolderName(event))
	//Cleanup
	defer os.Remove(tmpdir)
	//Check files exist
	_, err = os.Stat(filepath.Join(tmpdir, "metadata.json"))
	if err != nil {
		t.Error(err)
	}

	blobdir, err := os.Stat(filepath.Join(tmpdir, "blobs"))
	if err != nil {
		t.Error(err)
	}
	if !blobdir.IsDir() {
		t.Error("Blob dir not created")
	}

	_, err = os.Stat(filepath.Join(tmpdir, "blobs", "image0.png"))
	if err != nil {
		t.Error(err)
	}
}

func TestGetEventsFromStore_ListsEvents(t *testing.T) {
	//Read an existing event from the test data
	events, err := GetEventsFromDev("./testdata/.dev")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	event := events[event0ID]

	//Save it
	tmpdir := filepath.Join(os.TempDir(), "ionstore")
	err = os.MkdirAll(tmpdir, 0777)
	defer os.RemoveAll(tmpdir)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	err = SaveEvent("testmodule", "./testdata", tmpdir, event)
	if err != nil {
		t.Error(err)
	}

	//Test the method for list
	results, err := GetEventsFromStore(tmpdir)

	if err != nil {
		t.Error(err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 item got: %+v", len(results))
	}

	t.Log(results)
}
