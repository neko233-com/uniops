package monitor

import (
	"fmt"
	"strings"

	"golang.org/x/crypto/ssh"
)

type Collector struct{}

func NewCollector() *Collector {
	return &Collector{}
}

type SystemMetrics struct {
	CPU      CPUMetrics     `json:"cpu"`
	Memory   MemoryMetrics  `json:"memory"`
	Disk     []DiskMetrics  `json:"disk"`
	Network  []NetworkMetrics `json:"network"`
	Hostname string         `json:"hostname"`
	Uptime   string         `json:"uptime"`
	LoadAvg  []float64      `json:"load_avg"`
}

type CPUMetrics struct {
	Usage     float64 `json:"usage"`
	cores     int     `json:"-"`
	modelName string  `json:"-"`
}

type MemoryMetrics struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Free      uint64  `json:"free"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
}

type DiskMetrics struct {
	Mount string  `json:"mount"`
	Size  uint64  `json:"size"`
	Used  uint64  `json:"used"`
	Free  uint64  `json:"free"`
	Usage float64 `json:"usage"`
}

type NetworkMetrics struct {
	Interface string `json:"interface"`
	RxBytes   uint64 `json:"rx_bytes"`
	TxBytes   uint64 `json:"tx_bytes"`
}

func (c *Collector) Collect(host string, port int, username, password string) (*SystemMetrics, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	metrics := &SystemMetrics{}

	// Get hostname
	if output, err := c.runCommand(client, "hostname"); err == nil {
		metrics.Hostname = strings.TrimSpace(output)
	}

	// Get uptime
	if output, err := c.runCommand(client, "uptime -p"); err == nil {
		metrics.Uptime = strings.TrimSpace(output)
	}

	// Get load average
	if output, err := c.runCommand(client, "cat /proc/loadavg"); err == nil {
		parts := strings.Fields(output)
		if len(parts) >= 3 {
			metrics.LoadAvg = make([]float64, 3)
			fmt.Sscanf(parts[0], "%f", &metrics.LoadAvg[0])
			fmt.Sscanf(parts[1], "%f", &metrics.LoadAvg[1])
			fmt.Sscanf(parts[2], "%f", &metrics.LoadAvg[2])
		}
	}

	// Get CPU usage
	if output, err := c.runCommand(client, "top -bn1 | grep 'Cpu(s)' | awk '{print $2}'"); err == nil {
		var usage float64
		fmt.Sscanf(strings.TrimSpace(output), "%f", &usage)
		metrics.CPU.Usage = usage
	}

	// Get memory info
	if output, err := c.runCommand(client, "free -b | grep Mem"); err == nil {
		parts := strings.Fields(output)
		if len(parts) >= 4 {
			fmt.Sscanf(parts[1], "%d", &metrics.Memory.Total)
			fmt.Sscanf(parts[2], "%d", &metrics.Memory.Used)
			fmt.Sscanf(parts[3], "%d", &metrics.Memory.Free)
			metrics.Memory.Usage = float64(metrics.Memory.Used) / float64(metrics.Memory.Total) * 100
		}
	}

	// Get disk info
	if output, err := c.runCommand(client, "df -B1 | grep -E '^/dev/'"); err == nil {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 6 {
				var disk DiskMetrics
				disk.Mount = parts[5]
				fmt.Sscanf(parts[1], "%d", &disk.Size)
				fmt.Sscanf(parts[2], "%d", &disk.Used)
				fmt.Sscanf(parts[3], "%d", &disk.Free)
				disk.Usage = float64(disk.Used) / float64(disk.Size) * 100
				metrics.Disk = append(metrics.Disk, disk)
			}
		}
	}

	// Get network info
	if output, err := c.runCommand(client, "cat /proc/net/dev | grep -v 'face'"); err == nil {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 10 {
				iface := strings.Trim(parts[0], ":")
				if iface == "lo" {
					continue
				}
				var net NetworkMetrics
				net.Interface = iface
				fmt.Sscanf(parts[1], "%d", &net.RxBytes)
				fmt.Sscanf(parts[9], "%d", &net.TxBytes)
				metrics.Network = append(metrics.Network, net)
			}
		}
	}

	return metrics, nil
}

func (c *Collector) runCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	return string(output), err
}
