package ssserver

import (
	"fmt"
	"net"
	"redis-go/module/util"
	"sync/atomic"
	"time"
)

const TIMEOUT_SECONDS = 5
const MAX_SERVER_PACKETS = 64

type Server struct {
	serverConn       *NetcodeConn
	serverAddr       *net.UDPAddr
	shutdownCh       chan struct{}
	serverTime       float64
	running          bool
	maxClients       int
	connectedClients int
	timeout          float64

	clientManager  *ClientManager
	globalSequence uint64

	ignoreRequests  bool
	ignoreResponses bool
	allowedPackets  []byte
	protocolId      uint64

	privateKey   []byte
	challengeKey []byte

	challengeSequence uint64

	recvBytes int
	packetCh  chan *NetcodeData
}

func NewServer(serverAddress *net.UDPAddr, privateKey []byte, protocolId uint64, maxClients int) *Server {
	s := &Server{}
	s.serverAddr = serverAddress
	s.protocolId = protocolId
	s.privateKey = privateKey
	s.maxClients = maxClients

	s.globalSequence = uint64(1) << 63
	s.timeout = float64(TIMEOUT_SECONDS)
	s.clientManager = NewClientManager(s.timeout, maxClients)
	s.packetCh = make(chan *NetcodeData, s.maxClients*MAX_SERVER_PACKETS*2)
	s.shutdownCh = make(chan struct{})

	// set allowed packets for this server
	s.allowedPackets = make([]byte, ConnectionNumPackets)
	s.allowedPackets[ConnectionRequest] = 1
	s.allowedPackets[ConnectionResponse] = 1
	s.allowedPackets[ConnectionKeepAlive] = 1
	s.allowedPackets[ConnectionPayload] = 1
	s.allowedPackets[ConnectionDisconnect] = 1
	return s
}

func (s *Server) SetAllowedPackets(allowedPackets []byte) {
	s.allowedPackets = allowedPackets
}

func (s *Server) SetTimeout(duration time.Duration) {
	s.timeout = duration.Seconds()
	s.clientManager.setTimeout(s.timeout)
}

func (s *Server) SetIgnoreRequests(val bool) {
	s.ignoreRequests = val
}

func (s *Server) SetIgnoreResponses(val bool) {
	s.ignoreResponses = val
}

func (s *Server) incChallengeSequence() uint64 {
	val := atomic.AddUint64(&s.challengeSequence, 1)
	return val - 1
}

func (s *Server) incGlobalSequence() uint64 {
	val := atomic.AddUint64(&s.globalSequence, 1)
	return val - 1
}

func (s *Server) Init() error {
	var err error

	s.challengeKey, err = GenerateKey()
	if err != nil {
		return err
	}
	s.serverConn = NewNetcodeConn()
	s.serverConn.SetReadBuffer(SOCKET_RCVBUF_SIZE * s.maxClients)
	s.serverConn.SetWriteBuffer(SOCKET_SNDBUF_SIZE * s.maxClients)
	s.serverConn.SetRecvHandler(s.handleNetcodeData)
	return nil
}

