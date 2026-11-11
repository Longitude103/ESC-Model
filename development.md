# GO Installation

To install a program, use the binary file.

Install the binary file to `/usr/bin` and it will work without change to path since
`/usr/bin` is in the path.

If you put the binary file in another location you can add that location to the
path by running `export PATH=$PATH:/path/to/your/install/directory` to add it.

You can see what is in the path by running `echo $PATH`

In "GO" you can get the binary by running `go build` and if you want a different name
you can use `go build -o <name>` for a different file name.

There are several potential paths that can be used, but this is good for Linux, might be others for Windows.

# Docker Implementation
Docker gives everyone the same Go toolchain and C compiler, so builds don't depend on what is installed on the developer's machine.
To build the image use: `docker build -t esc-model:<version> .` (the version tag can be left off for local builds).

The image does not contain a database login. Mount your `.configjson` (created by `init`) along with the CropSim and output folders:

```
docker run -it \
  -v $(pwd)/.configjson:/app/.configjson \
  -v $(pwd)/CropSimOutput:/app/CropSimOutput \
  -v $(pwd)/OutputFiles:/app/OutputFiles \
  esc-model runModel --Desc "Test Run" --CSDir /app/CropSimOutput
```

- `-it` runs the container interactively so the prompts from `mfFiles` and `qcResults` work
- the `CropSimOutput` bind mount gives the container the CropSim monthly output files
- the `OutputFiles` bind mount puts the results on the host for the next steps

If you are using podman on RHEL, Fedora, or another SELinux distro, add `:Z` to each `-v` mount (or use `--privileged`) so the container can read the host files.

# Local Compiling
If you wish to compile and run the program locally. You can do this through the terminal using a few of the following commands.

These commands look like:
- `go run main.go runModel --CSDir "<path>"`
    - "--CSDir" is the path to the monthly output CropSim .txt files, you must qualify it in "<path>" if the path contains spaces
        - My example is `--CSDir "<path>/WWUMM2020/CropSim/Run005_WWUM2020/Output"`
- `go run main.go runModel --CSDir "<path>" --debug`