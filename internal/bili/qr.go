package bili

import (
	"fmt"
	"strings"

	"rsc.io/qr"
)

// qrSVG renders text as a standalone <svg>: black squares on transparent,
// crispEdges, with a 2-module quiet zone (like internal/demo/qr.go).
func qrSVG(text string) (string, error) {
	code, err := qr.Encode(text, qr.M)
	if err != nil {
		return "", err
	}
	size := code.Size
	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="-2 -2 %d %d" shape-rendering="crispEdges"><path d="`, size+4, size+4)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if code.Black(x, y) {
				fmt.Fprintf(&sb, "M%d %dh1v1h-1z", x, y)
			}
		}
	}
	sb.WriteString(`" fill="#000"/></svg>`)
	return sb.String(), nil
}
