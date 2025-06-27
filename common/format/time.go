/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/27 13:01
 */

// Package format

package format

import (
	"fmt"
	"strings"
)

func TimePretty(inputMs int64) string {
	var left = inputMs
	var step int64 = 0
	var builder strings.Builder
	const second = 1000
	if left >= second {
		const minute = 60 * second
		if left >= minute {
			const hour = 60 * minute
			if left >= hour {
				const day = 24 * hour
				if left >= day {
					const year = 365 * day
					if left >= year {
						step = left / year
						builder.WriteString(fmt.Sprintf(" %dyear", step))
						left = left % year
					}
					step = left / day
					if step > 0 {
						builder.WriteString(fmt.Sprintf(" %dday", step))
					}
					left = left % day
				}
				step = left / hour
				if step > 0 {
					builder.WriteString(fmt.Sprintf(" %dhour", step))
				}
				left = left % hour
			}
			step = left / minute
			if step > 0 {
				builder.WriteString(fmt.Sprintf(" %dmin", step))
			}
			left = left % minute
		}
		step = left / second
		if step > 0 {
			builder.WriteString(fmt.Sprintf(" %ds", step))
		}
		left = left % second
	}
	step = left
	if step > 0 {
		builder.WriteString(fmt.Sprintf(" %dms", step))
	}
	if builder.Len() == 0 {
		return " 0ms"
	}
	return builder.String()
}
