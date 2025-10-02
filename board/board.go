package board

import (
	"context"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
)

type Boarder interface {
	Run(ctx context.Context, args ...string) error
	CopyTo(ctx context.Context, src, dst string) error
	MkDirAll(ctx context.Context, path string) error
}

var OnBoard = sync.OnceValue(func() bool {
	var boardNames = []string{"UNO Q\n", "Imola\n", "Inc. Robotics RB1\n"}
	buf, err := os.ReadFile("/sys/class/dmi/id/product_name")
	if err == nil && slices.Contains(boardNames, string(buf)) {
		return true
	}
	return false
})()

// escapeArgs escapes arguments that contain spaces by wrapping them in quotes.
// This should allow argument forwarding on a remote shells.
func escapeArgs(args []string) []string {
	for i, arg := range args {
		if strings.Contains(arg, " ") {
			args[i] = strconv.Quote(arg)
		}
	}
	return args
}
