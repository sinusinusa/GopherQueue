package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Addr         string        `yaml:"addr" json:"addr"`
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`
}

type Scheduler struct {
	Timezone       string        `yaml:"timezone" json:"timezone"`
	Workers        int           `yaml:"workers" json:"workers"`
	DefaultTimeout time.Duration `yaml:"default_timeout" json:"default_timeout"`
}
type Task struct {
	ID      string        `yaml:"id" json:"id"`
	Cron    string        `yaml:"cron" json:"cron"`
	Command string        `yaml:"command" json:"command"`
	Args    []string      `yaml:"args" json:"args"`
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
	Retries int           `yaml:"retries" json:"retries"`
	Enabled bool          `yaml:"enabled" json:"enabled"`
}

type Config struct {
	Server    Server            `yaml:"server" json:"server"`
	Scheduler Scheduler         `yaml:"scheduler" json:"scheduler"`
	Tasks     []Task            `yaml:"tasks" json:"tasks"`
	Secrets   map[string]string `yaml:"secrets" json:"-"` // not fot external
}

func Load(path string) (*Config, error) {
	if path == "" {
		return nil, errors.New("config path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	switch strings.ToLower(filepath.Ext(path)) {
	case ".yaml", ".yml", "":
		dec := yaml.NewDecoder(strings.NewReader(string(data)))
		dec.KnownFields(true)
		if err := dec.Decode(&cfg); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse json: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format: %s", path)
	}

	applyDefaults(&cfg)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func applyDefaults(c *Config) {
	if c.Server.Addr == "" {
		c.Server.Addr = ":8080"
	}
	if c.Scheduler.Workers <= 0 {
		c.Scheduler.Workers = 1
	}
	if c.Scheduler.DefaultTimeout == 0 {
		c.Scheduler.DefaultTimeout = 30 * time.Second
	}
	for i := range c.Tasks {
		if c.Tasks[i].Timeout == 0 {
			c.Tasks[i].Timeout = c.Scheduler.DefaultTimeout
		}
	}
}

func (c *Config) Validate() error {
	ids := make(map[string]struct{})
	for _, t := range c.Tasks {
		if t.ID == "" {
			return errors.New("task.id is required")
		}
		if _, ok := ids[t.ID]; ok {
			return fmt.Errorf("duplicate task.id: %s", t.ID)
		}
		ids[t.ID] = struct{}{}

		if t.Cron == "" {
			return fmt.Errorf("task %s: cron is required", t.ID)
		}
		if t.Command == "" {
			return fmt.Errorf("task %s: command is required", t.ID)
		}
	}
	return nil
}

// Public copy without secrets
func (c *Config) PublicCopy() Config {
	cp := *c
	cp.Secrets = nil
	return cp
}
