package maven2

import (
	"fmt"
	"iscrie/utils"
	"path/filepath"
	"regexp"
	"strings"
)

// List of file extensions to ignore when processing Maven files
var IgnoredExtensions = []string{
	".md5",
	".sha1",
	".asc",  // PGP signature
	".sha256",
	".sha512",
}

// IsIgnoredExtension checks if a file has an extension that should be ignored
func IsIgnoredExtension(fileName string) bool {
	for _, ext := range IgnoredExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}
	return false
}

// ParseMavenFileName parses the Maven file name and extracts artifact details.
func ParseMavenFileName(fileName string, versionFromPath string) (artifactID, version, classifier, extension string, err error) {
	// Check if the file has an ignored extension
	if IsIgnoredExtension(fileName) {
		utils.LogDebug("Ignoring file with extension to be skipped: %s", fileName)
		return "", "", "", "", utils.LogAndReturnError("file '%s' has an extension that should be ignored", fileName)
	}

	extension = filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, extension)
	
	// The path is the source of truth for the version
	version = versionFromPath
	
	// Build a pattern to find the artifactID and classifier
	// We search for the exact version from the path in the file name
	versionPart := strings.ReplaceAll(versionFromPath, ".", "\\.")
	pattern := fmt.Sprintf(`^(.+?)-%s(?:-(.+))?$`, versionPart)
	regex := regexp.MustCompile(pattern)
	matches := regex.FindStringSubmatch(baseName)
	
	if matches == nil || len(matches) < 2 {
		utils.LogError("Failed to parse file name '%s' with version '%s'", fileName, versionFromPath)
		return "", "", "", "", utils.LogAndReturnError("failed to parse file name '%s'. Regex did not match", fileName)
	}
	
	// First group: artifactID
	artifactID = matches[1]
	
	// Second group (optional): classifier
	if len(matches) > 2 && matches[2] != "" {
		classifier = matches[2]
	} else {
		classifier = ""
	}
	
	utils.LogDebug("Parsed Maven file - ArtifactID: %s, Version: %s, Classifier: %s, Extension: %s",
		artifactID, version, classifier, extension)
	
	return artifactID, version, classifier, extension, nil
}

// GenerateMavenPath generates the path for the artifact in Maven repository format.
func GenerateMavenPath(groupID, artifactID, version, classifier, extension string) string {
	basePath := strings.ReplaceAll(groupID, ".", "/") + "/" + artifactID + "/" + version

	var formattedPath string
	if classifier != "" {
		formattedPath = fmt.Sprintf("%s/%s-%s-%s%s", basePath, artifactID, version, classifier, extension)
	} else {
		formattedPath = fmt.Sprintf("%s/%s-%s%s", basePath, artifactID, version, extension)
	}
	utils.LogDebug("Generated Maven Path: %s", formattedPath)
	return formattedPath
}
