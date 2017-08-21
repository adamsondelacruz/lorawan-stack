// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package main

import (
	"fmt"
	"os"

	"github.com/TheThingsNetwork/ttn/cmd/example/commands"
)

func main() {
	if err := commands.Root.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
