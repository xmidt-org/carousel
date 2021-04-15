package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xmidt-org/carousel/model"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// In the following tests, there's a lot going on.
// The following is based on:
// http://lucapette.me/writing-integration-tests-for-a-go-cli-application

var update = flag.Bool("update", false, "update golden files")

const binaryName = "carousel"

var binaryPath string

var prefixRegex = regexp.MustCompile(`(.*?)\.`)

/*
Test Cases
 - SimpleStep
 - N Module deep
 - No Output
 - Count leverage as module

*/

func diff(expected, actual interface{}) []string {
	return pretty.Diff(expected, actual)
}

// setupTestingDir create the testing_dir with the contents of the test case directory.
// The path of the testing directory is returned.
// If the testcase is not found, the test fails.
func setupTestingDir(t *testing.T, testcase string) string {
	wd, _ := os.Getwd()
	t.Helper()
	if testcase != "" {
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("problems recovering caller information")
		}

		testingDir := filepath.Join(filepath.Dir(filename), "testing_dir")
		// remove existing folder, if there
		_ = RemoveContents(testingDir)
		_ = os.Remove(testingDir)
		err := os.Mkdir(testingDir, 0755)
		if err != nil {
			t.Fatalf("failed to creating testing_folder %s", err)
		}
		wd = testingDir

		// copy .tf file and .yaml contents to testing dir
		terraformTestCaseDir := filepath.Join(filepath.Dir(filename), "terraform_test_cases", testcase)
		err = CopyDirectory(terraformTestCaseDir, testingDir, "tfstate", ".terraform")
		if err != nil {
			t.Fatalf("failed to creating copy test case %s: %s", testcase, err)
		}
	}
	return wd
}

// setupTerraformState prepares the terraform to be in a certain state by running terraform commands in a fixture file
// If any command fails the test fails.
func setupTerraformState(t *testing.T, fixtureFile string, wd string) {
	t.Helper()
	// setup with fixture file
	if fixtureFile != "" {
		fixture := newFixtureFile(t, fixtureFile)
		setupStepsContents := fixture.load()
		setupSteps := strings.Split(setupStepsContents, "\n")
		for _, setupStep := range setupSteps {
			args := strings.Split(setupStep, " ")
			cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
			cmd.Dir = wd

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("%s\nsetup file %s: %v", output, fixtureFile, err)
			}
		}
	}
}

func getCurrentHosts(t *testing.T, wd string) []string {
	require := require.New(t)

	fullStateCmd := exec.Command(binaryPath, "state", "--json", "--full")
	fullStateCmd.Dir = wd
	clusterStateBytes, err := fullStateCmd.CombinedOutput()
	require.NoError(err)
	cs := model.Cluster{}
	err = json.Unmarshal(clusterStateBytes, &cs)
	require.NoError(err, string(clusterStateBytes))
	currenthosts := make([]string, 0)
	for _, group := range cs {
		for _, host := range group.Hosts {
			if host != "" {
				currenthosts = append(currenthosts, host)
			} else {
				// TODO this should not happen
				t.Fatal(string(clusterStateBytes))
			}
		}
	}
	return currenthosts
}

