/*******************************************************************************
 * Copyright 2018 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/gorilla/context"
	objectmanager "github.com/object-manager"
	app "github.com/object-manager/app"
	"github.com/object-manager/internal"
	"github.com/object-manager/internal/correlation"
	"github.com/object-manager/internal/startup"
	"github.com/object-manager/internal/usage"
)

func main() {
	start := time.Now()
	var useRegistry bool
	var useProfile string

	flag.BoolVar(&useRegistry, "registry", false, "Indicates the service should use registry.")
	flag.BoolVar(&useRegistry, "r", false, "Indicates the service should use registry.")
	flag.StringVar(&useProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&useProfile, "p", "", "Specify a profile other than default.")
	flag.Usage = usage.HelpCallback
	flag.Parse()

	params := startup.BootParams{UseRegistry: useRegistry, UseProfile: useProfile, BootTimeout: internal.BootTimeoutDefault}
	startup.Bootstrap(params, app.Retry, logBeforeInit)

	ok := app.Init(useRegistry)
	if !ok {
		logBeforeInit(fmt.Errorf("%s: Service bootstrap failed!", "my-object"))
		os.Exit(1)
	}

	app.LoggingClient.Info("Service dependencies resolved...")
	app.LoggingClient.Info(fmt.Sprintf("Starting %s %s ", "my-object", objectmanager.Version))

	http.TimeoutHandler(nil, time.Millisecond*time.Duration(app.Configuration.Service.Timeout), "Request timed out")
	app.LoggingClient.Info(app.Configuration.Service.StartupMsg)

	errs := make(chan error, 2)
	listenForInterrupt(errs)
	startHttpServer(errs, app.Configuration.Service.Port)

	app.ManagerObjectInit()

	// Time it took to start service
	app.LoggingClient.Info("Service started in: " + time.Since(start).String())
	app.LoggingClient.Info("Listening on port: " + strconv.Itoa(app.Configuration.Service.Port))
	c := <-errs
	app.Destruct()
	app.LoggingClient.Warn(fmt.Sprintf("terminating: %v", c))

	os.Exit(0)
}

func logBeforeInit(err error) {
	app.LoggingClient = logger.NewClient("my-object", false, "", models.InfoLog)
	app.LoggingClient.Error(err.Error())
}

func listenForInterrupt(errChan chan error) {
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}

func startHttpServer(errChan chan error, port int) {
	go func() {
		correlation.LoggingClient = app.LoggingClient
		r := app.LoadRestRoutes()
		errChan <- http.ListenAndServe(":"+strconv.Itoa(port), context.ClearHandler(r))
	}()
}
