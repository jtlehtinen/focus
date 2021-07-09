package focus

import (
	"fmt"
	"net/http"
	"time"

	"github.com/urfave/cli/v2"
)

func init() {
	// Override the default help template
	cli.AppHelpTemplate = `DESCRIPTION:
	{{.Usage}}

USAGE:
   {{.HelpName}} {{if .UsageText}}{{ .UsageText }}{{end}}
{{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}{{end}}
{{if .Version}}
VERSION:
	 {{.Version}}{{end}}
{{if .VisibleFlags}}
FLAGS:{{range .VisibleFlags}}{{ if (eq .Name "find" "undo" "replace") }}
		 {{if .Aliases}}-{{range $element := .Aliases}}{{$element}},{{end}}{{end}} --{{.Name}} {{.DefaultText}}
				 {{.Usage}}
		 {{end}}{{end}}
OPTIONS:{{range .VisibleFlags}}{{ if not (eq .Name "find" "replace" "undo") }}
		 {{if .Aliases}}-{{range $element := .Aliases}}{{$element}},{{end}}{{end}} --{{.Name}} {{ .DefaultText }}
				 {{.Usage}}
		 {{end}}{{end}}{{end}}
DOCUMENTATION:
	https://github.com/ayoisaiah/f2/wiki

WEBSITE:
	https://github.com/ayoisaiah/f2
`

	// Override the default version printer
	oldVersionPrinter := cli.VersionPrinter
	cli.VersionPrinter = func(c *cli.Context) {
		oldVersionPrinter(c)
		checkForUpdates(GetApp())
	}
}

func checkForUpdates(app *cli.App) {
	fmt.Println("Checking for updates...")

	c := http.Client{Timeout: 20 * time.Second}
	resp, err := c.Get("https://github.com/ayoisaiah/f2/releases/latest")
	if err != nil {
		fmt.Println("HTTP Error: Failed to check for update")
		return
	}

	defer resp.Body.Close()

	var version string
	_, err = fmt.Sscanf(
		resp.Request.URL.String(),
		"https://github.com/ayoisaiah/f2/releases/tag/%s",
		&version,
	)
	if err != nil {
		fmt.Println("Failed to get latest version")
		return
	}

	if version == app.Version {
		fmt.Printf(
			"Congratulations, you are using the latest version of %s\n",
			app.Name,
		)
	} else {
		fmt.Printf("%s: %s at %s\n", printColor("green", "Update available"), version, resp.Request.URL.String())
	}
}

func newConfig() error {
	c := &Config{}

	return c.new()
}

// GetApp retrieves the focus app instance.
func GetApp() *cli.App {
	return &cli.App{
		Name: "Focus",
		Authors: []*cli.Author{
			{
				Name:  "Ayooluwa Isaiah",
				Email: "ayo@freshman.tech",
			},
		},
		Usage:                "Focus is a cross-platform pomodoro app for the command line",
		UsageText:            "FLAGS [OPTIONS] [PATHS...]",
		Version:              "v0.1.0",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:  "config",
				Usage: "Change the configuration",
				Action: func(c *cli.Context) error {
					return newConfig()
				},
			},
		},
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:    "long",
				Usage:   "Long break duration in minutes (default: 15)",
				Aliases: []string{"l"},
			},
			&cli.UintFlag{
				Name:    "short",
				Usage:   "Short break duration in minutes (default: 5)",
				Aliases: []string{"s"},
			},
			&cli.UintFlag{
				Name:    "pomodoro",
				Usage:   "Pomodoro interval duration in minutes (default: 25)",
				Aliases: []string{"p"},
			},
			&cli.UintFlag{
				Name:  "long-break-interval",
				Usage: "Set the number of pomodoro sessions before a long break (default: 4)",
			},
		},
		Action: func(c *cli.Context) error {
			config := &Config{}
			err := config.get()
			if err != nil {
				return newConfig()
			}

			t := newTimer(c, config)
			t.start(pomodoro)

			return nil
		},
	}
}
