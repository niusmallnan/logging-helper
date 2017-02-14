package helper

type LoggingFileUpdater interface {
	LinkContainer(containerID string) error
	LinkVolumeByContainerID(containerID string) error
	CleanDeadLinks()
}
