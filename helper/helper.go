package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

const (
	volumeSymlinkSuffix = ".log"
)

type Helper struct {
	dockerGraphDir  string
	containersDir   string
	volumesDir      string
	volumesPattern  string
	filesPattern    string
	dockerClient    *client.Client
	containersCache map[string]string
	volumesCache    map[string]string
}

func NewHelper(c *cli.Context) *Helper {
	dockerGraphDir := c.String("docker-graph-dir")
	containersDir := c.String("logging-containers-dir")
	volumesDir := c.String("logging-volumes-dir")
	volumesPattern := c.String("logging-volumes-pattern")
	filesPattern := c.String("logging-files-pattern")

	dockerClient, err := NewDockerClient()
	if err != nil {
		logrus.Fatalf("Failed to init docker client")
	}
	helper := &Helper{
		dockerGraphDir:  dockerGraphDir,
		containersDir:   containersDir,
		volumesDir:      volumesDir,
		volumesPattern:  volumesPattern,
		filesPattern:    filesPattern,
		dockerClient:    dockerClient,
		containersCache: make(map[string]string),
		volumesCache:    make(map[string]string),
	}
	mkdir(containersDir)
	mkdir(volumesDir)
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
	oldPath := filepath.Join(h.dockerGraphDir, "containers", containerID, jsonLoggingFile)
	err := h.addSymlink(containerID, oldPath, newPath)
	if err != nil {
		return err
	}
	logrus.Debugf("LinkContainer, ContainerID: %s, Linking", containerID)
	h.containersCache[newPath] = containerID
	return nil
}

func (h *Helper) LinkVolumeByContainerID(containerID string) error {
	container, err := h.dockerClient.ContainerInspect(context.Background(), containerID)
	if err != nil && !client.IsErrContainerNotFound(err) {
		return errors.Wrap(err, "Failed to inspect container")
	}

	for _, mount := range container.Mounts {
		matched, err := regexp.MatchString(h.volumesPattern, mount.Name)
		if err != nil {
			return errors.Wrap(err, "Failed to match volumes pattern")
		}
		if matched {
			oldPathes, err := filepath.Glob(filepath.Join(mount.Source, fmt.Sprintf("/%s", h.filesPattern)))
			if err != nil {
				return errors.Wrap(err, "Failed to gather volume logging files")
			}
			for _, oldPath := range oldPathes {
				_, oldFile := filepath.Split(oldPath)
				if strings.HasSuffix(oldFile, volumeSymlinkSuffix) == false {
					oldFile = fmt.Sprintf("%s%s", oldFile, volumeSymlinkSuffix)
				}
				newPath := filepath.Join(h.volumesDir, fmt.Sprintf("%s-%s-%s", containerID, mount.Name, oldFile))
				if _, ok := h.volumesCache[newPath]; ok {
					logrus.Debugf("LinkVolume, ContainerID: %s has been linked", h.volumesCache[newPath])
					return nil
				}
				err = h.addSymlink(containerID, oldPath, newPath)
				if err != nil {
					return err
				}
				logrus.Debugf("LinkVolume, ContainerID: %s, Linking", containerID)
				h.volumesCache[newPath] = containerID
			}
		}
	}
	return nil

}

func (h *Helper) CleanDeadLinks() {
	containerLogSymlinks, err := filepath.Glob(filepath.Join(h.containersDir, fmt.Sprintf("/%s", h.filesPattern)))
	if err == nil {
		h.removeSymlink(containerLogSymlinks)
	}
	volumesLogSymlinks, err := filepath.Glob(filepath.Join(h.volumesDir, "/*", fmt.Sprintf("/%s", h.filesPattern)))
	if err == nil {
		h.removeSymlink(volumesLogSymlinks)
	}
}
