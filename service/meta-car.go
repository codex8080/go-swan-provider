package service

import (
	"fmt"
	carv2 "github.com/ipld/go-car/v2"
	"github.com/urfave/cli/v2"
	"os"
)

func MetaCar() {
	app := &cli.App{
		Name:  "meta-car",
		Usage: "Utility for working with car files",
		Commands: []*cli.Command{
			{
				Name:   "root",
				Usage:  "Get the root CID of a car",
				Action: MetaCarRoot,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "Specify source car file",
					},
				},
			},
			{
				Name:   "list",
				Usage:  "List the CIDs in a car",
				Action: MetaCarList,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "verbose",
						Aliases: []string{"v"},
						Usage:   "Include verbose information about contained blocks",
					},
					&cli.BoolFlag{
						Name:  "unixfs",
						Usage: "List unixfs filesystem from the root of the car",
					},
					&cli.BoolFlag{
						Name:  "links",
						Usage: "List links from the root of the car",
					},
				},
			},
			{
				Name:  "build",
				Usage: "Generate CAR file",
				Flags: []cli.Flag{
					&cli.Uint64Flag{
						Name:  "slice-size",
						Value: 17179869184, // 16G
						Usage: "specify chunk piece size",
					},
					&cli.UintFlag{
						Name:  "parallel",
						Value: 2,
						Usage: "specify how many number of goroutines runs when generate file node",
					},
					&cli.StringFlag{
						Name:  "graph-name",
						Value: "meta",
						Usage: "specify graph name",
					},
					&cli.StringFlag{
						Name:     "car-dir",
						Required: true,
						Usage:    "specify output CAR directory",
					},
					&cli.StringFlag{
						Name:     "uuid",
						Required: true,
						Usage:    "Add uuid to filename suffix",
					},
					&cli.StringFlag{
						Name:  "parent-path",
						Value: "",
						Usage: "specify graph parent path",
					},
					&cli.BoolFlag{
						Name:  "save-manifest",
						Value: true,
						Usage: "create a mainfest.csv in car-dir to save mapping of data-cids and slice names",
					},
				},
				Action: MetaCarBuild,
			},
			{
				Name:  "restore",
				Usage: "Restore files from CAR files",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "car-path",
						Required: true,
						Usage:    "specify source car path, directory or file",
					},
					&cli.StringFlag{
						Name:     "output-dir",
						Required: true,
						Usage:    "specify output directory",
					},
					&cli.IntFlag{
						Name:  "parallel",
						Value: 2,
						Usage: "specify how many number of goroutines runs when generate file node",
					},
				},
				Action: MetaCarRestore,
			},
		},
	}

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

func MetaCarList(c *cli.Context) error {
	return nil
}

func MetaCarRoot(c *cli.Context) (err error) {
	inStream, err := os.Open(c.String("file"))
	if err != nil {
		return err
	}
	rd, err := carv2.NewBlockReader(inStream)
	if err != nil {
		return err
	}
	for _, r := range rd.Roots {
		fmt.Printf("root CID: %s\n", r.String())
	}

	return nil
}

func MetaCarBuild(c *cli.Context) error {
	return nil
}

func MetaCarRestore(c *cli.Context) error {
	return nil
}
