package dnsutil

import (
	"strings"

	"github.com/miekg/dns"
)

func NewMsgPrefixedLogger(msg *dns.Msg, logger Logger) PrefixedLogger {
	return NewPrefixedLogger(questionToString(msg)+": ", logger)
}

func questionToString(msg *dns.Msg) string {
	var result []string

	for _, question := range msg.Question {
		result = append(result, dns.Type(question.Qtype).String()+":"+question.Name)
	}

	return strings.Join(result, ",")
}
