package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	blockservice "github.com/ipfs/go-blockservice"
	filestore "github.com/ipfs/go-filestore"
	blockstore "github.com/ipfs/go-ipfs-blockstore"
	config "github.com/ipfs/go-ipfs-config"
	offline "github.com/ipfs/go-ipfs-exchange-offline"
	files "github.com/ipfs/go-ipfs-files"
	pin "github.com/ipfs/go-ipfs-pinner"
	cu "github.com/ipfs/go-ipfs/core/coreunix"
	"github.com/ipfs/go-ipfs/plugin/loader"
	repo "github.com/ipfs/go-ipfs/repo"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	ipld "github.com/ipfs/go-ipld-format"
	dag "github.com/ipfs/go-merkledag"

	ds "github.com/ipfs/go-datastore"
)

func openManifestFile(workdir string) (*os.File, error) {
	fi, err := os.Open(filepath.Join(workdir, ManifestFilename))
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return nil, fmt.Errorf("error: no %s found in %s", ManifestFilename, workdir)
		default:
			return nil, fmt.Errorf("error opening %s: %s", ManifestFilename, err)
		}
	}
	return fi, nil
}

func setupPlugins(externalPluginsPath string) error {
	// Load any external plugins if available on externalPluginsPath
	plugins, err := loader.NewPluginLoader(filepath.Join(externalPluginsPath, "plugins"))
	if err != nil {
		return fmt.Errorf("error loading plugins: %s", err)
	}

	// Load preloaded and external plugins
	if err := plugins.Initialize(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	if err := plugins.Inject(); err != nil {
		return fmt.Errorf("error initializing plugins: %s", err)
	}

	return nil
}

func getRepo(workdir string) (repo.Repo, error) {
	packpath := filepath.Join(workdir, ".ipfs-pack")

	if err := setupPlugins(packpath); err != nil {
		return nil, err

	}

	if !fsrepo.IsInitialized(packpath) {
		cfg, err := config.Init(ioutil.Discard, 2048)
		if err != nil {
			return nil, err
		}

		cfg.Addresses.API = config.Strings{""}
		cfg.Addresses.Gateway = config.Strings{"/ip4/127.0.0.1/tcp/0"}
		cfg.Addresses.Swarm = []string{"/ip4/0.0.0.0/tcp/0"}
		cfg.Datastore.NoSync = true
		cfg.Experimental.FilestoreEnabled = true
		cfg.Reprovider.Interval = "0"

		err = fsrepo.Init(packpath, cfg)
		if err != nil {
			return nil, err
		}
	}
	return fsrepo.Open(packpath)
}

func buildDagserv(dstore ds.Batching, fm *filestore.FileManager) (blockstore.Blockstore, ipld.DAGService) {
	var bs blockstore.Blockstore = blockstore.NewBlockstore(dstore)
	if fm != nil {
		bs = filestore.NewFilestore(bs, fm)
	}
	bserv := blockservice.New(bs, offline.Exchange(bs))
	return bs, dag.NewDAGService(bserv)
}

func getAdder(dstore ds.Batching, pinner pin.Pinner, fm *filestore.FileManager) (*cu.Adder, error) {
	bstore, dserv := buildDagserv(dstore, fm)

	gcbs := blockstore.NewGCBlockstore(bstore, blockstore.NewGCLocker())
	adder, err := cu.NewAdder(context.Background(), pinner, gcbs, dserv)
	if err != nil {
		return nil, err
	}
	adder.NoCopy = true
	adder.RawLeaves = true
	return adder, nil
}

func getFilteredDirectory(workdir string) (files.Directory, error) {
	contents, err := ioutil.ReadDir(workdir)
	if err != nil {
		return nil, err
	}

	var farr []files.DirEntry
	for _, ent := range contents {
		if ent.Name() == ManifestFilename || ent.Name() == ManifestFilename+".tmp" {
			continue
		}
		if strings.HasPrefix(ent.Name(), ".") {
			continue
		}
		f, err := files.NewSerialFile(filepath.Join(workdir, ent.Name()), false, ent)
		if err != nil {
			return nil, err
		}
		farr = append(farr, files.FileEntry(ent.Name(), f))
	}

	return files.NewSliceDirectory(farr), nil
}
