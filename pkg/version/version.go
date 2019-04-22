package version

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/spf13/pflag"
)

var (
	// RELEASE returns the release version
	RELEASE = "UNKNOWN"
	// REPO returns the git repository URL
	REPO = "UNKNOWN"
	// COMMIT returns the short sha from git
	COMMIT = "UNKNOWN"
)

const versionFlagName = "version"

var (
	// versionFlag make version flag global
	versionFlag = version(versionFlagName, VersionFalse, "Print version information,then quit!")
)

type versionValue int

const (
	VersionFalse versionValue = 0
	VersionTrue  versionValue = 1
	VersionRaw   versionValue = 2

	strRawVersion string = "raw"
)

func (v *versionValue) Get() interface{} {
	// Consider the thought.
	return versionValue(*v)
}

func (v *versionValue) Set(s string) error {
	if s == strRawVersion {
		*v = VersionRaw
		return nil
	}

	boolV, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	if boolV {
		*v = VersionTrue
	} else {
		*v = VersionFalse
	}
	return nil
}

func (v *versionValue) String() string {
	if *v == VersionRaw {
		return strRawVersion
	}
	return fmt.Sprintf("%v", bool(*v == VersionTrue))
}

func (v *versionValue) Type() string {
	return "version"
}

func versionVar(v *versionValue, name string, value versionValue, usage string) {
	*v = value
	pflag.Var(v, name, usage)
	// --version will be treated as "--version=true"
	pflag.Lookup(name).NoOptDefVal = "true"
}
func version(name string, value versionValue, usage string) *versionValue {
	v := new(versionValue)
	versionVar(v, name, value, usage)
	return v
}

func PrintAndExitIfRequested() {
	if *versionFlag == VersionRaw {
		fmt.Printf("%#v\n", Get())
		os.Exit(0)
	} else if *versionFlag == VersionTrue {
		fmt.Printf("Indagate Version: %s\n", Get())
		os.Exit(0)
	}
}

// AddFlags registers this package's flags on arbitrary FlagSets, such that they point to the
// same value as the global flags.
func AddFlags(fs *pflag.FlagSet) {
	fs.AddFlag(pflag.Lookup(versionFlagName))
}

// Info contains versioning information.
// TODO: Add []string of api versions supported? It's still unclear
// how we'll want to distribute that information.
type Version struct {
	Major        string `json:"major"`
	Minor        string `json:"minor"`
	GitVersion   string `json:"gitVersion"`
	GitCommit    string `json:"gitCommit"`
	GitTreeState string `json:"gitTreeState"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Compiler     string `json:"compiler"`
	Platform     string `json:"platform"`
}

// String returns info as a human-friendly version string.
func (v Version) String() string {
	return v.GitVersion
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was built from.
func Get() Version {
	// These variables typically come from -ldflags settings and in
	// their absence fallback to the settings in pkg/version/base.go
	return Version{
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

var (

	// semantic version, derived by build scripts (see
	// https://github.com/kubernetes/community/blob/master/contributors/design-proposals/release/versioning.md
	// for a detailed discussion of this field)
	//
	// TODO: This field is still called "gitVersion" for legacy
	// reasons. For prerelease versions, the build metadata on the
	// semantic version is a git hash, but the version itself is no
	// longer the direct output of "git describe", but a slight
	// translation to be semver compliant.

	// NOTE: The $Format strings are replaced during 'git archive' thanks to the
	// companion .gitattributes file containing 'export-subst' in this same
	// directory.  See also https://git-scm.com/docs/gitattributes
	gitVersion   = "v0.0.0-master+$Format:%h$"
	gitCommit    = "$Format:%H$" // sha1 from git, output of $(git rev-parse HEAD)
	gitTreeState = ""            // state of git tree, either "clean" or "dirty"

	buildDate = "1970-01-01T00:00:00Z" // build date in ISO8601 format, output of $(date -u +'%Y-%m-%dT%H:%M:%SZ')
)
