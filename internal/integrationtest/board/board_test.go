// This file is part of arduino-cli.
//
// Copyright 2022 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to
// modify or otherwise use the software for commercial activities involving the
// Arduino software without disclosing the source code of your own applications.
// To purchase a commercial license, send an email to license@arduino.cc.

package board_t_test

import (
	"os"
	"testing"

	"github.com/arduino/arduino-cli/internal/integrationtest"
	"github.com/stretchr/testify/require"
	"go.bug.st/testifyjson/requirejson"
)

func TestBoardList(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("VMs have no serial ports")
	}

	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	stdout, _, err := cli.Run("board", "list", "--format", "json")
	require.NoError(t, err)
	// check is a valid json and contains a list of ports
	requirejson.Query(t, stdout, ".[] | .port | .protocol!=\"\"", "true")
	requirejson.Query(t, stdout, ".[] | .port | .protocol_label!=\"\"", "true")
}

func TestBoardListWithInvalidDiscovery(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("core", "update-index")
	require.NoError(t, err)
	_, _, err = cli.Run("board", "list")
	require.NoError(t, err)

	// check that the CLI does not crash if an invalid discovery is installed
	// (for example if the installation fails midway).
	// https://github.com/arduino/arduino-cli/issues/1669
	toolDir := cli.DataDir().Join("packages", "builtin", "tools", "serial-discovery")
	dirToRemove, err := toolDir.ReadDir()
	require.NoError(t, err)
	content, err := dirToRemove[0].ReadDir()
	require.NoError(t, err)
	for _, d := range content {
		err = d.Remove()
		require.NoError(t, err)
	}

	_, stderr, err := cli.Run("board", "list")
	require.NoError(t, err)
	require.Contains(t, string(stderr), "builtin:serial-discovery")
}

func TestBoardListall(t *testing.T) {
	env, cli := integrationtest.CreateArduinoCLIWithEnvironment(t)
	defer env.CleanUp()

	_, _, err := cli.Run("update")
	require.NoError(t, err)
	_, _, err = cli.Run("core", "install", "arduino:avr@1.8.3")
	require.NoError(t, err)

	stdout, _, err := cli.Run("board", "listall", "--format", "json")
	require.NoError(t, err)

	requirejson.Query(t, stdout, ".boards | length", "26")

	requirejson.Contains(t, stdout, `{
		"boards":[
			{
				"name": "Arduino Yún",
				"fqbn": "arduino:avr:yun",
				"platform": {
					"id": "arduino:avr",
					"installed": "1.8.3",
					"name": "Arduino AVR Boards"
				}
			},
			{
				"name": "Arduino Uno",
				"fqbn": "arduino:avr:uno",
				"platform": {
					"id": "arduino:avr",
					"installed": "1.8.3",
					"name": "Arduino AVR Boards"
				}
			}
		]
	}`)

	// Check if the boards' "latest" value is not empty
	requirejson.Query(t, stdout, ".boards | .[] | .platform | .latest != \"\"", "true")
}
