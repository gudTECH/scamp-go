package scamp

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
)

// Config represents scamp config
type Config struct {
	// string key for easy equals, byte return for easy nil
	values map[string][]byte
}

// TODO: Will I regret using such a common name as a global variable?
var defaultConfig *Config

var defaultAnnounceInterval = 5

// DefaultConfigPath is the path at which the library will, by default, look for its configuration.
// var DefaultConfigPath = "/etc/SCAMP/soa.conf"

var globalConfig *Config

var defaultGroupIP = net.IPv4(239, 63, 248, 106)
var defaultGroupPort = 5555

// initConfig initializes the default scamp config struct and returns an
// error if it is unable to do so.
func initConfig(configPath string) (err error) {
	if len(configPath) == 0 {
		configPath = defaultConfigPath
	}

	defaultConfig = &Config{
		values: make(map[string][]byte),
	}
	err = defaultConfig.Load(configPath)
	if err != nil {
		err = fmt.Errorf("could not load config: %s", err)
		return
	}
	// TODO: review and possibly remove scamp debugger
	randomDebuggerString = scampDebuggerRandomString()
	return
}

// NewConfig creates a new configuration struct with default values initialized.
func NewConfig() (conf *Config) {
	conf = &Config{
		values: make(map[string][]byte),
	}
	return
}

// SetDefaultConfig sets the global configuration manually if need be.
// In general, users should use Initialize instead.
// func SetDefaultConfig(conf *Config) {
// 	initSCAMPLogger()
// 	defaultConfig = conf
// }

// DefaultConfig fetches the global configuration struct for use.
// This function panics if the global configuration is not initialized
// (with `Initialize()`).
func DefaultConfig() (conf *Config) {
	if defaultConfig == nil {
		log.Fatal("default Config{} is not initialized!")
	}
	return defaultConfig
}

// Load loads configuration k/v pairs from the file at the given path.
func (conf *Config) Load(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("could not open config file at `%s`: %s", configPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	err = conf.doLoad(scanner)
	if err != nil {
		return fmt.Errorf("couldn't load config info from file: %s", err)
	}

	return nil
}

func (conf *Config) doLoad(scanner *bufio.Scanner) (err error) {
	var read bool
	var configLine = regexp.MustCompile(`^\s*([\S^=]+)\s*=\s*([\S]+)`)
	for {
		read = scanner.Scan()
		if !read {
			break
		}

		re := configLine.FindSubmatch(scanner.Bytes())
		if re != nil {
			conf.values[string(re[1])] = re[2]
		}
	}
	return
}

// ServiceKeyPath uses the configuration to generate a path at which the key for the given service name should be found.
func (conf *Config) ServiceKeyPath(serviceName string) (keyPath []byte) {
	path := conf.values[serviceName+".soa_key"]
	if path == nil {
		//TODO: use default paths (and double check with Justin that they are correct)
		path = []byte(fmt.Sprintf("/etc/GT_private/services/%s.key", serviceName))
	}
	return path
}

// ServiceCertPath uses the configuration to generate a path at which the certificate for the given service name should be found.
func (conf *Config) ServiceCertPath(serviceName string) (certPath []byte) {
	path := conf.values[serviceName+".soa_cert"]
	if path == nil {
		path = []byte(fmt.Sprintf("/etc/GT_private/services/%s.crt", serviceName))
	}
	return path
}

// DiscoveryMulticastIP returns the configured discovery address, or the default one
// if there is no configured address (discovery.multicast_address)
func (conf *Config) DiscoveryMulticastIP() (ip net.IP) {
	rawAddr := conf.values["discovery.multicast_address"]
	if rawAddr != nil {
		return net.ParseIP(string(rawAddr))
	}

	return defaultGroupIP
}

// DiscoveryMulticastPort returns the configured discovery port, or the default one
// if there is no configured port (discovery.port)
func (conf *Config) DiscoveryMulticastPort() (port int) {
	portBytes := conf.values["discovery.port"]
	if portBytes != nil {
		port64, err := strconv.ParseInt(string(portBytes), 10, 0)
		if err != nil {
			Error.Printf("could not parse discovery.port `%s`. falling back to default", err)
			port = int(defaultGroupPort)
		} else {
			port = int(port64)
		}

		return
	}

	port = defaultGroupPort
	return
}

func (conf *Config) LocalDiscoveryMulticast() bool {
	_, ok := conf.values["discovery.local_multicast"]
	return ok
}

// Get returns the value of a given config option as a string, or false if it is not set.
func (conf *Config) Get(key string) (value string, ok bool) {
	valueBytes, ok := conf.values[key]
	value = string(valueBytes)
	return
}

// Set sets the given key to the given value in the configuration
func (conf *Config) Set(key string, value string) {
	valueBytes := []byte(value)
	conf.values[key] = valueBytes
}