func (s *Server) Listen() error {
	s.running = true

	addr := &net.UDPAddr{
		Port: s.serverAddr.Port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	if err := s.serverConn.Listen(addr); err != nil {
		return err
	}
	return nil
}

func (s *Server) SendPayloads(payloadData []byte) {
	if !s.running {
		return
	}
	s.clientManager.sendPayloads(payloadData, s.serverTime)
}

func (s *Server) SendPayloadToClient(clientId uint64, payloadData []byte) error {
	clientIndex, err := s.GetClientIndexByClientId(clientId)
	if err != nil {
		return err
	}

	s.clientManager.sendPayloadToInstance(clientIndex, payloadData, s.serverTime)
	return nil
}

func (s *Server) Update(time float64) error {
	if !s.running {
		return ErrServerShutdown
	}

	s.serverTime = time

	for {
		select {
		case recv := <-s.packetCh:
			s.OnPacketData(recv.data, recv.from)
		default:
			goto DONE
		}
	}
DONE:
	s.clientManager.SendKeepAlives(s.serverTime)
	s.clientManager.CheckTimeouts(s.serverTime)
	return nil
}

func (s *Server) DisconnectClient(clientId uint64, sendDisconnect bool) error {
	clientIndex, err := s.GetClientIndexByClientId(clientId)
	if err != nil {
		return err
	}

	s.clientManager.DisconnectClient(clientIndex, sendDisconnect, s.serverTime)
	return nil
}

func (s *Server) GetClientIndexByClientId(clientId uint64) (int, error) {
	if !s.running {
		return -1, ErrServerNotRunning
	}

	clientIndex := s.clientManager.FindClientIndexById(clientId)
	if clientIndex == -1 {
		return -1, fmt.Errorf("unknown client id %d", clientId)
	}
	return clientIndex, nil
}

func (s *Server) handleNetcodeData(packetData *NetcodeData) {
	s.packetCh <- packetData
}

func (s *Server) OnPacketData(packetData []byte, addr *net.UDPAddr) {
	var readPacketKey []byte
	var replayProtection *ReplayProtection

	if !s.running {
		return
	}

	size := len(packetData)

	encryptionIndex := -1
	clientIndex := s.clientManager.FindClientIndexByAddress(addr)
	if clientIndex != -1 {
		encryptionIndex = s.clientManager.FindEncryptionIndexByClientIndex(clientIndex)
	} else {
		encryptionIndex = s.clientManager.FindEncryptionEntryIndex(addr, s.serverTime)
	}
	readPacketKey = s.clientManager.GetEncryptionEntryRecvKey(encryptionIndex)

	timestamp := uint64(time.Now().Unix())

	packet := NewPacket(packetData)
	if clientIndex != -1 {
		client := s.clientManager.instances[clientIndex]
		replayProtection = client.replayProtection
	}

	if err := packet.Read(packetData, size, s.protocolId, timestamp, readPacketKey, s.privateKey, s.allowedPackets, replayProtection); err != nil {
		util.Zaplog().Debugf("error reading packet: %s from %s\n", err, addr)
		return
	}

	s.processPacket(clientIndex, encryptionIndex, packet, addr)
}

func (s *Server) processPacket(clientIndex, encryptionIndex int, packet Packet, addr *net.UDPAddr) {
	switch packet.GetType() {
	case ConnectionRequest:
		if s.ignoreRequests {
			return
		}
		util.Zaplog().Debugf("server received connection request from %s\n", addr.String())
		s.processConnectionRequest(packet, addr)
	case ConnectionResponse:
		if s.ignoreResponses {
			return
		}
		util.Zaplog().Debugf("server received connection response from %s\n", addr.String())
		s.processConnectionResponse(clientIndex, encryptionIndex, packet, addr)
	case ConnectionKeepAlive:
		if clientIndex == -1 {
			return
		}
		client := s.clientManager.instances[clientIndex]
		client.lastRecvTime = s.serverTime

		if !client.confirmed {
			client.confirmed = true
			util.Zaplog().Debugf("server confirmed connection to client %d:%s\n", client.clientId, client.address.String())
		}
	case ConnectionPayload:
		if clientIndex == -1 {
			return
		}
		client := s.clientManager.instances[clientIndex]
		client.lastRecvTime = s.serverTime

		if !client.confirmed {
			client.confirmed = true
			util.Zaplog().Debugf("server confirmed connection to client %d:%s\n", client.clientId, client.address.String())
		}

		client.packetQueue.Push(packet)
	case ConnectionDisconnect:
		if clientIndex == -1 {
			return
		}
		client := s.clientManager.instances[clientIndex]
		util.Zaplog().Debugf("server received disconnect packet from client %d:%s\n", client.clientId, client.address.String())
		s.clientManager.disconnectClient(client, false, s.serverTime)
	}
}

func (s *Server) processConnectionRequest(packet Packet, addr *net.UDPAddr) {
	requestPacket, ok := packet.(*RequestPacket)
	if !ok {
		return
	}

	if len(requestPacket.Token.ServerAddrs) == 0 {
		util.Zaplog().Debugf("server ignored connection request. server address not in connect token whitelist\n")
		return
	}

	addrFound := false
	for _, tokenAddr := range requestPacket.Token.ServerAddrs {
		if addressEqual(s.serverAddr, &tokenAddr) {
			addrFound = true
			break
		}
	}

	if !addrFound {
		util.Zaplog().Debugf("server ignored connection request. server address not in connect token whitelist\n")
		return
	}

	clientIndex := s.clientManager.FindClientIndexByAddress(addr)
	if clientIndex != -1 {
		util.Zaplog().Debugf("server ignored connection request. a client with this address is already connected\n")
		return
	}

	clientIndex = s.clientManager.FindClientIndexById(requestPacket.Token.ClientId)
	if clientIndex != -1 {
		util.Zaplog().Debugf("server ignored connection request. a client with this id has already been used\n")
		return
	}

	if !s.clientManager.FindOrAddTokenEntry(requestPacket.Token.Mac(), addr, s.serverTime) {
		util.Zaplog().Debugf("server ignored connection request. connect token has already been used\n")
		return
	}

	if s.clientManager.ConnectedClientCount() == s.maxClients {
		util.Zaplog().Debugf("server denied connection request. server is full\n")
		s.sendDeniedPacket(requestPacket.Token.ServerKey, addr)
		return
	}

	if !s.clientManager.AddEncryptionMapping(requestPacket.Token, addr, s.serverTime, s.serverTime+s.timeout) {
		util.Zaplog().Debugf("server ignored connection request. failed to add encryption mapping\n")
		return
	}

	s.sendChallengePacket(requestPacket, addr)
}

func (s *Server) sendChallengePacket(requestPacket *RequestPacket, addr *net.UDPAddr) {
	var bytesWritten int
	var err error

	challenge := NewChallengeToken(requestPacket.Token.ClientId)
	challengeBuf := challenge.Write(requestPacket.Token.UserData)
	challengeSequence := s.incChallengeSequence()

	if err := EncryptChallengeToken(challengeBuf, challengeSequence, s.challengeKey); err != nil {
		util.Zaplog().Debugf("server ignored connection request. failed to encrypt challenge token\n")
		return
	}

	challengePacket := &ChallengePacket{}
	challengePacket.ChallengeTokenData = challengeBuf
	challengePacket.ChallengeTokenSequence = challengeSequence

	buffer := make([]byte, MAX_PACKET_BYTES)
	if bytesWritten, err = challengePacket.Write(buffer, s.protocolId, s.incGlobalSequence(), requestPacket.Token.ServerKey); err != nil {
		util.Zaplog().Debugf("server error while writing challenge packet\n")
		return
	}

	s.sendGlobalPacket(buffer[:bytesWritten], addr)
}

func (s *Server) sendGlobalPacket(packetBuffer []byte, addr *net.UDPAddr) {
	if _, err := s.serverConn.WriteTo(packetBuffer, addr); err != nil {
		util.Zaplog().Debugf("error sending packet to %s\n", addr.String())
	}
}

func (s *Server) processConnectionResponse(clientIndex, encryptionIndex int, packet Packet, addr *net.UDPAddr) {
	var err error
	var tokenBuffer []byte
	var challengeToken *ChallengeToken

	responsePacket, ok := packet.(*ResponsePacket)
	if !ok {
		return
	}

	if tokenBuffer, err = DecryptChallengeToken(responsePacket.ChallengeTokenData, responsePacket.ChallengeTokenSequence, s.challengeKey); err != nil {
		util.Zaplog().Debugf("failed to decrypt challenge token: %s\n", err)
		return
	}

	if challengeToken, err = ReadChallengeToken(tokenBuffer); err != nil {
		util.Zaplog().Debugf("failed to read challenge token: %s\n", err)
		return
	}

	sendKey := s.clientManager.GetEncryptionEntrySendKey(encryptionIndex)
	if sendKey == nil {
		util.Zaplog().Debugf("server ignored connection response. no packet send key\n")
		return
	}

	if s.clientManager.FindClientIndexByAddress(addr) != -1 {
		util.Zaplog().Debugf("server ignored connection response. a client with this address is already connected")
		return
	}

	if s.clientManager.FindClientIndexById(challengeToken.ClientId) != -1 {
		util.Zaplog().Debugf("server ignored connection response. a client with this id is already connected")
		return
	}

	if s.clientManager.ConnectedClientCount() == s.maxClients {
		util.Zaplog().Debugf("server denied connection response. server is full\n")
		s.sendDeniedPacket(sendKey, addr)
		return
	}

	s.connectClient(encryptionIndex, challengeToken, addr)
	return

}

func (s *Server) sendDeniedPacket(sendKey []byte, addr *net.UDPAddr) {
	var bytesWritten int
	var err error

	deniedPacket := &DeniedPacket{}
	packetBuffer := make([]byte, MAX_PACKET_BYTES)
	if bytesWritten, err = deniedPacket.Write(packetBuffer, s.protocolId, s.incGlobalSequence(), sendKey); err != nil {
		util.Zaplog().Debugf("error creating denied packet: %s\n", err)
		return
	}

	s.sendGlobalPacket(packetBuffer[:bytesWritten], addr)
}

func (s *Server) connectClient(encryptionIndex int, challengeToken *ChallengeToken, addr *net.UDPAddr) {
	if s.clientManager.ConnectedClientCount() > s.maxClients {
		util.Zaplog().Debugf("maxium number of clients reached")
		return
	}

	s.clientManager.SetEncryptionEntryExpiration(encryptionIndex, -1)
	client := s.clientManager.ConnectClient(addr, challengeToken)
	if client == nil {
		return
	}
	client.serverConn = s.serverConn
	client.encryptionIndex = encryptionIndex
	client.protocolId = s.protocolId
	client.lastSendTime = s.serverTime
	client.lastRecvTime = s.serverTime
	util.Zaplog().Debugf("server accepted client %d from %s in slot: %d\n", client.clientId, addr.String(), client.clientIndex)
	s.sendKeepAlive(client)
}

func (s *Server) sendKeepAlive(client *ClientInstance) {
	clientIndex := client.clientIndex
	packet := &KeepAlivePacket{}
	packet.ClientIndex = uint32(clientIndex)
	packet.MaxClients = uint32(s.maxClients)

	if !s.clientManager.TouchEncryptionEntry(client.encryptionIndex, client.address, s.serverTime) {
		util.Zaplog().Debugf("error: encryption mapping is out of date for client %d encIndex: %d addr: %s\n", clientIndex, client.encryptionIndex, client.address.String())
		return
	}

	writePacketKey := s.clientManager.GetEncryptionEntrySendKey(client.encryptionIndex)
	if writePacketKey == nil {
		util.Zaplog().Debugf("error: unable to retrieve encryption key for client: %d\n", clientIndex)
		return
	}

	if err := client.SendPacket(packet, writePacketKey, s.serverTime); err != nil {
		util.Zaplog().Debugf("%s\n", err)
	}
}

func (s *Server) GetConnectedClientIds() []uint64 {
	return s.clientManager.ConnectedClients()
}

func (s *Server) MaxClients() int {
	return s.maxClients
}

func (s *Server) HasClients() int {
	return s.clientManager.ConnectedClientCount()
}

func (s *Server) GetClientUserData(clientId uint64) ([]byte, error) {
	i, err := s.GetClientIndexByClientId(clientId)

	if err != nil {
		return nil, err
	}

	return s.clientManager.instances[i].userData, nil
}

func (s *Server) Stop() error {
	if !s.running {
		return nil
	}
	s.clientManager.disconnectClients(s.serverTime)

	s.running = false
	s.maxClients = 0
	s.globalSequence = 0
	s.challengeSequence = 0
	s.challengeKey = make([]byte, KEY_BYTES)
	s.clientManager.resetCryptoEntries()
	s.clientManager.resetTokenEntries()
	close(s.shutdownCh)
	s.running = false
	s.serverConn.Close()

	return nil
}

func (s *Server) RecvPayload(clientIndex int) ([]byte, uint64) {
	packet := s.clientManager.instances[clientIndex].packetQueue.Pop()
	if packet == nil {
		return []byte{}, 0
	}
	p, ok := packet.(*PayloadPacket)
	if !ok {
		util.Zaplog().Debugf("not a payload packet")
		return []byte{}, 0
	}
	return p.PayloadData, p.sequence
}

func addressEqual(addr1, addr2 *net.UDPAddr) bool {
	if addr1 == nil || addr2 == nil {
		return false
	}
	return addr1.IP.Equal(addr2.IP) && addr1.Port == addr2.Port
}
