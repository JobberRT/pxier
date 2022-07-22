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
	ProviderTypeTSXPL = "TSX"
	// ProviderTypeSTRPL https://github.com/ShiftyTR/Proxy-List
	ProviderTypeSTRPL = "STR"
	// ProviderTypeIHuan https://ip.ihuan.me/ti.html
	ProviderTypeIHuan = "IHUAN"
	// ProviderTypeMix all type
	ProviderTypeMix = "MIX"
)

var (
	AllProviderType = []string{
		ProviderTypeCPL,
		ProviderTypeTSXPL,
		ProviderTypeSTRPL,
		ProviderTypeIHuan,
	}
	UserAvailableProviderType = []string{
		ProviderTypeCPL,
		ProviderTypeTSXPL,
		ProviderTypeSTRPL,
		ProviderTypeIHuan,
		ProviderTypeMix,
	}
)
