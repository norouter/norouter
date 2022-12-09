/*
   Copyright (C) NoRouter authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/mattn/go-isatty"
	"github.com/norouter/norouter/cmd/norouter/editorcmd"
	"github.com/norouter/norouter/pkg/manager"
	"github.com/norouter/norouter/pkg/manager/manifest"
	"github.com/norouter/norouter/pkg/manager/manifest/parsed"

	"github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var managerCommand = &cli.Command{
	Name:      "manager",
	Aliases:   []string{"m"},
	Usage:     "manager (default subcommand)",
	ArgsUsage: "[FILE]",

	Flags:  managerFlags,
	Action: managerAction,
}

var managerFlags = []cli.Flag{
	&cli.BoolFlag{
		Name:    "open-editor",
		Aliases: []string{"e"},
		Usage:   "open an editor for a temporary manifest file, with an example content",
	},
}

var sigCh = make(chan os.Signal)

func managerAction(clicontext *cli.Context) error {
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	openEditor := clicontext.Bool("open-editor")
	manifestPath := clicontext.Args().First()
	if openEditor {
		if manifestPath != "" {
			return errors.New("manifest file should not be specified when `--open-editor` is specified")
		}
		return runManagerWithEditor()
	}
	if manifestPath == "" {
		return fmt.Errorf("no manifest file path was specified, run `%s show-example` to show an example, or run `%s --open-editor` to open an editor with an example file",
			os.Args[0], os.Args[0])
	}
	err := runManager(manifestPath)
	if err == errInterrupted {
		logrus.Info("Interrupted. Exiting...")
		err = nil
	}
	return err
}

func runManagerWithEditor() error {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return errors.New("`--open-editor` requires stdout to be a terminal")
	}
	editor := editorcmd.Detect()
	if editor == "" {
		return errors.New("could not detect a text editor binary, try setting $EDITOR")
	}
	manifestFile, err := os.CreateTemp("", "norouter-editor-")
	if err != nil {
		return err
	}
	manifestPath := manifestFile.Name()
	defer os.RemoveAll(manifestPath)
	hdr := `# Example manifest for NoRouter.
`
	if err := os.WriteFile(manifestPath, []byte(exampleManifest(hdr)), 0o600); err != nil {
		return err
	}
	logrus.Debugf("Temporary manifest file: %q", manifestPath)
	for {
		fi, err := os.Stat(manifestPath)
		if err != nil {
			return err
		}
		modTime := fi.ModTime()
		editorCmd := exec.Command(editor, manifestPath)
		editorCmd.Env = os.Environ()
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr
		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("could not execute editor %q for a file %q: %w", editor, manifestPath, err)
		}
		fi, err = os.Stat(manifestPath)
		if err != nil {
			return err
		}
		if fi.ModTime() == modTime {
			logrus.Info("The manifest file was not modified. Exiting.")
			return nil
		}
		runErr := runManager(manifestPath)
		if runErr == nil {
			return nil
		}
		if runErr == errInterrupted {
			logrus.Warn("Interrupted. Press Ctrl-C (again) to exit, press RETURN to reopen the editor.")
		} else {
			logrus.WithError(runErr).Warn("Failed to run the manager. Press Ctrl-C (again) to exit, press RETURN to reopen the editor.")
		}
		waitPressReturn()
	}
}

func waitPressReturn() {
	ch := make(chan struct{})
	go func() {
		bufio.NewReader(os.Stdin).ReadString('\n')
		ch <- struct{}{}
	}()
	noInputCh := time.After(30 * time.Second)
	tickCh := time.Tick(3 * time.Second)
	for {
		select {
		case <-sigCh:
			logrus.Info("Exiting.")
			os.Exit(0)
		case <-tickCh:
			logrus.Info("Waiting for keyboard input, press RETURN or Ctrl-C")
		case <-noInputCh:
			logrus.Errorf("Did not get any keyboard input. Exiting.")
			os.Exit(1)
		case <-ch:
			return
		}
	}
}

var errInterrupted = errors.New("interrupted")

func runManager(manifestPath string) error {
	parsed, err := loadManifest(manifestPath)
	if err != nil {
		return err
	}
	logrus.Debugf("parsed: %s", spew.Sdump(parsed))
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	ccSet, err := manager.NewCmdClientSet(ctx, parsed)
	if err != nil {
		return err
	}
	for vip, client := range ccSet.ByVIP {
		logrus.Debugf("client for %q: %q", vip, client.String())
	}
	m, err := manager.New(ccSet)
	if err != nil {
		return err
	}
	errCh := make(chan error)
	go func() {
		errCh <- m.Run()
	}()
	select {
	case <-sigCh:
		cancel()
		return errInterrupted
	case err := <-errCh:
		return err
	}
}

func loadManifest(filePath string) (*parsed.ParsedManifest, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var raw manifest.Manifest
	if err := yaml.Unmarshal(b, &raw); err != nil {
		if strings.Contains(err.Error(), "found character that cannot start any token") {
			err = fmt.Errorf("failed to parse YAML, maybe you are mixing up tabs and spaces? YAML does not allow tabs.: %w", err)
		}
		return nil, err
	}
	var ignored manifest.Manifest
	if err := yaml.UnmarshalWithOptions(b, &ignored, yaml.Strict()); err != nil {
		logrus.WithError(err).Warn("The manifest seems to have unknown fields. Ignoring.")
	}
	return parsed.New(&raw)
}
