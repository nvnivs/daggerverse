package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/Excoriate/daggerverse/module-template/tests/internal/dagger"
)

// getGolangAlpineContainer returns a container with the Go toolchain installed.
//
// This function returns a container with the Go toolchain installed, which is
// suitable for testing Go related functionality.
func getGolangAlpineContainer(version string) *dagger.Container {
	if version == "" {
		version = "1.20.4"
	}

	return dag.Container().
		From("golang:" + version + "-alpine")
}

// TestGoWithGoPlatform tests the setting of different Go platforms within the target module's container.
//
// This method creates a target module with a Golang Alpine container and sets different Go platforms.
// It verifies if the Go platform is correctly set by running the `go version` command within the container
// for each defined platform.
//
// Arguments:
// - ctx (context.Context): The context for the test execution.
//
// Returns:
//   - error: Returns an error if there is an issue setting the Go platform, executing
//     commands in the container, or if the `go version` output does not match the expected pattern.
func (m *Tests) TestGoWithGoPlatform(ctx context.Context) error {
	platforms := map[dagger.Platform]string{
		"linux/amd64":   "go version go1.20.4 linux/amd64",
		"linux/arm64":   "go version go1.20.4 linux/arm64",
		"windows/amd64": "go version go1.20.4 windows/amd64",
	}

	for platform := range platforms {
		targetModule := dag.
			ModuleTemplate(dagger.ModuleTemplateOpts{
				Ctr: getGolangAlpineContainer(""),
			}).WithGoPlatform(dagger.ModuleTemplateWithGoPlatformOpts{
			Platform: platform,
		})

		// Check if the Go platform is set correctly.
		out, err := targetModule.Ctr().
			WithExec([]string{"go", "version"}).
			Stdout(ctx)

		if out == "" {
			return WrapErrorf(err, "failed to run go version for platform %s", platform)
		}

		if err != nil {
			return WrapErrorf(err, "failed to run go version for platform %s", platform)
		}

		// Validate the GOOS and GOARCH environment variables.
		goosOut, goosOutErr := targetModule.Ctr().
			WithExec([]string{"printenv", "GOOS"}).
			Stdout(ctx)

		if goosOutErr != nil {
			return WrapErrorf(goosOutErr, "failed to get GOOS for platform %s", platform)
		}

		goarchOut, goarchOutErr := targetModule.Ctr().
			WithExec([]string{"printenv", "GOARCH"}).
			Stdout(ctx)

		if goarchOutErr != nil {
			return WrapErrorf(goarchOutErr, "failed to get GOARCH for platform %s", platform)
		}

		platformStr := string(platform)

		expectedGOOS := strings.Split(platformStr, "/")[0]
		expectedGOARCH := strings.Split(platformStr, "/")[1]

		if !strings.Contains(goosOut, expectedGOOS) {
			return WrapErrorf(err, "expected GOOS=%s, got %s for platform %s", expectedGOOS, goosOut, platform)
		}

		if !strings.Contains(goarchOut, expectedGOARCH) {
			return WrapErrorf(err, "expected GOARCH=%s, got %s for platform %s", expectedGOARCH, goarchOut, platform)
		}
	}

	return nil
}

// TestGoWithCgoEnabled tests enabling CGO in a Go Alpine container.
//
// This function performs the following steps:
// 1. Sets up the Go Alpine container.
// 2. Enables the CGO feature in the Go environment.
// 3. Verifies that the CGO_ENABLED environment variable is set to "1".
//
// Parameters:
// - ctx: The context to control the execution.
//
// Returns:
// - error: If any step fails or produces an unexpected output, an error is returned.
func (m *Tests) TestGoWithCgoEnabled(ctx context.Context) error {
	// Setting the Go Alpine container.
	targetModule := dag.ModuleTemplate(dagger.ModuleTemplateOpts{
		Ctr: getGolangAlpineContainer(""),
	})

	// Enable CGO.
	targetModule = targetModule.
		WithGoCgoEnabled()

	out, err := targetModule.Ctr().
		WithExec([]string{"go", "env", "CGO_ENABLED"}).
		Stdout(ctx)

	if err != nil {
		return WrapError(err, "failed to get CGO_ENABLED environment variable")
	}

	if out == "" {
		return WrapError(err, "expected to have at least one folder, got empty output")
	}

	if !strings.Contains(out, "1") {
		return WrapErrorf(err, "expected CGO_ENABLED to be set to 1, got %s", out)
	}

	return nil
}

