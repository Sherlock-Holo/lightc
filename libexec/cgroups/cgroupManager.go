package cgroups

import (
	"github.com/Sherlock-Holo/lightc/libexec/cgroups/subsystems"
)

type CgroupManager struct {
	Path     string
	Resource *subsystems.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{Path: path}
}

func (cm *CgroupManager) Apply(pid int) error {
	for _, instance := range subsystems.Instances {
		_ = instance.Apply(cm.Path, pid)
	}
	return nil
}

func (cm *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	for _, instance := range subsystems.Instances {
		_ = instance.Set(cm.Path, res)
	}
	return nil
}

func (cm *CgroupManager) Destroy() error {
	for _, instance := range subsystems.Instances {
		_ = instance.Remove(cm.Path)
	}
	return nil
}
