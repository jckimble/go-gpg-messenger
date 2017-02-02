package main

import (
	"strings"
	"net"
)
func getAddr(addr string) (*Email){
	spt := strings.Split(addr,"@")
	var e=new(Email)
	e.Name=spt[0]
	e.Domain=spt[1]
	if !readDNSRecords(e) {
		return nil
	}
	return e
}
func readDNSRecords(email* Email) (bool){
	txts, err := net.LookupTXT(email.Name+"."+email.Domain)
	if err != nil {
		return false
	}
	if len(txts) == 0 {
		return false;
	}
	for _, txt := range txts {
		parts:=strings.Split(txt,";")
		for _, part := range parts {
			part=strings.Trim(part, " ")
			keyval:=strings.Split(part,"=")
			if keyval[0] == "key" {
				email.Key=keyval[1]
			}else if keyval[0] == "name" {
				email.Name=keyval[1]
			}
		}
	}
	_, addrs, err := net.LookupSRV(email.Name,"tcp",email.Domain)
	if err != nil {
		return false;
	}
	if len(addrs) == 0 {
		return false
	}
	for _, srv := range addrs {
		email.Domain=srv.Target
		email.Port=int(srv.Port)
	}
	return true
}
