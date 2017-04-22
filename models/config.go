package models

type TomlConfig struct {
	Owner DomainOwner
	Scopes []string
	Networks map[string][]Network
}

type DomainOwner struct {
	Domain string
	Organization string
}

type Network struct {
	Type string
	Ip []string
}

func (config *TomlConfig) GetAllIps() []string {
	var allIp []string
	for _, network := range config.Networks {
		for _, n := range network {
			allIp = append(allIp, n.Ip...)
		}
	}
	return allIp
}