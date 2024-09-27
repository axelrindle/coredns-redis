package redis

import (
	"encoding/json"
	"fmt"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

type RedisHandler struct {
	redis *Redis

	next plugin.Handler
}

// ServeDNS implements the plugin.Handler interface.
func (h *RedisHandler) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	qname := state.Name()
	qtype := state.Type()

	log.Debugf("Handling %s %s", qtype, qname)

	zones, err := h.redis.LoadZones()
	if err != nil {
		log.Errorf("Failed to load zone data: %v", err)
		return plugin.NextOrFailure(qname, h.next, ctx, w, r)
	}

	zone := plugin.Zones(zones).Matches(qname)
	log.Debugf("Zone: %s", zone)
	if zone == "" {
		return plugin.NextOrFailure(qname, h.next, ctx, w, r)
	}

	z, err := h.redis.LoadZone(zone)
	if err != nil {
		return h.errorResponse(state, zone, dns.RcodeServerFailure, err)
	} else {
		log.Debugf("Zone Resolved: %s", z.Name)
	}

	if qtype == "AXFR" {
		records := h.redis.AXFR(z)

		ch := make(chan *dns.Envelope)
		tr := new(dns.Transfer)
		tr.TsigSecret = nil

		go func(ch chan *dns.Envelope) {
			j, l := 0, 0

			for i, r := range records {
				l += dns.Len(r)
				if l > transferLength {
					ch <- &dns.Envelope{RR: records[j:i]}
					l = 0
					j = i
				}
			}
			if j < len(records) {
				ch <- &dns.Envelope{RR: records[j:]}
			}
			close(ch)
		}(ch)

		err := tr.Out(w, r, ch)
		if err != nil {
			fmt.Println(err)
		}
		w.Hijack()
		return dns.RcodeSuccess, nil
	}

	location := h.redis.Locate(qname, z)
	if len(location) == 0 { // empty, no results
		return h.errorResponse(state, zone, dns.RcodeNameError, nil)
	}

	answers := make([]dns.RR, 0, 10)
	extras := make([]dns.RR, 0, 10)

	record := h.redis.get(location, z)

	switch qtype {
	case "A":
		answers, extras = h.redis.A(qname, z, record)
	case "AAAA":
		answers, extras = h.redis.AAAA(qname, z, record)
	case "CNAME":
		answers, extras = h.redis.CNAME(qname, z, record)
	case "TXT":
		answers, extras = h.redis.TXT(qname, z, record)
	case "NS":
		answers, extras = h.redis.NS(qname, z, record)
	case "MX":
		answers, extras = h.redis.MX(qname, z, record)
	case "SRV":
		answers, extras = h.redis.SRV(qname, z, record)
	case "SOA":
		answers, extras = h.redis.SOA(qname, z, record)
	case "CAA":
		answers, extras = h.redis.CAA(qname, z, record)

	default:
		return h.errorResponse(state, zone, dns.RcodeNotImplemented, nil)
	}

	prettyAnswers, _ := json.Marshal(answers)
	prettyExtras, _ := json.Marshal(extras)
	log.Debugf("Answers: %s", prettyAnswers)
	log.Debugf("Extras: %s", prettyExtras)

	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	m.Answer = append(m.Answer, answers...)
	m.Extra = append(m.Extra, extras...)

	state.SizeAndDo(m)
	m = state.Scrub(m)
	_ = w.WriteMsg(m)
	return dns.RcodeSuccess, nil
}

// Name implements the Handler interface.
func (redis *RedisHandler) Name() string { return "redis" }

func (redis *RedisHandler) errorResponse(state request.Request, zone string, rcode int, err error) (int, error) {
	m := new(dns.Msg)
	m.SetRcode(state.Req, rcode)
	m.Authoritative, m.RecursionAvailable, m.Compress = true, false, true

	state.SizeAndDo(m)
	_ = state.W.WriteMsg(m)
	// Return success as the rcode to signal we have written to the client.
	return dns.RcodeSuccess, err
}
