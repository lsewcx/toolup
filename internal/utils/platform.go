package utils

import (
	"fmt"
	"strings"

	"github.com/shirou/gopsutil/v4/host"
)

type OSInfo struct {
	OS             string
	Arch           string
	Distro         string
	Family         string
	PackageManager string
}

func detectPackageManager(platform, family string) string {
	ids := strings.Fields(strings.ToLower(platform + " " + family))

	has := func(names ...string) bool {
		for _, name := range names {
			for _, id := range ids {
				if id == name {
					return true
				}
			}
		}
		return false
	}

	switch {
	case has("ubuntu", "debian", "linuxmint", "pop", "kali"):
		return "apt"
	case has("fedora", "rhel", "centos", "rocky", "almalinux"):
		return "dnf"
	default:
		return ""
	}
}

func GetOSInfo() (OSInfo, error) {
	info, err := host.Info()
	if err != nil {
		return OSInfo{}, fmt.Errorf("获取系统信息: %w", err)
	}

	return OSInfo{
		OS:             info.OS,
		Arch:           info.KernelArch,
		Distro:         info.Platform,
		Family:         info.PlatformFamily,
		PackageManager: detectPackageManager(info.Platform, info.PlatformFamily),
	}, nil
}
