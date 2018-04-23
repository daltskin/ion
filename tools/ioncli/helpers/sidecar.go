package helpers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/lawrencegripper/ion/tools/ioncli/types"
)

// SidecarRunner Runs and controls a sidecar process
type SidecarRunner struct {
	binaryLocation  string
	cmd             *exec.Cmd
	combindedOutput *bytes.Buffer
}

//NewBlankSidecar creates a sidecar for the first event in a pipeline with no input data.
func NewBlankSidecar(binaryLocation, iondir, workingdir, modulename, publishesevents string) (runner SidecarRunner, err error) {
	runner = SidecarRunner{binaryLocation: binaryLocation}
	cmd := exec.Command(binaryLocation)
	cmd.Dir = workingdir
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	runner.combindedOutput = &b

	runner.cmd = cmd
	args := []string{
		"--loglevel=debug",
		"--development",
		"--context.correlationid=corrid1",
		"--context.eventid=eventid1",
		"--printconfig", "--sharedsecret=dev",
		"--basedir=" + iondir,
	}
	args = append(args, "--valideventtypes="+publishesevents)
	args = append(args, "--context.name="+modulename)
	runner.cmd.Args = args
	return
}

//NewSidecarRunnerFromEvent Responsible for starting and watching the sidecar with the right event information.
func NewSidecarRunnerFromEvent(binaryLocation, iondir, workingdir, modulename, publishesevents string, eventContext types.SavedEventInfo) (runner SidecarRunner, err error) {
	runner = SidecarRunner{binaryLocation: binaryLocation}
	cmd := exec.Command(binaryLocation)
	cmd.Dir = workingdir
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	runner.combindedOutput = &b

	runner.cmd = cmd

	args, err := GetArgs(eventContext)
	if err != nil {
		return
	}
	args = append(args, "--valideventtypes="+publishesevents)
	args = append(args, "--context.name="+modulename)
	args = append(args, "--basedir="+iondir)
	args = append(args, "--context.correlationid=corrid1")
	runner.cmd.Args = args

	inDir := filepath.Join(iondir, "in")
	err = os.MkdirAll(inDir, 0777)
	if err != nil {
		return
	}
	CopyDir(eventContext.AbsFolderPath, inDir)

	return
}

//Start starts the sidecar
func (runner *SidecarRunner) Start() error {
	return runner.cmd.Start()
}

//Logs gets logs from stdio
func (runner *SidecarRunner) Logs() string {
	return string(runner.combindedOutput.Bytes())
}

//Wait wait until "Done" is called by the module
// then return the output from the stdio
func (runner *SidecarRunner) Wait() (string, error) {
	defer os.RemoveAll(filepath.Join(runner.cmd.Dir, ".dev"))

	if runner.cmd.Process == nil {
		return "", fmt.Errorf("process not started for sidecar or crashed")
	}
	for {
		if runner.cmd.ProcessState != nil && runner.cmd.ProcessState.Exited() {
			return string(runner.combindedOutput.Bytes()), fmt.Errorf("Process completed with Exited when not expected")
		}
		glob := filepath.Join(runner.cmd.Dir, ".dev", "*", "dev.done")
		fmt.Println("Checking glob: " + glob)
		files, err := filepath.Glob(glob)
		if err != nil {
			return "", err
		}
		if len(files) > 0 {
			err = runner.cmd.Process.Kill()
			if err != nil {
				return "", err
			}
			return string(runner.combindedOutput.Bytes()), nil
		}
		fmt.Println("Sidecar running, waiting for `done` to be called...")
		if runner.cmd.Process != nil {
			fmt.Printf("PID: %v", runner.cmd.Process.Pid)
		}
		time.Sleep(10 * time.Second)
	}

}
