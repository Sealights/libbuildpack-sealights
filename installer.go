package sealights

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
)

type Installer struct {
	Log                *libbuildpack.Logger
	Options            *SealightsOptions
	MaxDownloadRetries int
}

func NewInstaller(log *libbuildpack.Logger, options *SealightsOptions) *Installer {
	return &Installer{Log: log, Options: options, MaxDownloadRetries: 3}
}

func (inst *Installer) InstallAgent() {
	inst.downloadPackage()
}

func (inst *Installer) downloadPackage() {
	url := inst.getDownloadUrl()

	inst.Log.Info("Sealights. Download package started. From '%s'", url)

	tempAgentFile := filepath.Join(os.TempDir(), "sealights-agent.tar.gz")
	err := downloadFileWithRetry(url, tempAgentFile, inst.MaxDownloadRetries)
	if err != nil {
		inst.Log.Error("Sealights. Failed to download package.")
	}

	inst.Log.Info("Sealights. Download finished.")
}

func (inst *Installer) extractPackage() {
	libbuildpack.GetBuildpackDir()
}

func (inst *Installer) getDownloadUrl() string {
	if inst.Options.CustomAgentUrl != "" {
		return inst.Options.CustomAgentUrl
	}

	labId := "agents"
	if inst.Options.LabId != "" {
		labId = inst.Options.LabId
	}

	version := "latest"
	if inst.Options.Version != "" {
		version = inst.Options.Version
	}

	url := fmt.Sprintf("https://%s.sealights.co/dotnetcore/sealights-dotnet-agent-%s.tar.gz", labId, version)

	return url
}

func downloadFileWithRetry(url string, filePath string, MaxDownloadRetries int) error {
	const baseWaitTime = 3 * time.Second

	var err error
	for i := 0; i < MaxDownloadRetries; i++ {
		err = downloadFile(url, filePath)
		if err == nil {
			return nil
		}

		waitTime := baseWaitTime + time.Duration(math.Pow(2, float64(i)))*time.Second
		time.Sleep(waitTime)
	}

	return err
}

func downloadFile(url, destFile string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("could not download: %d", resp.StatusCode)
	}

	return writeToFile(resp.Body, destFile, 0666)
}

func writeToFile(source io.Reader, destFile string, mode os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(destFile), 0755)
	if err != nil {
		return err
	}

	fh, err := os.OpenFile(destFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer fh.Close()

	_, err = io.Copy(fh, source)
	if err != nil {
		return err
	}

	return nil
}