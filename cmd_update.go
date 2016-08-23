package main

import (
	"os"
	"path"
	"strings"

	"github.com/Masterminds/vcs"
	"github.com/Sirupsen/logrus"
	"github.com/Unknwon/com"
	"github.com/spf13/cobra"
)

var sourcesFile string

func init() {
	RootCmd.AddCommand(updateCmd)

	updateCmd.Flags().StringVar(&sourcesFile, "sources", "sources.yaml", "File with base16 sources")
}

// buildCmd represents the build command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Pull in updates from the source repos",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Updating sources")
		dirs, ok := downloadSourceList(sourcesFile, sourcesDir)
		if !ok {
			log.Fatal("Failed to update sources")
		}

		for _, dir := range []string{"schemes", "templates"} {
			if !com.IsSliceContainsStr(dirs, dir) {
				log.Fatalf("%q location is missing from sources file", dir)
			}
		}

		log.Info("Updating schemes")
		_, ok = downloadSourceList(path.Join(sourcesDir, "schemes", "list.yaml"), schemesDir)
		if !ok {
			log.Fatal("Failed to update schemes")
		}

		log.Info("Updating templates")
		_, ok = downloadSourceList(path.Join(sourcesDir, "templates", "list.yaml"), templatesDir)
		if !ok {
			log.Fatal("Failed to update templates")
		}
	},
}

func downloadSourceList(sourceFile, targetDir string) ([]string, bool) {
	sources, err := readSourcesList(sourceFile)
	if err != nil {
		log.Error(err)
		return nil, false
	}

	err = os.MkdirAll(targetDir, 0777)
	if err != nil {
		log.Error(err)
		return nil, false
	}

	ok := true
	var ret []string
	for _, source := range sources {
		ret = append(ret, source.Key.(string))

		key := source.Key.(string)

		sourceDir := path.Join(targetDir, key)
		sourceLocation := source.Value.(string)

		repo, err := vcs.NewRepo(sourceLocation, sourceDir)
		if err != nil {
			log.Error(err)
			ok = false
			continue
		}

		if ok := repo.CheckLocal(); !ok {
			log.Debugf("Cloning %q to %q", sourceLocation, sourceDir)
			err = repo.Get()
			if err != nil {
				handleVcsError(log.WithField("source", key), err)
				ok = false
				continue
			}
		} else {
			log.Debugf("Updating %q", sourceDir)
			err = repo.Update()
			if err != nil {
				handleVcsError(log.WithField("source", key), err)
				ok = false
				continue
			}
		}
	}

	return ret, ok
}

func handleVcsError(logger *logrus.Entry, err error) {
	logger.Error(err)

	if lErr, ok := err.(*vcs.LocalError); ok {
		logger.Error(strings.TrimSpace(lErr.Out()))
		logger.Error(lErr.Original())
	}

	if rErr, ok := err.(*vcs.RemoteError); ok {
		logger.Error(strings.TrimSpace(rErr.Out()))
		logger.Error(rErr.Original())
	}
}