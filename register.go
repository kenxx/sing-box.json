package main

import (
	"slices"

	"github.com/kenxx/sing-box.json/schema"
	"github.com/sagernet/sing-box/constant"
)

// registerAndGenerateOneOf 注册所有类型到顶层 $defs，返回 oneOf 引用数组
func registerAndGenerateOneOf(reflector *schema.Reflector, defs map[string]*schema.Schema, category string) []*schema.Schema {
	var oneOfSchemas []*schema.Schema

	var typeList []TypeDef

	switch category {
	case "inbound":
		typeList = getInboundTypes()
	case "outbound":
		typeList = getOutboundTypes()
	default:
		return nil
	}

	for _, t := range typeList {
		var typeSchema *schema.Schema
		if t.Options != nil {
			typeSchema = reflector.Reflect(t.Options)
		} else {
			// 对于没有特殊选项的类型，创建空 schema
			typeSchema = &schema.Schema{
				Type:       "object",
				Properties: make(map[string]*schema.Schema),
			}
		}

		// 添加 type 字段约束
		typeSchema.AddProperty("type", &schema.Schema{
			Type:    "string",
			Const:   t.TypeName,
			Default: t.TypeName,
		})

		// 添加 tag 字段（如果不存在）
		if _, ok := typeSchema.Properties["tag"]; !ok {
			typeSchema.AddProperty("tag", &schema.Schema{
				Type: "string",
			})
		}

		// 确保 type 是必需的
		if !slices.Contains(typeSchema.Required, "type") {
			typeSchema.Required = append(typeSchema.Required, "type")
		}

		// 将子 schema 的 $defs 合并到顶层
		mergeDefinitions(typeSchema, defs)

		// 注册到顶层 $defs
		defs[t.DefName] = typeSchema

		// 创建 oneOf 引用
		oneOfSchemas = append(oneOfSchemas, &schema.Schema{
			Ref: "#/$defs/" + t.DefName,
		})
	}

	return oneOfSchemas
}

// registerRuleTypes 注册 Rule 类型
func registerRuleTypes(reflector *schema.Reflector, defs map[string]*schema.Schema) []*schema.Schema {
	var oneOfSchemas []*schema.Schema

	for _, ruleType := range getRuleTypes() {
		typeSchema := reflector.Reflect(ruleType.Options)

		// 添加 type 字段约束
		if ruleType.TypeName == constant.RuleTypeDefault {
			typeSchema.AddProperty("type", &schema.Schema{
				Type: "string",
				Enum: []interface{}{"", constant.RuleTypeDefault},
			})
		} else {
			typeSchema.AddProperty("type", &schema.Schema{
				Type:    "string",
				Const:   ruleType.TypeName,
				Default: ruleType.TypeName,
			})
		}

		// 合并子 $defs
		mergeDefinitions(typeSchema, defs)

		defs[ruleType.DefName] = typeSchema
		oneOfSchemas = append(oneOfSchemas, &schema.Schema{
			Ref: "#/$defs/" + ruleType.DefName,
		})
	}

	return oneOfSchemas
}

// registerRuleSetTypes 注册 RuleSet 类型
func registerRuleSetTypes(reflector *schema.Reflector, defs map[string]*schema.Schema) []*schema.Schema {
	var oneOfSchemas []*schema.Schema

	for _, ruleSetType := range getRuleSetTypes() {
		typeSchema := reflector.Reflect(ruleSetType.Options)

		typeSchema.AddProperty("type", &schema.Schema{
			Type:    "string",
			Const:   ruleSetType.TypeName,
			Default: ruleSetType.TypeName,
		})

		// 确保 type, tag, format 是必需的
		requiredFields := []string{"type", "tag", "format"}
		for _, field := range requiredFields {
			if !slices.Contains(typeSchema.Required, field) {
				typeSchema.Required = append(typeSchema.Required, field)
			}
		}

		// 合并子 $defs
		mergeDefinitions(typeSchema, defs)

		defs[ruleSetType.DefName] = typeSchema
		oneOfSchemas = append(oneOfSchemas, &schema.Schema{
			Ref: "#/$defs/" + ruleSetType.DefName,
		})
	}

	return oneOfSchemas
}

// registerDNSRuleTypes 注册 DNSRule 类型
func registerDNSRuleTypes(reflector *schema.Reflector, defs map[string]*schema.Schema) []*schema.Schema {
	var oneOfSchemas []*schema.Schema

	for _, dnsRuleType := range getDNSRuleTypes() {
		typeSchema := reflector.Reflect(dnsRuleType.Options)

		if dnsRuleType.TypeName == constant.RuleTypeDefault {
			typeSchema.AddProperty("type", &schema.Schema{
				Type: "string",
				Enum: []interface{}{"", constant.RuleTypeDefault},
			})
		} else {
			typeSchema.AddProperty("type", &schema.Schema{
				Type:    "string",
				Const:   dnsRuleType.TypeName,
				Default: dnsRuleType.TypeName,
			})
		}

		// 合并子 $defs
		mergeDefinitions(typeSchema, defs)

		defs[dnsRuleType.DefName] = typeSchema
		oneOfSchemas = append(oneOfSchemas, &schema.Schema{
			Ref: "#/$defs/" + dnsRuleType.DefName,
		})
	}

	return oneOfSchemas
}

// registerDNSServerTypes 注册 DNSServer 类型
func registerDNSServerTypes(reflector *schema.Reflector, defs map[string]*schema.Schema) []*schema.Schema {
	var oneOfSchemas []*schema.Schema

	for _, serverType := range getDNSServerTypes() {
		typeSchema := reflector.Reflect(serverType.Options)

		// 添加 type 字段约束
		typeSchema.AddProperty("type", &schema.Schema{
			Type:    "string",
			Const:   serverType.TypeName,
			Default: serverType.TypeName,
		})

		// 添加 tag 字段（如果不存在）
		if _, ok := typeSchema.Properties["tag"]; !ok {
			typeSchema.AddProperty("tag", &schema.Schema{
				Type: "string",
			})
		}

		// 合并子 $defs
		mergeDefinitions(typeSchema, defs)

		defs[serverType.DefName] = typeSchema
		oneOfSchemas = append(oneOfSchemas, &schema.Schema{
			Ref: "#/$defs/" + serverType.DefName,
		})
	}

	return oneOfSchemas
}

// mergeDefinitions 将 typeSchema 的 $defs 合并到顶层 defs
func mergeDefinitions(typeSchema *schema.Schema, defs map[string]*schema.Schema) {
	if typeSchema.Definitions != nil {
		for name, def := range typeSchema.Definitions {
			if _, exists := defs[name]; !exists {
				defs[name] = def
			}
		}
		typeSchema.Definitions = nil
	}
}

