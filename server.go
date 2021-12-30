package adb

import (
	stderrors "errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/hemeda3/goadb/public/errors"
	"github.com/hemeda3/goadb/wire"
)

const (
	AdbExecutableName = "adb"

	// Default port the adb Server listens on.
	AdbPort = 5037
)

type ServerConfig struct {
	// Path to the adb executable. If empty, the PATH environment variable will be searched.
	PathToAdb string

	// Host and port the adb Server is listening on.
	// If not specified, will use the default port on localhost.
	Host string
	Port int

	// Dialer used to connect to the adb Server.
	Dialer

	fs *filesystem
}

// Server knows how to start the adb Server and connect to it.
type server interface {
	Start() error
	Root() error
	Install(apkPath string) error

	Dial() (*wire.Conn, error)
}

func roundTripSingleResponse(s server, req string) ([]byte, error) {
	conn, err := s.Dial()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return conn.RoundTripSingleResponse([]byte(req))
}

type realServer struct {
	config ServerConfig

	// Caches Host:Port so they don't have to be concatenated for every dial.
	address string
}

func newServer(config ServerConfig) (server, error) {
	if config.Dialer == nil {
		config.Dialer = tcpDialer{}
	}

	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = AdbPort
	}

	if config.fs == nil {
		config.fs = localFilesystem
	}

	if config.PathToAdb == "" {
		path, err := config.fs.LookPath(AdbExecutableName)
		if err != nil {
			return nil, errors.WrapErrorf(err, errors.ServerNotAvailable, "could not find %s in PATH", AdbExecutableName)
		}
		config.PathToAdb = path
	}
	if err := config.fs.IsExecutableFile(config.PathToAdb); err != nil {
		return nil, errors.WrapErrorf(err, errors.ServerNotAvailable, "invalid adb executable: %s", config.PathToAdb)
	}

	return &realServer{
		config:  config,
		address: fmt.Sprintf("%s:%d", config.Host, config.Port),
	}, nil
}

// Dial tries to connect to the Server. If the first attempt fails, tries starting the Server before
// retrying. If the second attempt fails, returns the error.
func (s *realServer) Dial() (*wire.Conn, error) {
	conn, err := s.config.Dial(s.address)
	if err != nil {
		// Attempt to start the Server and try again.
		if err = s.Start(); err != nil {
			return nil, errors.WrapErrorf(err, errors.ServerNotAvailable, "error starting Server for dial")
		}

		conn, err = s.config.Dial(s.address)
		if err != nil {
			return nil, err
		}
	}
	return conn, nil
}

// StartServer ensures there is a Server running.
func (s *realServer) Start() error {
	output, err := s.config.fs.CmdCombinedOutput(s.config.PathToAdb, fmt.Sprintf("tcp:%s", s.address), "start-server")
	if err != nil {
		output, err = s.config.fs.CmdCombinedOutput(s.config.PathToAdb, "start-server")

	}

	outputStr := strings.TrimSpace(string(output))

	return errors.WrapErrorf(err, errors.ServerNotAvailable, "error starting Server: %s\noutput:\n%s", err, outputStr)
}

// root ensures there is a Server running as root.
func (s *realServer) Root() error {
	output, err := s.config.fs.CmdCombinedOutput(s.config.PathToAdb, "root")
	outputStr := strings.TrimSpace(string(output))
	fmt.Println("rooting result ", string(output))
	return errors.WrapErrorf(err, errors.ServerNotAvailable, "error rooting Server: %s\noutput:\n%s", err, outputStr)
}

// root ensures there is a Server running as root.
func (s *realServer) Install(apkPath string) error {
	output, err := s.config.fs.CmdCombinedOutput(s.config.PathToAdb, "-e", "install", apkPath)
	outputStr := strings.TrimSpace(string(output))
	fmt.Println("outputStr " + string(outputStr))

	fmt.Println("apk installing result %s, path %s ", string(output), apkPath)
	return errors.WrapErrorf(err, errors.ServerNotAvailable, "error apk installing Server: %s\noutput:\n%s", err, outputStr)
}

// filesystem abstracts interactions with the local filesystem for testability.
type filesystem struct {
	// Wraps exec.LookPath.
	LookPath func(string) (string, error)

	// Returns nil if path is a regular file and executable by the current user.
	IsExecutableFile func(path string) error

	// Wraps exec.Command().CombinedOutput()
	CmdCombinedOutput func(name string, arg ...string) ([]byte, error)
}

var localFilesystem = &filesystem{
	LookPath: exec.LookPath,
	IsExecutableFile: func(path string) error {
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return stderrors.New("not a regular file")
		}
		return isExecutable(path)
	},
	CmdCombinedOutput: func(name string, arg ...string) ([]byte, error) {
		return exec.Command(name, arg...).CombinedOutput()
	},
}
