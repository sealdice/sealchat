package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type Permission struct {
	Key  string
	Desc string
}

func processFile(filePath string) []Permission {
	// 读取权限文件
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var perms []Permission

	// 正则表达式匹配权限定义
	permRegex := regexp.MustCompile(`NewStdPermission\("([^"]+)"\).*\/\/\s*(.+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := permRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			perms = append(perms, Permission{
				Key:  matches[1],
				Desc: strings.TrimSpace(matches[2]),
			})
		}
	}

	return perms
}

type InputInfo struct {
	Files     []string
	OutFile   string
	TypeName  string
	ArrayName string
}

func solveOne(info *InputInfo, tmpl *template.Template) {
	// 接受多个文件路径作为输入
	files := info.Files

	var allPerms []Permission

	// 处理每个输入文件
	for _, file := range files {
		perms := processFile(file)
		allPerms = append(allPerms, perms...)
	}

	// 生成新文件
	outFn := info.OutFile
	if outFn == "" {
		fileName := filepath.Base(files[0])
		ext := filepath.Ext(fileName)

		// 获取不带扩展名的文件名
		nameWithoutExt := strings.TrimSuffix(fileName, ext)
		// 生成新的文件名
		newFileName := nameWithoutExt + "_generated" + ext

		outFn = filepath.Join("./pm/gen", newFileName)
	}

	output, err := os.Create(outFn)
	if err != nil {
		panic(err)
	}
	defer output.Close()

	if info.ArrayName == "" {
		base := info.TypeName
		// 如果TypeName以Map结尾,去掉Map后加上Array
		if strings.HasSuffix(info.TypeName, "Map") {
			base = strings.TrimSuffix(info.TypeName, "Map")
		}
		info.ArrayName = base + "Array"
	}

	// 解析并执行模板
	if err := tmpl.Execute(output, map[string]interface{}{
		"Perms":     allPerms,
		"MapName":   info.TypeName,
		"ArrayName": info.ArrayName,
	}); err != nil {
		panic(err)
	}

}

var data = []*InputInfo{
	{
		Files:    []string{"./pm/perm_channel.go"},
		OutFile:  "",
		TypeName: "PermChannelMap",
	},
	{
		Files:    []string{"./pm/perm_system.go"},
		OutFile:  "",
		TypeName: "PermSystemMap",
	},
}

var data2 = []*InputInfo{
	{
		Files:    []string{"./pm/perm_channel.go"},
		OutFile:  "./ui/src/types-perm-channel.ts",
		TypeName: "ChannelRolePermSheet",
	},
	{
		Files:    []string{"./pm/perm_system.go"},
		OutFile:  "./ui/src/types-perm-system.ts",
		TypeName: "SystemRolePermSheet",
	},
}

func main() {
	for _, i := range data {
		solveOne(i, tmplGo)
	}

	for _, i := range data2 {
		solveOne(i, tmplTs)
	}
}