// TestGoWithCgoDisabled tests disabling CGO in a Go Alpine container.
//
// This function performs the following steps:
// 1. Sets up the Go Alpine container.
// 2. Disables the CGO feature in the Go environment.
// 3. Verifies that the CGO_ENABLED environment variable is set to "0".
//
// Parameters:
// - ctx: The context to control the execution.
//
// Returns:
// - error: If any step fails or produces an unexpected output, an error is returned.
func (m *Tests) TestGoWithCgoDisabled(ctx context.Context) error {
	// Setting the Go Alpine container.
	targetModule := dag.ModuleTemplate(dagger.ModuleTemplateOpts{
		Ctr: getGolangAlpineContainer(""),
	})

	// Disable CGO.
	targetModule = targetModule.
		WithCgoDisabled()

	out, err := targetModule.Ctr().
		WithExec([]string{"go", "env", "CGO_ENABLED"}).
		Stdout(ctx)

	if err != nil {
		return WrapError(err, "failed to get CGO_ENABLED environment variable")
	}

	if out == "" {
		return WrapError(err, "expected to have at least one folder, got empty output")
	}

	if !strings.Contains(out, "0") {
		return WrapErrorf(err, "expected CGO_ENABLED to be set to 0, got %s", out)
	}

	return nil
}

// TestGoWithGoBuildCache verifies that the Go build cache (GOCACHE) is set correctly
// in the provided Alpine container.
//
// This function performs the following steps:
// 1. Sets up the Go Alpine container without specifying a particular Go version.
// 2. Configures the Go build cache.
// 3. Executes the `go env GOCACHE` command to retrieve the GOCACHE environment variable.
// 4. Validates that the GOCACHE environment variable is set to the expected path.
//
// Parameters:
// - ctx: The context to control the execution.
//
// Returns:
// - error: If any of the steps fail, an error is returned indicating what went wrong.
func (m *Tests) TestGoWithGoBuildCache(ctx context.Context) error {
	// Setting the Go Alpine container.
	targetModule := dag.ModuleTemplate(dagger.ModuleTemplateOpts{
		Ctr: getGolangAlpineContainer(""),
	})

	// Set the Go build cache.
	targetModule = targetModule.WithGoBuildCache()
	out, err := targetModule.Ctr().
		WithExec([]string{"go", "env", "GOCACHE"}).
		Stdout(ctx)

	if err != nil {
		return WrapError(err, "failed to get GOCACHE environment variable")
	}

	if out == "" {
		return WrapError(err, "expected to have at least one folder, got empty output")
	}

	if !strings.Contains(out, "/root/.cache/go-build") {
		return WrapErrorf(err, "expected GOCACHE to be set to /root/.cache/go-build, got %s", out)
	}

	return nil
}

// TestGoWithGoModCache verifies that the Go module cache (GOMODCACHE) is set correctly
// in the provided Alpine container.
//
// This function performs the following steps:
// 1. Sets up the Go Alpine container without specifying a particular Go version.
// 2. Configures the Go module cache.
// 3. Executes the `go env GOMODCACHE` command to retrieve the GOMODCACHE environment variable.
// 4. Validates that the GOMODCACHE environment variable is set to the expected path.
//
// Parameters:
// - ctx: The context to control the execution.
//
// Returns:
// - error: If any of the steps fail, an error is returned indicating what went wrong.
func (m *Tests) TestGoWithGoModCache(ctx context.Context) error {
	// Setting the Go Alpine container.
	targetModule := dag.ModuleTemplate(dagger.ModuleTemplateOpts{
		Ctr: getGolangAlpineContainer(""),
	})

	// Set the Go mod cache.
	targetModule = targetModule.
		WithGoModCache()

	out, err := targetModule.Ctr().
		WithExec([]string{"go", "env", "GOMODCACHE"}).
		Stdout(ctx)

	if err != nil {
		return WrapError(err, "failed to get GOMODCACHE environment variable")
	}

	if out == "" {
		return WrapError(err, "expected to have at least one folder, got empty output")
	}

	if !strings.Contains(out, "/go/pkg/mod") {
		return WrapErrorf(err, "expected GOMODCACHE to be set to /go/pkg/mod, got %s", out)
	}

	return nil
}

