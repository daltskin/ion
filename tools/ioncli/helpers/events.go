package helpers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/lawrencegripper/ion/common"

	"github.com/lawrencegripper/ion/tools/ioncli/types"
)

const blobFolder = "blobs"
const metadataFileName = "metadata.json"
const argsFileName = ".args"

//SaveEvent will persist the event as it would be in the "ion/in" folder of a module receiving it.
func SaveEvent(moduleName, sidecarbasedir, savedir string, eventBundle types.EventBundle) error {
	eventFolder := filepath.Join(savedir, getSaveEventFolderName(eventBundle))
	err := os.MkdirAll(eventFolder, 0777)
	if err != nil {
		return err
	}

	absBlobFolder := filepath.Join(eventFolder, blobFolder)
	err = os.MkdirAll(absBlobFolder, 0777)
	if err != nil {
		return err
	}

	//copy all referenced blob folders
	for _, absFilePath := range eventBundle.DataFiles {
		_, filename := filepath.Split(absFilePath)
		CopyFile(absFilePath, filepath.Join(absBlobFolder, filename))
	}

	//create metadata
	data := eventBundle.EventContext.Data
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(eventFolder, metadataFileName), bytes, 0777)
	if err != nil {
		return err
	}

	//create args file
	// it still needs.... --valideventtypes=face_detected --context.name=modulename
	args := "--loglevel=debug --development --printconfig --sharedsecret=dev " +
		"--context.eventid=" + eventBundle.EventContext.EventID + " "

	argsFilePath := filepath.Join(eventFolder, argsFileName)
	err = ioutil.WriteFile(argsFilePath, []byte(args), 0777)
	if err != nil {
		return err
	}

	return nil
}

//GetEventsFromStore Lists the events available in the local store
func GetEventsFromStore(storeDir string) ([]types.SavedEventInfo, error) {
	files, err := ioutil.ReadDir(storeDir)
	if err != nil {
		return []types.SavedEventInfo{}, err
	}

	events := make([]types.SavedEventInfo, 0, len(files))

	for _, f := range files {
		modulename, eventtype, eventID := getDetailsFromSaveFolderName(f.Name())
		absFolderPath, err := filepath.Abs(filepath.Join(storeDir, f.Name()))
		if err != nil {
			return []types.SavedEventInfo{}, err
		}
		events = append(events, types.SavedEventInfo{
			EventID:       eventID,
			EventType:     eventtype,
			ModuleName:    modulename,
			AbsFolderPath: absFolderPath,
		})
	}

	return events, nil
}

//GetArgs gets the sidecar args for a stored event
func GetArgs(s types.SavedEventInfo) ([]string, error) {
	file, err := os.Open(filepath.Join(s.AbsFolderPath, argsFileName))
	if err != nil {
		return []string{}, err
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return []string{}, err
	}

	args := strings.Split(string(bytes), " ")

	return args, nil
}

//GetEventsFromDev enumerates the dev output from the sidecar and extracts event bundles
func GetEventsFromDev(devDir string) (map[string]types.EventBundle, error) {
	bundles := map[string]types.EventBundle{}

	files, err := filepath.Glob(filepath.Join(devDir, "/*/*.event_event*.json"))
	if err != nil {
		log.Fatal(err)
		return bundles, err
	}

	for _, eventFilename := range files {
		event := &common.Event{}
		dir, err := readFile(eventFilename, event)
		if err != nil {
			return bundles, err
		}

		contextFileName := strings.Replace(eventFilename, "event_", "context_", -1)
		context := &types.EventContext{}
		_, err = readFile(contextFileName, context)
		if err != nil {
			return bundles, err
		}

		dataFiles, err := getDataFilePaths(context, dir)
		if err != nil {
			return bundles, err
		}

		bundles[event.Context.EventID] = types.EventBundle{
			Event:        event,
			EventContext: context,
			DataFiles:    dataFiles,
		}
	}

	return bundles, nil
}

func getSaveEventFolderName(e types.EventBundle) string {
	return strings.Join([]string{e.Event.Name, e.Event.Type, e.Event.EventID}, "__")
}

func getDetailsFromSaveFolderName(foldername string) (modulename, eventtype, eventID string) {
	parts := strings.Split(foldername, "__")
	if len(parts) > 3 {
		panic("Saved event Folder name incorrectly formatted")
	}
	return parts[0], parts[1], parts[2]
}

func getDataFilePaths(eventContext *types.EventContext, dir string) ([]string, error) {
	blobDir := filepath.Join(dir, blobFolder)
	files, err := ioutil.ReadDir(blobDir)
	if err != nil {
		return []string{}, err
	}

	lookup := map[string]*string{}
	for _, exportedFileName := range eventContext.Files {
		lookup[exportedFileName] = nil
	}
	result := make([]string, 0, len(lookup))

	for _, blobFile := range files {
		if blobFile.IsDir() {
			continue
		}
		filename := blobFile.Name()
		fullFilePath, err := filepath.Abs(filepath.Join(blobDir, filename))
		if err != nil {
			return []string{}, err
		}
		_, exists := lookup[filename]
		if exists {
			result = append(result, fullFilePath)
		}
	}

	return result, nil

}

func readFile(filename string, toType interface{}) (directory string, err error) {
	fileReader, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	bytes, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(bytes, toType)
	if err != nil {
		return "", err
	}

	dir, _ := filepath.Split(filename)
	return dir, nil
}
