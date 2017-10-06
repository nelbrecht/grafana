package setting

func readXMPPSettings(){
	sec := Cfg.Section("xmpp")
	XMPPEnabled = sec.Key("enabled").MustBool(false)
	XMPP.Host = sec.Key("host").String()
	XMPP.User = sec.Key("username").String()
	XMPP.Password = sec.Key("password").String()
	XMPP.NoTLS = sec.Key("no_tls").MustBool(false)
	XMPP.Debug = sec.Key("debug").MustBool(false)
	XMPP.Session = sec.Key("session").MustBool(false)
	XMPP.Status = sec.Key("status").String()
	XMPP.StatusMessage = sec.Key("status_message").String()
}


