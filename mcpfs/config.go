package mcpfs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	std "github.com/jlrickert/go-std/pkg"
	"gopkg.in/yaml.v3"
)

var AppName = "mcpfs"

// Configuration for filesystem access controls for an MCP server.
// The configuration is YAML-backed and supports environment-variable
// substitution in all string fields and in string slices.
//
// Typical usage:
//
//	cfg, err := ReadAndParseConfig("path/to/config.yaml")
//	if err != nil { ... }
//	allowed := cfg.IsAllowed(PermWrite, "/var/data/foo")
//
// Important: by default, if a rule omits explicit permissions it will be
// treated as read-only (no write) to align with the typical "no write by default"
// preference.
const (
	ConfigVersionV1       = "1.0"
	DefaultConfigFilename = "config.yaml"
	defaultAllowSubpaths  = true
)

// Permission is a bitmask for path operations.
type Permission uint8

const (
	PermNone Permission = 0
	PermRead Permission = 1 << iota
	PermWrite
	PermExec
)

func (p Permission) String() string {
	var parts []string
	if p&PermRead != 0 {
		parts = append(parts, "read")
	}
	if p&PermWrite != 0 {
		parts = append(parts, "write")
	}
	if p&PermExec != 0 {
		parts = append(parts, "exec")
	}
	if len(parts) == 0 {
		return "none"
	}
	return strings.Join(parts, "|")
}

// PathRule describes one allowed path and which permissions are granted.
// YAML schema:
//
//   - path: "/var/www"              # path on disk; may contain env vars like ${HOME}
//     perms: ["read", "exec"]       # permitted operations; default: ["read"]
//     users: ["alice", "bob"]       # optional list of users allowed
//     roles: ["admin"]              # optional list of roles allowed
//     allow_subpaths: true          # whether subpaths are covered (default true)
//     description: "web content dir"
type PathRule struct {
	Path          string   `yaml:"path" json:"path"`
	Perms         []string `yaml:"perms,omitempty" json:"perms,omitempty"`
	AllowSubpaths *bool    `yaml:"allow_subpaths,omitempty" json:"allow_subpaths,omitempty"`
	Description   string   `yaml:"description,omitempty" json:"description,omitempty"`

	// runtime fields (not marshaled)
	parsedPerms Permission `yaml:"-" json:"-"`
	cleanPath   string     `yaml:"-" json:"-"`
}

// Config is the top-level configuration.
type Config struct {
	Version  string     `yaml:"version,omitempty" json:"version,omitempty"`
	Paths    []PathRule `yaml:"paths" json:"paths"`
	LogLevel string     `yaml:"log_level" json:"log_level"`
	LogPath  string     `yaml:"log_path" json:"log_path"`
}

