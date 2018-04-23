package helpers

import (
	"net/http"
	"os"
	"os/user"
	"path"
	"sync"
	"testing"
	"time"
)

func TestSidcarRunner_blank(t *testing.T) {
	defer os.RemoveAll(".dev")
	usr, _ := user.Current()
	ionDir := path.Join(usr.HomeDir, ".ion")

	runner, err := NewBlankSidecar("./../../../sidecar/sidecar", ionDir, "testmodule", "face_detected")
	if err != nil {
		t.Error(err)
	}

	runner.Start()
	time.Sleep(time.Second * 3)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		result, err := runner.Wait()
		t.Log(result)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/ready", nil)
	req.Close = true
	req.Header.Add("secret", "dev")
	res, err := client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		t.Error("Ready call failed")
		return
	}

	req, _ = http.NewRequest("GET", "http://localhost:8080/done", nil)
	req.Close = true
	req.Header.Add("secret", "dev")
	res, err = client.Do(req)
	if err != nil {
		t.Error(err)
		return
	}
	if res.StatusCode != http.StatusOK {
		t.Error("done call failed")
		return
	}

	wg.Wait()
}

func TestSidcarRunner_existing(t *testing.T) {
	defer os.RemoveAll(".dev")
	usr, _ := user.Current()
	ionDir := path.Join(usr.HomeDir, ".ion")

	events, err := GetEventsFromStore("testdata/.store")
	if err != nil {
		t.Error(err)
		return
	}
	event := events[0]
	runner, err := NewSidecarRunnerFromEvent("./../../../sidecar/sidecar", ionDir, "testmodule", "face_detected", event)
	if err != nil {
		t.Error(err)
		return
	}

	runner.Start()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		result, err := runner.Wait()
		t.Log(result)
		if err != nil {
			t.Error(err)
		}
		wg.Done()
	}()
	time.Sleep(time.Second * 4)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8080/ready", nil)
	req.Close = true
	req.Header.Add("secret", "dev")
	res, err := client.Do(req)
	if err != nil {
		t.Error(err)
		t.Log(runner.Logs())
	}
	if res.StatusCode != http.StatusOK {
		t.Error("Ready call failed")
	}

	req, _ = http.NewRequest("GET", "http://localhost:8080/done", nil)
	req.Close = true
	req.Header.Add("secret", "dev")
	res, err = client.Do(req)
	if err != nil {
		t.Error(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Error("done call failed")
	}

	wg.Wait()
}
func TestSidcarRunner_missingBinary(t *testing.T) {
	defer os.RemoveAll(".dev")
	usr, _ := user.Current()
	ionDir := path.Join(usr.HomeDir, ".ion")

	events, err := GetEventsFromStore("testdata/.store")
	if err != nil {
		t.Error(err)
		return
	}
	event := events[0]
	runner, err := NewSidecarRunnerFromEvent("doesntexist", ionDir, "testmodule", "face_detected", event)
	if err != nil {
		t.Error(err)
		return
	}

	err = runner.Start()
	_, err = runner.Wait()
	if err == nil {
		t.Fail()
	}
}
