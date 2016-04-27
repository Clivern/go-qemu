// Copyright 2016 The go-qemu Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package qemu

import (
	"testing"
	"time"

	"github.com/digitalocean/go-qemu/qmp"
)

const defaultTestTimeout = 5 * time.Second

func TestNew(t *testing.T) {
	m := mockMonitor{}

	_, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}
}

func TestNewError(t *testing.T) {
	m := mockMonitor{alwaysFail: true}

	_, err := NewDomain(m, "foo")
	if err == nil {
		t.Errorf("expected monitor failure")
	}
}

func TestBlockDevice(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	device := "drive-virtio-disk0"
	bd, err := d.BlockDevice(device)
	if err != nil {
		t.Error(err)
	}

	if bd.Device != device {
		t.Errorf("expected device %q, got %q", device, bd.Device)
	}
}

func TestBlockDeviceNotFound(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	device := "foo"
	_, err = d.BlockDevice(device)
	if err == nil {
		t.Errorf("expected block device %q to not exist", device)
	}

	if err != ErrBlockDeviceNotFound {
		t.Errorf("expected ErrBlockDeviceNotFound")
	}
}

func TestBlockDeviceMonitorFailure(t *testing.T) {
	m := &mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	m.alwaysFail = true
	if _, err = d.BlockDevice("foo"); err == nil {
		t.Error("expected monitor failure")
	}
}

func TestBlockJobs(t *testing.T) {
	m := mockMonitor{activeJobs: true}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	jobs, err := d.BlockJobs()
	if err != nil {
		t.Error(err)
	}

	if len(jobs) != 1 {
		t.Error("expected running backup job")
	}

	expected := "ok"
	if jobs[0].IOStatus != expected {
		t.Errorf("expected i/o status %q, got %q", expected, jobs[0].IOStatus)
	}

	expected = "drive-virtio-disk0"
	if jobs[0].Device != expected {
		t.Errorf("expected device %q, got %q", expected, jobs[0].Device)
	}
}

func TestBlockStats(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	stats, err := d.BlockStats()
	if err != nil {
		t.Error(err)
	}

	if len(stats) != 4 {
		t.Error("expected 4 block stats")
	}

	expected := "ide0-hd0"
	if stats[0].Device != expected {
		t.Errorf("expected device %q, got %q", expected, stats[0].Device)
	}

	expectedBytes := uint64(9786368)
	if stats[0].WriteBytes != expectedBytes {
		t.Errorf("expected %d write bytes, got %d", expectedBytes, stats[0].WriteBytes)
	}
}

func TestBlockJobsMonitorFail(t *testing.T) {
	m := &mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	m.alwaysFail = true
	_, err = d.BlockJobs()
	if err == nil {
		t.Errorf("expected monitor failure")
	}
}

func TestBlockJobsInvalidJSON(t *testing.T) {
	m := mockMonitor{invalidJSON: true}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	_, err = d.BlockJobs()
	if err == nil {
		t.Errorf("expected invalid json to cause failure")
	}
}

func TestClose(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	if err := d.Close(); err != nil {
		t.Error(err)
	}

	if _, ok := <-d.done; ok {
		t.Error("domain should be closed")
	}

	if err := m.Disconnect(); err != nil {
		t.Error(err)
	}
}

func TestCommands(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	cmds, err := d.Commands()
	if err != nil {
		t.Error(err)
	}

	expected := 135
	actual := len(cmds)
	if actual != expected {
		t.Errorf("expected number of supported commands to be %d, got %d", expected, actual)
	}

	found := false
	search := "query-block"
	for _, c := range cmds {
		if c == search {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected command %q to be returned", search)
	}
}

func TestCommandsMonitorFailure(t *testing.T) {
	m := &mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	m.alwaysFail = true
	if _, err := d.Commands(); err == nil {
		t.Error("expected monitor failure")
	}
}

func TestCommandsInvalidJSON(t *testing.T) {
	m := mockMonitor{invalidJSON: true}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	if _, err := d.Commands(); err == nil {
		t.Error("expected invalid json to cause failure")
	}
}

func TestPCIDevices(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	devices, err := d.PCIDevices()
	if err != nil {
		t.Error(err)
	}

	if len(devices) != 2 {
		t.Error("expected two PCI devices")
	}

	expected := 0
	if devices[0].Bus != expected {
		t.Errorf("expected device bus %d, got %q", expected, devices[0].Bus)
	}

	expectedDesc := "Intel Ethernet controller"
	if devices[1].ClassInfo.Desc != expectedDesc {
		t.Errorf("expected device %q, got %q", expectedDesc, devices[1].ClassInfo.Desc)
	}
}

func TestStatusRunning(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	status, err := d.Status()
	if err != nil {
		t.Error(err)
	}

	if status != StatusRunning {
		t.Error("expected domain to be running")
	}
}

func TestStatusShutdown(t *testing.T) {
	m := mockMonitor{poweredOff: true}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	status, err := d.Status()
	if err != nil {
		t.Error(err)
	}

	if status != StatusShutdown {
		t.Error("expected domain to be powered off")
	}
}

func TestStatusFail(t *testing.T) {
	m := &mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	m.alwaysFail = true
	_, err = d.Status()
	if err == nil {
		t.Errorf("expected monitor failure")
	}
}

func TestStatusInvalidJSON(t *testing.T) {
	m := mockMonitor{invalidJSON: true}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	_, err = d.Status()
	if err == nil {
		t.Errorf("expected invalid json to cause failure")
	}
}

func TestRunInvalidCommand(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	_, err = d.Run(qmp.Cmd{})
	if err == nil {
		t.Error("expected invalid command to fail")
	}
}

func TestSupported(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	cmd := "query-block"
	supported, err := d.Supported(cmd)
	if err != nil {
		t.Error(err)
	}

	if !supported {
		t.Errorf("expected command %q to be supported", cmd)
	}
}

func TestSupportedFalse(t *testing.T) {
	m := mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	cmd := "query-foo"
	supported, err := d.Supported(cmd)
	if err != nil {
		t.Error(err)
	}

	if supported {
		t.Errorf("expected command %q to be unsupported", cmd)
	}
}

func TestSupportedMonitorFailure(t *testing.T) {
	m := &mockMonitor{}

	d, err := NewDomain(m, "foo")
	if err != nil {
		t.Error(err)
	}

	m.alwaysFail = true
	if _, err := d.Supported("foo"); err == nil {
		t.Error("expected monitor failure")
	}
}