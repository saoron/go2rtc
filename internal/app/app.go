package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

var (
	Version    string
	UserAgent  string
	ConfigPath string
	Info       = make(map[string]any)
)

const usage = `Usage of go2rtc:

  -c, --config   Path to config file or config string as YAML or JSON, support multiple
  -d, --daemon   Run in background
  -v, --version  Print version and exit
`

func Init() {
	var config flagConfig
	var daemon bool
	var version bool

	flag.Var(&config, "config", "")
	flag.Var(&config, "c", "")
	flag.BoolVar(&daemon, "daemon", false, "")
	flag.BoolVar(&daemon, "d", false, "")
	flag.BoolVar(&version, "version", false, "")
	flag.BoolVar(&version, "v", false, "")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()

	revision, vcsTime := readRevisionTime()

	if version {
		fmt.Printf("go2rtc version %s (%s) %s/%s\n", Version, revision, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if daemon && os.Getppid() != 1 {
		if runtime.GOOS == "windows" {
			fmt.Println("Daemon mode is not supported on Windows")
			os.Exit(1)
		}

		// Re-run the program in background and exit
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		if err := cmd.Start(); err != nil {
			fmt.Println("Failed to start daemon:", err)
			os.Exit(1)
		}
		fmt.Println("Running in daemon mode with PID:", cmd.Process.Pid)
		os.Exit(0)
	}

	UserAgent = "go2rtc/" + Version

	Info["version"] = Version
	Info["revision"] = revision

	initConfig(config)
	initLogger()

	platform := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	Logger.Info().Str("version", Version).Str("platform", platform).Str("revision", revision).Msg("go2rtc")
	Logger.Debug().Str("version", runtime.Version()).Str("vcs.time", vcsTime).Msg("build")

	if ConfigPath != "" {
		Logger.Info().Str("path", ConfigPath).Msg("config")
	}
}

func readRevisionTime() (revision, vcsTime string) {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) > 7 {
					revision = setting.Value[:7]
				} else {
					revision = setting.Value
				}
			case "vcs.time":
				vcsTime = setting.Value
			case "vcs.modified":
				if setting.Value == "true" {
					revision = "mod." + revision
				}
			}
		}
	}
	return
}


type AssetToken struct {
	Token     string
	ExpiresAt int64
}

func VerifyAssetToken(tokenToCompare string) bool {
	if tokenToCompare == "" {
		return false
	}

	tokens := DBRead("assetTokens")
	if tokens == "" {
		tokens = "[]"
	}
	var tokensObject []AssetToken
	err := json.Unmarshal([]byte(tokens), &tokensObject)
	if err != nil {
		fmt.Println("Failed to unmarshal asset token " + err.Error())
		return false
	}

	for _, token := range tokensObject {
		if (token.Token == tokenToCompare) && (token.ExpiresAt-time.Now().Unix()) > 0 {
			return true
		}
	}
	fmt.Println("token not found!")
	return false
}

func DBRead(key string) string {
	key = strings.ToLower(key)
	resp, err := Get("http://127.0.0.1:4000/read?key="+key, 1)

	if err != nil {
		return ""
	}
	return resp
}

func Get(url string, retry int) (string, error) {
	if retry <= 0 {
		return "", errors.New("Max retries reached " + url)
	}

	resp, err := http.Get(url)

	if err != nil {
		time.Sleep(time.Millisecond * 100)
		return Get(url, retry-1)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		time.Sleep(time.Millisecond * 250)
		return Get(url, retry-1)
	}
	return string(body), nil
}



func IsProtectedPath(path string) bool {
	return  strings.Contains((path), ".html") ||
		strings.Contains((path), ".jpeg") 
}