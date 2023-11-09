package shared

import (
	"fmt"
	"strings"
)

func FormatKernelCmdLine(args map[string]string) string {
	output := []string{}

	for key, value := range args {
		if value == "" {
			output = append(output, key)
		} else {
			output = append(output, fmt.Sprintf("%s=%s", key, value))
		}
	}

	return strings.Join(output, " ")
}
