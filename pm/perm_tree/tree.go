package perm_tree

import (
	"sealchat/pm/gen"
	"strings"
)

// PermTreeNode 定义权限树节点结构
type PermTreeNode struct {
	Name      string         `json:"name"`
	ModelName string         `json:"modelName,omitempty"`
	Children  []PermTreeNode `json:"children,omitempty"`
}

type OneItem = []map[string]string

// BuildPermTree 构建权限树
func BuildPermTree(selectedKeys []string, permArrayLst []OneItem) []PermTreeNode {
	// 创建根节点映射
	treeMap := make(map[string]PermTreeNode)

	// 创建selectedKeys的map用于快速查询
	selectedMap := make(map[string]bool)
	for _, key := range selectedKeys {
		selectedMap[key] = true
	}

	// 遍历权限映射
	for _, permArr := range permArrayLst {
		for _, m := range permArr {
			perm := m["key"]
			desc := m["desc"]
			// 如果 selectedKeys 为空,显示所有权限;否则只显示选中的权限
			if len(selectedKeys) > 0 && !selectedMap[perm] {
				continue
			}

			parts := strings.Split(desc, " - ")
			if len(parts) != 3 {
				continue
			}

			cat1 := strings.TrimSpace(parts[0])
			cat2 := strings.TrimSpace(parts[1])
			cat3 := strings.TrimSpace(parts[2])

			// 确保一级分类存在
			if _, exists := treeMap[cat1]; !exists {
				treeMap[cat1] = PermTreeNode{
					Name:     cat1,
					Children: []PermTreeNode{},
				}
			}

			// 在一级分类下查找/创建二级分类
			cat1Node := treeMap[cat1]
			var cat2Node *PermTreeNode
			for i := range cat1Node.Children {
				if cat1Node.Children[i].Name == cat2 {
					cat2Node = &cat1Node.Children[i]
					break
				}
			}
			if cat2Node == nil {
				cat1Node.Children = append(cat1Node.Children, PermTreeNode{
					Name:     cat2,
					Children: []PermTreeNode{},
				})
				cat2Node = &cat1Node.Children[len(cat1Node.Children)-1]
			}

			// 添加叶子节点
			cat2Node.Children = append(cat2Node.Children, PermTreeNode{
				Name:      cat3,
				ModelName: perm,
			})

			treeMap[cat1] = cat1Node
		}
	}

	// 将map转换为切片
	result := make([]PermTreeNode, 0, len(treeMap))
	for _, node := range treeMap {
		result = append(result, node)
	}

	return result
}

var PermTreeChannel = BuildPermTree(nil, []OneItem{gen.PermChannelArray})
var PermTreeSystem = BuildPermTree(nil, []OneItem{gen.PermSystemArray})

// func main() {
// 	ret := BuildPermTree(nil, []map[string]string{gen.PermChannelMap})
// 	jsonBytes, err := json.MarshalIndent(ret, "", "  ")
// 	if err != nil {
// 		fmt.Println("JSON marshal error:", err)
// 		return
// 	}
// 	fmt.Println(string(jsonBytes))
// }
