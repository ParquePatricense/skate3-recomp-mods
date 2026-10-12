//go:build !windows

package main

import "errors"

func runGUI() error { return errors.New("the window is Windows-only") }

func consoleSetup(bool) {}

func uiMonitorSetting() int { return 0 }
