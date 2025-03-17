package maven2

import (
	"bytes"
	"fmt"
	"io"
	"iscrie/core/importer"
	"iscrie/network"
	"iscrie/utils"
	"os"
	"path/filepath"
	"strings"
)

// Maven2Importer handles importing Maven2 files into Nexus.
type Maven2Importer struct {
	BaseURL      string
	Repository   string
	HTTPClient   *network.HTTPClientAdapter
	RootPath     string
	ForceReplace bool
}

// NewMaven2Importer creates a new Maven2Importer instance.
func NewMaven2Importer(baseURL, repository, rootPath string, httpClient *network.HTTPClient, forceReplace bool) *Maven2Importer {
	adapter := network.NewHTTPClientAdapter(httpClient, baseURL, repository, forceReplace)

	return &Maven2Importer{
		BaseURL:      baseURL,
		Repository:   repository,
		HTTPClient:   adapter,
		RootPath:     rootPath,
		ForceReplace: forceReplace,
	}
}

// BuildFullTargetURL constructs the full target URL for Maven2 files.
func (mi *Maven2Importer) BuildFullTargetURL(filePath string) (string, error) {
	// Check if the file has an extension that should be ignored
	if IsIgnoredExtension(filepath.Base(filePath)) {
		utils.LogDebug("Skipping file with ignored extension: %s", filePath)
		return "", fmt.Errorf("file '%s' has an extension that should be ignored", filePath)
	}

	// Log initial path
	utils.LogDebug("File Path: %s", filePath)

	// Calculate relative path from rootPath
	relativePath, err := filepath.Rel(mi.RootPath, filePath)
	if err != nil {
		utils.LogError("Failed to compute relative path: %v", err)
		return "", fmt.Errorf("failed to compute relative path: %w", err)
	}
	utils.LogDebug("Relative Path: %s", relativePath)

	// Normalize and extract path segments
	normalizedPath := filepath.ToSlash(relativePath)
	segments := strings.Split(normalizedPath, "/")

	if len(segments) < 3 {
		utils.LogError("Invalid Maven2 path: %s", normalizedPath)
		return "", fmt.Errorf("invalid Maven2 path: %s", normalizedPath)
	}

	// Extract Maven details
	groupID := strings.Join(segments[:len(segments)-3], ".")
	version := segments[len(segments)-2]
	fileName := segments[len(segments)-1]

	// Parse the filename
	artifactID, parsedVersion, classifier, extension, err := ParseMavenFileName(fileName, version)
	if err != nil {
		utils.LogError("Failed to parse Maven file name '%s': %v", fileName, err)
		return "", fmt.Errorf("failed to parse Maven file name '%s': %w", fileName, err)
	}

	utils.LogDebug("Parsed File - GroupID: %s, ArtifactID: %s, Version: %s, Classifier: %s, Extension: %s",
		groupID, artifactID, parsedVersion, classifier, extension)

	// Construct the Maven2 path
	mavenPath := GenerateMavenPath(groupID, artifactID, parsedVersion, classifier, extension)
	utils.LogDebug("Maven Path: %s", mavenPath)

	// Construct the full URL
	fullURL := fmt.Sprintf("%s/repository/%s/%s", mi.BaseURL, mi.Repository, mavenPath)
	utils.LogDebug("Full Target URL: %s", fullURL)

	return fullURL, nil
}

// UploadMaven2File uploads a Maven2 artifact to Nexus with detailed logging.
func (mi *Maven2Importer) UploadMaven2File(filePath string, retryAttempts int, debugLogger, errorLogger func(format string, args ...interface{})) error {
	// Default no-op loggers if nil
	if debugLogger == nil {
		debugLogger = func(format string, args ...interface{}) {}
	}
	if errorLogger == nil {
		errorLogger = func(format string, args ...interface{}) {}
	}

	// Check if we should skip this file based on its extension
	if IsIgnoredExtension(filepath.Base(filePath)) {
		debugLogger("Skipping file with ignored extension: %s", filePath)
		return nil // Return nil as this is not an error, just a file we're intentionally skipping
	}

	// Step 1: Build full URL
	fullURL, err := mi.BuildFullTargetURL(filePath)
	if err != nil {
		errorLogger("Failed to build full URL for file '%s': %v", filePath, err)
		return fmt.Errorf("failed to build full URL: %w", err)
	}

	// Step 2: Open the file
	file, err := os.Open(filePath)
	if err != nil {
		errorLogger("Failed to open file '%s': %v", filePath, err)
		return fmt.Errorf("failed to open file '%s': %w", filePath, err)
	}
	defer file.Close()

	// Step 3: Log file preview
	var preview bytes.Buffer
	_, _ = io.CopyN(&preview, file, 100)
	file.Seek(0, io.SeekStart) // Reset file cursor after preview

	debugLogger("Uploading file: %s", filePath)
	debugLogger("Target URL: %s", fullURL)
	debugLogger("File preview (first 100 bytes): %q", preview.String())

	// Step 4: Call `UploadFileWithRetry` with both loggers
	return importer.UploadFileWithRetry(mi.HTTPClient, fullURL, filePath, retryAttempts, debugLogger, errorLogger)
}
