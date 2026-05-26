package services

import (
	"Backend/internal/database/app"
	"log"
	"time"
)

type VersionUpdater struct {
	VersionService *VersionService
}

func NewVersionUpdater(versionService *VersionService) *VersionUpdater {
	return &VersionUpdater{VersionService: versionService}
}

func (v *VersionUpdater) runOnce() {
	log.Println("VersionUpdater: Version update loop started")
	latestVersion, err := v.VersionService.FetchVersion()
	if err != nil {
		log.Println("Error fetching latest version:", err)
		return
	}

	isLatest, err := app.CheckVersion(latestVersion.TagName)
	if err != nil {
		log.Println("Error checking version:", err)
		return
	}

	if !isLatest {
		err := app.UpdateLatestVersion(latestVersion.TagName)
		if err != nil {
			log.Println("Error updating version:", err)
			return
		}

		err = app.UpdateChangelog(latestVersion.TagName, latestVersion.Body)
		if err != nil {
			log.Println("Error updating changelog:", err)
			return
		}
	}

	log.Println("VersionUpdater: Version update loop completed")
}

func (v *VersionUpdater) Run() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run immediately on startup so version is populated right away,
	// without waiting for the first 1-hour tick.
	v.runOnce()

	for {
		select {
		case <-ticker.C:
			v.runOnce()
		}
	}
}
