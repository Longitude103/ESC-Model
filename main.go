// Copyright 2022 Longitude103. All rights reserved.
// Use of this source code is governed by GPLv3
// license that can be found in the LICENSE file.

// Package wwum2020 implements the Western Water Use Management model, 2020 update.
//
// The model calculates the crop water use and subsequent
// soil water balance that is needed to operate the MODFLOW
// modeling used in administration decisions by the NRDs of
// Nebraska. The model is a CLI based model that uses two basic
// commands to operate it. The first is runModel that
// tells the model to operate and calculate all the required
// information and saves it in an output database locally.
// Once the output database and log files are created, the
// other command is run mfFiles that produces the MODFLOW
// WEL and RCH files needed to be imported into that model
// for operations. There are a series of questions that
// are asked as to the model options for the output files
// that allow some flexibility for the creation of the
// files.
package main

import "bitbucket.org/longitude-103/wwumm-esc/cmd"

// main function is the entry for the application that sets up the CLI and sets the flags needed for the application. This
// function also has an error checking to deal with flags not set correctly.
func main() {
	cmd.Execute()
}
