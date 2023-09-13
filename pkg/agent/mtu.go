package agent

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func getMTU(ip string) (int, error) {
	cmd := exec.Command("bash", "-c", fmt.Sprintf("bash ../../mtu_test.sh %s", ip))
	output, err := cmd.Output()

	if err != nil {
		return 0, err
	}

	result := string(output)

	fields := strings.Fields(result)
	if len(fields) < 5 {
		return 0, fmt.Errorf("Invalid output format")
	}

	mtuStr := fields[len(fields)-1]
	mtu, err := strconv.Atoi(mtuStr)
	if err != nil {
		return 0, err
	}

	return mtu, nil
}
