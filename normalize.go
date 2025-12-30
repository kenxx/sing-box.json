package main

import (
	"strings"

	"github.com/kenxx/sing-box.json/schema"
)

// normalizeDefNames 规范化 $defs 中的名称，移除特殊字符
func normalizeDefNames(s *schema.Schema) {
	if s.Definitions == nil {
		return
	}

	// 构建名称映射：旧名称 -> 新名称
	nameMapping := make(map[string]string)
	for name := range s.Definitions {
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
	newDefs := make(map[string]*schema.Schema)
	for name, def := range s.Definitions {
		newName := name
		if mapped, ok := nameMapping[name]; ok {
			newName = mapped
		}
		newDefs[newName] = def
	}
	s.Definitions = newDefs

	// 更新所有 $ref 引用
	updateRefs(s, nameMapping)
}

// normalizeDefName 规范化单个定义名称
func normalizeDefName(name string) string {
	result := name

	// 处理 Listable[xxx] 格式
	if strings.HasPrefix(result, "Listable[") && strings.HasSuffix(result, "]") {
		inner := result[9 : len(result)-1]

		// 提取最后一个类型名（去除包路径）
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
	result = strings.ReplaceAll(result, "*", "")

	return result
}

// updateRefs 递归更新所有 $ref 引用
func updateRefs(s *schema.Schema, nameMapping map[string]string) {
	if s == nil {
		return
	}

	// 更新当前 schema 的 $ref
	if s.Ref != "" {
		for oldName, newName := range nameMapping {
			oldRef := "#/$defs/" + oldName
			newRef := "#/$defs/" + newName
			if s.Ref == oldRef {
				s.Ref = newRef
				break
			}
		}
	}

	// 递归更新 $defs 中的引用
	if s.Definitions != nil {
		for _, def := range s.Definitions {
			updateRefs(def, nameMapping)
		}
	}

	// 递归更新 properties
	if s.Properties != nil {
		for _, prop := range s.Properties {
			updateRefs(prop, nameMapping)
		}
	}

	// 递归更新 oneOf/anyOf/allOf
	for _, sub := range s.OneOf {
		updateRefs(sub, nameMapping)
	}
	for _, sub := range s.AnyOf {
		updateRefs(sub, nameMapping)
	}
	for _, sub := range s.AllOf {
		updateRefs(sub, nameMapping)
	}

	// 递归更新 items
	if s.Items != nil {
		updateRefs(s.Items, nameMapping)
	}

	// 递归更新 additionalProperties
	if addProps, ok := s.AdditionalProperties.(*schema.Schema); ok {
		updateRefs(addProps, nameMapping)
	}
}

