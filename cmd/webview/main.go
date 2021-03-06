//
// Copyright (c) 2019-present Codist <countstarlight@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// Written by Codist <countstarlight@gmail.com>, March 2019
//

package main

import (
	"fmt"
	"github.com/countstarlight/homo/cmd/webview/config"
	"github.com/countstarlight/homo/module/sphinx"
	"github.com/countstarlight/homo/module/view"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"runtime"
	"strings"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//setting.MaxConcurrency = runtime.NumCPU() * 2

	// initialize global pseudo random generator
	// rand.Seed(time.Now().Unix())
}

var flags = []cli.Flag{
	cli.BoolFlag{
		EnvVar: "HOMO_WEBVIEW_DEBUG",
		Name:   "debug, d",
		Usage:  "start homo webview in debug mode",
	},
	cli.BoolFlag{
		EnvVar:      "HOMO_WEBVIEW_OFFLINE",
		Name:        "offline, o",
		Usage:       "disable speech recognition and text to speech",
		Destination: &config.OfflineMode,
	},
	cli.BoolFlag{
		EnvVar:      "HOMO_WEBVIEW_FALL",
		Name:        "fall, f",
		Usage:       "disable wakeup",
		Destination: &config.WakeUpDisabled,
	},
	cli.BoolFlag{
		EnvVar:      "HOMO_WEBVIEW_INTERRUPT",
		Name:        "interrupt, i",
		Usage:       "can interrupt the playing voice",
		Destination: &config.InterruptMode,
	},
}

// Greeting list
var Greetings = []string{"你好，我是你的智能助理", "有什么我能帮你的吗？"}

func main() {
	app := cli.NewApp()
	app.Name = config.AppName
	app.Version = config.AppVersion
	app.Usage = "Help"
	app.Action = lanchWebview
	app.Flags = flags
	app.Before = before
	app.After = config.Terminal
	if err := app.Run(os.Args); err != nil {
		logrus.Fatalf("[homo-webview]%s", err.Error())
	}
}

func lanchWebview(ctx *cli.Context) {
	if ctx.Bool("debug") {
		config.DebugMode = true
		// Set logrus format
		// Print file name and line code
		logrus.SetReportCaller(true)
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "15:04:05",
			// Show colorful on windows
			ForceColors:   true,
			FullTimestamp: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				repopath := fmt.Sprintf("%s/src/github.com/countstarlight/homo/", os.Getenv("GOPATH"))
				filename := strings.Replace(f.File, repopath, "", -1)
				r := strings.Split(f.Function, ".")
				return fmt.Sprintf("%s()", r[len(r)-1]), fmt.Sprintf("%s:%d", filename, f.Line)
			},
		})
		logrus.Infof("Running in debug mode")
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "15:04:05",
			// Show colorful on windows
			ForceColors:   true,
			FullTimestamp: true,
		})
	}

	if config.OfflineMode {
		logrus.Warnf("注意：当前处于离线模式，语音识别和语音合成将不可用")
	}
	// Init webview
	view.InitWebView(config.AppName, config.DebugMode)
	//defer w.Exit()
	//
	// Prepare wake up function
	//
	if config.WakeUpDisabled {
		if !config.OfflineMode {
			go sphinx.LoadCMUSphinx()
			config.WakeUpd = true
		}
		config.IsPlayingVoice = true
	} else {
		config.WakeUpWait.Add(1)
		go sphinx.LoadCMUSphinx()
		config.WakeUpWait.Wait()
	}

	logrus.Infof("唤醒成功，开始唤起界面...")

	go func() {
		//Greeting := Greetings[rand.Intn(len(Greetings))]
		if !config.OfflineMode {
			view.SendReplyWithVoice(Greetings)
		} else {
			view.SendReply(Greetings)
		}
	}()

	view.Run()
}

func before(c *cli.Context) error {
	config.LoadConfig()
	return nil
}
