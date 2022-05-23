// SPDX-License-Identifier: AGPL-3.0-or-later
package main

import (
	"fmt"
)

//go:noinline
func Divide(dividend int, divisor int) (quotient int, remainder int) {
	quotient = dividend / divisor
	remainder = dividend % divisor
	return
}

func main() {
	dividend, divisor := 5, 3
	quotient, remainder := Divide(dividend, divisor)
	fmt.Printf("dividend=%d divisor=%d quotient=%d remainder=%d\n", dividend, divisor, quotient, remainder)
}
