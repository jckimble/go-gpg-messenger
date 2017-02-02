package main

import (
	"strings"
	"fmt"
	"encoding/json"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)
type UserAuth struct {
	Username string
	Password string
}
type Email struct {
	Name string
	Domain string
	Key string
	Port int
}
type Message struct {
	From* Email
	To* Email
	Message string
	Time int
}
func parseMessage(str string) (*Message) {
	var msg = new(Message)
	err := json.Unmarshal([]byte(str),msg)
	if err != nil {
		return nil
	}
	return msg
}
func checkMessage(msg* Message) (bool) {
	var p packet.Packet
	var err error
	var key string
	var str=strings.NewReader(msg.Message)
	block, err := armor.Decode(str)
	if err != nil {
		return false
	}
	packets := packet.NewReader(block.Body)
	ParsePackets:
	for {
		p, err = packets.Next()
		if err != nil{
			fmt.Printf(err.Error())
			break ParsePackets
		}
		switch p := p.(type) {
		case *packet.Compressed, *packet.LiteralData, *packet.OnePassSignature:
			break ParsePackets
		case *packet.SymmetricallyEncrypted:
			break ParsePackets
		case *packet.EncryptedKey:
			key = fmt.Sprintf("%016X",p.KeyId)
			if msg.To.Key != key && msg.From.Key != key {
				return false
			}
		}
	}
	return true
}
