/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/20 10:46
 */

// Package log
package log

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"strings"
)

// 设置颜色
var levelDec = [7]string{
	logrus.PanicLevel: "\u001B[1;35m[PANIC]\u001B[0m", // 紫色
	logrus.FatalLevel: "\u001B[1;31m[FATAL]\u001B[0m", // 红色
	logrus.ErrorLevel: "\u001B[31m[ERROR]\u001B[0m",   // 深红
	logrus.WarnLevel:  "\u001B[33m[WARN ]\u001B[0m",   // 黄色
	logrus.InfoLevel:  "\u001B[32m[INFO ]\u001B[0m",   // 绿色
	logrus.DebugLevel: "\u001B[36m[DEBUG]\u001B[0m",   // 青色
	logrus.TraceLevel: "\u001B[34m[TRACE]\u001B[0m",   // 蓝色
}

type customFormatter struct {
	middleInfos [][]byte
}

func newFormatter(loggerName string) *customFormatter {
	formatter := new(customFormatter)
	formatter.preSetMiddles(loggerName)
	return formatter
}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(entry.Time.Format("2006-01-02 15:04:05.000"))
	b.WriteString(fmt.Sprintf(" %8d", entry.Data[FieldGoId]))
	b.Write(f.middleInfos[entry.Level])
	b.WriteString(entry.Message)
	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *customFormatter) preSetMiddles(loggerName string) {
	f.middleInfos = make([][]byte, len(levelDec))
	for i := 0; i < len(levelDec); i++ {
		buffer := bytes.NewBuffer(make([]byte, 0, 128))
		buffer.WriteString(" -- ")
		buffer.WriteString(fmt.Sprintf("\u001B[36;1m%-30s\u001B[0m ", shorterName(loggerName, 30)))
		buffer.WriteString(levelDec[i])
		buffer.Write([]byte(" "))
		f.middleInfos[i] = buffer.Bytes()
	}
}

func shorterName(input string, threshold int) string {
	handled := 0
	for len(input) > threshold {
		paths := strings.Split(input, ".")
		if len(paths)-1 <= handled {
			input = input[len(input)-threshold:]
		}
		current := paths[handled]
		paths[handled] = current[:1]
		var builder strings.Builder
		for i, v := range paths {
			builder.WriteString(v)
			if i != len(paths)-1 {
				builder.WriteString(".")
			}
		}
		input = builder.String()
		handled++
	}
	return input
}
