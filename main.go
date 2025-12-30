package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
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

	// 确保顶层 $defs 存在
	if baseSchema.Definitions == nil {
		baseSchema.Definitions = make(jsonschema.Definitions)
	}

	// 1. 先注册所有 Inbound 类型到顶层 $defs，然后生成 oneOf 引用
	inboundOneOf := registerAndGenerateOneOf(&reflector, baseSchema.Definitions, "inbound")
	if len(inboundOneOf) > 0 {
		if inboundsProp, ok := baseSchema.Properties.Get("inbounds"); ok && inboundsProp != nil {
			if inboundsProp.Items != nil {
				inboundsProp.Items.OneOf = inboundOneOf
				inboundsProp.Items.Ref = "" // 清除原来的 $ref
			}
		}
	}

	// 2. 注册所有 Outbound 类型到顶层 $defs，然后生成 oneOf 引用
	outboundOneOf := registerAndGenerateOneOf(&reflector, baseSchema.Definitions, "outbound")
	if len(outboundOneOf) > 0 {
		if outboundsProp, ok := baseSchema.Properties.Get("outbounds"); ok && outboundsProp != nil {
			if outboundsProp.Items != nil {
				outboundsProp.Items.OneOf = outboundOneOf
				outboundsProp.Items.Ref = "" // 清除原来的 $ref
			}
		}
	}

	// 3. 注册 Rule 类型
	ruleOneOf := registerRuleTypes(&reflector, baseSchema.Definitions)
	if len(ruleOneOf) > 0 {
		if routeProp, ok := baseSchema.Properties.Get("route"); ok && routeProp != nil {
			if routeProp.Properties != nil {
				if rulesProp, ok := routeProp.Properties.Get("rules"); ok && rulesProp != nil {
					if rulesProp.Items != nil {
						rulesProp.Items.OneOf = ruleOneOf
						rulesProp.Items.Ref = ""
					}
				}
			}
		}
	}

	// 4. 注册 RuleSet 类型
	ruleSetOneOf := registerRuleSetTypes(&reflector, baseSchema.Definitions)
	if len(ruleSetOneOf) > 0 {
		if routeProp, ok := baseSchema.Properties.Get("route"); ok && routeProp != nil {
			if routeProp.Properties != nil {
				if ruleSetProp, ok := routeProp.Properties.Get("rule_set"); ok && ruleSetProp != nil {
					if ruleSetProp.Items != nil {
						ruleSetProp.Items.OneOf = ruleSetOneOf
						ruleSetProp.Items.Ref = ""
					}
				}
			}
		}
	}

	// 5. 注册 DNSRule 类型
	dnsRuleOneOf := registerDNSRuleTypes(&reflector, baseSchema.Definitions)
	if len(dnsRuleOneOf) > 0 {
		if dnsProp, ok := baseSchema.Properties.Get("dns"); ok && dnsProp != nil {
			if dnsProp.Properties != nil {
				if dnsRulesProp, ok := dnsProp.Properties.Get("rules"); ok && dnsRulesProp != nil {
					if dnsRulesProp.Items != nil {
						dnsRulesProp.Items.OneOf = dnsRuleOneOf
						dnsRulesProp.Items.Ref = ""
					}
				}
			}
		}
	}

	// 设置 schema 的元数据
	baseSchema.ID = "https://kenxx.github.io/sing-box.json/sing-box-config-schema.json"
	baseSchema.Title = "sing-box Configuration Schema"
	baseSchema.Description = "JSON Schema for sing-box configuration file"

	// 清理：移除所有嵌套的 $defs（已经合并到顶层）
	cleanupNestedDefs(baseSchema)

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

