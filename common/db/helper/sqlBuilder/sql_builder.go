/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/7/31 11:04
 */

// Package sqlBuilder
package sqlBuilder

import (
	"errors"
	"fmt"
	"github.com/real-uangi/allingo/common/db/helper/page"
	"strconv"
	"strings"
)

type SqlBuilder struct {
	fields []string
	from   []string
	where  []string
	order  []string
	limit  int
	offset int
	group  []string
}

func New() *SqlBuilder {
	return &SqlBuilder{
		fields: make([]string, 0, 1),
		from:   make([]string, 0, 1),
		where:  make([]string, 0, 8),
		order:  make([]string, 0, 2),
		group:  make([]string, 0, 2),
	}
}

func NewWithCap(fields, from, where, order, group int) *SqlBuilder {
	return &SqlBuilder{
		fields: make([]string, 0, fields),
		from:   make([]string, 0, from),
		where:  make([]string, 0, where),
		order:  make([]string, 0, order),
		group:  make([]string, 0, group),
	}
}

func (builder *SqlBuilder) Field(fields ...string) *SqlBuilder {
	for _, f := range fields {
		if f != "" {
			builder.fields = append(builder.fields, f)
		}
	}
	return builder
}

func (builder *SqlBuilder) FieldAs(field, as string) *SqlBuilder {
	s := field + ` as "` + as + `"`
	builder.fields = append(builder.fields, s)
	return builder
}

func (builder *SqlBuilder) From(from string) *SqlBuilder {
	if from != "" {
		builder.from = append(builder.from, from)
	}
	return builder
}

func (builder *SqlBuilder) Fromf(format string, args ...interface{}) *SqlBuilder {
	return builder.From(fmt.Sprintf(format, args...))
}

func (builder *SqlBuilder) Where(where string) *SqlBuilder {
	if where != "" {
		builder.where = append(builder.where, where)
	}
	return builder
}

func (builder *SqlBuilder) Wheref(format string, args ...interface{}) *SqlBuilder {
	return builder.Where(fmt.Sprintf(format, args...))
}

func (builder *SqlBuilder) Order(order ...string) *SqlBuilder {
	for _, o := range order {
		if o != "" {
			builder.order = append(builder.order, o)
		}
	}
	return builder
}

func (builder *SqlBuilder) GroupBy(fields ...string) *SqlBuilder {
	for _, g := range fields {
		if g != "" {
			builder.group = append(builder.group, g)
		}
	}
	return builder
}

func (builder *SqlBuilder) Limit(limit int) *SqlBuilder {
	builder.limit = limit
	return builder
}

func (builder *SqlBuilder) Offset(offset int) *SqlBuilder {
	builder.offset = offset
	return builder
}

func (builder *SqlBuilder) UsePage(pi page.InputInterface) *SqlBuilder {
	return builder.Limit(pi.GetPageSize()).Offset(pi.GetOffset())
}

func (builder *SqlBuilder) String() string {
	if len(builder.from) == 0 {
		panic(errors.New("no table specified"))
	}
	var sb strings.Builder
	sb.WriteString("select ")
	if len(builder.fields) > 0 {
		for i, field := range builder.fields {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(field)
		}
	} else {
		sb.WriteString("*")
	}
	sb.WriteString(" from ")
	for _, from := range builder.from {
		sb.WriteString(from)
		sb.WriteString(" ")
	}
	if len(builder.where) > 0 {
		sb.WriteString(" where ")
		for i, where := range builder.where {
			if i > 0 {
				sb.WriteString(" and ")
			}
			sb.WriteString(where)
		}
	}
	if len(builder.group) > 0 {
		sb.WriteString(" group by ")
		for i, group := range builder.group {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(group)
		}
	}
	if len(builder.order) > 0 {
		sb.WriteString(" order by ")
		for i, order := range builder.order {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(order)
		}
	}
	if builder.limit > 0 {
		sb.WriteString(" limit ")
		sb.WriteString(strconv.Itoa(builder.limit))
	}
	if builder.offset > 0 {
		sb.WriteString(" offset ")
		sb.WriteString(strconv.Itoa(builder.offset))
	}
	return sb.String()
}

func (builder *SqlBuilder) StringForCount() string {
	if len(builder.from) == 0 {
		panic(errors.New("no table specified"))
	}
	var sb strings.Builder
	sb.WriteString("select count(1) from ")
	for _, from := range builder.from {
		sb.WriteString(from)
		sb.WriteString(" ")
	}
	if len(builder.where) > 0 {
		sb.WriteString(" where ")
		for i, where := range builder.where {
			if i > 0 {
				sb.WriteString(" and ")
			}
			sb.WriteString(where)
		}
	}

	return sb.String()
}