func TestRolloutCLI(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		golden            string
		fixture           string
		terraformTestCase string
		wantErr           bool
	}{
		{
			name:              "simple state",
			args:              []string{"rollout", "-q", "3", "0.3.1"},
			golden:            "rollout.golden",
			fixture:           "rollout.tmp",
			terraformTestCase: "basic",
			wantErr:           false,
		},
		{
			name:              "invalid state",
			args:              []string{"rollout", "-q", "3", "0.3.1"},
			golden:            "invalid_tf_file.golden",
			terraformTestCase: "invalid",
			wantErr:           true,
		},
		{
			name:              "simple state with count",
			args:              []string{"rollout", "-q", "3", "0.3.1"},
			golden:            "rollout.golden",
			fixture:           "rollout.tmp",
			terraformTestCase: "basicCount",
			wantErr:           false,
		},
		{
			name:              "nested",
			args:              []string{"rollout", "-q", "3", "0.3.1"},
			golden:            "rollout.golden",
			fixture:           "rollout.tmp",
			terraformTestCase: "nestedModules",
		},
		{
			name:              "dry run",
			args:              []string{"rollout", "-d", "3", "0.3.1"},
			golden:            "dry_run.golden",
			fixture:           "rollout.tmp",
			terraformTestCase: "basic",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// setup workspace
			wd := setupTestingDir(t, tt.terraformTestCase)
			// get to known state
			setupTerraformState(t, tt.fixture, wd)

			// execute the cli
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = wd
			output, err := cmd.CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", output, tt.wantErr, err != nil, err)
			}
			actual := string(output)

			// get the state from the cli
			stateCmd := exec.Command(binaryPath, "state", "--json")
			stateCmd.Dir = wd
			stateOutput, err := stateCmd.CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", stateOutput, tt.wantErr, err != nil, err)
			}
			// actualState := string(stateOutput)

			golden := newGoldenFile(t, tt.golden)

			// TODO: Figure out better way of validating output
			// Output can be random since hostnames have a random generator build in.

			if *update {
				golden.write(actual)
			}
			expected := golden.load()
			assert.Equal(expected, actual)
		})
	}
}

func TestResumeCLI(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		golden            string
		fixture           string
		terraformTestCase string
		wantErr           bool
	}{
		{
			name:              "simple state resume",
			args:              []string{"resume", "err.json", "-q"},
			golden:            "stateResume.golden",
			fixture:           "stateResume.tmp",
			terraformTestCase: "basicResume",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			// setup workspace
			wd := setupTestingDir(t, tt.terraformTestCase)
			// get to known state
			setupTerraformState(t, tt.fixture, wd)

			// create err file

			// execute the cli
			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = wd
			output, err := cmd.CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", output, tt.wantErr, err != nil, err)
			}
			actual := string(output)

			// get the state from the cli
			stateCmd := exec.Command(binaryPath, "state", "--json")
			stateCmd.Dir = wd
			stateOutput, err := stateCmd.CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", stateOutput, tt.wantErr, err != nil, err)
			}
			// actualState := string(stateOutput)

			golden := newGoldenFile(t, tt.golden)

			// TODO: Figure out better way of validating output
			// Output can be random since hostnames have a random generator build in.

			if *update {
				golden.write(actual)
			}
			expected := golden.load()
			assert.Equal(expected, actual)
		})
	}
}

func TestStateCLI(t *testing.T) {
	tests := []struct {
		name              string
		args              []string
		golden            string
		fixture           string
		terraformTestCase string
		wantErr           bool
	}{
		{
			name:    "no arguments",
			args:    []string{},
			golden:  "help.golden",
			wantErr: true,
		},
		{
			name:              "simple state",
			args:              []string{"state", "--json"},
			golden:            "state.golden",
			fixture:           "state.tmp",
			terraformTestCase: "basic",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)
			// setup workspace
			wd := setupTestingDir(t, tt.terraformTestCase)

			// setup with fixture file
			if tt.fixture != "" {
				fixture := newFixtureFile(t, tt.fixture)
				setupStepsContents := fixture.load()
				setupSteps := strings.Split(setupStepsContents, "\n")
				for _, setupStep := range setupSteps {
					args := strings.Split(setupStep, " ")
					cmd := exec.Command(args[0], args[1:]...) //nolint:gosec
					cmd.Dir = wd

					output, err := cmd.CombinedOutput()
					if err != nil {
						t.Fatalf("%s\nexpected no error or setup file %s", output, tt.fixture)
					}
				}
			}

			cmd := exec.Command(binaryPath, tt.args...)
			cmd.Dir = wd
			output, err := cmd.CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", output, tt.wantErr, err != nil, err)
			}
			actual := string(output)

			golden := newGoldenFile(t, tt.golden)

			if *update {
				golden.write(actual)
			}
			expected := golden.load()

			// TODO: Figure out better way of comparing strings
			if !reflect.DeepEqual(expected, actual) {
				if !assert.JSONEq(expected, actual) {
					diff(expected, actual)
				}
			}
		})
	}
}

