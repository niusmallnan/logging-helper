package main

func NewHelper(containersDir string, volumesDir string) *Helper {
	return &Helper{
		containersDir: containersDir,
		volumesDir:    volumesDir,
	}
}

type Helper struct {
	containersDir string
	volumesDir    string
}

func (h *Helper) LinkContainer(containerID string) {

}

func (h *Helper) LinkVolume(volumeName string) {

}

func (h *Helper) CleanDeadLinks() {

}
