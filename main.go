// +build linux

package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const dataDir = "/var/lib/vscode-updater"

const pkgext = ".pkg.tar.xz"
const dbext = ".db.tar.gz"

var repopath string
var reponame string

func init() {
	flag.StringVar(&repopath, "repopath", "/usr/local/repo/vscode", "path to the repository where packages will be stored")
	flag.StringVar(&reponame, "reponame", "vscode", "name of the repository")
}

type edition struct {
	Name        string
	FullName    string
	Description string
	WMClass     string
	Channel     string
}

var editions = []edition{
	{
		Name:        "code",
		FullName:    "Visual Studio Code",
		Description: "Code Editing. Redefined.",
		WMClass:     "code",
		Channel:     "stable",
	},
	{
		Name:        "code-insiders",
		FullName:    "Visual Studio Code - Insiders",
		Description: "Code Editing. Redefined.",
		WMClass:     "code - insiders",
		Channel:     "insider",
	},
}

type metaData struct {
	URL        string `json:"url"`
	Version    string `json:"version"`
	Sha256Hash string `json:"sha256hash"`
}

func fetchMetadata(channel string) (*metaData, error) {
	url := fmt.Sprintf("https://vscode-update.azurewebsites.net/api/update/linux-x64/%s/0", channel)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	meta := &metaData{}
	err = json.NewDecoder(res.Body).Decode(meta)
	if err != nil {
		return nil, err
	}

	meta.Version, err = parseVersion(meta.URL)
	if err != nil {
		return nil, err
	}

	return meta, nil
}

func shouldUpdate(meta *metaData, edition *edition) bool {
	f, err := os.Open(path.Join(dataDir, edition.Name))
	if os.IsNotExist(err) {
		return true
	} else if err != nil {
		log.Errorf("could not read last saved metadata: %v", err)
		return false
	}
	defer f.Close()

	var lastMeta metaData
	err = json.NewDecoder(f).Decode(&lastMeta)
	if err != nil {
		log.Errorf("could not decode last saved metadata: %v", err)
		return false
	}

	order, err := compareVersions(meta.Version, lastMeta.Version)
	if err != nil {
		log.Errorf("could not compare versions: %v", err)
		return false
	}

	return order == 1
}

func copyFile(dst, src string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()

	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}

	return io.Copy(df, sf)
}

func buildPackage(meta *metaData, edition *edition) error {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("could not create source directory: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	// Render desktop entry file.
	f, err := os.Create(filepath.Join(tmpdir, edition.Name+".desktop"))
	if err != nil {
		return err
	}
	h := sha256.New()
	w := io.MultiWriter(f, h)
	err = desktopTmpl.Execute(w,
		&desktopData{
			Name:        edition.Name,
			FullName:    edition.FullName,
			Description: edition.Description,
			WMClass:     edition.WMClass,
		})
	f.Close()
	if err != nil {
		return err
	}

	// Render PKGBUID.
	f, err = os.Create(filepath.Join(tmpdir, "PKGBUILD"))
	if err != nil {
		return fmt.Errorf("could not create PKGBUILD: %v", err)
	}
	err = pkgbuildTmpl.Execute(f,
		&pkgbuildData{
			Name:        edition.Name,
			Description: edition.Description,
			Version:     meta.Version,
			URL:         meta.URL,
			ArchiveHash: meta.Sha256Hash,
			DesktopHash: fmt.Sprintf("%x", h.Sum(nil)),
		})
	f.Close()
	if err != nil {
		return fmt.Errorf("could not render PKGBUILD: %v", err)
	}

	// Start package building.
	cmd := exec.Command("makepkg", "--clean")
	cmd.Dir = tmpdir
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("makepkg failed: %v", err)
	}

	files, err := ioutil.ReadDir(tmpdir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), pkgext) {
			continue
		}

		// Copy package to the repo directory.
		src := filepath.Join(tmpdir, file.Name())
		dst := filepath.Join(repopath, file.Name())
		_, err := copyFile(dst, src)
		if err != nil {
			return fmt.Errorf("could not copy package to the repo dir: %v", err)
		}

		// Update package database.
		db := filepath.Join(repopath, reponame+dbext)
		cmd := exec.Command("repo-add", db, dst)
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("repo-add failed: %v", err)
		}
	}

	return nil
}

func update(edition *edition) {
	log.Infof("checking for %s updates", edition.Name)

	meta, err := fetchMetadata(edition.Channel)
	if err != nil {
		log.Errorf("could not fetch metadata: %v", err)
		return
	}

	if !shouldUpdate(meta, edition) {
		return
	}

	log.Infof("there is a new update for %s (%s)", edition.Name, meta.Version)

	err = buildPackage(meta, edition)
	if err != nil {
		log.Errorf("could not build package: %v", err)
		return
	}

	f, err := os.Create(path.Join(dataDir, edition.Name))
	if err != nil {
		log.Errorf("could not save state: %v", err)
		return
	}
	defer f.Close()
	json.NewEncoder(f).Encode(meta)

	log.Infof("%v was successfully built", edition.Name)
}

func updateAll() {
	for _, e := range editions {
		update(&e)
	}
}

func main() {
	period := flag.Int("period", 1, "check update period (in hours)")
	flag.Parse()

	quit := make(chan bool)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)

		<-sig
		log.Warning("shutting down...")
		quit <- true
	}()

	ticker := time.NewTicker(time.Duration(*period) * time.Hour)
	defer ticker.Stop()

	for {
		updateAll()
		select {
		case <-ticker.C:
			continue
		case <-quit:
			return
		}
	}
}
