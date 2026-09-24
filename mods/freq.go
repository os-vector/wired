package mods

import (
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/os-vector/wired/vars"
)

var FreqPreset int = 0
var FreqPresetStr string = "0"
var FreqName = "FreqChange"
var FreqSaveFile string = filepath.Join(vars.GetModDir(FreqName), "freq3")

type FreqChange struct {
	vars.Modification
}

func NewFreqChange() *FreqChange {
	return &FreqChange{}
}

func (m *FreqChange) Name() string {
	return FreqName
}

func (m *FreqChange) Load() error {
	var freqInt int
	contents, err := vars.ReadFile(FreqSaveFile)
	if err == nil {
		freqInt, err = strconv.Atoi(contents)
	}
	if err != nil {
		fmt.Println("(freq) Error loading contents, using FreqPreset of 1")
		freqInt = FreqPreset
		contents = FreqPresetStr
	}
	DoFreqChange(freqInt, contents)
	return nil
}

func (m *FreqChange) HTTP(w http.ResponseWriter, r *http.Request) {
	if vars.IsEndpoint(r, "set") {
		f := r.FormValue("freq")
		if f == "" {
			vars.HTTPError(w, r, "empty freq")
			return
		}
		fi, err := strconv.Atoi(f)
		if err != nil {
			vars.HTTPError(w, r, "given freq is not an int")
			return
		}
		if fi != 0 && fi != 1 && fi != 2 {
			vars.HTTPError(w, r, "given freq is not 0, 1, or 2")
			return
		}
		DoFreqChange(fi, f)
		vars.HTTPSuccess(w, r)
		return
	} else if vars.IsEndpoint(r, "get") {
		c, err := vars.ReadFile(FreqSaveFile)
		if err != nil || c == "" {
			c = FreqPresetStr
		}
		w.Write([]byte(c))
		return
	}
}

func DoFreqChange(freq int, freqStr string) {
	err := vars.SaveFile(freqStr, FreqSaveFile)
	if err != nil {
		fmt.Println("freqchange save error:", err)
	}
	var cpufreq string
	var ramfreq string
	var gov string
	switch {
	case freq == 0:
		cpufreq = "533333"
		ramfreq = "400000"
		gov = "interactive"
	case freq == 1:
		cpufreq = "729600"
		ramfreq = "400000"
		gov = "interactive"
	case freq == 2:
		cpufreq = "1267200"
		ramfreq = "800000"
		gov = "performance"
	}
	fmt.Println("FreqChange done!: " + cpufreq + " " + ramfreq + " " + gov)
	RunCmd("echo " + cpufreq + " > /sys/devices/system/cpu/cpu0/cpufreq/scaling_max_freq")
	RunCmd("echo disabled > /sys/kernel/debug/msm_otg/bus_voting")
	RunCmd("echo 0 > /sys/kernel/debug/msm-bus-dbg/shell-client/update_request")
	RunCmd("echo 1 > /sys/kernel/debug/msm-bus-dbg/shell-client/mas")
	RunCmd("echo 512 > /sys/kernel/debug/msm-bus-dbg/shell-client/slv")
	RunCmd("echo 0 > /sys/kernel/debug/msm-bus-dbg/shell-client/ab")
	RunCmd("echo active clk2 0 1 max " + ramfreq + " > /sys/kernel/debug/rpm_send_msg/message")
	RunCmd("echo " + gov + " > /sys/devices/system/cpu/cpu0/cpufreq/scaling_governor")
	RunCmd("echo 1 > /sys/kernel/debug/msm-bus-dbg/shell-client/update_request")
}

func RunCmd(cmd string) ([]byte, error) {
	return exec.Command("/bin/bash", "-c", cmd).Output()
}