// TestGoWithGoInstall tests the installation of various Go packages
// in the provided Alpine container.
//
// This function performs the following steps:
// 1. Sets up the Go Alpine container with the expected Go version.
// 2. Installs a list of specified Go packages.
// 3. Verifies the installation by checking if the installed packages are in the PATH.
// 4. Ensures the Go module cache is correctly set.
//
// Parameters:
// - ctx: The context to control the execution.
//
// Returns:
// - error: If any of the installation or verification steps fail, an error is returned.
//
// longer function.
func (m *Tests) TestGoWithGoInstall(ctx context.Context) error {
	// Setting the Go Alpine container.
	targetModule := dag.ModuleTemplate(dagger.ModuleTemplateOpts{
		Ctr: getGolangAlpineContainer("1.22.3"),
	})

	// List of packages to install
	packages := []string{
		"gotest.tools/gotestsum@latest",
		"github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.1",
		"github.com/go-delve/delve/cmd/dlv@latest",
	}

	// Prepare the installation commands for each package
	targetModule = targetModule.WithGoInstall(packages)
	// Sync, to execute the installation commands
	_, err := targetModule.Ctr().Stdout(ctx)
	if err != nil {
		return WrapError(err, "failed to get the standard output of the commands to install Go packages")
	}

	// Verify installations
	for _, pkg := range packages {
		pkgNameSplit := strings.Split(pkg, "/")
		pkgName := pkgNameSplit[len(pkgNameSplit)-1]
		pkgName = strings.Split(pkgName, "@")[0]

		out, err := targetModule.Ctr().WithExec([]string{"which", pkgName}).Stdout(ctx)
		if err != nil {
			return WrapErrorf(err, "failed to verify installation of %s", pkg)
		}

		if out == "" {
			return WrapErrorf(err, "expected to find %s in PATH, got empty output", pkgName)
		}
	}

	// Verify Go module cache
	out, goCacheErr := targetModule.Ctr().WithExec([]string{"go", "env", "GOMODCACHE"}).Stdout(ctx)
	if goCacheErr != nil {
		return WrapError(goCacheErr, "failed to get GOMODCACHE environment variable")
	}

	if !strings.Contains(out, "/go/pkg/mod") {
		return Errorf("expected GOMODCACHE to be set to /go/pkg/mod, got %s", strings.TrimSpace(out))
	}

	return nil
}

