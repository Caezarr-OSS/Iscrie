package scripts

import (
	"fmt"
	"iscrie/core/importer/maven2"
	"iscrie/utils"
	"os"
)

// TestMavenParser tests the Maven parser with different filename combinations
// This can be called from another package to test the parser
func TestMavenParser() {
	// Initialize logger
	err := utils.InitLogger("./logs", "debug")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer utils.CloseLogger()

	// Test cases
	testCases := []struct {
		fileName        string
		versionFromPath string
	}{
		// Standard cases
		{"my-artifact-1.0.0.jar", "1.0.0"},
		{"my-artifact-1.0.0-SNAPSHOT.jar", "1.0.0-SNAPSHOT"},
		{"my-artifact-1.0.0-javadoc.jar", "1.0.0"},
		{"my-artifact-1.0.0-sources.jar", "1.0.0"},
		
		// Artifacts with hyphens in name
		{"my-cool-artifact-1.0.0.jar", "1.0.0"},
		{"my-cool-artifact-1.0.0-SNAPSHOT.jar", "1.0.0-SNAPSHOT"},
		
		// Complex extensions
		{"my-artifact-1.0.0.jar.md5", "1.0.0"},
		{"my-artifact-1.0.0.jar.sha1", "1.0.0"},
		{"my-artifact-1.0.0-javadoc.jar.sha1", "1.0.0"},
		
		// Non-standard versions
		{"my-artifact-1.0.RC1.jar", "1.0.RC1"},
		{"my-artifact-1.0-alpha-1.jar", "1.0-alpha-1"},
		
		// Complex versions with classifiers
		{"my-artifact-1.0.0-SNAPSHOT-sources.jar", "1.0.0-SNAPSHOT"},
		{"my-artifact-1.0.0-SNAPSHOT-javadoc.jar", "1.0.0-SNAPSHOT"},
		{"my-artifact-1.0.0-with-deps.jar", "1.0.0-with-deps"},
		{"my-artifact-1.0.0-with-deps-test.jar", "1.0.0-with-deps"},
		{"my-artifact-1.0.0-with-deps-sources.jar", "1.0.0-with-deps"},
	}

	fmt.Println("Maven Parser Test")
	fmt.Println("=================")
	
	for i, tc := range testCases {
		fmt.Printf("\nTest case %d:\n", i+1)
		fmt.Printf("  File: %s, Path Version: %s\n", tc.fileName, tc.versionFromPath)
		
		artifactID, version, classifier, extension, err := maven2.ParseMavenFileName(tc.fileName, tc.versionFromPath)
		
		if err != nil {
			fmt.Printf("  ERROR: %v\n", err)
		} else {
			fmt.Printf("  Result: \n")
			fmt.Printf("    ArtifactID: %s\n", artifactID)
			fmt.Printf("    Version: %s\n", version)
			fmt.Printf("    Classifier: %s\n", classifier)
			fmt.Printf("    Extension: %s\n", extension)
			
			// Generate expected Maven path
			mavenPath := maven2.GenerateMavenPath("com.example", artifactID, version, classifier, extension)
			fmt.Printf("    Maven Path: %s\n", mavenPath)
		}
	}
}
