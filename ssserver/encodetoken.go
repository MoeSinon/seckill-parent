package ssserver

import (
	"errors"
	"net"
	"redis-go/module/util"
)

func (shared *commonToken) GenerateShared() {
	var err error
	if shared.ClientKey, err = RandomKey(KEY_BYTES); err != nil {
		util.Zaplog().Errorf("error generating client key: %s", err)
	}
	if shared.ServerKey, err = RandomKey(KEY_BYTES); err != nil {
		util.Zaplog().Errorf("error generating server key: %s", err)
	}
}

type ConnectTokenPrivate struct {
	commonToken
	ClientId  uint64
	UserData  []byte
	mac       []byte
	TokenData *Buffer
}

func NewConnectTokenPrivate(clientId uint64, Timeoutstamp int32, serverAddrs []net.UDPAddr, userData []byte) *ConnectTokenPrivate {
	p := &ConnectTokenPrivate{}
	p.TokenData = NewBuffer(make([]byte, 0), CONNECT_TOKEN_PRIVATE_BYTES)
	p.Timeoutstamp = Timeoutstamp
	p.ClientId = clientId
	p.UserData = userData
	p.ServerAddrs = serverAddrs
	p.mac = make([]byte, MAC_BYTES)
	return p
}

func NewConnectTokenPrivateEncrypted(buffer []byte) *ConnectTokenPrivate {
	p := &ConnectTokenPrivate{}
	p.mac = make([]byte, MAC_BYTES)
	p.TokenData = NewBuffer(buffer)
	return p
}

func (p *ConnectTokenPrivate) Buffer() []byte {
	p.GenerateShared()
	return p.TokenData.Buf
}

func (p *ConnectTokenPrivate) Mac() []byte {
	p.GenerateShared()
	return p.mac
}

func (p *ConnectTokenPrivate) Read() error {
	p.GenerateShared()
	var err error
	if ClientId, err := p.TokenData.Readbytebytype(p.ClientId); err != nil {
		p.ClientId = ClientId.(uint64)
		return err
	}
	p.ReadToken(p.TokenData)

	if p.UserData, err = p.TokenData.GetBytes(USER_DATA_BYTES); err != nil {
		util.Zaplog().Errorf("ErrReadingUserData", err)
	}

	return nil
}

func (p *ConnectTokenPrivate) Write() ([]byte, error) {
	p.GenerateShared()
	p.TokenData.Writebytebytype(p.ClientId)

	if err := p.WriteToken(p.TokenData); err != nil {
		return nil, err
	}

	p.TokenData.BeByteafterNew(p.UserData, USER_DATA_BYTES)
	return p.TokenData.Buf, nil
}

func (p *ConnectTokenPrivate) Encrypt(aead AEAD, protocolId, expireTimestamp, sequence uint64, privateKey []byte) {
	p.GenerateShared()
	additionalData, nonce := buildTokenCryptData(protocolId, expireTimestamp, sequence)
	encBuf := p.TokenData.Buf[:CONNECT_TOKEN_PRIVATE_BYTES-MAC_BYTES]
	if err := Encrypt(aead, encBuf, additionalData, nonce, privateKey); err != nil {
		if len(p.TokenData.Buf) != CONNECT_TOKEN_PRIVATE_BYTES {
			util.Zaplog().Errorf("ErrInvalidTokenPrivateByteSize", err)
		}

	}

	copy(p.mac, p.TokenData.Buf[CONNECT_TOKEN_PRIVATE_BYTES-MAC_BYTES:])
}

func (p *ConnectTokenPrivate) Decrypt(aead AEAD, protocolId, expireTimestamp, sequence uint64, privateKey []byte) ([]byte, error) {
	var err error
	p.GenerateShared()
	if len(p.TokenData.Buf) != CONNECT_TOKEN_PRIVATE_BYTES {
		return nil, errors.New("ErrInvalidTokenPrivateByteSize")
	}

	copy(p.mac, p.TokenData.Buf[CONNECT_TOKEN_PRIVATE_BYTES-MAC_BYTES:])
	additionalData, nonce := buildTokenCryptData(protocolId, expireTimestamp, sequence)
	if p.TokenData.Buf, err = Decrypt(aead, p.TokenData.Buf, additionalData, nonce, privateKey); err != nil {
		return nil, err
	}
	p.TokenData.position = 0 // reset for reads
	return p.TokenData.Buf, nil
}

func buildTokenCryptData(protocolId, expireTimestamp, sequence uint64) ([]byte, []byte) {
	additionalData := NewBuffer(make([]byte, 0), VERSION_INFO_BYTES+8+8)
	additionalData.BeByteafterNew([]byte(VERSION_INFO))
	additionalData.Writebytebytype(protocolId)
	additionalData.Writebytebytype(expireTimestamp)

	nonce := NewBuffer(make([]byte, 0), Uint64+Uint32)
	nonce.Writebytebytype(0)
	nonce.Writebytebytype(sequence)
	return additionalData.Buf, nonce.Buf
}