func TestRolloutWithPlugin(t *testing.T) {
	// If using workspaces tainting will failing in terraform < 0.14.x
	tests := []struct {
		name              string
		args              []string
		golden            string
		fixture           string
		terraformTestCase string
		wantErr           bool
	}{
		{
			name:              "simple state",
			args:              []string{"rollout", "-q", "3", "0.3.1"},
			golden:            "rolloutPlugin.golden",
			fixture:           "rolloutPlugin.tmp",
			terraformTestCase: "basic",
			wantErr:           false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// setup workspace
			wd := setupTestingDir(t, tt.terraformTestCase)
			// get to known state
			setupTerraformState(t, tt.fixture, wd)

			// execute the cli
			cmd := exec.Command(binaryPath, append(tt.args, "-p", "../plugin/evenHostValidator.so")...)
			cmd.Dir = wd
			output, err := cmd.CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", output, tt.wantErr, err != nil, err)
			}
			// actual := string(output)

			// get the state from the cli
			stateCmd := exec.Command(binaryPath, "state", "--json")
			stateCmd.Dir = wd
			stateOutput, err := stateCmd.CombinedOutput()
			if (err != nil) != tt.wantErr {
				t.Fatalf("%s\nexpected (err != nil) to be %v, but got %v. err: %v", stateOutput, tt.wantErr, err != nil, err)
			}
			actualState := string(stateOutput)

			golden := newGoldenFile(t, tt.golden)

			// TODO: Figure out better way of validating output
			// Output can be random since hostnames have a random generator build in.

			if *update {
				golden.write(actualState)
			}
			expected := golden.load()
			assert.JSONEq(expected, actualState)

			if !tt.wantErr {
				// check that host only have an even ending
				hosts := getCurrentHosts(t, wd)
				for _, host := range hosts {
					prefixMatches := prefixRegex.FindStringSubmatch(host)

					c := prefixMatches[1][len(prefixMatches[1])-1:]
					val, err := strconv.Atoi(c)
					assert.NoError(err, "last character should be a number")
					assert.True(val%2 == 0, "last character should be even")
				}
			}
		})
	}
}

func TestTaintCLI(t *testing.T) {
	// If using workspaces tainting will failing in terraform < 0.14.x
	tests := []struct {
		name              string
		fixture           string
		terraformTestCase string
		wantErr           bool
	}{
		{
			name:              "simple state",
			fixture:           "taint.tmp",
			terraformTestCase: "basic",
			wantErr:           false,
		},
		{
			name:              "simple state with count",
			fixture:           "taint.tmp",
			terraformTestCase: "basicCount",
			wantErr:           false,
		},
		{
			name:              "nested",
			fixture:           "taint.tmp",
			terraformTestCase: "nestedModules",
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := assert.New(t)

			// setup workspace
			wd := setupTestingDir(t, tt.terraformTestCase)
			// get to known state
			setupTerraformState(t, tt.fixture, wd)

			// get the current hosts
			currentHosts := getCurrentHosts(t, wd)
			// Test the Taint Command
			var lastOutput string
			for _, host := range currentHosts {
				cmd := exec.Command(binaryPath, "taint", host)
				cmd.Dir = wd
				output, err := cmd.CombinedOutput()
				if !assert.NoError(err) {
					fmt.Println(string(output))
				}
				lastOutput = string(output)
			}
			// rerun setup
			setupTerraformState(t, tt.fixture, wd)

			// get the current hosts
			newHosts := getCurrentHosts(t, wd)
			// from tainting
			for _, host := range newHosts {
				if !assert.NotContains(currentHosts, host) {
					t.Log(lastOutput)
				}
			}
		})
	}
}

func TestMain(m *testing.M) {
	err := os.Chdir("..")
	if err != nil {
		fmt.Printf("could not change dir: %v", err)
		os.Exit(1)
	}

	abs, err := filepath.Abs(binaryName)
	if err != nil {
		fmt.Printf("could not get abs path for %s: %v", binaryName, err)
		os.Exit(1)
	}

	binaryPath = abs

	if err := exec.Command("make").Run(); err != nil {
		fmt.Printf("could not make binary for %s: %v", binaryName, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}
