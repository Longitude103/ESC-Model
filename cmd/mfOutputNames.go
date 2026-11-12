package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

var nonFileChars = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

// cleanSuffix turns what the user typed into something safe for a file name: spaces and other characters become
// underscores, and leading or trailing underscores are dropped.
func cleanSuffix(s string) string {
	s = nonFileChars.ReplaceAllString(strings.TrimSpace(s), "_")
	return strings.Trim(s, "_")
}

// outputName builds an output file name (without extension) from a base name and an optional scenario suffix,
// e.g. outputName("AggregateWEL", "task3_npnrd") is "AggregateWEL_task3_npnrd".
func outputName(base, suffix string) string {
	base = nonFileChars.ReplaceAllString(base, "_")
	suffix = cleanSuffix(suffix)
	if suffix == "" {
		return base
	}
	return base + "_" + suffix
}

// outputQuestions asks for the scenario suffix added to every output file name, and for the description written
// in the file headers. The description defaults to the one stored in the results database.
func outputQuestions(defaultDesc string) (suffix, desc string, err error) {
	qs := []*survey.Question{
		{
			Name: "Suffix",
			Prompt: &survey.Input{
				Message: "Suffix for the output file names (scenario and NRD, e.g. npnrd or task3_npnrd; blank for none):",
				Help:    "Added to the end of every WEL and RCH file name so files from different scenarios don't overwrite each other.",
			},
		},
		{
			Name: "Desc",
			Prompt: &survey.Input{
				Message: "Description for the file headers:",
				Default: defaultDesc,
				Help:    "Written at the top of each WEL and RCH file. Defaults to the description stored in the results database.",
			},
		},
	}

	a := struct {
		Suffix string
		Desc   string
	}{}
	if err := survey.Ask(qs, &a); err != nil {
		return "", "", err
	}

	return cleanSuffix(a.Suffix), strings.TrimSpace(a.Desc), nil
}

// setReduceCSVName rewrites the AUTO_FLOW_REDUCE_CSV FILEOUT option in a WEL6 file so it points at csvName.
// The option is in the OPTIONS block at the top of the file; the rest of the file is copied unchanged.
func setReduceCSVName(welPath, csvName string) error {
	in, err := os.Open(welPath)
	if err != nil {
		return err
	}
	defer in.Close()

	tmpPath := welPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	if err := rewriteReduceCSV(in, out, csvName); err != nil {
		out.Close()
		os.Remove(tmpPath)
		return err
	}

	if err := out.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}
	in.Close()

	return os.Rename(tmpPath, welPath)
}

func rewriteReduceCSV(r io.Reader, w io.Writer, csvName string) error {
	reader := bufio.NewReader(r)
	writer := bufio.NewWriter(w)
	inOptions := true

	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			if inOptions {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(strings.ToUpper(trimmed), "AUTO_FLOW_REDUCE_CSV") {
					line = fmt.Sprintf("  AUTO_FLOW_REDUCE_CSV FILEOUT %s\n", csvName)
				}
				if strings.EqualFold(trimmed, "END OPTIONS") {
					inOptions = false
				}
			}
			if _, werr := writer.WriteString(line); werr != nil {
				return werr
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}
