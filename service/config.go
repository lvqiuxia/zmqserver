package service

type AppConfig struct {
	Name     string
	Type     string
	State    ComponentState
	Actor    bool
	Children []AppConfig
}