// ReadConfigData reads the file at configPath and returns its contents.
// Returns an error if the file does not exist or cannot be read.
func ReadConfigData(configPath string) ([]byte, error) {
	info, err := os.Stat(configPath)
	if err != nil {
		return nil, fmt.Errorf("stat config file: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("config path %q is a directory", configPath)
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	return data, nil
}

// ReadDefaultConfigData builds the default config path from the user's config directory
// and the default filename and reads it.
func ReadDefaultConfigData(env std.Env) ([]byte, error) {
	dir, err := std.UserConfigPath(AppName, env)
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(dir, DefaultConfigFilename)
	return ReadConfigData(configPath)
}

// ParseConfigData parses YAML data into FSConfig. It expands environment variables,
// normalizes paths, and parses permission strings. If a rule omits perms, default to read-only.
//
// Note: This implementation does not perform comment-preserving roundtrips; it focuses on
// configuration semantics suitable for enforcement.
func ParseConfigData(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("%w: yaml unmarshal: %v", ErrParse, err)
	}

	// Default version if empty
	if cfg.Version == "" {
		cfg.Version = ConfigVersionV1 // assume legacy if absent
	}

	// // Migrate v1 -> v2 if needed (simple migration rule: just set version to v2).
	// if cfg.Version == ConfigVersionV1 {
	// 	// In v1 we might have slightly different semantics; for now, migrate by
	// 	// stamping the version and leaving contents as-is.
	// 	cfg.Version = ConfigVersionV2
	// }

	// Expand environment variables throughout the struct
	if err := expandEnvInValue(reflect.ValueOf(&cfg)); err != nil {
		return nil, fmt.Errorf("expand env: %w", err)
	}

	// Normalize and parse each path rule
	for i := range cfg.Paths {
		r := &cfg.Paths[i]
		if r.AllowSubpaths == nil {
			// default to allowing subpaths; explicit false must be set to disable
			v := defaultAllowSubpaths
			r.AllowSubpaths = &v
		}
		// Clean and make absolute if possible (do not require existence)
		clean := filepath.Clean(r.Path)
		if !filepath.IsAbs(clean) {
			// attempt to make it absolute relative to current working dir
			abs, err := filepath.Abs(clean)
			if err == nil {
				clean = abs
			}
		}
		r.cleanPath = clean

		// Parse perms
		if len(r.Perms) == 0 {
			// default to read-only
			r.parsedPerms = PermRead
		} else {
			mask, err := parsePerms(r.Perms)
			if err != nil {
				return nil, fmt.Errorf("invalid perms for path %q: %w", r.Path, err)
			}
			r.parsedPerms = mask
		}
	}

	return &cfg, nil
}

// Parse a slice of permission strings like ["read", "write"] into a Permission mask.
func parsePerms(perms []string) (Permission, error) {
	var mask Permission
	for _, p := range perms {
		switch strings.ToLower(strings.TrimSpace(p)) {
		case "read", "r":
			mask |= PermRead
		case "write", "w":
			mask |= PermWrite
		case "exec", "execute", "x":
			mask |= PermExec
		default:
			return 0, fmt.Errorf("unknown permission %q", p)
		}
	}
	return mask, nil
}

// IsAllowed returns true if the given principal (user with roles) is allowed to perform
// op on targetPath according to the configured rules. The first matching rule grants access;
// more sophisticated merging or deny-list semantics are intentionally omitted for simplicity.
//
// Matching rules:
//   - The rule path is matched exactly, or if allow_subpaths=true then any path under
//     the rule path is considered a match.
//   - A rule without Users/Roles applies to all principals.
func (c *Config) IsAllowed(op Permission, targetPath string) bool {
	if c == nil {
		return false
	}
	cleanTarget := filepath.Clean(targetPath)
	if !filepath.IsAbs(cleanTarget) {
		abs, err := filepath.Abs(cleanTarget)
		if err == nil {
			cleanTarget = abs
		}
	}

	for i := range c.Paths {
		r := &c.Paths[i]
		// permission check
		if r.parsedPerms&op == 0 {
			// this rule doesn't grant the requested operation
			continue
		}

		// path match
		if r.cleanPath == cleanTarget {
			// exact match passes; fall through to identity checks below
		} else if *r.AllowSubpaths {
			rel, err := filepath.Rel(r.cleanPath, cleanTarget)
			if err != nil {
				// cannot compute relation; skip this rule
				continue
			}
			if rel == "." {
				// same path; ok
			} else if strings.HasPrefix(rel, "..") {
				// target is outside of rule path; no match
				continue
			}
			// else target is a subpath under rule path -> match
		} else {
			// not exact and subpaths aren't allowed
			continue
		}

		// passed all checks -> allowed
		return true
	}

	return false
}

// ExpandEnv walks the config and expands environment variables in all string fields
// and in elements of []string slices. It mutates the config in place.
func (c *Config) ExpandEnv() error {
	if c == nil {
		return nil
	}
	return expandEnvInValue(reflect.ValueOf(c))
}

// ToYAML serializes the configuration to a YAML string.
func (c *Config) ToYAML() (string, error) {
	if c == nil {
		return "", errors.New("nil config")
	}
	out, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ToJSON serializes the configuration to a YAML string.
func (c *Config) ToJSON() (string, error) {
	if c == nil {
		return "", errors.New("nil config")
	}
	out, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// ReadAndParseConfig is a convenience that reads a config file and parses it.
func ReadAndParseConfig(path string) (*Config, error) {
	data, err := ReadConfigData(path)
	if err != nil {
		return nil, err
	}
	return ParseConfigData(data)
}

// Utility helpers

// expandEnvInValue recursively walks a value and applies os.ExpandEnv
// to all string fields and to all elements of []string slices/maps[string]string.
func expandEnvInValue(v reflect.Value) error {
	if !v.IsValid() {
		return nil
	}
	// If pointer, get the element (allocate if nil)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			// allocate a new zero value for the element so we can set fields inside it
			v.Set(reflect.New(v.Type().Elem()))
		}
		return expandEnvInValue(v.Elem())
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			// only settable/exported fields will be modified; skip unexported
			f := v.Field(i)
			// skip unexported fields
			if !v.Type().Field(i).IsExported() {
				continue
			}
			if err := expandEnvInValue(f); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		// special-case []byte: do not treat as []string
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return nil
		}
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.String {
				expanded := os.ExpandEnv(elem.String())
				if elem.CanSet() {
					elem.SetString(expanded)
				} else {
					// can't set (maybe it's from an unexported field); skip
				}
			} else {
				if err := expandEnvInValue(elem); err != nil {
					return err
				}
			}
		}
	case reflect.Map:
		// only handle map[string]string conveniently
		if v.Type().Key().Kind() == reflect.String && v.Type().Elem().Kind() == reflect.String {
			for _, key := range v.MapKeys() {
				orig := v.MapIndex(key).String()
				expanded := os.ExpandEnv(orig)
				v.SetMapIndex(key, reflect.ValueOf(expanded))
			}
		} else {
			// for other map types, recursively walk values
			for _, key := range v.MapKeys() {
				val := v.MapIndex(key)
				if err := expandEnvInValue(val); err != nil {
					return err
				}
			}
		}
	case reflect.String:
		if v.CanSet() {
			v.SetString(os.ExpandEnv(v.String()))
		}
	}
	return nil
}
