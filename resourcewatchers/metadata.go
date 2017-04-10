package resourcewatchers

import (
	"fmt"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/niusmallnan/logging-helper/helper"
	"github.com/rancher/go-rancher-metadata/metadata"
)

func WatchMetadata(client metadata.Client, updater helper.LoggingFileUpdater) error {
	logrus.Infof("Subscribing to metadata changes.")

	selfHost, err := client.GetSelfHost()
	if err != nil {
		logrus.Fatalf("Can not get self host UUID from rancher metadata.")
	}

	watcher := &metadataWatcher{
		client:      client,
		hostUUID:    selfHost.UUID,
		fileUpdater: updater,
	}
	return client.OnChangeWithError(5, watcher.updateFromMetadata)
}

type metadataWatcher struct {
	client                metadata.Client
	consecutiveErrorCount int
	updateCount           int
	hostUUID              string
	fileUpdater           helper.LoggingFileUpdater
	mu                    sync.Mutex
}

func (w *metadataWatcher) updateFromMetadata(mdVersion string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	stack, err := w.client.GetSelfStack()
	if err != nil {
		w.checkError(err)
	}

	containers, err := w.client.GetContainers()
	if err != nil {
		w.checkError(err)
	}
	w.updateCount++
	logrus.Debugf("Metadata update count: %d", w.updateCount)

	for _, c := range containers {
		if c.HostUUID == w.hostUUID && c.State == "running" && c.StackName != stack.Name {
			containerID := c.ExternalId
			err = w.fileUpdater.LinkContainer(containerID)
			if err != nil {
				logrus.Error(err)
			}
			err = w.fileUpdater.LinkVolumeByContainerID(containerID)
			if err != nil {
				logrus.Error(err)
			}

		}
	}

	if w.updateCount%50 == 0 {
		logrus.Infof("Clean dead links, update count:%d", w.updateCount)
		w.fileUpdater.CleanDeadLinks()
	}
}

func (w *metadataWatcher) checkError(err error) {
	w.consecutiveErrorCount++
	if w.consecutiveErrorCount > 5 {
		panic(fmt.Sprintf("%v consecutive errors attempting to reach metadata. Panicing. Error: %v", w.consecutiveErrorCount, err))
	}
	logrus.Errorf("Error %v getting metadata: %v", w.consecutiveErrorCount, err)
}
