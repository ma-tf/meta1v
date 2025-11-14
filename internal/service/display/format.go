package display

import (
	"bytes"
	"fmt"
	"math"
)

//nolint:gochecknoglobals // not exported anyway
var fpBits = [8]int{7, 2, 8, 3, 8, 2, 8, 7}

func byteToBox(b byte, l int) (string, error) {
	const emptyBox, filledBox = "\u25AF", "\u25AE"

	var buf bytes.Buffer

	for i, mask := 0, byte(math.MaxInt8+1); i < l; i, mask = i+1, mask>>1 {
		var (
			topOrBottomRow  = l == 7
			startOfFullByte = l == 8 && i == 0
			endOfBits       = l != 8 && i == l-1
			err             error
		)

		switch {
		case b&mask != 0:
			_, err = buf.WriteString("\033[31m" + filledBox + "\033[0m ")
		case topOrBottomRow || startOfFullByte || endOfBits:
			_, err = buf.WriteString("\033[31m" + emptyBox + "\033[0m ")
		default:
			_, err = buf.WriteString(emptyBox + " ")
		}

		if err != nil {
			return "", fmt.Errorf("could not print focus point: %w", err)
		}
	}

	return buf.String(), nil
}
