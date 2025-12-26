package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/option"
)

func main() {
	// 创建 Reflector 实例
	reflector := jsonschema.Reflector{
		// 设置一些选项以生成更好的 schema
		DoNotReference: false,
		ExpandedStruct: true,
		// 允许额外的属性，因为 Inbound 和 Outbound 会根据 type 动态添加字段
		AllowAdditionalProperties: true,
	}

	// 首先生成所有类型的 schema 定义
	baseSchema := reflector.Reflect(&option.Options{})

	// 手动为 Inbound 生成完整的 oneOf schema
	inboundSchema := generateInboundSchema(&reflector)
	if len(inboundSchema) > 0 {
		// 更新 inbounds 数组项的 schema
		if inboundsProp, ok := baseSchema.Properties.Get("inbounds"); ok && inboundsProp != nil {
			if inboundsProp.Items != nil {
				itemsSchema := inboundsProp.Items
				itemsSchema.OneOf = inboundSchema
			}
		}
	}

	// 手动为 Outbound 生成完整的 oneOf schema
	outboundSchema := generateOutboundSchema(&reflector)
	if len(outboundSchema) > 0 {
		// 更新 outbounds 数组项的 schema
		if outboundsProp, ok := baseSchema.Properties.Get("outbounds"); ok && outboundsProp != nil {
			if outboundsProp.Items != nil {
				itemsSchema := outboundsProp.Items
				itemsSchema.OneOf = outboundSchema
			}
		}
	}

	// 手动为 Rule 生成完整的 oneOf schema
	ruleSchema := generateRuleSchema(&reflector)
	if len(ruleSchema) > 0 {
		// 更新 route.rules 数组项的 schema
		if routeProp, ok := baseSchema.Properties.Get("route"); ok && routeProp != nil {
			if routeProp.Properties != nil {
				if rulesProp, ok := routeProp.Properties.Get("rules"); ok && rulesProp != nil {
					if rulesProp.Items != nil {
						itemsSchema := rulesProp.Items
						itemsSchema.OneOf = ruleSchema
					}
				}
			}
		}
	}

	// 手动为 RuleSet 生成完整的 oneOf schema
	ruleSetSchema := generateRuleSetSchema(&reflector)
	if len(ruleSetSchema) > 0 {
		// 更新 route.rule_set 数组项的 schema
		if routeProp, ok := baseSchema.Properties.Get("route"); ok && routeProp != nil {
			if routeProp.Properties != nil {
				if ruleSetProp, ok := routeProp.Properties.Get("rule_set"); ok && ruleSetProp != nil {
					if ruleSetProp.Items != nil {
						itemsSchema := ruleSetProp.Items
						itemsSchema.OneOf = ruleSetSchema
					}
				}
			}
		}
	}

	// 手动为 DNSRule 生成完整的 oneOf schema
	dnsRuleSchema := generateDNSRuleSchema(&reflector)
	if len(dnsRuleSchema) > 0 {
		// 更新 dns.rules 数组项的 schema
		if dnsProp, ok := baseSchema.Properties.Get("dns"); ok && dnsProp != nil {
			if dnsProp.Properties != nil {
				if dnsRulesProp, ok := dnsProp.Properties.Get("rules"); ok && dnsRulesProp != nil {
					if dnsRulesProp.Items != nil {
						itemsSchema := dnsRulesProp.Items
						itemsSchema.OneOf = dnsRuleSchema
					}
				}
			}
		}
	}

	// 设置 schema 的元数据
	baseSchema.ID = "https://kenxx.github.io/sing-box.json/sing-box-config-schema.json"
	baseSchema.Title = "sing-box Configuration Schema"
	baseSchema.Description = "JSON Schema for sing-box configuration file"

	// 更新所有 $defs 中的 $id，使其指向我们的 GitHub Pages
	updateSchemaIDs(baseSchema, "https://kenxx.github.io/sing-box.json")

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

// generateInboundSchema 为所有 inbound 类型生成 oneOf schema
func generateInboundSchema(reflector *jsonschema.Reflector) []*jsonschema.Schema {
	var schemas []*jsonschema.Schema

	// 使用反射动态获取类型映射关系
	typeMapping := getInboundTypeMapping()

	// 使用反射获取 _Inbound 结构体的所有字段
	inboundStructType := reflect.TypeOf((*option.Inbound)(nil)).Elem()

	// 遍历所有字段（跳过 Type 和 Tag）
	for i := 0; i < inboundStructType.NumField(); i++ {
		field := inboundStructType.Field(i)
		fieldName := field.Name

		// 跳过 Type 和 Tag 字段
		if fieldName == "Type" || fieldName == "Tag" {
			continue
		}

		// 获取类型名称（通过动态映射）
		typeName, ok := typeMapping[fieldName]
		if !ok {
			continue
		}

		// 获取字段的类型
		fieldType := field.Type

		// 创建该类型的零值实例
		fieldValue := reflect.New(fieldType).Interface()

		// 生成 schema
		typeSchema := reflector.Reflect(fieldValue)
		// 添加 type 字段的约束
		typeSchema.Properties.Set("type", &jsonschema.Schema{
			Type:    "string",
			Const:   typeName,
			Default: typeName,
		})
		// 确保 type 是必需的
		if typeSchema.Required == nil {
			typeSchema.Required = []string{"type"}
		} else {
			// 检查是否已包含 type
			hasType := false
			for _, req := range typeSchema.Required {
				if req == "type" {
					hasType = true
					break
				}
			}
			if !hasType {
				typeSchema.Required = append(typeSchema.Required, "type")
			}
		}
		schemas = append(schemas, typeSchema)
	}

	return schemas
}

// getInboundTypeMapping 通过反射动态获取 inbound 类型映射关系
// 通过遍历所有字段，尝试从字段名推断类型名，然后验证
func getInboundTypeMapping() map[string]string {
	mapping := make(map[string]string)

	// 创建一个临时的 Inbound 实例
	inbound := &option.Inbound{}
	inboundValue := reflect.ValueOf(inbound).Elem()
	inboundType := inboundValue.Type()

	// 获取所有字段的地址，用于后续比较
	fieldAddrs := make(map[uintptr]string)
	fieldNames := make([]string, 0)
	for i := 0; i < inboundType.NumField(); i++ {
		field := inboundType.Field(i)
		fieldName := field.Name
		if fieldName == "Type" || fieldName == "Tag" {
			continue
		}
		fieldValue := inboundValue.Field(i)
		if fieldValue.CanAddr() {
			addr := fieldValue.Addr().Pointer()
			fieldAddrs[addr] = fieldName
			fieldNames = append(fieldNames, fieldName)
		}
	}

	// 对每个字段，尝试推断可能的类型名
	// 规则：去掉 "Options" 后缀，然后转换为小写
	// 特殊处理：HTTPOptions -> http, MixedOptions -> mixed, SOCKS -> socks 等
	for _, fieldName := range fieldNames {
		// 尝试从字段名推断类型名
		typeName := inferTypeNameFromFieldName(fieldName)

		// 设置 Type 字段
		typeField := inboundValue.FieldByName("Type")
		if !typeField.IsValid() || !typeField.CanSet() {
			continue
		}

		typeField.SetString(typeName)

		// 调用 RawOptions() 验证
		rawOptions, err := inbound.RawOptions()
		if err != nil {
			// 如果失败，尝试其他可能的类型名变体
			typeNameVariants := getTypeNameVariants(fieldName)
			for _, variant := range typeNameVariants {
				typeField.SetString(variant)
				rawOptions, err = inbound.RawOptions()
				if err == nil {
					typeName = variant
					break
				}
			}
			if err != nil {
				continue
			}
		}

		// 如果返回 nil，跳过
		if rawOptions == nil {
			continue
		}

		// 获取返回值的指针并验证
		optionsValue := reflect.ValueOf(rawOptions)
		if optionsValue.Kind() == reflect.Ptr {
			optionsPtr := optionsValue.Pointer()

			// 查找这个指针对应的字段
			if mappedFieldName, ok := fieldAddrs[optionsPtr]; ok && mappedFieldName == fieldName {
				mapping[fieldName] = typeName
			}
		}
	}

	return mapping
}

// inferTypeNameFromFieldName 从字段名推断类型名
func inferTypeNameFromFieldName(fieldName string) string {
	// 去掉 "Options" 后缀
	if len(fieldName) > 7 && fieldName[len(fieldName)-7:] == "Options" {
		baseName := fieldName[:len(fieldName)-7]
		// 转换为小写
		return toLowerCamelCase(baseName)
	}
	return ""
}

// toLowerCamelCase 将驼峰命名转换为小写加下划线（或直接小写）
func toLowerCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	// 特殊映射
	specialCases := map[string]string{
		"SOCKS":        "socks",
		"HTTP":         "http",
		"TLS":          "tls",
		"TUIC":         "tuic",
		"VMess":        "vmess",
		"VLESS":        "vless",
		"ShadowTLS":    "shadowtls",
		"Shadowsocks":  "shadowsocks",
		"ShadowsocksR": "shadowsocksr",
		"TProxy":       "tproxy",
		"Tun":          "tun",
		"Hysteria":     "hysteria",
		"Hysteria2":    "hysteria2",
	}

	if mapped, ok := specialCases[s]; ok {
		return mapped
	}

	// 通用转换：首字母小写，其他保持原样
	result := make([]rune, 0, len(s))
	for i, r := range s {
		if i == 0 {
			result = append(result, toLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// toLower 将单个字符转换为小写
func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

// getTypeNameVariants 获取类型名的可能变体
func getTypeNameVariants(fieldName string) []string {
	variants := []string{}

	// 尝试不同的命名规则
	baseName := fieldName
	if len(baseName) > 7 && baseName[len(baseName)-7:] == "Options" {
		baseName = baseName[:len(baseName)-7]
	}

	// 尝试全小写
	variants = append(variants, toLowerCamelCase(baseName))

	// 尝试特殊映射
	specialMappings := map[string][]string{
		"SocksOptions":        {"socks"},
		"HTTPOptions":         {"http"},
		"MixedOptions":        {"mixed"},
		"ShadowTLSOptions":    {"shadowtls"},
		"ShadowsocksOptions":  {"shadowsocks"},
		"ShadowsocksROptions": {"shadowsocksr"},
		"TProxyOptions":       {"tproxy"},
		"TUICOptions":         {"tuic"},
		"Hysteria2Options":    {"hysteria2"},
	}

	if mapped, ok := specialMappings[fieldName]; ok {
		variants = append(variants, mapped...)
	}

	return variants
}

// generateOutboundSchema 为所有 outbound 类型生成 oneOf schema
func generateOutboundSchema(reflector *jsonschema.Reflector) []*jsonschema.Schema {
	var schemas []*jsonschema.Schema

	// 使用反射动态获取类型映射关系
	typeMapping := getOutboundTypeMapping()

	// 使用反射获取 _Outbound 结构体的所有字段
	outboundStructType := reflect.TypeOf((*option.Outbound)(nil)).Elem()

	// 特殊处理：没有选项的类型（通过检查 RawOptions() 返回 nil 来识别）
	noOptionsTypes := []string{
		constant.TypeBlock,
		constant.TypeDNS,
	}

	// 先添加没有选项的类型
	for _, typeName := range noOptionsTypes {
		props := jsonschema.NewProperties()
		props.Set("type", &jsonschema.Schema{
			Type:    "string",
			Const:   typeName,
			Default: typeName,
		})
		props.Set("tag", &jsonschema.Schema{
			Type: "string",
		})
		typeSchema := &jsonschema.Schema{
			Type:       "object",
			Properties: props,
			Required:   []string{"type"},
		}
		schemas = append(schemas, typeSchema)
	}

	// 遍历所有字段（跳过 Type 和 Tag）
	for i := 0; i < outboundStructType.NumField(); i++ {
		field := outboundStructType.Field(i)
		fieldName := field.Name

		// 跳过 Type 和 Tag 字段
		if fieldName == "Type" || fieldName == "Tag" {
			continue
		}

		// 获取类型名称（通过动态映射）
		typeName, ok := typeMapping[fieldName]
		if !ok {
			continue
		}

		// 获取字段的类型
		fieldType := field.Type

		// 创建该类型的零值实例
		fieldValue := reflect.New(fieldType).Interface()

		// 生成 schema
		typeSchema := reflector.Reflect(fieldValue)
		// 添加 type 字段的约束
		typeSchema.Properties.Set("type", &jsonschema.Schema{
			Type:    "string",
			Const:   typeName,
			Default: typeName,
		})
		// 确保 type 是必需的
		if typeSchema.Required == nil {
			typeSchema.Required = []string{"type"}
		} else {
			// 检查是否已包含 type
			hasType := false
			for _, req := range typeSchema.Required {
				if req == "type" {
					hasType = true
					break
				}
			}
			if !hasType {
				typeSchema.Required = append(typeSchema.Required, "type")
			}
		}
		schemas = append(schemas, typeSchema)
	}

	return schemas
}

// getOutboundTypeMapping 通过反射动态获取 outbound 类型映射关系
// 通过遍历所有字段，尝试从字段名推断类型名，然后验证
func getOutboundTypeMapping() map[string]string {
	mapping := make(map[string]string)

	// 创建一个临时的 Outbound 实例
	outbound := &option.Outbound{}
	outboundValue := reflect.ValueOf(outbound).Elem()
	outboundType := outboundValue.Type()

	// 获取所有字段的地址，用于后续比较
	fieldAddrs := make(map[uintptr]string)
	fieldNames := make([]string, 0)
	for i := 0; i < outboundType.NumField(); i++ {
		field := outboundType.Field(i)
		fieldName := field.Name
		if fieldName == "Type" || fieldName == "Tag" {
			continue
		}
		fieldValue := outboundValue.Field(i)
		if fieldValue.CanAddr() {
			addr := fieldValue.Addr().Pointer()
			fieldAddrs[addr] = fieldName
			fieldNames = append(fieldNames, fieldName)
		}
	}

	// 对每个字段，尝试推断可能的类型名
	for _, fieldName := range fieldNames {
		// 尝试从字段名推断类型名
		typeName := inferTypeNameFromFieldName(fieldName)

		// 设置 Type 字段
		typeField := outboundValue.FieldByName("Type")
		if !typeField.IsValid() || !typeField.CanSet() {
			continue
		}

		typeField.SetString(typeName)

		// 调用 RawOptions() 验证
		rawOptions, err := outbound.RawOptions()
		if err != nil {
			// 如果失败，尝试其他可能的类型名变体
			typeNameVariants := getTypeNameVariants(fieldName)
			for _, variant := range typeNameVariants {
				typeField.SetString(variant)
				rawOptions, err = outbound.RawOptions()
				if err == nil {
					typeName = variant
					break
				}
			}
			if err != nil {
				continue
			}
		}

		// 如果返回 nil，跳过（block, dns 等）
		if rawOptions == nil {
			continue
		}

		// 获取返回值的指针并验证
		optionsValue := reflect.ValueOf(rawOptions)
		if optionsValue.Kind() == reflect.Ptr {
			optionsPtr := optionsValue.Pointer()

			// 查找这个指针对应的字段
			if mappedFieldName, ok := fieldAddrs[optionsPtr]; ok && mappedFieldName == fieldName {
				mapping[fieldName] = typeName
			}
		}
	}

	return mapping
}

// generateRuleSchema 为所有 rule 类型生成 oneOf schema
func generateRuleSchema(reflector *jsonschema.Reflector) []*jsonschema.Schema {
	var schemas []*jsonschema.Schema

	// Rule 有两种类型：default 和 logical
	ruleTypes := []struct {
		typeName string
		options  interface{}
	}{
		{constant.RuleTypeDefault, &option.DefaultRule{}},
		{constant.RuleTypeLogical, &option.LogicalRule{}},
	}

	for _, ruleType := range ruleTypes {
		typeSchema := reflector.Reflect(ruleType.options)
		// 添加 type 字段的约束
		// default 类型的 type 字段可以为空字符串或 "default"
		if ruleType.typeName == constant.RuleTypeDefault {
			typeSchema.Properties.Set("type", &jsonschema.Schema{
				Type: "string",
				Enum: []interface{}{"", constant.RuleTypeDefault},
			})
		} else {
			typeSchema.Properties.Set("type", &jsonschema.Schema{
				Type:    "string",
				Const:   ruleType.typeName,
				Default: ruleType.typeName,
			})
		}
		schemas = append(schemas, typeSchema)
	}

	return schemas
}

// generateRuleSetSchema 为所有 rule_set 类型生成 oneOf schema
func generateRuleSetSchema(reflector *jsonschema.Reflector) []*jsonschema.Schema {
	var schemas []*jsonschema.Schema

	// RuleSet 有两种类型：local 和 remote
	ruleSetTypes := []struct {
		typeName string
		options  interface{}
	}{
		{constant.RuleSetTypeLocal, &option.LocalRuleSet{}},
		{constant.RuleSetTypeRemote, &option.RemoteRuleSet{}},
	}

	for _, ruleSetType := range ruleSetTypes {
		typeSchema := reflector.Reflect(ruleSetType.options)
		// 添加 type 字段的约束
		typeSchema.Properties.Set("type", &jsonschema.Schema{
			Type:    "string",
			Const:   ruleSetType.typeName,
			Default: ruleSetType.typeName,
		})
		// RuleSet 需要 tag 和 format 字段
		typeSchema.Properties.Set("tag", &jsonschema.Schema{
			Type: "string",
		})
		typeSchema.Properties.Set("format", &jsonschema.Schema{
			Type: "string",
			Enum: []interface{}{constant.RuleSetFormatSource, constant.RuleSetFormatBinary},
		})
		// 确保 type, tag, format 是必需的
		if typeSchema.Required == nil {
			typeSchema.Required = []string{"type", "tag", "format"}
		} else {
			requiredMap := make(map[string]bool)
			for _, req := range typeSchema.Required {
				requiredMap[req] = true
			}
			if !requiredMap["type"] {
				typeSchema.Required = append(typeSchema.Required, "type")
			}
			if !requiredMap["tag"] {
				typeSchema.Required = append(typeSchema.Required, "tag")
			}
			if !requiredMap["format"] {
				typeSchema.Required = append(typeSchema.Required, "format")
			}
		}
		schemas = append(schemas, typeSchema)
	}

	return schemas
}

// generateDNSRuleSchema 为所有 dns_rule 类型生成 oneOf schema
func generateDNSRuleSchema(reflector *jsonschema.Reflector) []*jsonschema.Schema {
	var schemas []*jsonschema.Schema

	// DNSRule 有两种类型：default 和 logical
	dnsRuleTypes := []struct {
		typeName string
		options  interface{}
	}{
		{constant.RuleTypeDefault, &option.DefaultDNSRule{}},
		{constant.RuleTypeLogical, &option.LogicalDNSRule{}},
	}

	for _, dnsRuleType := range dnsRuleTypes {
		typeSchema := reflector.Reflect(dnsRuleType.options)
		// 添加 type 字段的约束
		// default 类型的 type 字段可以为空字符串或 "default"
		if dnsRuleType.typeName == constant.RuleTypeDefault {
			typeSchema.Properties.Set("type", &jsonschema.Schema{
				Type: "string",
				Enum: []interface{}{"", constant.RuleTypeDefault},
			})
		} else {
			typeSchema.Properties.Set("type", &jsonschema.Schema{
				Type:    "string",
				Const:   dnsRuleType.typeName,
				Default: dnsRuleType.typeName,
			})
		}
		schemas = append(schemas, typeSchema)
	}

	return schemas
}

// updateSchemaIDs 递归更新 schema 中所有 $id 字段，使其指向我们的 GitHub Pages
func updateSchemaIDs(schema *jsonschema.Schema, baseURL string) {
	// 更新当前 schema 的 $id（如果存在且指向 sing-box 仓库）
	idStr := string(schema.ID)
	if idStr != "" && strings.Contains(idStr, "github.com/sagernet/sing-box") {
		// 从原始 $id 中提取类型名称
		// 例如: https://github.com/sagernet/sing-box/option/direct-outbound-options
		// 转换为: DirectOutboundOptions (驼峰命名)
		if strings.Contains(idStr, "/option/") {
			parts := strings.Split(idStr, "/option/")
			if len(parts) > 1 {
				typeNameKebab := parts[1]
				// 将 kebab-case 转换为 PascalCase
				// direct-outbound-options -> DirectOutboundOptions
				typeNameParts := strings.Split(typeNameKebab, "-")
				var typeNamePascal strings.Builder
				for _, part := range typeNameParts {
					if len(part) > 0 {
						typeNamePascal.WriteString(strings.ToUpper(part[:1]) + part[1:])
					}
				}
				// 指向主 schema 文件，使用片段标识符指向 $defs 中的定义
				// 这样链接可以点击，并且指向正确的定义
				schema.ID = jsonschema.ID(baseURL + "/sing-box-config-schema.json#/$defs/" + typeNamePascal.String())
			}
		} else {
			// 其他情况，直接指向主文件
			schema.ID = jsonschema.ID(baseURL + "/sing-box-config-schema.json")
		}
	}

	// 递归更新 $defs 中的所有 schema
	if schema.Definitions != nil {
		for _, defSchema := range schema.Definitions {
			if defSchema != nil {
				updateSchemaIDs(defSchema, baseURL)
			}
		}
	}

	// 递归更新 oneOf 中的 schema
	if schema.OneOf != nil {
		for _, oneOfSchema := range schema.OneOf {
			if oneOfSchema != nil {
				updateSchemaIDs(oneOfSchema, baseURL)
			}
		}
	}

	// 递归更新 anyOf 中的 schema
	if schema.AnyOf != nil {
		for _, anyOfSchema := range schema.AnyOf {
			if anyOfSchema != nil {
				updateSchemaIDs(anyOfSchema, baseURL)
			}
		}
	}

	// 递归更新 allOf 中的 schema
	if schema.AllOf != nil {
		for _, allOfSchema := range schema.AllOf {
			if allOfSchema != nil {
				updateSchemaIDs(allOfSchema, baseURL)
			}
		}
	}
}