// registerAndGenerateOneOf 注册所有类型到顶层 $defs，返回 oneOf 引用数组
func registerAndGenerateOneOf(reflector *jsonschema.Reflector, defs jsonschema.Definitions, category string) []*jsonschema.Schema {
	var oneOfSchemas []*jsonschema.Schema

	var typeMapping map[string]string
	var structType reflect.Type

	switch category {
	case "inbound":
		typeMapping = getInboundTypeMapping()
		structType = reflect.TypeOf((*option.Inbound)(nil)).Elem()
	case "outbound":
		typeMapping = getOutboundTypeMapping()
		structType = reflect.TypeOf((*option.Outbound)(nil)).Elem()
	default:
		return nil
	}

	// 遍历所有字段
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldName := field.Name

		if fieldName == "Type" || fieldName == "Tag" {
			continue
		}

		typeName, ok := typeMapping[fieldName]
		if !ok {
			continue
		}

		// 生成定义名称，如 "HTTPInboundOptions"
		defName := fieldName
		if !strings.HasSuffix(defName, "Options") {
			defName = fieldName + "Options"
		}

		// 创建该类型的零值实例并生成 schema
		fieldType := field.Type
		fieldValue := reflect.New(fieldType).Interface()
		typeSchema := reflector.Reflect(fieldValue)

		// 添加 type 字段约束
		typeSchema.Properties.Set("type", &jsonschema.Schema{
			Type:    "string",
			Const:   typeName,
			Default: typeName,
		})

		// 确保 type 是必需的
		if !slices.Contains(typeSchema.Required, "type") {
			typeSchema.Required = append(typeSchema.Required, "type")
		}

		// 清理 schema 的元数据（$id, $schema）
		typeSchema.ID = ""
		typeSchema.Version = ""

		// 将子 schema 的 $defs 合并到顶层
		if typeSchema.Definitions != nil {
			for name, def := range typeSchema.Definitions {
				if _, exists := defs[name]; !exists {
					defs[name] = def
				}
			}
			typeSchema.Definitions = nil
		}

		// 注册到顶层 $defs
		defs[defName] = typeSchema

		// 创建 oneOf 引用
		oneOfSchemas = append(oneOfSchemas, &jsonschema.Schema{
			Ref: "#/$defs/" + defName,
		})
	}

	return oneOfSchemas
}

// registerRuleTypes 注册 Rule 类型
func registerRuleTypes(reflector *jsonschema.Reflector, defs jsonschema.Definitions) []*jsonschema.Schema {
	var oneOfSchemas []*jsonschema.Schema

	ruleTypes := []struct {
		typeName string
		defName  string
		options  interface{}
	}{
		{constant.RuleTypeDefault, "DefaultRule", &option.DefaultRule{}},
		{constant.RuleTypeLogical, "LogicalRule", &option.LogicalRule{}},
	}

	for _, ruleType := range ruleTypes {
		typeSchema := reflector.Reflect(ruleType.options)

		// 添加 type 字段约束
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

		// 清理元数据
		typeSchema.ID = ""
		typeSchema.Version = ""

		// 合并子 $defs
		if typeSchema.Definitions != nil {
			for name, def := range typeSchema.Definitions {
				if _, exists := defs[name]; !exists {
					defs[name] = def
				}
			}
			typeSchema.Definitions = nil
		}

		defs[ruleType.defName] = typeSchema
		oneOfSchemas = append(oneOfSchemas, &jsonschema.Schema{
			Ref: "#/$defs/" + ruleType.defName,
		})
	}

	return oneOfSchemas
}

