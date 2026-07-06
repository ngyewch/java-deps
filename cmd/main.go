package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/jwalton/go-supportscolor"
	"github.com/phsym/console-slog"
	"github.com/urfave/cli/v3"
)

var (
	version string

	hideUnchangedFlag = &cli.BoolFlag{
		Name:    "hide-unchanged",
		Usage:   "hide unchanged",
		Sources: cli.EnvVars("HIDE_UNCHANGED"),
	}

	inputFileArg = &cli.StringArg{
		Name:      "input-file",
		UsageText: "(input-file)",
	}
	inputFile2Arg = &cli.StringArg{
		Name:      "input-file2",
		UsageText: "(input-file2)",
	}
	outputFileArg = &cli.StringArg{
		Name:      "output-file",
		UsageText: "(output-file)",
	}
	pathArg = &cli.StringArg{
		Name:      "path",
		UsageText: "(path)",
	}

	app = &cli.Command{
		Name:    "java-deps",
		Usage:   "Java dependencies",
		Version: version,
		Commands: []*cli.Command{
			{
				Name:   "scan",
				Usage:  "scan",
				Action: doScan,
				Arguments: []cli.Argument{
					outputFileArg,
					pathArg,
				},
			},
			{
				Name:   "compare",
				Usage:  "compare",
				Action: doCompare,
				Flags: []cli.Flag{
					hideUnchangedFlag,
				},
				Arguments: []cli.Argument{
					inputFileArg,
					inputFile2Arg,
				},
			},
		},
	}
)

func main() {
	err := app.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	logLevel := slog.LevelInfo
	_ = logLevel.UnmarshalText([]byte(os.Getenv("LOG_LEVEL")))
	addSource := os.Getenv("LOG_ADD_SOURCE") == "true"
	noColor := os.Getenv("NO_COLOR") == "true"

	var logger *slog.Logger
	if supportscolor.Stderr().SupportsColor {
		logger = slog.New(
			console.NewHandler(os.Stderr, &console.HandlerOptions{
				Level:     logLevel,
				AddSource: addSource,
				NoColor:   noColor,
			}),
		)
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     logLevel,
			AddSource: addSource,
		}))
	}
	slog.SetDefault(logger)
}
