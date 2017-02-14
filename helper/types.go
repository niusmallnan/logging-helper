package helper

type LoggingFileUpdater interface {
	LinkContainer(containerID string)
	LinkVolume(volumeName string)
	CleanDeadLinks()
}
