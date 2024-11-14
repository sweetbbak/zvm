// Copyright 2022 Tristan Isham. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

var ErrNoSettings = errors.New("settings.json not found")

func Initialize() *ZVM {
	zvm_path, ok := os.LookupEnv("ZVM_PATH")

	if !ok {
		home, ok := os.LookupEnv("XDG_DATA_DIR")
		if !ok {
			home, _ = os.UserHomeDir()
			zvm_path = filepath.Join(home, ".zvm")
		} else {
			zvm_path = filepath.Join(home, "zvm")
		}
	}

	if _, err := os.Stat(zvm_path); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(filepath.Join(zvm_path, "self"), 0775); err != nil {
			log.Fatal(err)
		}
	}

	zvm := &ZVM{
		baseDir: zvm_path,
	}

	zvm.Settings.path = filepath.Join(zvm_path, "settings.json")

	if err := zvm.loadSettings(); err != nil {
		if errors.Is(err, ErrNoSettings) {
			zvm.Settings = Settings{
				UseColor:           true,
				VersionMapUrl:      "https://ziglang.org/download/index.json",
				AlwaysForceInstall: false,
				ZlsVMU:             "https://releases.zigtools.org/",
			}

			out_settings, err := json.MarshalIndent(&zvm.Settings, "", "    ")
			if err != nil {
				log.Warn("Unable to generate settings.json file", err)
			}

			if err := os.WriteFile(filepath.Join(zvm_path, "settings.json"), out_settings, 0755); err != nil {
				log.Warn("Unable to create settings.json file", err)
			}
		}
	}

	return zvm
}

type ZVM struct {
	baseDir  string
	Settings Settings
}

// A representaiton of the offical json schema for Zig versions
type zigVersionMap = map[string]zigVersion

// LoadMasterVersion takes a zigVersionMap and returns the master disto's version if it's present.
// If it's not, this function returns an empty string.
func LoadMasterVersion(zigMap *zigVersionMap) string {
	if ver, ok := (*zigMap)["master"]["version"].(string); ok {
		return ver
	}
	return ""
}

// A representation of individual Zig versions
type zigVersion = map[string]any

type ZigOnlVersion = map[string][]map[string]string

func validVmuAlis(version string) bool {
	return version == "default" || version == "mach"
}

func (z ZVM) getVersion(version string) error {
	if _, err := os.Stat(filepath.Join(z.baseDir, version)); err != nil {
		return err
	}

	targetZig := strings.TrimSpace(filepath.Join(z.baseDir, version, "zig"))
	var zigVersion strings.Builder

	cmd := exec.Command(targetZig, "version")
	cmd.Stdout = &zigVersion

	err := cmd.Run()
	if err != nil {
		log.Warn(err)
	}

	outputVersion := strings.TrimSpace(zigVersion.String())

	log.Debug("getVersion:", "output", outputVersion, "version", version, "program", targetZig)

	if version == outputVersion {
		return nil
	} else {
		if _, statErr := os.Stat(targetZig); statErr == nil || version == "master" {
			return nil
		}
		return fmt.Errorf("version %s is not a released version", version)
	}
}

func (z *ZVM) loadSettings() error {
	set_path := z.Settings.path
	if _, err := os.Stat(set_path); errors.Is(err, os.ErrNotExist) {
		return ErrNoSettings
	}

	data, err := os.ReadFile(set_path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &z.Settings)
}
