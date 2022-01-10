// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows
// +build windows

// Package winuntil contains misc Windows/win32 helper functions.
package winutil

import (
	"log"
	"syscall"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	RegBase       = `SOFTWARE\Tailscale IPN`
	regPolicyBase = `SOFTWARE\Policies\Tailscale`
)

// GetDesktopPID searches the PID of the process that's running the
// currently active desktop and whether it was found.
// Usually the PID will be for explorer.exe.
func GetDesktopPID() (pid uint32, ok bool) {
	hwnd := windows.GetShellWindow()
	if hwnd == 0 {
		return 0, false
	}
	windows.GetWindowThreadProcessId(hwnd, &pid)
	return pid, pid != 0
}

// GetPolicyString looks up a registry value in our local machine's path for
// system policies, or returns the given default if it can't.
//
// This function will only work on GOOS=windows. Trying to run it on any other
// OS will always return the default value.
func GetPolicyString(name, defval string) string {
	s, err := getRegString(regPolicyBase, name)
	if err != nil {
		// Fall back to the legacy path for policies
		return GetRegString(name, defval)
	}
	return s
}

// GetPolicyInteger looks up a registry value in our local machine's path for
// system policies, or returns the given default if it can't.
//
// This function will only work on GOOS=windows. Trying to run it on any other
// OS will always return the default value.
func GetPolicyInteger(name string, defval uint64) uint64 {
	i, err := getRegInt(regPolicyBase, name)
	if err != nil {
		// Fall back to the legacy path for policies
		return GetRegInteger(name, defval)
	}
	return i
}

// GetRegString looks up a registry path in our local machine path, or returns
// the given default if it can't.
//
// This function will only work on GOOS=windows. Trying to run it on any other
// OS will always return the default value.
func GetRegString(name, defval string) string {
	s, err := getRegString(RegBase, name)
	if err != nil {
		return defval
	}
	return s
}

// GetRegInteger looks up a registry path in our local machine path, or returns
// the given default if it can't.
//
// This function will only work on GOOS=windows. Trying to run it on any other
// OS will always return the default value.
func GetRegInteger(name string, defval uint64) uint64 {
	i, err := getRegInt(RegBase, name)
	if err != nil {
		return defval
	}
	return i
}

func getRegString(subKey, valueName string) (string, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, subKey, registry.READ)
	if err != nil {
		log.Printf("registry.OpenKey(%v): %v", subKey, err)
		return "", err
	}
	defer key.Close()

	val, _, err := key.GetStringValue(name)
	if err != nil {
		if err != registry.ErrNotExist {
			log.Printf("registry.GetStringValue(%v): %v", name, err)
		}
		return "", err
	}
	return val, nil
}

func getRegInt(subKey, valueName string) (uint64, error) {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, subKey, registry.READ)
	if err != nil {
		log.Printf("registry.OpenKey(%v): %v", subKey, err)
		return 0, err
	}
	defer key.Close()

	val, _, err := key.GetIntegerValue(name)
	if err != nil {
		log.Printf("registry.GetIntegerValue(%v): %v", name, err)
		return 0, err
	}
	return val, nil
}

var (
	kernel32                         = syscall.NewLazyDLL("kernel32.dll")
	procWTSGetActiveConsoleSessionId = kernel32.NewProc("WTSGetActiveConsoleSessionId")
)

// TODO(crawshaw): replace with x/sys/windows... one day.
// https://go-review.googlesource.com/c/sys/+/331909
func WTSGetActiveConsoleSessionId() uint32 {
	r1, _, _ := procWTSGetActiveConsoleSessionId.Call()
	return uint32(r1)
}