// TestGoWithGoExec tests the execution of various Go commands and ensures they produce the expected
// results in the provided Alpine container.
//
// This function performs the following steps:
// 1. Sets up the Go Alpine container with the expected Go version.
// 2. Executes the `go version` command and verifies the output against the expected Go version.
// 3. Runs additional Go commands (e.g., `go env GOPATH`) and checks their
// output against expected values.
// 4. Verifies specific Go environment variables (e.g., `GOPROXY`) to ensure
// they are set correctly.
//
// Parameters:
// - ctx: The context to control the execution.
//
// Returns:
// - error: If any of the commands fail or produce unexpected output, an error is returned.
//
//nolint:cyclop // The test handles multiple commands and environments, requiring a longer function.
func (m *Tests) TestGoWithGoExec(ctx context.Context) error {
	// Setting the Go Alpine container.
	expectedGoVersion := "1.22.3"
	targetModule := dag.ModuleTemplate(dagger.ModuleTemplateOpts{
		Ctr: getGolangAlpineContainer(expectedGoVersion),
	})

	// Execute `go version` command and capture the output.
	targetModule = targetModule.
		WithGoExec([]string{"version"})

	out, err := targetModule.Ctr().Stdout(ctx)
	if err != nil {
		return WrapErrorf(err, "failed to get the standard output of the "+
			"version command with expected Go version %s", expectedGoVersion)
	}

	if out == "" {
		return Errorf("expected to have Go version %s, got empty output", expectedGoVersion)
	}

	// Verify the Go version.
	expectedVersionString := "go" + expectedGoVersion
	if !strings.Contains(out, expectedVersionString) {
		return Errorf("expected Go version %s, but got %s", expectedVersionString, out)
	}

	// Additional Go commands to verify.
	commands := map[string]string{
		"go env GOPATH":  "/go",
		"go env GOROOT":  "/usr/local/go",
		"go env GOPROXY": "https://proxy.golang.org,direct",
	}

	for cmd, expectedOutput := range commands {
		out, execErr := targetModule.Ctr().
			WithExec(strings.Split(cmd, " ")).
			Stdout(ctx)
		if execErr != nil {
			return WrapErrorf(execErr, "failed to execute command: %s", cmd)
		}

		if out == "" {
			return Errorf("expected output for command %s, got empty output", cmd)
		}

		if !strings.Contains(out, expectedOutput) {
			return Errorf("expected output for command '%s' to contain %s, but got %s", cmd, expectedOutput, out)
		}
	}

	// Verify 'go env' command to ensure environment variables are set correctly.
	envVars := map[string]string{
		"GOPATH":  "/go",
		"GOROOT":  "/usr/local/go",
		"GOPROXY": "https://proxy.golang.org,direct",
	}

	for envVar, expectedVal := range envVars {
		out, envErr := targetModule.Ctr().
			WithExec([]string{"go", "env", envVar}).
			Stdout(ctx)
		if envErr != nil {
			return WrapErrorf(envErr, "failed to get %s environment variable", envVar)
		}

		if out == "" {
			return Errorf("expected %s environment variable to be set, got empty output", envVar)
		}

		if !strings.Contains(out, expectedVal) {
			return Errorf("expected %s to be set to %s, but got %s", envVar, expectedVal, out)
		}
	}

	return nil
}

// TestGoWithGoBuild tests the Go build process using the provided Alpine container.
//
// This function performs the following steps:
//
// 1. Sets up the Go Alpine container with the expected Go version.
// 2. Configures the build process with specific options including the source
// directory, target platform, package to build, and output binary name.
// 3. Executes the build process and checks for errors.
// 4. Verifies the presence of the output binary in the container's directory.
// 5. Runs the binary and verifies the output against the expected string.
//
// Parameters:
// - ctx: The context to control the execution.
func (m *Tests) TestGoWithGoBuild(ctx context.Context) error {
	// Setting the Go Alpine container.
	expectedGoVersion := "1.22.3"
	targetModule := dag.ModuleTemplate(dagger.ModuleTemplateOpts{
		Ctr: getGolangAlpineContainer(expectedGoVersion),
	})

	// Configure the build process with specific options.
	targetModule = targetModule.
		WithSource(m.TestDir).
		WithGoBuild(dagger.ModuleTemplateWithGoBuildOpts{
			Platform: "linux/amd64",
			Pkg:      "golang/main.go",
			Verbose:  true,
			// Output the binary called dagger to the current directory.
			Output: "./dagger",
		})

	// Execute the build process and capture the output.
	_, err := targetModule.Ctr().Stdout(ctx)

	if err != nil {
		return WrapErrorf(err, "failed to get the standard output of "+
			"the version command with expected Go version %s", expectedGoVersion)
	}

	// Verify the presence of files in the directory.
	lsOut, lsErr := targetModule.Ctr().WithExec([]string{"ls"}).Stdout(ctx)
	if lsErr != nil {
		return WrapErrorf(lsErr, "failed to get the standard output of the "+
			"version command with expected Go version %s", expectedGoVersion)
	}

	if lsOut == "" {
		return Errorf("expected to have files listed, got empty output")
	}

	// Verify the presence of the output binary named 'dagger'.
	daggerOut, daggerErr := targetModule.Ctr().WithExec([]string{"ls", "dagger"}).Stdout(ctx)
	if daggerErr != nil {
		return WrapErrorf(daggerErr, "failed to get the standard output of the version "+
			"command with expected Go version %s", expectedGoVersion)
	}

	if daggerOut == "" {
		return Errorf("expected to have a file called 'dagger', got empty output")
	}

	// Run the binary and verify the output.
	expectedOutput := "Hello, Dagger!"
	out, binaryErr := targetModule.
		Ctr().
		WithExec([]string{"./dagger"}).Stdout(ctx)

	if binaryErr != nil {
		return WrapErrorf(binaryErr, "failed to get the standard output of the version "+
			"command with expected Go version %s", expectedGoVersion)
	}

	if out == "" {
		return Errorf("expected to have output, got empty output")
	}

	if !strings.Contains(out, expectedOutput) {
		return Errorf("expected output to contain %s, got %s", expectedOutput, out)
	}

	return nil
}

