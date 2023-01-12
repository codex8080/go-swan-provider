package service

import (
	"fmt"
	meta_car "github.com/FogMeta/meta-lib/module/ipfs"
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
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "Specify source car file",
					},
				},
			},
			{
				Name:  "build",
				Usage: "Generate CAR file",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "dir",
						Value: true,
						Usage: "",
					},
					&cli.Uint64Flag{
						Name:  "slice-size",
						Value: 17179869184, // 16G
						Usage: "specify chunk piece size",
					},
					&cli.StringFlag{
						Name:     "output-dir",
						Required: true,
						Usage:    "specify output CAR directory",
					},
				},
				Action: MetaCarBuildFromDir,
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
	carFile := c.String("file")

	info, err := meta_car.ListCarFile(carFile)
	if err != nil {
		return err
	}

	fmt.Println("List CAR :", carFile)
	for index, val := range info {
		fmt.Println(index, val)
	}

	return nil
}

func MetaCarRoot(c *cli.Context) error {
	carFile := c.String("file")

	root, err := meta_car.GetCarRoot(carFile)
	if err != nil {
		return err
	}

	fmt.Println("CAR :", carFile)
	fmt.Println("CID :", root)
	return nil
}

func MetaCarBuildFromDir(c *cli.Context) error {
	outputDir := c.String("output-dir")
	sliceSize := c.Uint64("slice-size")
	srcDir := c.Args().First()

	carFileName, err := meta_car.GenerateCarFromDir(outputDir, srcDir, int64(sliceSize))
	if err != nil {
		return err
	}

	fmt.Println("Build CAR :", carFileName)
	return nil
}

func MetaCarRestore(c *cli.Context) error {
	return nil
}