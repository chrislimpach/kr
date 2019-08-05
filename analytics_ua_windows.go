package kr

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

var analytics_user_agent = fmt.Sprintf("Mozilla/5.0 (Windows; Windows) (KHTML, like Gecko) Version/%s kr/%s", CURRENT_VERSION, CURRENT_VERSION)

const analytics_os = "Windows"

var cachedAnalyticsOSVersion *string
var osVersionMutex sync.Mutex

func getAnalyticsOSVersion() *string {
	osVersionMutex.Lock()
	defer osVersionMutex.Unlock()
	if cachedAnalyticsOSVersion != nil {
		return cachedAnalyticsOSVersion
	}

	name, _ := exec.Command("reg", "query", "HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "-v", "ProductName").Output()
	nameRX := regexp.MustCompile(`(?s)ProductName\s+REG_SZ\s+(.*?)\s*$`)
	nameLoc := nameRX.FindSubmatchIndex(name)
	if len(nameLoc) == 0 {
		name = []byte("ProductName REG_SZ Windows Unknown")
		nameLoc = nameRX.FindSubmatchIndex(name)
	}

	rel, _ := exec.Command("reg", "query", "HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "-v", "ReleaseId").Output()
	relRX := regexp.MustCompile(`(?s)ReleaseId\s+REG_SZ\s+(.*?)\s*$`)
	relLoc := relRX.FindSubmatchIndex(rel)
	if len(relLoc) == 0 {
		rel = []byte("ReleaseId REG_SZ 0000")
		relLoc = relRX.FindSubmatchIndex(rel)
	}
	ver := fmt.Sprintf("%s %s", string(name[nameLoc[2]:nameLoc[3]]),
		string(rel[relLoc[2]:relLoc[3]]))

	stripped := strings.TrimSpace(ver)
	cachedAnalyticsOSVersion = &stripped
	return cachedAnalyticsOSVersion
}