// TestGoWithGoPrivate tests the configuration of the GOPRIVATE environment variable
// using the WithGoPrivate method. It ensures that the GOPRIVATE environment variable
// is set correctly within the specified context.
//
// This function performs the following steps:
// 1. Sets up the Go module container using the default module template.
// 2. Configures the GOPRIVATE environment variable to the specified value.
// 3. Executes the `go env GOPRIVATE` command to retrieve the GOPRIVATE environment variable.
// 4. Validates that the GOPRIVATE environment variable is set to the expected value.
//
// Parameters:
// - ctx: The context to control the execution.
//
// Returns:
// - error: If any of the steps fail, an error is returned indicating what went wrong.
func (m *Tests) TestGoWithGoPrivate(ctx context.Context) error {
	// Setting up the Go module container using the default module template.
	targetModule := dag.ModuleTemplate(
		dagger.ModuleTemplateOpts{
			Ctr: getGolangAlpineContainer(""),
		},
	)

	// Set the GOPRIVATE environment variable to the specified value.
	targetModule = targetModule.
		WithGoPrivate("github.com/Excoriate")

	// Execute the `go env GOPRIVATE` command to retrieve the GOPRIVATE environment variable.
	envOut, err := targetModule.Ctr().
		WithExec([]string{"go", "env", "GOPRIVATE"}).
		Stdout(ctx)

	if err != nil {
		return WrapError(err, "failed to get GOPRIVATE environment variable")
	}

	// Validate that the GOPRIVATE environment variable is set to the expected value.
	if envOut == "" {
		return Errorf("expected to have at least one folder, got empty output")
	}

	if !strings.Contains(envOut, "github.com/Excoriate") {
		return WrapErrorf(err, "expected GOPRIVATE to be set "+
			"to github.com/Excoriate, got %s", envOut)
	}

	return nil
}

// TestGoWithGCCCompiler verifies the installation and presence of the GCC compiler
// within the specified context, using the provided Alpine container setup.
//
// This function performs the following steps:
// 1. Sets up the Go module container using the default module template.
// 2. Installs the GCC compiler and associated development tools.
// 3. Executes the `gcc --version` command to verify the GCC compiler installation.
// 4. Validates the output to confirm the presence of the GCC compiler.
//
// Parameters:
// - ctx: The context to control the execution.
//
// Returns:
//   - error: If any of the installation or verification steps fail, an error is returned
//     indicating what went wrong.
func (m *Tests) TestGoWithGCCCompiler(ctx context.Context) error {
	// Setting up the Go module container using the default module template.
	targetModule := dag.ModuleTemplate(
		dagger.ModuleTemplateOpts{
			Ctr: getGolangAlpineContainer(""),
		},
	)

	// Install the GCC compiler and development tools.
	targetModule = targetModule.
		WithGoGcccompiler()

	// Execute the `gcc --version` command to check if the GCC compiler is installed.
	cmdOut, cmdErr := targetModule.Ctr().
		WithExec([]string{"gcc", "--version"}).
		Stdout(ctx)

	if cmdErr != nil {
		return WrapError(cmdErr, "failed to run gcc --version command")
	}

	// Validate that the GCC compiler is installed.
	if cmdOut == "" {
		return Errorf("expected to have output when running gcc --version, " +
			"got empty output")
	}

	if !strings.Contains(cmdOut, "gcc") {
		return Errorf("expected to have found gcc in the output, got %s", cmdOut)
	}

	return nil
}

