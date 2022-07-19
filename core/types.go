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
	ProviderTypeCPL = "CPL"
	// ProviderTypeTSXPL https://github.com/TheSpeedX/PROXY-List
	ProviderTypeTSXPL = "TSXPL"
	// ProviderTypeSTRPL https://github.com/ShiftyTR/Proxy-List
	ProviderTypeSTRPL = "STRPL"
	// ProviderTypeMix all type
	ProviderTypeMix = "MIX"
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
		ProviderTypeMix,
	}
)
