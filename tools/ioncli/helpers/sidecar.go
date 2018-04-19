package helpers

import (
	"os/exec"
)

// SidecarRunner Runs and controls a sidecar process
type SidecarRunner struct {
	binaryLocation string
	cmd            *exec.Cmd
}

func NewSidecarRunner(binaryLocation, eventdir string) (runner SidecarRunner, err error) {
	runner = SidecarRunner{binaryLocation: binaryLocation}
	cmd := exec.Command(binaryLocation)
	runner.cmd = cmd
	return
}

// func getArgsFromEventDir(eventdir string) (string[], error) {

// }
