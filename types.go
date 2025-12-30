package main

import (
	"github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

// TypeDef 定义类型信息
type TypeDef struct {
	TypeName string
	DefName  string
	Options  any
}

// getInboundTypes 返回所有 inbound 类型定义
func getInboundTypes() []TypeDef {
	return []TypeDef{
		{constant.TypeTun, "TunInbound", &option.TunInboundOptions{}},
		{constant.TypeRedirect, "RedirectInbound", &option.RedirectInboundOptions{}},
		{constant.TypeTProxy, "TProxyInbound", &option.TProxyInboundOptions{}},
		{constant.TypeDirect, "DirectInbound", &option.DirectInboundOptions{}},
		{constant.TypeSOCKS, "SocksInbound", &option.SocksInboundOptions{}},
		{constant.TypeHTTP, "HTTPInbound", &option.HTTPMixedInboundOptions{}},
		{constant.TypeMixed, "MixedInbound", &option.HTTPMixedInboundOptions{}},
		{constant.TypeShadowsocks, "ShadowsocksInbound", &option.ShadowsocksInboundOptions{}},
		{constant.TypeVMess, "VMessInbound", &option.VMessInboundOptions{}},
		{constant.TypeTrojan, "TrojanInbound", &option.TrojanInboundOptions{}},
		{constant.TypeNaive, "NaiveInbound", &option.NaiveInboundOptions{}},
		{constant.TypeHysteria, "HysteriaInbound", &option.HysteriaInboundOptions{}},
		{constant.TypeShadowTLS, "ShadowTLSInbound", &option.ShadowTLSInboundOptions{}},
		{constant.TypeAnyTLS, "AnyTLSInbound", &option.AnyTLSInboundOptions{}},
		{constant.TypeVLESS, "VLESSInbound", &option.VLESSInboundOptions{}},
		{constant.TypeTUIC, "TUICInbound", &option.TUICInboundOptions{}},
		{constant.TypeHysteria2, "Hysteria2Inbound", &option.Hysteria2InboundOptions{}},
	}
}

// getOutboundTypes 返回所有 outbound 类型定义
func getOutboundTypes() []TypeDef {
	return []TypeDef{
		{constant.TypeDirect, "DirectOutbound", &option.DirectOutboundOptions{}},
		{constant.TypeBlock, "BlockOutbound", nil}, // Block 没有特殊选项，使用 nil 标记
		{constant.TypeDNS, "DNSOutbound", nil},     // DNS 没有特殊选项，使用 nil 标记
		{constant.TypeSOCKS, "SocksOutbound", &option.SOCKSOutboundOptions{}},
		{constant.TypeHTTP, "HTTPOutbound", &option.HTTPOutboundOptions{}},
		{constant.TypeShadowsocks, "ShadowsocksOutbound", &option.ShadowsocksOutboundOptions{}},
		{constant.TypeVMess, "VMessOutbound", &option.VMessOutboundOptions{}},
		{constant.TypeTrojan, "TrojanOutbound", &option.TrojanOutboundOptions{}},
		{constant.TypeWireGuard, "WireGuardOutbound", &option.LegacyWireGuardOutboundOptions{}},
		{constant.TypeHysteria, "HysteriaOutbound", &option.HysteriaOutboundOptions{}},
		{constant.TypeTor, "TorOutbound", &option.TorOutboundOptions{}},
		{constant.TypeSSH, "SSHOutbound", &option.SSHOutboundOptions{}},
		{constant.TypeShadowTLS, "ShadowTLSOutbound", &option.ShadowTLSOutboundOptions{}},
		{constant.TypeAnyTLS, "AnyTLSOutbound", &option.AnyTLSOutboundOptions{}},
		{constant.TypeShadowsocksR, "ShadowsocksROutbound", &option.ShadowsocksROutboundOptions{}},
		{constant.TypeVLESS, "VLESSOutbound", &option.VLESSOutboundOptions{}},
		{constant.TypeTUIC, "TUICOutbound", &option.TUICOutboundOptions{}},
		{constant.TypeHysteria2, "Hysteria2Outbound", &option.Hysteria2OutboundOptions{}},
		{constant.TypeSelector, "SelectorOutbound", &option.SelectorOutboundOptions{}},
		{constant.TypeURLTest, "URLTestOutbound", &option.URLTestOutboundOptions{}},
	}
}

// getRuleTypes 返回所有 rule 类型定义
func getRuleTypes() []TypeDef {
	return []TypeDef{
		{constant.RuleTypeDefault, "DefaultRule", &option.DefaultRule{}},
		{constant.RuleTypeLogical, "LogicalRule", &option.LogicalRule{}},
	}
}

// getRuleSetTypes 返回所有 rule_set 类型定义
func getRuleSetTypes() []TypeDef {
	return []TypeDef{
		{constant.RuleSetTypeLocal, "LocalRuleSet", &option.LocalRuleSet{}},
		{constant.RuleSetTypeRemote, "RemoteRuleSet", &option.RemoteRuleSet{}},
	}
}

// getDNSRuleTypes 返回所有 dns rule 类型定义
func getDNSRuleTypes() []TypeDef {
	return []TypeDef{
		{constant.RuleTypeDefault, "DefaultDNSRule", &option.DefaultDNSRule{}},
		{constant.RuleTypeLogical, "LogicalDNSRule", &option.LogicalDNSRule{}},
	}
}

// getDNSServerTypes 返回所有 dns server 类型定义
func getDNSServerTypes() []TypeDef {
	return []TypeDef{
		{constant.DNSTypeLegacy, "LegacyDNSServer", &option.LegacyDNSServerOptions{}},
		{constant.DNSTypeUDP, "UDPDNSServer", &option.RemoteDNSServerOptions{}},
		{constant.DNSTypeTCP, "TCPDNSServer", &option.RemoteDNSServerOptions{}},
		{constant.DNSTypeTLS, "TLSDNSServer", &option.RemoteTLSDNSServerOptions{}},
		{constant.DNSTypeHTTPS, "HTTPSDNSServer", &option.RemoteHTTPSDNSServerOptions{}},
		{constant.DNSTypeQUIC, "QUICDNSServer", &option.RemoteTLSDNSServerOptions{}},
		{constant.DNSTypeHTTP3, "H3DNSServer", &option.RemoteHTTPSDNSServerOptions{}},
		{constant.DNSTypeLocal, "LocalDNSServer", &option.LocalDNSServerOptions{}},
		{constant.DNSTypeHosts, "HostsDNSServer", &option.HostsDNSServerOptions{}},
		{constant.DNSTypeFakeIP, "FakeIPDNSServer", &option.FakeIPDNSServerOptions{}},
		{constant.DNSTypeDHCP, "DHCPDNSServer", &option.DHCPDNSServerOptions{}},
		{constant.DNSTypeTailscale, "TailscaleDNSServer", &option.TailscaleDNSServerOptions{}},
	}
}
