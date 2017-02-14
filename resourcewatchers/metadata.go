package resourcewatchers

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/niusmallnan/helper"
	"github.com/rancher/go-rancher-metadata/metadata"
)

func WatchMetadata(client metadata.Client, updater helper.LoggingFileUpdater) error {
	logrus.Infof("Subscribing to metadata changes.")

	selfHost, err = w.client.GetSelfHost()
	if err != nil {
		logrus.Fatalf("Can not get self host UUID from rancher metadata.")
	}

	watcher := metadataWatcher{
		client:      client,
		hostUUID:    selfHost.UUID,
		fileUpdater: updater,
	}
	return client.OnChangeWithError(5, watcher.updateFromMetadata)
}

type metadataWatcher struct {
	client                metadata.Client
	consecutiveErrorCount int
	hostUUID              string
	fileUpdater           helper.LoggingFileUpdater
}

func (w *metadataWatcher) updateFromMetadata(mdVersion string) {
	containers, err := w.client.GetContainers()
	if err != nil {
		w.checkError(err)
	}

	for _, c := range containers {
		if c.HostUUID == w.hostUUID && c.State == "running" {
			containerID = c.ExternalId
		}
	}
}

func (w *metadataWatcher) checkError(err error) {
	w.consecutiveErrorCount++
	if w.consecutiveErrorCount > 5 {
		panic(fmt.Sprintf("%v consecutive errors attempting to reach metadata. Panicing. Error: %v", w.consecutiveErrorCount, err))
	}
	logrus.Errorf("Error %v getting metadata: %v", w.consecutiveErrorCount, err)
}
