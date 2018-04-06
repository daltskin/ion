package providers

import (
	apiv1 "k8s.io/api/core/v1"
	"strings"
	"testing"
)

func getDockerCommand() string {
	containers := []apiv1.Container{
		{
			Name:  "sidecar",
			Image: "doesntmatter",
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ionvolume",
					MountPath: "/ion",
				},
			},
		},
		{
			Name:            "worker",
			Image:           "doesntmatter",
			ImagePullPolicy: apiv1.PullAlways,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      "ionvolume",
					MountPath: "/ion",
				},
			},
		},
	}

	// Todo: Pull this out into a standalone package once stabilized
	podCommand, err := getPodCommand(batchPodComponents{
		Containers: containers,
		PodName:    mockMessageID,
		TaskID:     mockMessageID,
		Volumes: []apiv1.Volume{
			{
				Name: "ionvolume",
				VolumeSource: apiv1.VolumeSource{
					EmptyDir: &apiv1.EmptyDirVolumeSource{},
				},
			},
		},
	})

	if err != nil {
		panic(err)
	}

	return podCommand
}

func TestPod2DockerVolumeGenerated(t *testing.T) {

	podCommand := getDockerCommand()

	// Todo: Very basic smoke test that shared volume path is present in batch command
	if !strings.Contains(podCommand, "/ion") {
		t.Log(podCommand)
		t.Error("Missing shared volume")
	}

	// Todo: Very basic smoke test that shared volume path is present in batch command
	if !strings.Contains(podCommand, "docker volume create examplemessageID_ionvolume") {
		t.Log(podCommand)
		t.Error("Missing shared volume")
	}
}

func TestPod2DockerVolumeGenerated(t *testing.T) {

	podCommand := getDockerCommand()

	// Todo: Very basic smoke test that shared volume path is present in batch command
	if !strings.Contains(podCommand, "/ion") {
		t.Log(podCommand)
		t.Error("Missing shared volume")
	}

	// Todo: Very basic smoke test that shared volume path is present in batch command
	if !strings.Contains(podCommand, "docker volume create examplemessageID_ionvolume") {
		t.Log(podCommand)
		t.Error("Missing shared volume")
	}
}

func TestPod2DockerGeneratesValidOutputEncoding(t *testing.T) {

	podCommand := getDockerCommand()

	t.Log(podCommand)
	if strings.Contains(podCommand, "&lt;") {
		t.Error("output contains incorrect encoding")
	}
}

func TestExecutionOfPod2DockerCommand(t *Testing.T) {
	panic("Finish me!")
}