// registerRuleSetTypes 注册 RuleSet 类型
func registerRuleSetTypes(reflector *jsonschema.Reflector, defs jsonschema.Definitions) []*jsonschema.Schema {
	var oneOfSchemas []*jsonschema.Schema

	ruleSetTypes := []struct {
		typeName string
		defName  string
		options  interface{}
	}{
		{constant.RuleSetTypeLocal, "LocalRuleSet", &option.LocalRuleSet{}},
		{constant.RuleSetTypeRemote, "RemoteRuleSet", &option.RemoteRuleSet{}},
	}

	for _, ruleSetType := range ruleSetTypes {
		typeSchema := reflector.Reflect(ruleSetType.options)

		typeSchema.Properties.Set("type", &jsonschema.Schema{
			Type:    "string",
			Const:   ruleSetType.typeName,
			Default: ruleSetType.typeName,
		})

		// 确保 type, tag, format 是必需的
		requiredFields := []string{"type", "tag", "format"}
		for _, field := range requiredFields {
			if !slices.Contains(typeSchema.Required, field) {
				typeSchema.Required = append(typeSchema.Required, field)
			}
		}

		// 清理元数据
		typeSchema.ID = ""
		typeSchema.Version = ""

		// 合并子 $defs
		if typeSchema.Definitions != nil {
			for name, def := range typeSchema.Definitions {
				if _, exists := defs[name]; !exists {
					defs[name] = def
				}
			}
			typeSchema.Definitions = nil
		}

		defs[ruleSetType.defName] = typeSchema
		oneOfSchemas = append(oneOfSchemas, &jsonschema.Schema{
			Ref: "#/$defs/" + ruleSetType.defName,
		})
	}

	return oneOfSchemas
}

// registerDNSRuleTypes 注册 DNSRule 类型
func registerDNSRuleTypes(reflector *jsonschema.Reflector, defs jsonschema.Definitions) []*jsonschema.Schema {
	var oneOfSchemas []*jsonschema.Schema

	dnsRuleTypes := []struct {
		typeName string
		defName  string
		options  interface{}
	}{
		{constant.RuleTypeDefault, "DefaultDNSRule", &option.DefaultDNSRule{}},
		{constant.RuleTypeLogical, "LogicalDNSRule", &option.LogicalDNSRule{}},
	}

	for _, dnsRuleType := range dnsRuleTypes {
		typeSchema := reflector.Reflect(dnsRuleType.options)

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

		// 清理元数据
		typeSchema.ID = ""
		typeSchema.Version = ""

		// 合并子 $defs
		if typeSchema.Definitions != nil {
			for name, def := range typeSchema.Definitions {
				if _, exists := defs[name]; !exists {
					defs[name] = def
				}
			}
			typeSchema.Definitions = nil
		}

		defs[dnsRuleType.defName] = typeSchema
		oneOfSchemas = append(oneOfSchemas, &jsonschema.Schema{
			Ref: "#/$defs/" + dnsRuleType.defName,
		})
	}

	return oneOfSchemas
}

// cleanupNestedDefs 清理所有嵌套的 $defs
func cleanupNestedDefs(schema *jsonschema.Schema) {
	if schema == nil {
		return
	}

	// 递归处理 properties
	if schema.Properties != nil {
		for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
			cleanupNestedDefs(pair.Value)
		}
	}

	// 递归处理 oneOf
	for _, s := range schema.OneOf {
		cleanupNestedDefs(s)
	}

	// 递归处理 anyOf
	for _, s := range schema.AnyOf {
		cleanupNestedDefs(s)
	}

	// 递归处理 allOf
	for _, s := range schema.AllOf {
		cleanupNestedDefs(s)
	}

	// 递归处理 items
	if schema.Items != nil {
		cleanupNestedDefs(schema.Items)
	}

	// 递归处理 $defs 中的每个定义
	if schema.Definitions != nil {
		for _, def := range schema.Definitions {
			// 递归清理子定义中的嵌套 $defs
			cleanupNestedDefsInDef(def)
		}
	}
}

// cleanupNestedDefsInDef 清理定义中的嵌套 $defs（但不清理顶层 $defs）
func cleanupNestedDefsInDef(schema *jsonschema.Schema) {
	if schema == nil {
		return
	}

	// 清理当前 schema 的 $defs
	schema.Definitions = nil

	// 递归处理 properties
	if schema.Properties != nil {
		for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
			cleanupNestedDefsInDef(pair.Value)
		}
	}

	// 递归处理 oneOf/anyOf/allOf
	for _, s := range schema.OneOf {
		cleanupNestedDefsInDef(s)
	}
	for _, s := range schema.AnyOf {
		cleanupNestedDefsInDef(s)
	}
	for _, s := range schema.AllOf {
		cleanupNestedDefsInDef(s)
	}

	// 递归处理 items
	if schema.Items != nil {
		cleanupNestedDefsInDef(schema.Items)
	}
}

