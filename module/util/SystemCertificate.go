package util

/*****
 * @Author: http://www.itheima.com
 * @Project: seckill
 * @Description: com.seckill.util.Audience
 ****/
type Audience struct {
	ClientId      string
	Base64Secret  string
	Name          string
	ExpiresSecond int
}

type Audienceutl interface {
	getClientId() string

	setClientId() string

	getBase64Secret() string

	setBase64Secret(Base64Secret string)

	getName() string

	setName(Name string)

	getExpiresSecond()

	setExpiresSecond(ExpiresSecond int)
}

func (A *Audience) getClientId() string {
	return A.ClientId
}

func (A *Audience) setClientId(ClientId string) {
	A.ClientId = ClientId
}

func (A *Audience) getBase64Secret() string {
	return A.Base64Secret
}

func (A *Audience) setBase64Secret(Base64Secret string) {
	A.Base64Secret = Base64Secret
}

func (A *Audience) getName() string {
	return A.Name
}

func (A *Audience) setName(Name string) {
	A.Name = Name
}

func (A *Audience) getExpiresSecond() int {
	return A.ExpiresSecond
}

func (A *Audience) setExpiresSecond(ExpiresSecond int) {
	A.ExpiresSecond = ExpiresSecond
}