// TestGoWithGoTestSum verifies the installation and presence of the GoTestSum tool
// within the specified context, using the provided Alpine container setup.
//
// This function performs the following steps:
// 1. Sets up the Go module container using the default module template.
// 2. Installs the GoTestSum tool and its dependency `tparse` in the container.
// 3. Executes the `gotestsum --version` command to verify the GoTestSum tool installation.
// 4. Validates the output to confirm the presence of the GoTestSum tool.
//
// Parameters:
// - ctx: The context to control the execution.
//
// Returns:
//   - error: If any of the installation or verification steps fail, an error is returned
//     indicating what went wrong.
//
// TestGoWithGoTestSum verifies the installation and presence of the GoTestSum tool
// within the specified context, using the provided Alpine container setup.
func (m *Tests) TestGoWithGoTestSum(ctx context.Context) error {
	baseModule := dag.ModuleTemplate(
		dagger.ModuleTemplateOpts{
			Ctr: getGolangAlpineContainer("1.21"),
		},
	)

	testCases := []struct {
		name             string
		goTestSumVersion string
		tParseVersion    string
		skipTParse       bool
	}{
		{"Default versions", "", "", false},
		{"Specific GoTestSum version", "v1.10.0", "", false},
		{"Specific versions for both", "v1.10.0", "v0.11.0", false},
		{"Skip TParse", "v1.10.0", "", true},
	}

	for _, testCase := range testCases {
		if err := m.runGoTestSumTestCase(ctx, baseModule, testCase); err != nil {
			return err
		}
	}

	return nil
}

func (m *Tests) runGoTestSumTestCase(ctx context.Context, baseModule *dagger.ModuleTemplate, testCase struct {
	name             string
	goTestSumVersion string
	tParseVersion    string
	skipTParse       bool
}) error {
	targetModule := baseModule.
		WithGoTestSum(dagger.ModuleTemplateWithGoTestSumOpts{
			GoTestSumVersion: testCase.goTestSumVersion,
			TParseVersion:    testCase.tParseVersion,
			SkipTparse:       testCase.skipTParse,
		})

	if err := m.checkToolInstallation(ctx, targetModule, "gotestsum",
		testCase.goTestSumVersion, testCase.name); err != nil {
		return err
	}

	if !testCase.skipTParse {
		if err := m.checkToolInstallation(ctx, targetModule, "tparse", testCase.tParseVersion, testCase.name); err != nil {
			return err
		}
	} else {
		if err := m.verifyToolNotInstalled(ctx, targetModule, "tparse", testCase.name); err != nil {
			return err
		}
	}

	return nil
}

func (m *Tests) checkToolInstallation(ctx context.Context, targetModule *dagger.ModuleTemplate,
	toolName, expectedVersion,
	testCaseName string) error {
	toolOut, toolErr := targetModule.Ctr().
		WithExec([]string{toolName, "--version"}).
		Stdout(ctx)

	if toolErr != nil {
		return WrapError(toolErr, fmt.Sprintf("%s: failed to run %s --version command", testCaseName, toolName))
	}

	if toolOut == "" {
		return Errorf("%s: expected to have output when running %s --version, got empty output", testCaseName, toolName)
	}

	if !strings.Contains(toolOut, toolName) {
		return Errorf("%s: expected to have found %s in the output, got %s", testCaseName, toolName, toolOut)
	}

	installedVersion := extractVersion(toolOut)
	fmt.Printf("%s: Installed %s version: %s\n", testCaseName, toolName, installedVersion)

	if expectedVersion != "" && expectedVersion != "latest" {
		if installedVersion == "dev" {
			fmt.Printf("%s: Warning: %s is installed with 'dev' version\n", testCaseName, toolName)
		} else if !strings.HasPrefix(installedVersion, expectedVersion) {
			return Errorf("%s: expected %s version starting with %s, but got %s",
				testCaseName, toolName, expectedVersion, installedVersion)
		}
	}

	return nil
}

