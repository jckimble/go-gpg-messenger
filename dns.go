package main

import (
	"strings"
	"net"
	"time"
	"log"
	"fmt"
	"regexp"
	"database/sql"
	"github.com/miekg/dns"
	"github.com/spf13/viper"
)
func serveDNSRequest(db *sql.DB, w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress=false
	viper.SetDefault("dns.records.A.default","127.0.0.1")
	var (
	domain string
	name string
	key string
	s []string
	)
	if r.Question[0].Qtype == dns.TypeSRV {
		re:=regexp.MustCompile("([a-zA-Z0-9-_]+)._tcp.(.*).")
		if re.MatchString(m.Question[0].Name) {
			s =re.FindStringSubmatch(m.Question[0].Name)
		}
	}else if r.Question[0].Qtype == dns.TypeTXT || r.Question[0].Qtype == dns.TypeA {
		re:=regexp.MustCompile("([a-zA-Z0-9-_]+).(.*).")
		if re.MatchString(m.Question[0].Name) {
			s =re.FindStringSubmatch(m.Question[0].Name)
		}
	}
	if len(s) == 3 {
		rows, err := db.Query("SELECT name,domain,key FROM users WHERE name=? AND domain=? AND password!=''", s[1], s[2])
		if err != nil {
			log.Printf("SQL Error: %s",err)
		}
		defer rows.Close()
		for rows.Next(){
			err:= rows.Scan(&name,&domain,&key)
			if err != nil {
				log.Printf("Error Pulling User: %s",err)
			}
		}
	}
	rr := &dns.A{
		Hdr: dns.RR_Header{Name: m.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
		A: net.ParseIP(viper.GetString("dns.records.A.default")).To4(),
	}
	srv := &dns.SRV{
		Hdr: dns.RR_Header{Name: m.Question[0].Name, Rrtype: dns.TypeSRV, Class: dns.ClassINET, Ttl: 0},
		Port: uint16(viper.GetInt("port")),
		Target: fmt.Sprintf("%s.",domain),
	}
	t := &dns.TXT{
		Hdr: dns.RR_Header{Name: m.Question[0].Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 0},
		Txt: []string{fmt.Sprintf("name=%s;key=%s",name,key)},
	}
	switch r.Question[0].Qtype {
	case dns.TypeSRV:
		if domain != "" {
			m.Answer = append(m.Answer, srv)
			m.Extra = append(m.Extra, rr)
			m.Extra = append(m.Extra, t)
		}
	case dns.TypeA:
			m.Answer = append(m.Answer, rr)
		if domain != "" {
			m.Extra = append(m.Extra, t)
			m.Extra = append(m.Extra, srv)
		}
	case dns.TypeTXT:
		if domain != "" {
			m.Answer = append(m.Answer, t)
			m.Extra = append(m.Extra, rr)
			m.Extra = append(m.Extra, srv)
		}
	default:
		log.Printf("Unsupported Record Type: %s",dns.TypeToString[r.Question[0].Qtype])
		fallthrough
	case dns.TypeAXFR, dns.TypeIXFR:
		c := make(chan *dns.Envelope)
		tr := new(dns.Transfer)
		defer close(c)
		if err := tr.Out(w,r,c); err != nil {
			log.Printf("Error: %s",err)
			return
		}
		soa, _ := dns.NewRR(fmt.Sprintf(`%s. 0 IN SOA %s. %s. 2009032802 21600 7200 604800 3600`,m.Question[0].Name,m.Question[0].Name,m.Question[0].Name))
		c <- &dns.Envelope{RR: []dns.RR{soa, t, rr, soa}}
		w.Hijack()
		return
	}
	if r.IsTsig() != nil {
		if w.TsigStatus() == nil {
			m.SetTsig(r.Extra[len(r.Extra)-1].(*dns.TSIG).Hdr.Name, dns.HmacMD5, 300, time.Now().Unix())
		}else{
			log.Printf("Status: %s", w.TsigStatus().Error())
		}
		if m.Question[0].Name == fmt.Sprintf("tc.%s.",viper.GetString("dns.domain")) {
			m.Truncated = true
			buf, _ := m.Pack()
			w.Write(buf[:len(buf)/2])
			return
		}
	}
	w.WriteMsg(m)
}
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
