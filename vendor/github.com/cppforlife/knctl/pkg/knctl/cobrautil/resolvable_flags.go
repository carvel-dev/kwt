/*
Copyright 2018 The Knative Authors

Licensed under the Apache License, Open 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cobrautil

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type ResolvableFlag interface {
	Resolve() error
}

func ResolveFlagsForCmd(cmd *cobra.Command) {
	origRunE := cmd.RunE
	cmd.RunE = func(cmd2 *cobra.Command, args []string) error {
		var lastFlagErr error
		cmd2.Flags().VisitAll(func(flag *pflag.Flag) {
			if flag.Value == nil {
				return
			}
			if resolvableVal, ok := flag.Value.(ResolvableFlag); ok {
				err := resolvableVal.Resolve()
				if err != nil {
					lastFlagErr = err
				}
			}
		})
		if lastFlagErr != nil {
			return lastFlagErr
		}
		return origRunE(cmd2, args)
	}
}
