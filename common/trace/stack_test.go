/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/28 12:47
 */

// Package trace
package trace

import (
	"errors"
	"fmt"
	"testing"
)

func TestStack(t *testing.T) {
	p()
}

func p() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Println(err, Stack(3))
		}
	}()
	panic(errors.New("this is a panic"))
}
