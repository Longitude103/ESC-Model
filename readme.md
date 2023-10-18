# WWUMM 2020 CLI
[![Go Reference](https://pkg.go.dev/badge/github.com/Longitude103/wwum2020.svg)](https://pkg.go.dev/github.com/Longitude103/wwum2020)
[![GPLv3 license](https://img.shields.io/badge/License-GPLv3-blue.svg)](http://perso.crans.org/besson/LICENSE.html)
[![Release](https://img.shields.io/github/v/release/Longitude103/wwum2020?display_name=tag)](https://bitbucket.org/longitude-103/wwumm-esc/releases)

This CLI is the major starting point for the 2020 model operations. It's written in [![Go](https://img.shields.io/badge/--00ADD8?style=plastic&logo=go&logoColor=ffffff)](https://golang.org/)
and the binaries are compiled with [![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/Longitude103/wwum2020)](https://github.com/Longitude103/wwum2020) and are available in the latest release.

## Database
The CLI uses the WWUM postgres database hosted on [Amazon Web Services RDS](https://aws.amazon.com/rds/?nc2=h_ql_prod_db_rds) using
a Postgresql database. The postgresql database uses the [PostGIS](https://postgis.net/) extension to calculate the spatial
components of the data.

## CLI functions
There are several CLI functions called with the CLI and this is a list of them. You can 
also get a list by running the program without anything. The "flag" package is used for the CLI which enables the use
of flags in the CLI. This is a core go package.

- runModel => is the operation to start the main model run function, and it includes the following flags:
  - Required Flags: 
    - --Desc: A description of the model being run
    - --CSDir: The directory path to the CropSim Results text files
  - Optional Flags
    - --debug: run in debug mode with more log output and limited write operations
    - --StartYr: start year of the distribution, defaults to: 1997
    - --EndYr: end year of the distribution, defaults to: 2020
    - --NoExcessFlow: Removes the excess flow recharge from the model
    - --oldGrid: runs the model with the original 40-acre grid pattern from previous models
    - --mf6Grid40: runs the model with the 40 acre grids using MODFLOW 6 node numbers
    - --post97: sets the model to post 1997 mode where it holds 1997 acres and crop types constant, no excess flows


- mfFiles => is the CLI function to use the results from the `runModel` function and results database to create
  new MODFLOW files for MODFLOW 6 in the form of RCH6 and WEL6 files. It creates an "OutputFiles" directory in the
  same location as the binary for the result files.


- qcResults => is a CLI function to retrieve information from the results' database for information about the run. The initial
  function asks questions about which results database to analyze and then produces an Annual Recharge Summary by recharge
  type. In addition, it can also create a GeoJSON file for further analysis in the "OutputFiles" directory
### Common CLI Examples
- Production runModel:
  - Windows -> In Powershell: `wwum2020-amd64-windows.exe runModel --Desc "Test Run" --CSDir "<path>\WWUMM2020\CropSim\Run005_WWUM2020\Output"`
  - MacOS (Intel Mac)-> In a Terminal: `./wwum2020-amd64-darwin runModel --Desc "Test Run" --CSDir "<path>/WWUMM2020/CropSim/Run005_WWUM2020/Output"`
  - MacOS (Apple Silicon)-> In a Terminal: `./wwum2020-arm64-darwin runModel --Desc "Test Run" --CSDir "<path>/WWUMM2020/CropSim/Run005_WWUM2020/Output"`
  - Linux -> In a Terminal: `./wwum2020-amd64-linux runModel --Desc "Test Run" --CSDir "<path>/WWUMM2020/CropSim/Run005_WWUM2020/Output"`
   
- Production mfFiles
  - Windows -> In Powershell: `wwum2020-amd64-windows.exe mfFiles` -> Follow Prompts
  - MacOS (Intel Mac) -> In a Terminal: `./wwum2020-amd64-darwin mfFiles` -> Follow Prompts
  - MacOS (Apple Silicon)-> In a Terminal: `./wwum2020-arm64-darwin mfFiles` -> Follow Prompts
  - Linux -> In a Terminal: `./wwum2020-amd64-linux mfFIles` -> Follow Prompts


#### Development runs:
These runs have been replaced locally by running in ![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=plastic&logo=docker&logoColor=white). The Docker notes are found in the "development.md" file in this source 
directory and give guidance on compiling and running locally.

### Output
To store the large volume of output from `runModel`, the CLI uses a SQLite3 file as the storage container and is named
`results<timestamp>.sqlite` that will be in the same dir as the executable. This gives the ability to still be SQL enabled, but also allow compressing and store the data in Binary Format.
The text file output from the `runModel` is `results<timestamp>.log` file. This file includes run details and 
any additional information about errors that might occur while operating. The `mfFiles` function creates two text files that 
are the rch and wel files. The wel6 and rch6 files are stored
in an `OutputFiles` directory created in the same location as the executable file and then in sub-folders that correspond with
the results database and are named the same.

### ModFlow Output Files
The second command that is used by the app is the `mfFiles` command that enables you to write out ModFlow 6 files needed for the 
further model runs.

- Create the MODFLOW 6 files by using `./wwum2020-amd64-windows.exe mfFiles` (Windows Example)

### QC Results
This function is intended to analyze the results database and produce some totaling and output files that can make it easier
to spot problems and issues with the runs. This creates a recharge table section that outputs the selected annual amount of 
recharge by type. In addition, there is an optional recharge output that will allow you to create a GeoJSON file that can be used
by that can be added to GIS desktop software or online mapping software for spatial analysis of the results by model grid. 
We've tested this on [QGIS,](https://qgis.org) and it should also work on ESRI products.

- Create the QC Results by using `./wwum2020-amd64-windows.exe qcResults` (Windows Example)
    