/*
 * Copyright © 2022 Andrey Kuvshinov. Contacts: <syslinux@protonmail.com>
 * Copyright © 2022 Eltaline OU. Contacts: <eltaline.ou@gmail.com>
 *
 * This file is part of eCrond.
 *
 * eCrond is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * eCrond is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"flag"
	"fmt"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
	"github.com/gookit/validate"
	"github.com/rjeczalik/notify"
	"github.com/rs/zerolog"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
)

// Global types/methods

type Settings struct {
	Recursive bool
	Commands  []string
}

func (c *Settings) AddCommand(comm string) {
	c.Commands = append(c.Commands, comm)
}

// Global variables

var (

	// Config

	configfile string = "/etc/ecrond/ecrond.yaml"

	// Variables

	shutdown bool = false

	tracemode bool = false
	debugmode bool = false
	testmode  bool = false

	loglevel string = "warn"

	logdir  string      = "/var/log/ecrond"
	logmode os.FileMode = 0640

	pidfile string = "/run/ecrond/ecrond.pid"

	esettings = make(map[string]Settings)

)

// Main function

func main() {

	var err error

	var version string = "1.0.0"
	var vprint bool = false
	var help bool = false

	// Command line options

	flag.StringVar(&configfile, "config", configfile, "--config=/etc/ecrond/ecrond.yaml")
	flag.BoolVar(&tracemode, "trace", tracemode, "--trace - trace mode")
	flag.BoolVar(&debugmode, "debug", debugmode, "--debug - debug mode")
	flag.BoolVar(&testmode, "test", testmode, "--test - test mode")
	flag.BoolVar(&vprint, "version", vprint, "--version - print version")
	flag.BoolVar(&help, "help", help, "--help - displays help")

	flag.Parse()

	switch {
	case vprint:
		fmt.Printf("eCrond Version: %s\n", version)
		os.Exit(0)
	case help:
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Load configuration

	// config.WithOptions(config.ParseEnv)

	config.AddDriver(yaml.Driver)

	err = config.LoadFiles(configfile)
	if err != nil {
		fmt.Printf("Can`t decode config file | File [%s] | %v\n", configfile, err)
		os.Exit(1)
	}

	// fmt.Printf("config data: \n %#v\n", config.Data())

	// Validate configuration

	v := validate.Map(config.Data())
	v.StringRule("tracemode", "bool")
	v.StringRule("debugmode", "bool")
	v.StringRule("pidfile", "required|string|unixPath")
	v.StringRule("loglevel", "required|string|in:trace,debug,info,warn,error,fatal,panic")
	v.StringRule("logdir", "required|string|unixPath")
	v.StringRule("logmode", "required|uint")

	if !v.Validate() {
		fmt.Println(v.Errors)
		os.Exit(1)
	}

	for cpath, msettings := range config.Get("paths").(map[interface{}]interface{}) {

		pathOptions := make(map[string]interface{}, len(msettings.(map[interface{}]interface{})))
		for key, val := range msettings.(map[interface{}]interface{}) {
			pathOptions[key.(string)] = val
		}

		cpathName := make(map[string]interface{})
		cpathName["cpath"] = cpath.(string)

		v := validate.Map(cpathName)
		v.StringRule("cpath", "string|unixPath")

		if !v.Validate() {
			fmt.Println(v.Errors)
			os.Exit(1)
		}

		v = validate.Map(pathOptions)
		v.StringRule("recursive", "bool")

		if !v.Validate() {
			fmt.Println(v.Errors)
			os.Exit(1)
		}

		pathComm := make(map[string]interface{})

		for _, val := range pathOptions["commands"].([]interface{}) {
			pathComm["command"] = val.(string)
			v := validate.Map(pathComm)
			v.StringRule("command", "string")
			if !v.Validate() {
				fmt.Println(v.Errors)
				os.Exit(1)
			}
		}

	}

	// Test mode

	if testmode {
		os.Exit(0)
	}

	// Logging

	loglevel = config.String("loglevel")
	logdir = filepath.Clean(config.String("logdir"))
	logmode = os.FileMode(config.Uint("logmode"))

	logfile := filepath.Clean(logdir + "/" + "app.log")
	applogfile, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, logmode)
	if err != nil {
		fmt.Printf("Can`t open/create app log file | File [%s] | %v", logfile, err)
		os.Exit(1)
	}
	defer applogfile.Close()

	err = os.Chmod(logfile, logmode)
	if err != nil {
		fmt.Printf("Can`t chmod log file | File [%s] | %v", logfile, err)
		os.Exit(1)
	}

	switch loglevel {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	}

	if debugmode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if tracemode {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}

	zerolog.TimeFieldFormat = "02/Jan/2006:15:04:05"
	appLogger := zerolog.New(applogfile).With().Timestamp().Logger()

	// System handling

	// Get pid

	gpid, fpid := GetPID()

	// Pid file

	pidfile = filepath.Clean(config.String("pidfile"))

	switch {
	case FileExists(pidfile):

		err = os.Remove(pidfile)
		if err != nil {
			appLogger.Error().Msgf("Can`t remove pid file | File [%s] | %v", pidfile, err)
			fmt.Printf("Can`t remove pid file | File [%s] | %v\n", pidfile, err)
			os.Exit(1)
		}

		fallthrough

	default:

		err = ioutil.WriteFile(pidfile, []byte(fpid), 0644)
		if err != nil {
			appLogger.Error().Msgf("Can`t create pid file | File [%s] | %v", pidfile, err)
			fmt.Printf("Can`t create pid file | File [%s] | %v\n", pidfile, err)
			os.Exit(1)
		}

	}

	appLogger.Info().Msgf("Starting eCrond service [%s]", version)

	appLogger.Info().Msgf("Trace mode: [%t]", tracemode)
	appLogger.Info().Msgf("Debug mode: [%t]", debugmode)

	appLogger.Info().Msgf("Pid file: [%s]", pidfile)

	appLogger.Info().Msgf("Log level: [%s]", loglevel)
	appLogger.Info().Msgf("Log directory: [%s]", logdir)
	appLogger.Info().Msgf("Log mode: [%v]", logmode)

	// Populate path settings

	for cpath, msettings := range config.Get("paths").(map[interface{}]interface{}) {

		scpath := cpath.(string)

		appLogger.Info().Msgf("Path: [%s]", scpath)

		var rcrsv bool
		var comms []string

		for key, val := range msettings.(map[interface{}]interface{}) {

			if key == "recursive" {
				rcrsv = val.(bool)
			}

			if key == "commands" {
				for _, command := range val.([]interface{}) {
					scommand := command.(string)
					comms = append(comms, scommand)
					appLogger.Info().Msgf("Command: [%s]", scommand)
				}
			}
		}

		mset := esettings[scpath]

		mset.Recursive = rcrsv
		mset.Commands = comms

		esettings[scpath] = mset

	}

	// Main waitGroup

	var wg sync.WaitGroup

	// Notify directory watchers

	nc := make(chan notify.EventInfo, 1)

	for cpath, options := range esettings {

		ipath := filepath.Clean(cpath)
		iinfo, err := os.Stat(ipath)
		if err != nil {
			appLogger.Error().Msgf("Can`t stat directory from config | Directory [%s] | %v", ipath, err)
			os.Exit(1)
		}

		if iinfo.Mode().IsDir() || iinfo.Mode().IsRegular() {

			if options.Recursive {
				ipath = ipath + "/..."
			}

			err = notify.Watch(ipath, nc, notify.InCloseWrite, notify.InMovedTo)
			if err != nil {
				appLogger.Error().Msgf("Can`t create notify.InCloseWrite and notify.InMovedTo watcher | Directory [%s] | %v", ipath, err)
				fmt.Printf("Can`t create notify.InCloseWrite and notify.InMovedTo watcher | Directory [%s] | %v\n", ipath, err)
				os.Exit(1)
			}
			defer notify.Stop(nc)

			appLogger.Info().Msgf("Started watch path via notify watcher | Directory [%s]", ipath)

		} else {

			appLogger.Error().Msgf("Unknown type object | Object [%s] | %v", ipath, err)
			fmt.Printf("Unknown type object | Object [%s] | %v\n", ipath, err)
			os.Exit(1)

		}

	}

	appLogger.Info().Msgf("eCrond service running with a pid: %s", gpid)

	// Daemon channel

	// done := make(chan bool)

	// Interrupt handler

	InterruptHandler := func() {

		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		go func() {

			<-c
			shutdown = true

			// Wait go routines

			appLogger.Info().Msgf("Awaiting all go routines")

			wg.Wait()

			appLogger.Info().Msgf("Finished all go routines")

			// Shutdown message

			appLogger.Info().Msgf("Shutdown eCrond service completed")

			// Remove pid file

			if FileExists(pidfile) {
				err = os.Remove(pidfile)
				if err != nil {
					appLogger.Error().Msgf("Can`t remove pid file error | File [%s] | %v", pidfile, err)
					fmt.Printf("Can`t remove pid file error | File [%s] | %v\n", pidfile, err)
					// os.Exit(1)
				}
			}

			// done <- true
			os.Exit(0)

		}()

	}

	// Interrupt routine

	InterruptHandler()

	// Watcher handler

	WatcherHandler := func(nc chan notify.EventInfo, wg *sync.WaitGroup) {

		// Wait Group

		for {

			if shutdown {
				break
			}

			event := <-nc

			wg.Add(1)

			evpath := event.Path()
			sfpath := evpath

			_, err := os.Stat(sfpath)
			if err != nil {
				appLogger.Error().Msgf("Can`t stat file or directory via event from watcher | Path [%s] | %v", sfpath, err)
				wg.Done()
				continue
			}

			for cpath, _ := range esettings {

				if strings.Contains(sfpath, cpath) {

					mset := esettings[cpath]

					for _, command := range mset.Commands {

						output, err := QuickExec(command)
						if err != nil {
							appLogger.Error().Msgf("Run command | Monitored Path [%s] | Changed Path [%s] | Command [%s] | Output [%s] | %v", cpath, sfpath, command, output, err)
							break
						}

						appLogger.Info().Msgf("Run command | Monitored Path [%s] | Changed Path [%s] | Command [%s] | Output [%s] | %v", cpath, sfpath, command, output, err)

					}

				}

			}

			wg.Done()

		}

	}

	// Main routine

	WatcherHandler(nc, &wg)

	// Daemon channel

	// <-done

}
