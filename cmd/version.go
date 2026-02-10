// meta1v is a command-line tool for viewing and manipulating metadata for Canon EOS-1V files of the EFD format.
// Copyright (C) 2026  Matt F
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  `Display the version, commit hash, and build date of meta1v.`,
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Fprintf(
				os.Stdout,
				`                         _               __         
   _ _____      ____    FJ_      ___ _  / J  _    _ 
  J '_  _ `+"`"+`,   F __ J  J  _|    F __`+"`"+` L LFJ J |  | L
  | |_||_| |  | _____J | |-'   | |--| | J  LJ J  F L
  F L LJ J J  F L___--.F |__-. F L__J J J  LJ\ \/ /F
 J__L LJ J__LJ\______/F\_____/J\____,__LJ__L \\__// 
 |__L LJ J__| J______F J_____F J____,__F|__|  \__/  

meta1v %s (commit: %s, built: %s)
`,
				buildVersion,
				buildCommit,
				buildDate,
			)
		},
	}
}
