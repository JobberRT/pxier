package core

const (
	httpSuccess = 0
	httpFailed  = -1
)

const (
	ProxyTypeHttp   = "http"
	ProxyTypeSocks5 = "socks5"
)

const (
	// ProviderTypeCPL https://github.com/clarketm/proxy-list
	ProviderTypeCPL = "cPL"
	// ProviderTypeTSXPL https://github.com/TheSpeedX/PROXY-List
	ProviderTypeTSXPL = "tsxPL"
	// ProviderTypeSTRPL https://github.com/ShiftyTR/Proxy-List
	ProviderTypeSTRPL = "strPL"
)

var (
	AllProviderType = []string{
		ProviderTypeCPL,
		ProviderTypeTSXPL,
		ProviderTypeSTRPL,
	}
	UserAvailableProviderType = []string{
		ProviderTypeCPL,
		ProviderTypeTSXPL,
		ProviderTypeSTRPL,
		"mix",
	}
)
