package jobs

import (
	"github.com/containerd/cgroups/v3/cgroup2"
)

const groupName string = "teleport-group.slice"
const memLimit int64 = 10 * 1024 * 1024
const cpuPeriod uint64 = 1000000
const cpuQuota int64 = 200000

func GetOrCreateGroup() (*cgroup2.Manager, error) {
	m, err := cgroup2.LoadSystemd("/", groupName)
	if err == nil {
		return m, nil
	}

	memMax := memLimit
	period := cpuPeriod
	quota := cpuQuota
	cpu := cgroup2.CPU{
		Max: cgroup2.NewCPUMax(&quota, &period),
	}
	mem := cgroup2.Memory{
		Max: &memMax,
	}
	res := cgroup2.Resources{
		CPU:    &cpu,
		Memory: &mem,
	}
	// dummy PID of -1 is used for creating a "general slice" to be used as a parent cgroup.
	return cgroup2.NewSystemd("/", groupName, -1, &res)

}
