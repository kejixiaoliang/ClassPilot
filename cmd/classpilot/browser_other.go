//go:build !windows

package main

import "errors"

func openBrowser(_ string) error {
	return errors.New("当前版本仅支持 Windows")
}
