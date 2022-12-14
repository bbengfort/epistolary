package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/bbengfort/epistolary/pkg"
	"github.com/bbengfort/epistolary/pkg/api/v1"
	"github.com/bbengfort/epistolary/pkg/server"
	"github.com/bbengfort/epistolary/pkg/server/config"
	"github.com/bbengfort/epistolary/pkg/server/db/schema"
	"github.com/bbengfort/epistolary/pkg/server/fetch"
	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func main() {
	// Load the dotenv file if it exists
	godotenv.Load()

	// Create the CLI application
	app := cli.App{
		Name:    "epistolary",
		Version: pkg.Version(),
		Usage:   "the epistolary api server and management tools",
		Flags:   []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:     "serve",
				Usage:    "serve the epistolary api",
				Category: "server",
				Action:   serve,
				Flags:    []cli.Flag{},
			},
			{
				Name:     "migrate",
				Usage:    "migrate the database to the latest schema version",
				Category: "database",
				Action:   migrate,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dsn",
						Aliases: []string{"d", "db"},
						Usage:   "database dsn to connect to the database on",
						EnvVars: []string{"DATABASE_URL", "EPISTOLARY_DATABASE_URL"},
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "force the latest schema version to be applied",
					},
					&cli.BoolFlag{
						Name:    "drop",
						Aliases: []string{"D"},
						Usage:   "drop the database schema before migrating (force must be true)",
					},
				},
			},
			{
				Name:     "schema",
				Usage:    "get the current version of the database schema",
				Category: "database",
				Action:   schemaVersion,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dsn",
						Aliases: []string{"d", "db"},
						Usage:   "database dsn to connect to the database on",
						EnvVars: []string{"DATABASE_URL", "EPISTOLARY_DATABASE_URL"},
					},
					&cli.BoolFlag{
						Name:    "verify",
						Aliases: []string{"v"},
						Usage:   "verify the schema has been correctly loaded",
					},
				},
			},
			{
				Name:     "status",
				Usage:    "send a status request to the epistolary api",
				Category: "client",
				Action:   status,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "url",
						Aliases: []string{"u"},
						Value:   "http://localhost:8000",
					},
				},
			},
			{
				Name:     "fetch",
				Usage:    "fetch a webpage to see how it is parsed",
				Category: "debug",
				Action:   fetchURL,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

//===========================================================================
// Server Actions
//===========================================================================

func serve(c *cli.Context) (err error) {
	var srv *server.Server
	if srv, err = server.New(config.Config{}); err != nil {
		return cli.Exit(err, 1)
	}

	if err = srv.Serve(); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

//===========================================================================
// Database Actions
//===========================================================================

func migrate(c *cli.Context) (err error) {
	if err = schema.Configure(c.String("dsn")); err != nil {
		return cli.Exit(err, 1)
	}
	defer schema.Close()

	if c.Bool("drop") {
		if !c.Bool("force") {
			return cli.Exit("cannot drop without forcing", 1)
		}
		if err = schema.Drop(); err != nil {
			return cli.Exit(err, 1)
		}
	}

	if c.Bool("force") {
		if err = schema.Force(); err != nil {
			return cli.Exit(err, 1)
		}
	} else {
		if err = schema.Migrate(); err != nil {
			return cli.Exit(err, 1)
		}
	}
	return nil
}

func schemaVersion(c *cli.Context) (err error) {
	defer schema.Close()
	if c.Bool("verify") {
		if err = schema.Verify(c.String("dsn")); err != nil {
			return cli.Exit(err, 1)
		}
	}

	var vers *schema.Version
	if vers, err = schema.CurrentVersion(c.String("dsn")); err != nil {
		return cli.Exit(err, 1)
	}

	if err = json.NewEncoder(os.Stdout).Encode(vers); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

//===========================================================================
// Client Actions
//===========================================================================

func status(c *cli.Context) (err error) {
	var client api.EpistolaryClient
	if client, err = api.New(c.String("url")); err != nil {
		return cli.Exit(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var rep *api.StatusReply
	if rep, err = client.Status(ctx); err != nil {
		return cli.Exit(err, 1)
	}

	if err = json.NewEncoder(os.Stdout).Encode(rep); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

//===========================================================================
// Debug Actions
//===========================================================================

func fetchURL(c *cli.Context) (err error) {
	if c.NArg() == 0 {
		return cli.Exit("specify at least one URL to fetch", 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	docs := make([]*fetch.Document, 0, c.NArg())
	for i := 0; i < c.NArg(); i++ {
		var doc *fetch.Document
		if doc, err = fetch.Fetch(ctx, c.Args().Get(i)); err != nil {
			return cli.Exit(err, 1)
		}

		docs = append(docs, doc)
	}

	if err = json.NewEncoder(os.Stdout).Encode(docs); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}
