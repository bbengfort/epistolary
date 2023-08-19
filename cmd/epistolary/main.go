package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
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
	ulid "github.com/oklog/ulid/v2"
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
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "url",
				Aliases: []string{"u"},
				Usage:   "the endpoint for client requests",
				Value:   "https://api.epistolary.app",
				EnvVars: []string{"EPISTOLARY_ENDPOINT"},
			},
			&cli.StringFlag{
				Name:    "username",
				Aliases: []string{"U"},
				Usage:   "username for authenticating client requests",
				EnvVars: []string{"EPISTOLARY_USERNAME"},
			},
			&cli.StringFlag{
				Name:    "password",
				Aliases: []string{"P"},
				Usage:   "password for authenticating client requests",
				EnvVars: []string{"EPISTOLARY_PASSWORD"},
			},
		},
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
				Name:     "tokenkey",
				Usage:    "generate an RSA token key pair and ksuid for JWT token signing",
				Category: "admin",
				Action:   generateTokenKey,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "out",
						Aliases: []string{"o"},
						Usage:   "path to write keys out to (optional, will be saved as ksuid.pem by default)",
					},
					&cli.IntFlag{
						Name:    "size",
						Aliases: []string{"s"},
						Usage:   "number of bits for the generated keys",
						Value:   4096,
					},
				},
			},
			{
				Name:     "status",
				Usage:    "send a status request to the epistolary api",
				Category: "client",
				Action:   status,
				Flags:    []cli.Flag{},
			},
			{
				Name:      "fetch",
				Usage:     "fetch a webpage or icon to see how it is parsed",
				ArgsUsage: "url [url ...]",
				Category:  "debug",
				Action:    fetchURL,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "icon",
						Aliases: []string{"i"},
						Usage:   "check if an icon exists",
					},
				},
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
// Admin Actions
//===========================================================================

func generateTokenKey(c *cli.Context) (err error) {
	// Create ksuid and determine outpath
	keyid := ulid.Make()

	var out string
	if out = c.String("out"); out == "" {
		out = fmt.Sprintf("%s.pem", keyid)
	}

	// Generate RSA keys using crypto random
	var key *rsa.PrivateKey
	if key, err = rsa.GenerateKey(rand.Reader, c.Int("size")); err != nil {
		return cli.Exit(err, 1)
	}

	// Open file to PEM encode keys to
	var f *os.File
	if f, err = os.OpenFile(out, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600); err != nil {
		return cli.Exit(err, 1)
	}

	if err = pem.Encode(f, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}); err != nil {
		return cli.Exit(err, 1)
	}

	fmt.Printf("RSA key id: %s -- saved with PEM encoding to %s\n", keyid, out)
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

		if c.Bool("icon") {
			if doc.FaviconCheck, err = fetch.CheckIcon(ctx, doc.Favicon); err != nil {
				return cli.Exit(err, 1)
			}

			if !doc.FaviconCheck {
				doc.Favicon = ""
			}
		}

		docs = append(docs, doc)
	}

	if err = json.NewEncoder(os.Stdout).Encode(docs); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}