// normalizeDefNames 规范化 $defs 中的名称，移除特殊字符
func normalizeDefNames(schema *jsonschema.Schema) {
	if schema.Definitions == nil {
		return
	}

	// 构建名称映射：旧名称 -> 新名称
	nameMapping := make(map[string]string)
	for name := range schema.Definitions {
		normalizedName := normalizeDefName(name)
		if normalizedName != name {
			nameMapping[name] = normalizedName
		}
	}

	// 如果没有需要重命名的，直接返回
	if len(nameMapping) == 0 {
		return
	}

	// 重命名 $defs 中的 key
	newDefs := make(jsonschema.Definitions)
	for name, def := range schema.Definitions {
		newName := name
		if mapped, ok := nameMapping[name]; ok {
			newName = mapped
		}
		newDefs[newName] = def
	}
	schema.Definitions = newDefs

	// 更新所有 $ref 引用
	updateRefs(schema, nameMapping)
}

// normalizeDefName 规范化单个定义名称
func normalizeDefName(name string) string {
	// 替换特殊字符
	// Listable[net/netip.Prefix] -> ListableNetipPrefix
	// Listable[string] -> ListableString
	// Listable[uint32] -> ListableUint32

	result := name

	// 处理 Listable[xxx] 格式
	if strings.HasPrefix(result, "Listable[") && strings.HasSuffix(result, "]") {
		inner := result[9 : len(result)-1] // 提取 [] 中的内容

		// 提取最后一个类型名（去除包路径）
		// github.com/sagernet/sing-box/option.DNSQueryType -> DNSQueryType
		// net/netip.Prefix -> Prefix
		if lastDot := strings.LastIndex(inner, "."); lastDot != -1 {
			inner = inner[lastDot+1:]
		} else if lastSlash := strings.LastIndex(inner, "/"); lastSlash != -1 {
			inner = inner[lastSlash+1:]
		}

		// 首字母大写
		if len(inner) > 0 {
			inner = strings.ToUpper(inner[:1]) + inner[1:]
		}
		result = "Listable" + inner
	}

	// 移除所有特殊字符
	result = strings.ReplaceAll(result, "[", "")
	result = strings.ReplaceAll(result, "]", "")
	result = strings.ReplaceAll(result, "/", "")
	result = strings.ReplaceAll(result, ".", "")
	result = strings.ReplaceAll(result, "-", "")

	return result
}

// updateRefs 递归更新所有 $ref 引用
func updateRefs(schema *jsonschema.Schema, nameMapping map[string]string) {
	if schema == nil {
		return
	}

	// 更新当前 schema 的 $ref
	if schema.Ref != "" {
		for oldName, newName := range nameMapping {
			oldRef := "#/$defs/" + oldName
			newRef := "#/$defs/" + newName
			if schema.Ref == oldRef {
				schema.Ref = newRef
				break
			}
		}
	}

	// 递归更新 $defs 中的引用
	if schema.Definitions != nil {
		for _, def := range schema.Definitions {
			updateRefs(def, nameMapping)
		}
	}

	// 递归更新 properties
	if schema.Properties != nil {
		for pair := schema.Properties.Oldest(); pair != nil; pair = pair.Next() {
			updateRefs(pair.Value, nameMapping)
		}
	}

	// 递归更新 oneOf/anyOf/allOf
	for _, s := range schema.OneOf {
		updateRefs(s, nameMapping)
	}
	for _, s := range schema.AnyOf {
		updateRefs(s, nameMapping)
	}
	for _, s := range schema.AllOf {
		updateRefs(s, nameMapping)
	}

	// 递归更新 items
	if schema.Items != nil {
		updateRefs(schema.Items, nameMapping)
	}

	// 递归更新 additionalProperties
	if schema.AdditionalProperties != nil {
		updateRefs(schema.AdditionalProperties, nameMapping)
	}
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
		if optionsValue.Kind() == reflect.Pointer {
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