func (m *Tests) verifyToolNotInstalled(ctx context.Context,
	targetModule *dagger.ModuleTemplate, toolName, testCaseName string) error {
	_, toolErr := targetModule.Ctr().
		WithExec([]string{toolName, "--version"}).
		Stdout(ctx)

	if toolErr == nil {
		return Errorf("%s: expected %s to not be installed, but it was found", testCaseName, toolName)
	}

	return nil
}

// extractVersion returns the version from the output string.
func extractVersion(output string) string {
	const minParts = 2

	parts := strings.Fields(output)

	if len(parts) >= minParts {
		return parts[len(parts)-1]
	}

	return ""
}

// TestGoWithGoReleaserAndGolangCILint tests the installation and setup
// of GoReleaser and GoLangCILint using gotoolbox.
//
// This function sets up the Go toolbox with a specified Go version, installs
// GoReleaser and GoLangCILint, and verifies their installation.
//
// ctx: The context for the test execution, to control cancellation and deadlines.
//
// Returns an error if the installation or verification of GoReleaser or GoLangCILint fails.
func (m *Tests) TestGoWithGoReleaserAndGolangCILint(ctx context.Context) error {
	goContainerWithGit := dag.
		Container().
		From("golang:1.23-alpine").
		WithExec([]string{"apk", "add", "--no-cache", "git"})

	// Initialize the Go toolbox with the specified version.
	targetModDefault := dag.
		ModuleTemplate(dagger.ModuleTemplateOpts{
			Ctr: goContainerWithGit,
		})

	// Set the installation commands for GoReleaser and GoLangCILint.
	targetModDefault = targetModDefault.
		WithGoReleaser().
		WithGoLint("v1.60.1")

	// Execute the container to install the tools.
	_, ctrErr := targetModDefault.
		Ctr().
		Stdout(ctx)

	if ctrErr != nil {
		return WrapError(ctrErr, "failed to install GoReleaser and GoLangCILint")
	}

	// Check golangci-lint binary installed in path.
	golangciPathOut, golangciPathErr := targetModDefault.
		Ctr().
		WithExec([]string{"which", "golangci-lint"}).
		Stdout(ctx)

	if golangciPathErr != nil {
		return WrapError(golangciPathErr, "failed to get golangci-lint path")
	}

	if golangciPathOut == "" {
		return Errorf("expected to have golangci-lint in path /go/bin/golangci-lint, got empty output")
	}

	if !strings.Contains(golangciPathOut, "/go/bin/golangci-lint") {
		return Errorf("expected to have golangci-lint "+
			"in path /go/bin/golangci-lint, got %s", golangciPathOut)
	}

	// Check goreleaser binary installed in path.
	goreleaserPathOut, goreleaserPathErr := targetModDefault.
		Ctr().
		WithExec([]string{"which", "goreleaser"}).
		Stdout(ctx)

	if goreleaserPathErr != nil {
		return WrapError(goreleaserPathErr, "failed to get goreleaser path")
	}

	if goreleaserPathOut == "" {
		return Errorf("expected to have goreleaser in path /go/bin/goreleaser, got empty output")
	}

	if !strings.Contains(goreleaserPathOut, "/go/bin/goreleaser") {
		return Errorf("expected to have goreleaser "+
			"in path /go/bin/goreleaser, got %s", goreleaserPathOut)
	}

	return nil
}
