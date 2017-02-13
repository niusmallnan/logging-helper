package main

type LoggingFileUpdater interface {
	LinkContainer(containerID string)
	LinkVolume(volumeName string)
	CleanDeadLinks()
}
