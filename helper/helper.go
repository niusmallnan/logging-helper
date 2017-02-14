package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

const (
	rancherLoggingVolumeName = "rancher-logging"
)

type Helper struct {
	dockerGraphDir  string
	containersDir   string
	volumesDir      string
	dockerClient    *client.Client
	containersCache map[string]string
	volumesCache    map[string]string
}

func NewHelper(dockerGraphDir string, containersDir string, volumesDir string) *Helper {
	dockerClient, err := NewDockerClient()
	if err != nil {
		logrus.Fatalf("Failed to init docker client")
	}
	helper := &Helper{
		dockerGraphDir:  dockerGraphDir,
		containersDir:   containersDir,
		volumesDir:      volumesDir,
		dockerClient:    dockerClient,
		containersCache: make(map[string]string),
		volumesCache:    make(map[string]string),
	}
	return helper
}

func (h *Helper) addSymlink(containerID string, oldPath string, newPath string) error {
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		err = os.Symlink(oldPath, newPath)
		if err != nil {
			return errors.Wrap(err, "Failed to create symlink file for json log")
		}
	}
	return nil
}

func (h *Helper) removeSymlink(logSymlinks []string) error {
	for _, logSymlink := range logSymlinks {
		if _, err := os.Stat(logSymlink); os.IsNotExist(err) {
			err = os.Remove(logSymlink)
			if err != nil {
				return errors.Wrap(err, "Failed to clean LogSymlinks.")
			}
			delete(h.containersCache, logSymlink)
			delete(h.volumesCache, logSymlink)
		}
	}
	return nil
}

func (h *Helper) LinkContainer(containerID string) error {
	jsonLoggingFile := fmt.Sprintf("%s-json.log", containerID)
	newPath := filepath.Join(h.containersDir, jsonLoggingFile)
	if _, ok := h.containersCache[newPath]; ok {
		logrus.Debugf("LinkContainer, ContainerID: %s has been linked", h.containersCache[newPath])
		return nil
	}
	oldPath := filepath.Join(h.dockerGraphDir, "containers", jsonLoggingFile)
	err := h.addSymlink(containerID, oldPath, newPath)
	if err != nil {
		return err
	}
	h.containersCache[newPath] = containerID
	return nil
}

func (h *Helper) LinkVolumeByContainerID(containerID string) error {
	container, err := h.dockerClient.ContainerInspect(context.Background(), containerID)
	if err != nil && !client.IsErrContainerNotFound(err) {
		return errors.Wrap(err, "Failed to inspect container")
	}

	for _, mount := range container.Mounts {
		if strings.Contains(mount.Name, rancherLoggingVolumeName) {
			oldPathes, err := filepath.Glob(filepath.Join(mount.Source, "/*.log"))
			if err != nil {
				return errors.Wrap(err, "Failed to gather volume logging files")
			}
			for _, oldPath := range oldPathes {
				_, oldFile := filepath.Split(oldPath)
				newDir := filepath.Join(h.volumesDir, mount.Name)
				newPath := filepath.Join(newDir, oldFile)
				if _, ok := h.volumesCache[newPath]; ok {
					logrus.Debugf("LinkVolume, ContainerID: %s has been linked", h.volumesCache[newPath])
					return nil
				}
				if _, err = os.Stat(newDir); os.IsNotExist(err) {
					err = os.Mkdir(newDir, 664)
					if err != nil {
						return errors.Wrap(err, "Failed to mkdir")
					}
				}
				err = h.addSymlink(containerID, oldPath, newPath)
				if err != nil {
					return err
				}
				h.volumesCache[newPath] = containerID
			}
		}
	}
	return nil

}

func (h *Helper) CleanDeadLinks() {
	containerLogSymlinks, err := filepath.Glob(filepath.Join(h.containersDir, "/*.log"))
	if err == nil {
		h.removeSymlink(containerLogSymlinks)
	}
	volumesLogSymlinks, err := filepath.Glob(filepath.Join(h.volumesDir, "/*", "/*.log"))
	if err == nil {
		h.removeSymlink(volumesLogSymlinks)
	}
}
