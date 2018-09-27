//
// Last.Backend LLC CONFIDENTIAL
// __________________
//
// [2014] - [2018] Last.Backend LLC
// All Rights Reserved.
//
// NOTICE:  All information contained herein is, and remains
// the property of Last.Backend LLC and its suppliers,
// if any.  The intellectual and technical concepts contained
// herein are proprietary to Last.Backend LLC
// and its suppliers and may be covered by Russian Federation and Foreign Patents,
// patents in process, and are protected by trade secret or copyright law.
// Dissemination of this information or reproduction of this material
// is strictly forbidden unless prior written permission is obtained
// from Last.Backend LLC.
//

package runtime

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/lastbackend/lastbackend/pkg/distribution/errors"

	"io/ioutil"

	"github.com/lastbackend/ingress/pkg/ingress/envs"
	"github.com/lastbackend/lastbackend/pkg/distribution/types"
	"github.com/lastbackend/lastbackend/pkg/log"
)

const (
	ConfigName      = "haproxy.cfg"
	logConfigPrefix = "runtime:config"
)

func configCheck() error {

	log.Debug("config check")
	var (
		_, path, name, _ = envs.Get().GetTemplate()
	)

	cfgPath := filepath.Join(path, name)
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		log.Debug("config not found: create new")
		return configSync()
	}

	return nil
}

func configSync() error {

	log.Debug("config sync")

	var (
		routes             = envs.Get().GetState().Routes().GetRoutes()
		tpl, path, name, _ = envs.Get().GetTemplate()
	)

	log.Debugf("Update routes: %d", len(routes))

	var cfg = struct {
		Resolver string
		Routes map[string]*types.RouteManifest
	}{
		Resolver: envs.Get().GetDNSResolver(),
		Routes: routes,
	}

	buf := &bytes.Buffer{}
	tpl.Execute(buf, cfg)
	log.Debugf("config path: %s", path)

	var (
		f   *os.File
		err error
	)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Debugf("config direcotry not exists: %s", path)
		if err := os.MkdirAll(path, 0644); err != nil {
			log.Errorf("can not create config dir: %s", err.Error())
			return err
		}
	}

	if name == types.EmptyString {
		name = ConfigName
	}

	cfgPath := filepath.Join(path, name)
	testPath := fmt.Sprintf("%s.test", cfgPath)

	f, err = os.Open(testPath)
	if os.IsNotExist(err) {
		log.Debugf("config file not exists: %s", testPath)
		f, err = os.Create(testPath)
		if err != nil {
			log.Errorf("can not create config file: %s", err.Error())
		}
	}
	f.Close()

	if err := ioutil.WriteFile(testPath, buf.Bytes(), 0644); err != nil {
		log.Errorf("can no write test config: %s", err.Error())
		return err
	}

	if err := configValidate(testPath); err != nil {
		log.Errorf("config is not working (%s)", err.Error())
		return err
	}

	f, err = os.Open(cfgPath)
	if os.IsNotExist(err) {
		log.Debugf("config file not exists: %s", cfgPath)
		f, err = os.Create(cfgPath)
		if err != nil {
			log.Errorf("can not create config file: %s", err.Error())
		}
	}
	f.Close()

	return ioutil.WriteFile(cfgPath, buf.Bytes(), 0644)
}

func configValidate(path string) error {

	log.Debugf("%s:> config validate", logConfigPrefix)

	var hpbin = envs.Get().GetHaproxy()

	cmd := exec.Command(hpbin, "-c", "-V", "-f", path)
	err := cmd.Start()

	if err != nil {
		log.Errorf("can not check config: %s", err.Error())
		return err
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// This works on both Unix and Windows. Although package
			// syscall is generally platform dependent, WaitStatus is
			// defined for both Unix and Windows and in both cases has
			// an ExitStatus() method with the same signature.
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() != 0 {
					return errors.New(string(exiterr.Stderr))
				}
			}
		} else {
			log.Fatalf("cmd.Wait: %v", err)
		}
	}

	return nil
}
