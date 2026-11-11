# Windows Compile Notes

Windows takes a couple of special items to compile so the exe can execute. Using the sqlite3 lib requires that a GCC compiler is present so that CGO can run which is a compiler
for C which SQLite is written in and is compiled with GO so that everything is in one binary.

## Setup

We used the MinGW program for the GCC compiler. This was downloaded and installed from [TDM-gcc](https://jmeubank.github.io/tdm-gcc/) so that we could run the gcc compiler with GO.

## Steps to compile

- Open the PowerShell App from the start menu and navigate to the dir with the source code
- To navigate to Z dir type: `Z:\` in prompt
- Navigate to the directory with source files
- Run `go build -o bin/wwumm-amd64-win.exe` in the prompt
