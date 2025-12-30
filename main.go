package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/kenxx/sing-box.json/schema"
	"github.com/sagernet/sing-box/option"
)

func main() {
	// 创建 Reflector 实例
	reflector := schema.NewReflector()
	reflector.AllowAdditionalProperties = true

	// 首先生成所有类型的 schema 定义
	baseSchema := reflector.Reflect(&option.Options{})

	// 确保顶层 $defs 存在
	if baseSchema.Definitions == nil {
		baseSchema.Definitions = make(map[string]*schema.Schema)
	}

	// 1. 先注册所有 Inbound 类型到顶层 $defs，然后生成 oneOf 引用
	inboundOneOf := registerAndGenerateOneOf(reflector, baseSchema.Definitions, "inbound")
	if len(inboundOneOf) > 0 {
		if inboundsProp, ok := baseSchema.Properties["inbounds"]; ok && inboundsProp != nil {
			if inboundsProp.Items != nil {
				inboundsProp.Items.OneOf = inboundOneOf
				inboundsProp.Items.Ref = "" // 清除原来的 $ref
			}
		}
	}

	// 2. 注册所有 Outbound 类型到顶层 $defs，然后生成 oneOf 引用
	outboundOneOf := registerAndGenerateOneOf(reflector, baseSchema.Definitions, "outbound")
	if len(outboundOneOf) > 0 {
		if outboundsProp, ok := baseSchema.Properties["outbounds"]; ok && outboundsProp != nil {
			if outboundsProp.Items != nil {
				outboundsProp.Items.OneOf = outboundOneOf
				outboundsProp.Items.Ref = "" // 清除原来的 $ref
			}
		}
	}

	// 3. 注册 Rule 类型
	ruleOneOf := registerRuleTypes(reflector, baseSchema.Definitions)
	if len(ruleOneOf) > 0 {
		if routeProp, ok := baseSchema.Properties["route"]; ok && routeProp != nil {
			if routeProp.Properties != nil {
				if rulesProp, ok := routeProp.Properties["rules"]; ok && rulesProp != nil {
					if rulesProp.Items != nil {
						rulesProp.Items.OneOf = ruleOneOf
						rulesProp.Items.Ref = ""
					}
				}
			}
		}
	}

	// 4. 注册 RuleSet 类型
	ruleSetOneOf := registerRuleSetTypes(reflector, baseSchema.Definitions)
	if len(ruleSetOneOf) > 0 {
		if routeProp, ok := baseSchema.Properties["route"]; ok && routeProp != nil {
			if routeProp.Properties != nil {
				if ruleSetProp, ok := routeProp.Properties["rule_set"]; ok && ruleSetProp != nil {
					if ruleSetProp.Items != nil {
						ruleSetProp.Items.OneOf = ruleSetOneOf
						ruleSetProp.Items.Ref = ""
					}
				}
			}
		}
	}

	// 5. 注册 DNSRule 类型
	dnsRuleOneOf := registerDNSRuleTypes(reflector, baseSchema.Definitions)
	if len(dnsRuleOneOf) > 0 {
		if dnsProp, ok := baseSchema.Properties["dns"]; ok && dnsProp != nil {
			if dnsProp.Properties != nil {
				if dnsRulesProp, ok := dnsProp.Properties["rules"]; ok && dnsRulesProp != nil {
					if dnsRulesProp.Items != nil {
						dnsRulesProp.Items.OneOf = dnsRuleOneOf
						dnsRulesProp.Items.Ref = ""
					}
				}
			}
		}
	}

	// 6. 注册 DNSServer 类型
	dnsServerOneOf := registerDNSServerTypes(reflector, baseSchema.Definitions)
	if len(dnsServerOneOf) > 0 {
		if dnsProp, ok := baseSchema.Properties["dns"]; ok && dnsProp != nil {
			if dnsProp.Properties != nil {
				if dnsServersProp, ok := dnsProp.Properties["servers"]; ok && dnsServersProp != nil {
					if dnsServersProp.Items != nil {
						dnsServersProp.Items.OneOf = dnsServerOneOf
						dnsServersProp.Items.Ref = ""
					}
				}
			}
		}
	}

	// 设置 schema 的元数据
	baseSchema.Schema = "http://json-schema.org/draft-07/schema#"
	baseSchema.ID = "https://kenxx.github.io/sing-box.json/sing-box-config-schema.json"
	baseSchema.Title = "sing-box Configuration Schema"
	baseSchema.Description = "JSON Schema for sing-box configuration file"

	// 规范化 $defs 名称（移除特殊字符如 [], /）
	normalizeDefNames(baseSchema)

	// 将 Schema 转换为 JSON 格式
	schemaJSON, err := json.MarshalIndent(baseSchema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating JSON Schema: %v\n", err)
		os.Exit(1)
	}

	// 将 JSON Schema 写入文件
	outputFile := "sing-box-config-schema.json"
	err = os.WriteFile(outputFile, schemaJSON, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON Schema to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("JSON Schema generated successfully: %s\n", outputFile)
}
