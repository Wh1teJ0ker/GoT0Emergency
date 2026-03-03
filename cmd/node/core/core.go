package core

import "time"

// Module interface for all agent modules
type Module interface {
	Name() string
	Run() (interface{}, error)
}

// AgentStatus defines the overall status structure
type AgentStatus struct {
	Hostname     string                 `json:"hostname"`
	OS           string                 `json:"os"`
	Platform     string                 `json:"platform"`
	Uptime       uint64                 `json:"uptime"`
	Modules      map[string]interface{} `json:"modules"`
	AgentVersion string                 `json:"agent_version"`
	Timestamp    int64                  `json:"timestamp"`
}

const Version = "0.0.2"

var registeredModules = make(map[string]Module)

func RegisterModule(m Module) {
	registeredModules[m.Name()] = m
}

func GetModules() map[string]Module {
	return registeredModules
}

// Collector logic
func Collect(status *AgentStatus) {
	status.AgentVersion = Version
	status.Timestamp = time.Now().Unix()
	status.Modules = make(map[string]interface{})

	for name, module := range registeredModules {
		data, err := module.Run()
		if err == nil {
			status.Modules[name] = data
		} else {
			status.Modules[name] = map[string]string{"error": err.Error()}
		}
	}
}
