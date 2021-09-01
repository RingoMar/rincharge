package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/gen2brain/beeep"
)

var (
	user32           = syscall.MustLoadDLL("user32.dll")
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	getLastInputInfo = user32.MustFindProc("GetLastInputInfo")
	getTickCount     = kernel32.MustFindProc("GetTickCount")
	lastInputInfo    struct {
		cbSize uint32
		dwTime uint32
	}
)
var lastNoti time.Time

var BatteryStatus = map[int]string{1: "discharging", 2: "Plugged in", 3: "Fully Charged", 4: "low", 5: "critical",
	6: "Charging", 7: "Charging and High", 8: "Charging and Low", 9: "Charging and Critical", 10: "Undefined", 11: "Partially Charged"}

func getBattryStatus() int {
	out, err := exec.Command("WMIC", "PATH", "Win32_Battery", "Get", "BatteryStatus").Output()
	if err != nil {
		fmt.Println(out)
	}
	output := string(out)
	v := strings.Split(output, "\r\n")
	i, _ := strconv.Atoi(strings.Replace(strings.Replace(v[1], " ", "", -1), "\r", "", -1))
	if err != nil {
		fmt.Println(err)
	}

	return i

}

func getBattryLevel() int {
	out, err := exec.Command("WMIC", "PATH", "Win32_Battery", "Get", "EstimatedChargeRemaining").Output()
	if err != nil {
		fmt.Println(out)
	}
	output := string(out)
	v := strings.Split(output, "\r\n")
	i, _ := strconv.Atoi(strings.Replace(strings.Replace(v[1], " ", "", -1), "\r", "", -1))
	return i

}

func IdleTime() int {
	lastInputInfo.cbSize = uint32(unsafe.Sizeof(lastInputInfo))
	currentTickCount, _, _ := getTickCount.Call()
	r1, _, err := getLastInputInfo.Call(uintptr(unsafe.Pointer(&lastInputInfo)))
	if r1 == 0 {
		panic("error getting last input info: " + err.Error())
	}
	getTime := (time.Duration((uint32(currentTickCount) - lastInputInfo.dwTime)) * time.Millisecond)
	return int(getTime) / 1000000000

}

func main() {
	t := time.NewTicker(1 * time.Second)
	fmt.Println("Rin Batt Check")
	for range t.C {
		battStatus := (getBattryStatus())
		battryInt := getBattryLevel()

		idleTimeData := IdleTime()
		if idleTimeData > 40 { // if computer is idle for about 1 or 2 mins
			if battStatus == 2 && battryInt >= 98 { // If charging and Battery is more thatn 98
				lastNotiTime := int(time.Since(lastNoti)) / 1000000000
				if lastNotiTime > 120 {
					beeep.Alert("Battery Status: "+BatteryStatus[battStatus], "Charging Complete", "asset.png")
					fmt.Println(battStatus, BatteryStatus[battStatus], battryInt, idleTimeData, lastNotiTime)
					lastNoti = time.Now()
				}
			}
		}
	}
}
