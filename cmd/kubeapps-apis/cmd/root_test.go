/*
Copyright 2021 VMware. All Rights Reserved.

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

package cmd

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/kubeapps/kubeapps/cmd/kubeapps-apis/server"
)

func TestParseFlagsCorrect(t *testing.T) {
	var tests = []struct {
		name string
		args []string
		conf server.ServeOptions
	}{
		{
			"all arguments are captured",
			[]string{
				"--config", "file",
				"--port", "901",
				"--plugin-dir", "foo01",
				"--clusters-config-path", "foo02",
				"--pinniped-proxy-url", "foo03",
				"--unsafe-use-demo-sa", "true",
				"--unsafe-local-dev-kubeconfig", "true",
			},
			server.ServeOptions{
				Port:                     901,
				PluginDirs:               []string{"foo01"},
				ClustersConfigPath:       "foo02",
				PinnipedProxyURL:         "foo03",
				UnsafeUseDemoSA:          true,
				UnsafeLocalDevKubeconfig: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newRootCmd()
			b := bytes.NewBufferString("")
			cmd.SetOut(b)
			cmd.SetErr(b)
			setFlags(cmd)
			cmd.SetArgs(tt.args)
			cmd.Execute()
			if got, want := serveOpts, tt.conf; !cmp.Equal(want, got) {
				t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(want, got))
			}
		})
	}
}
