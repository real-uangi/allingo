/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/25 16:12
 */

// Package api
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/real-uangi/allingo/common/business"
	"github.com/real-uangi/allingo/common/result"
	"net/http"
)

func JsonFunc[I any, O any](f func(I) (O, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		jsonPrecess(c, f)
	}
}

func jsonPrecess[I any, O any](c *gin.Context, f func(I) (O, error)) {
	var input I
	err := c.BindJSON(&input)
	if err != nil {
		c.Render(http.StatusBadRequest, result.BadRequest(err))
		return
	}
	output, err := f(input)
	if err != nil {
		c.Render(HandleErr(err))
	} else {
		c.Render(http.StatusOK, result.Ok(output))
	}
}

func QueryFunc[I any, O any](f func(I) (O, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		queryPrecess(c, f)
	}
}

func queryPrecess[I any, O any](c *gin.Context, f func(I) (O, error)) {
	var input I
	err := c.BindQuery(&input)
	if err != nil {
		c.Render(http.StatusBadRequest, result.BadRequest(err))
		return
	}
	output, err := f(input)
	if err != nil {
		c.Render(HandleErr(err))
	} else {
		c.Render(http.StatusOK, result.Ok(output))
	}
}

func SingleQueryFunc[O any](f func(string) (O, error), name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		singleQueryPrecess(c, f, name)
	}
}

func singleQueryPrecess[O any](c *gin.Context, f func(string) (O, error), name string) {
	s := c.Query(name)
	if s == "" {
		c.Render(http.StatusBadRequest, result.BadRequest(business.NewErrorWithCode(name+" 不能为空", http.StatusBadRequest)))
		return
	}
	output, err := f(s)
	if err != nil {
		c.Render(HandleErr(err))
	} else {
		c.Render(http.StatusOK, result.Ok(output))
	}
}

func NoArgsFunc[O any](f func() (O, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		noArgsProcess(c, f)
	}
}

func noArgsProcess[O any](c *gin.Context, f func() (O, error)) {
	output, err := f()
	if err != nil {
		c.Render(HandleErr(err))
	} else {
		c.Render(http.StatusOK, result.Ok(output))
	}
}

func HandleErr(err error) (int, *result.Result[result.RawField]) {
	return result.FromError(err)
}
