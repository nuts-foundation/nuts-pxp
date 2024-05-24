/*
 * Nuts node
 * Copyright (C) 2021 Nuts community
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"context"
	"errors"
	"github.com/nuts-foundation/nuts-pxp/policy"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nuts-foundation/nuts-pxp/api/opa"
	"github.com/nuts-foundation/nuts-pxp/api/pip"
	"github.com/nuts-foundation/nuts-pxp/config"
	"github.com/nuts-foundation/nuts-pxp/db"
)

func main() {
	// Listen for interrupt signals (CTRL/CMD+C, OS instructing the process to stop) to cancel context.
	ctx, cancelNotify := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancelNotify()

	// read config using koanf
	cfg := config.Config{}
	if err := cfg.Load(); err != nil {
		panic(err)
	}

	// init DB
	sqlDb, err := db.New(cfg)
	if err != nil {
		panic(err)
	}
	defer func() {
		err := sqlDb.Close()
		if err != nil {
			panic(err)
		}
	}()

	// init OPA
	decisionMaker, err := policy.New(cfg, sqlDb)
	if err != nil {
		panic(err)
	}

	// init http server and bind to localhost:8080
	echoServer := echo.New()
	echoServer.HTTPErrorHandler = errorHandlerfunc
	echoServer.HideBanner = true
	echoServer.HidePort = true

	// init API & register routes
	pipController := &pip.Wrapper{DB: sqlDb}
	opaController := &opa.Wrapper{DecisionMaker: decisionMaker}
	pip.RegisterHandlers(echoServer, pip.NewStrictHandler(pipController, []pip.StrictMiddlewareFunc{}))
	opa.RegisterHandlers(echoServer, opa.NewStrictHandler(opaController, []opa.StrictMiddlewareFunc{}))

	// Start server
	go func() {
		if err := echoServer.Start(":8080"); err != nil && errors.Is(err, http.ErrServerClosed) {
			panic("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := echoServer.Shutdown(ctx); err != nil {
		panic(err)
	}
}

func errorHandlerfunc(err error, ctx echo.Context) {
	if !ctx.Response().Committed {
		ctx.Response().Write([]byte(err.Error()))
	}
}
