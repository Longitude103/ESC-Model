# ESC Model (WWUMM)

[![CI](https://github.com/Longitude103/ESC-Model/actions/workflows/ci.yml/badge.svg)](https://github.com/Longitude103/ESC-Model/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Longitude103/ESC-Model.svg)](https://pkg.go.dev/github.com/Longitude103/ESC-Model)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Longitude103/ESC-Model)](https://github.com/Longitude103/ESC-Model/blob/master/go.mod)
[![Release](https://img.shields.io/github/v/release/Longitude103/ESC-Model?display_name=tag)](https://github.com/Longitude103/ESC-Model/releases)
[![GPLv3 license](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

The ESC Model is the command line tool for the Western Water Use Management Model (WWUMM). It reads parcel, well,
canal and CropSim data, runs the soil water balance for each parcel, and writes the recharge and pumping inputs that the
groundwater (MODFLOW 6) model uses to analyze the stream effects of pumping.

It is written in [Go](https://go.dev/). Pre-built binaries for Windows, macOS and Linux are attached to each
[release](https://github.com/Longitude103/ESC-Model/releases).

## Quick start

1. Download the binary for your system from the latest release.
2. Save your database login (once per computer, see [Database](#database)):

   ```
   EscModel init -u <username> -p <password> --endYear 2024
   ```

3. Run the model, pointing it at the folder of CropSim monthly output files:

   ```
   EscModel runModel --Desc "Test Run" --CSDir "<path>/CropSim/Output"
   ```

4. Turn the results into MODFLOW 6 files and answer the prompts:

   ```
   EscModel mfFiles
   ```

In these examples `EscModel` stands for the binary you downloaded, for example `wwumm-amd64-win_v1.6.15.exe` in
PowerShell or `./wwumm-arm64-darwin_v1.6.15` in a macOS Terminal.

## Database

The model reads its source data from the WWUMM PostgreSQL database, hosted on a
[Linode (Akamai) managed PostgreSQL](https://www.linode.com/products/postgresql/) server with the
[PostGIS](https://postgis.net/) extension for the spatial work. Connections require SSL.

Before running any other command, save your database login with `init`. Each office has its own login:

```
EscModel init -u <username> -p <password>
```

This writes a `.configjson` file in the folder you run the command from. Keep that file private; it holds your password.
The host, port, database name and SSL mode default to the WWUMM server. Override them with `--host`, `--port`,
`--dbname` and `--sslmode` if needed (for example `--sslmode disable` for a local development database).
If the database server moves, run `init` again with the new `--host`; you don't need to change or rebuild the code.

`--endYear` saves the last year with model data. `runModel` uses it as its default end year and `runSteadyState` as the
default end of its averaging period, so extending the model only needs `init --endYear <year>`, not a code change. A
config file without it falls back to 2020 and prints a reminder. External (outside-NRD) recharge and pumping use their
data through the last year in the `ext_recharge` and `ext_pumping` tables and repeat that year after it.

## Commands

Run the binary with no arguments, or with `--help`, to see the full list. Add `--help` after any command to see its flags.

| Command | What it does |
| --- | --- |
| `init` | Saves your database login to `.configjson`. |
| `runModel` | Runs the model and writes a results database (see flags below). |
| `runSteadyState` | Runs the steady state model, averaging CropSim results over a span of years. |
| `mfFiles` | Creates MODFLOW 6 RCH and WEL files from a results database. Asks which results to use, a suffix for the file names (for example `task3_npnrd`, giving `AggregateWEL_task3_npnrd.WEL6`) and the description for the file headers (pre-filled from the scenario's notes: a `project` scenario's description, otherwise the run's). Each WEL file's flow-reduction CSV is named after that file (`<file>_reduction.csv`). |
| `qcResults` | Summarizes a results database (annual recharge by type) and can write GeoJSON for GIS review. |
| `project` | Projects pumping and recharge forward for the Robust Review analysis from an existing results database. |
| `parcel-delete` | Deletes a parcel, and its junction table rows, from the land use data. |
| `parcel-copy` | Copies a parcel from one year to another. |
| `jct-delete` | Deletes a single parcel to well junction record. |

### `runModel` flags

| Flag | Meaning |
| --- | --- |
| `-d, --Desc` | **Required.** Description of the run. |
| `-c, --CSDir` | **Required.** Folder with the CropSim monthly output text files. |
| `-s, --StartYr` | First year to run (default 1997). |
| `-e, --EndYr` | Last year to run (default: `endYear` from the config). |
| `-b, --debug` | More log output and limited writes. |
| `-n, --noExcessFlow` | Leaves out excess flow recharge. |
| `-o, --oldGrid` | Uses the original 40 acre grid instead of the USG grid. |
| `-m, --mf6Grid40` | Uses the 40 acre grid with MODFLOW 6 node numbers. |
| `-p, --post97` | Post 1997 run: holds 1997 acres and crop types constant, no excess flows. |
| `-r, --noRetirements` | Puts retired parcels back using net pumping. |
| `-a, --AllocAnalysis` | Allocation analysis mode. |
| `-t, --TotalDep` | Total depletion mode (no GW-only parcels and no municipal or industrial pumping). |
| `-z, --P97Dep` | Post 1997 depletion analysis mode. |

### `runSteadyState` flags

`--Desc` and `--CSDir` are required, as above. `--StartYr` and `--EndYr` set the steady state period (defaults 1893 to
1952), and `--AvgStart` and `--AvgEnd` set the years used to average the CropSim results and canal diversions (defaults 1953
to the config `endYear`).
`--oldGrid` and `--mf6Grid40` work as in `runModel`.

### `project` flags

`--Years` (required) is a quoted, comma separated list of years to repeat (for example `"2018, 2020, 2018"`),
`--ProjectYears` is how many years to project forward (default 50), and `--FileNames` and `--Desc` name the output. The
projection starts the year after the results database ends.

## Output

Everything is written under an `OutputFiles` folder in the folder you run the binary from, with one time-stamped
sub-folder per run.

- `runModel` and `runSteadyState` write a SQLite results database (`.sqlite`) and a `.log` file with run details and
  any errors into an `OutputFiles/results<timestamp>` folder.
- `mfFiles`, `qcResults` and `project` read a results database and write their RCH, WEL, summary and GeoJSON files
  into their own time-stamped sub-folder.
- The GeoJSON from `qcResults` opens in [QGIS](https://qgis.org) and should also work in ESRI products.

## Building from source

You need Go 1.26 or newer and a C compiler, because the SQLite driver uses CGO.

```
git clone https://github.com/Longitude103/ESC-Model.git
cd ESC-Model
go build -o bin/esc-model .
```

On Windows, install a GCC such as [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) first; see
[windows_compile.md](windows_compile.md). The `Makefile` has `build` (macOS and Linux) and `build-win` targets for
local builds.

### Releases

Push a version tag (for example `git tag v1.7.0 && git push origin v1.7.0`). The Release workflow builds the Windows,
macOS (Intel and Apple silicon) and Linux binaries on GitHub and attaches them to a new release, so nobody has to build
on their own machine.

### Tests

```
go test ./...
```

Most tests are unit tests and run anywhere. Tests that read from the PostgreSQL database need a `.env` file at the
repository root with `user`, `password`, `host`, `port` and `dbname` entries (and optionally `sslmode`, which defaults
to `require`); without it they are skipped. Never commit
`.env` or `.configjson`: this repository is public.

GitHub Actions runs `gofmt`, `go vet`, the tests and a build on every pull request.

### Docker

The [Dockerfile](Dockerfile) builds a Linux image. See [development.md](development.md) for running it with your
CropSim and output folders mounted.

## Model notes

Notes on how specific processes are handled (fallow, pumping outside the NRDs, unit conversions to MODFLOW) are in
[notes.md](notes.md).

## License

GPLv3. See [LICENSE](LICENSE).
