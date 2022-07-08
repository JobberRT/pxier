package core

type Proxy struct {
	Id        int `gorm:"primaryKey; autoIncrement"`
	Address   string
	Provider  string
	CreatedAt int64
	UpdatedAt int64
	ErrTimes  int
	DialType  string
}

func (p *Proxy) TableName() string {
	return "proxy"
}
