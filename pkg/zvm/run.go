// Copyright 2022 Tristan Isham. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/sweetbbak/zvm/pkg/meta"
)

// Run the given Zig compiler with the provided arguments
func (z *ZVM) Run(version string, cmd []string) error {
	if len(version) == 0 {
		return fmt.Errorf("no zig version provided.")
	}

	installedVersions, err := z.GetInstalledVersions()
	if err != nil {
		return err
	}

	if slices.Contains(installedVersions, version) {
		return z.runBin(version, cmd)
	}

	rawVersionStructure, err := z.fetchVersionMap()
	if err != nil {
		return err
	}

	_, err = getTarPath(version, &rawVersionStructure)
	if err != nil {
		if errors.Is(err, ErrUnsupportedVersion) {
			return fmt.Errorf("%s: %q", err, version)
		}
		return err
	}

	fmt.Printf("It looks like %s isn't installed. Would you like to install it? [y/n]\n", version)

	if !getConfirmation() {
		return fmt.Errorf("version %s is not installed", version)
	}

	if err = z.Install(version, false); err != nil {
		return err
	}

	return z.runBin(version, cmd)
}

func (z *ZVM) runBin(version string, cmd []string) error {
	zigExe := "zig"
	if runtime.GOOS == "windows" {
		zigExe = "zig.exe"
	}

	bin := strings.TrimSpace(filepath.Join(z.baseDir, version, zigExe))

	stat, err := os.Stat(bin)
	if err != nil {
		return fmt.Errorf("%w: %s", err, stat.Name())
	}

	if err := meta.Exec(bin, cmd); err != nil {
		return err
	}

	return nil

}
