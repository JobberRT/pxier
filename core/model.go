package core

type Proxy struct {
	Id        int    `gorm:"primaryKey; autoIncrement" json:"id"`
	Address   string `json:"address"`
	Provider  string `json:"provider"`
	CreatedAt int64  `json:"-"`
	UpdatedAt int64  `json:"-"`
	ErrTimes  int    `json:"-"`
	DialType  string `json:"dial_type"`
}

func (p *Proxy) TableName() string {
	return "proxy"
}
