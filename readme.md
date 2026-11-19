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

In these examples `EscModel` stands for the binary you downloaded, for example `wwumm-amd64-win_v1.8.0.exe` in
PowerShell or `./wwumm-arm64-darwin_v1.8.0` in a macOS Terminal.

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

Run the binary with no arguments, or with `--help`, to see the full list. Add `--help` after any command to see its
flags. Every command except `init`, `help` and `completion` reads the database login from `.configjson`, so run
`init` first.

| Command | What it does |
| --- | --- |
| [`init`](#init) | Saves your database login to `.configjson`. |
| [`runModel`](#runmodel) | Runs the model for a range of years and writes a results database. |
| [`runSteadyState`](#runsteadystate) | Runs the steady state model, averaging CropSim results over a span of years. |
| [`mfFiles`](#mffiles) | Creates MODFLOW 6 RCH and WEL files from a results database. |
| [`qcResults`](#qcresults) | Summarizes a results database and can write graphs and GeoJSON for GIS review. |
| [`project`](#project) | Projects pumping and recharge forward for the Robust Review analysis. |
| [`parcel-delete`](#parcel-delete) | Deletes a parcel, and its junction table rows, from the land use data. |
| [`parcel-copy`](#parcel-copy) | Copies a parcel, and its junction table rows, from one year to another. |
| [`jct-delete`](#jct-delete) | Deletes a single parcel to well junction record. |
| `completion` | Writes a shell autocompletion script (`bash`, `zsh`, `fish` or `powershell`). |
| `help` | Shows help for any command, for example `EscModel help runModel`. |

### `init`

Saves the database login and settings to `.configjson` in the current folder. Run it once per computer, and again
whenever any of these settings change.

```
EscModel init -u <username> -p <password> --endYear 2024
```

| Flag | Meaning |
| --- | --- |
| `-u, --username` | Database user name. |
| `-p, --password` | Database password. |
| `--endYear` | Last year with model data. The default end year for `runModel` and for `runSteadyState` averaging. |
| `--host` | Database host (default: the WWUMM server). |
| `--port` | Database port (default: the WWUMM server's port). |
| `--dbname` | Database name (default `wwumm`). |
| `--sslmode` | PostgreSQL SSL mode (default `require`). |

### `runModel`

The main model run. It reads the source data from the database, runs the soil water balance for each parcel and writes
a SQLite results database to `OutputFiles/results<timestamp>/`.

```
EscModel runModel --Desc "2024 update" --CSDir "<path>/CropSim/Output" --EndYr 2024
```

| Flag | Meaning |
| --- | --- |
| `-d, --Desc` | **Required.** Description of the run. It is stored in the results database and used for file headers. |
| `-c, --CSDir` | **Required.** Folder with the CropSim monthly output text files (`*_MON.TXT`). |
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

### `runSteadyState`

Runs the steady state model and writes a results database the same way as `runModel`. CropSim results and canal
diversions are averaged over the averaging period and applied to the steady state period.

```
EscModel runSteadyState --Desc "Steady state" --CSDir "<path>/CropSim/Output"
```

| Flag | Meaning |
| --- | --- |
| `-d, --Desc` | **Required.** Description of the run. |
| `-c, --CSDir` | **Required.** Folder with the CropSim monthly output text files. |
| `-s, --StartYr` | First year of the steady state period (default 1893). |
| `-e, --EndYr` | Last year of the steady state period (default 1952). |
| `-a, --AvgStart` | First year used for averaging (default 1953). |
| `-q, --AvgEnd` | Last year used for averaging (default: `endYear` from the config). |
| `-o, --oldGrid` | Uses the original 40 acre grid instead of the USG grid. |
| `-m, --mf6Grid40` | Uses the 40 acre grid with MODFLOW 6 node numbers. |

### `mfFiles`

Creates MODFLOW 6 RCH and WEL files from a results database. It has no flags; it asks these questions instead:

1. **Which results DB should be used?** The path to a `.sqlite` results file. Press Tab to pick from the files under
   `OutputFiles`.
2. **Choose WEL files to exclude** and **Choose which RCH files to exclude.** Each source you pick is written to its
   own file and left out of the aggregate file. A steady state database only makes the aggregate RCH file, so these
   are skipped.
3. **Output Row-Column?** Writes row and column instead of node numbers (node is the default).
4. **Suffix for the output file names.** For example `task3_npnrd` gives `AggregateWEL_task3_npnrd.WEL6`. Leave it
   blank for no suffix.
5. **Description for the file headers.** Pre-filled from the results database: a `project` scenario's description,
   otherwise the run's.

Files are written to `OutputFiles/mfFiles<timestamp>/`. Each WEL file's flow-reduction CSV is named after that file
(`<file>_reduction.csv`).

### `qcResults`

Summarizes a results database for review. It has no flags; it asks which results database to use, then whether to
write each of these:

- a recharge balance for one year,
- graphs of the results,
- a GeoJSON file of one year of recharge results,
- a GeoJSON file of one year of well pumping results.

For the GeoJSON files it asks whether to include monthly or only annual data, and for the year-based outputs it asks
which year. Files and a log are written to `OutputFiles/qc<timestamp>/`. The GeoJSON opens in
[QGIS](https://qgis.org) and should also work in ESRI products.

### `project`

Projects pumping and recharge forward from an existing results database for the Robust Review analysis. It asks which
results database to use. The projection starts the year after that database ends and repeats the listed years in
order until it has covered `--ProjectYears` years.

```
EscModel project --Years "2018, 2020, 2018" --ProjectYears 50
```

| Flag | Meaning |
| --- | --- |
| `--Years` | **Required.** Quoted, comma separated list of years to repeat, in order. |
| `--ProjectYears` | How many years to project forward (default 50). |
| `--FileNames` | Name for the projection output files (default `RR`). |
| `--Desc` | Description of the projection run (default `Robust Review Analysis`). |

### Land use editing commands

These change the land use tables in the PostgreSQL database directly (`<nrd>.t<year>_irr` and `<nrd>.t<year>_jct`),
so the change applies to every later model run. `--nrd` must be `np` or `sp`.

#### `parcel-delete`

Deletes a parcel and all of its junction table (parcel to well) rows for one year.

```
EscModel parcel-delete --nrd np --year 2020 --id 12345
```

| Flag | Meaning |
| --- | --- |
| `--nrd` | **Required.** `np` or `sp`. |
| `--id` | **Required.** Parcel ID to delete. |
| `--year` | Land use year (default 2020). |

#### `parcel-copy`

Copies a parcel and its junction table rows from one year to another. If the parcel ID already exists in the target
year, the copy gets the next free ID and the command prints it.

```
EscModel parcel-copy --nrd sp --parcel_id 12345 --year_from 2020 --year_to 2021
```

| Flag | Meaning |
| --- | --- |
| `--nrd` | **Required.** `np` or `sp`. |
| `--parcel_id` | **Required.** Parcel ID to copy. |
| `--year_from` | Year to copy from (default 2020). |
| `--year_to` | Year to copy to (default 2021). |

#### `jct-delete`

Deletes one parcel to well junction record, leaving the parcel and the well in place.

```
EscModel jct-delete --nrd np --year 2020 --parcel_id 12345 --well_id 678
```

| Flag | Meaning |
| --- | --- |
| `--nrd` | **Required.** `np` or `sp`. |
| `--parcel_id` | **Required.** Parcel ID of the record. |
| `--well_id` | **Required.** Well ID of the record. |
| `--year` | Land use year (default 2020). |

## Output

Everything is written under an `OutputFiles` folder in the folder you run the binary from, with one time-stamped
sub-folder per run.

- `runModel` and `runSteadyState` write a SQLite results database (`.sqlite`) and a `.log` file with run details and
  any errors into an `OutputFiles/results<timestamp>` folder.
- `mfFiles`, `qcResults` and `project` read a results database and write their RCH, WEL, summary and GeoJSON files
  into their own time-stamped sub-folder.

## Building from source

You need Go 1.27 or newer and a C compiler, because the SQLite driver uses CGO.

```
git clone https://github.com/Longitude103/ESC-Model.git
cd ESC-Model
go build -o bin/esc-model .
```

On Windows, install a GCC such as [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) first; see
[windows_compile.md](windows_compile.md). The `Makefile` has `build` (macOS and Linux) and `build-win` targets for
local builds.

### Packages

The model is built with Go 1.27 (toolchain `go1.27.4`). Its direct dependencies are:

| Package | Version | Used for |
| --- | --- | --- |
| [spf13/cobra](https://github.com/spf13/cobra) | v1.10.2 | The command line commands and flags. |
| [AlecAivazis/survey](https://github.com/AlecAivazis/survey) | v2.3.7 | The interactive questions in `mfFiles`, `qcResults` and `project`. The project is archived upstream. |
| [pterm/pterm](https://github.com/pterm/pterm) | v0.12.84 | Coloured terminal output and progress bars. |
| [lib/pq](https://github.com/lib/pq) | v1.12.3 | PostgreSQL driver for the WWUMM database. |
| [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3) | v1.14.53 | SQLite driver for the results databases (needs CGO). |
| [jmoiron/sqlx](https://github.com/jmoiron/sqlx) | v1.4.0 | Query helpers on top of `database/sql`. |
| [paulmach/orb](https://github.com/paulmach/orb) | v0.14.0 | Geometry and GeoJSON for the `qcResults` files. |
| [longitude-103/flogo](https://bitbucket.org/longitude-103/flogo) | v0.1.20 | Writes the MODFLOW 6 WEL and RCH files. |
| [longitude-103/gisUtils](https://bitbucket.org/longitude-103/gisutils) | v1.1.1 | Distance and inverse distance weighting for weather stations. |
| [DATA-DOG/go-sqlmock](https://github.com/DATA-DOG/go-sqlmock) | v1.5.2 | Mock database in tests. |
| [joho/godotenv](https://github.com/joho/godotenv) | v1.5.1 | Reads the `.env` file for database tests. |

`spf13/pflag` (pulled in by cobra) is held at v1.0.10, because v1.1.0 changes the help text. To update the packages,
run `go get -u ./...` and `go mod tidy`, then set pflag back with `go get github.com/spf13/pflag@v1.0.10`.

### Releases

1. Write the release notes for modelers in `release-notes/<tag>.md` (for example `release-notes/v1.8.0.md`) and merge
   them to master.
2. Push a version tag from master (for example `git tag v1.8.0 && git push origin v1.8.0`). Don't create the release by
   hand on GitHub first.

The Release workflow builds the Windows, macOS (Intel and Apple silicon) and Linux binaries on GitHub and creates the
release with them, using the notes file. Without a notes file GitHub writes notes from the merged PRs. If the release
already exists, the workflow only adds the binaries to it. To add binaries to an older release, run the Release
workflow by hand from the Actions tab and give it the tag.

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
