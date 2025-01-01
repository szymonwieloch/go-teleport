package jobs

import (
	"fmt"

	"github.com/containerd/cgroups"
	"github.com/opencontainers/runtime-spec/specs-go"
)

const groupName string = "/teleport"

func GetOrCreateGroup() (cgroups.Cgroup, error) {
	// Create a new cgroup group
	// If the group already exists, return it
	// Otherwise, create a new group and return it
	// If the group can't be created, return an error

	control, err := cgroups.Load(cgroups.V1, cgroups.StaticPath(groupName))
	if err == nil {
		return control, nil
	}
	if err != cgroups.ErrCgroupDeleted && err != cgroups.ErrMountPointNotExist {
		return nil, fmt.Errorf("could not load cgroup: %w", err)
	}
	// Limit the group resources to 20% of CPU and 10 MB of memory
	// this is a very simplified soludion but adequate for the purpose of this project
	period := uint64(1000000)
	quota := int64(200000)
	mem := int64(10 * 1024 * 1024) // 10MB
	return cgroups.New(cgroups.V1, cgroups.StaticPath(groupName), &specs.LinuxResources{
		CPU: &specs.LinuxCPU{
			Period: &period,
			Quota:  &quota,
		},
		Memory: &specs.LinuxMemory{
			Limit: &mem,
		},
	})
}
