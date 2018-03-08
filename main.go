package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	version string = ""
)

func main() {

	_ = godotenv.Load()

	app := cli.NewApp()
	app.Name = "Zeppelin client"
	app.Usage = "Zeppelin client"
	app.Action = run
	app.Version = fmt.Sprintf("%s", version)
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{

		cli.StringFlag{
			Name:   "plugin.endpoint",
			Usage:  "Zeppelin Url",
			EnvVar: "PLUGIN_ZEPPELIN_ENDPOINT",
			Value:  "/zeppelin",
		},
		cli.StringFlag{
			Name:   "plugin.username",
			Usage:  "API Username",
			EnvVar: "PLUGIN_ZEPPELIN_USERNAME",
		},
		cli.StringFlag{
			Name:   "plugin.password",
			Usage:  "API Password",
			EnvVar: "PLUGIN_ZEPPELIN_PASSWORD",
		},
		cli.StringFlag{
			Name:   "plugin.notebook.name",
			Usage:  "Zeppelin notebook name",
			EnvVar: "PLUGIN_NOTEBOOK_NAME",
		},
		cli.StringFlag{
			Name:   "plugin.notebook.filePath",
			Usage:  "The path to the zepplein notebook",
			EnvVar: "PLUGIN_NOTEBOOK_FILE_PATH",
		},
		cli.StringFlag{
			Name:   "plugin.notebook.state",
			Usage:  "The state of the notebook",
			EnvVar: "PLUGIN_NOTEBOOK_STATE",
		},
		cli.StringFlag{
			Name:   "plugin.log.level",
			Usage:  "Specific log level (debug,info,warn)",
			EnvVar: "PLUGIN_LOG_LEVEL",
		},
		cli.StringFlag{
			Name:   "plugin.log.format",
			Usage:  "Specific log format (text, json) default is text",
			EnvVar: "PLUGIN_LOG_FORMAT",
			Value:  "text",
		},
	}
	app.Run(os.Args)
}

func run(c *cli.Context) error {

	plugin := Plugin{

		Config: Config{
			Endpoint: c.String("plugin.endpoint"),
			Username: c.String("plugin.username"),
			Password: c.String("plugin.password"),
			Notebook: Notebook{
				Name:     c.String("plugin.notebook.name"),
				FilePath: c.String("plugin.notebook.filePath"),
				State:    c.String("plugin.notebook.state"),
			},
		},
	}

	processLogLevel(c)

	err := plugin.Exec()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func processLogLevel(c *cli.Context) {
	switch strings.ToUpper(c.String("plugin.log.level")) {
	case "INFO":
		log.SetLevel(log.InfoLevel)
	case "DEBUG":
		log.SetLevel(log.DebugLevel)
	case "WARN":
		log.SetLevel(log.WarnLevel)
	case "ERROR":
		log.SetLevel(log.ErrorLevel)
	case "PANIC":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}
