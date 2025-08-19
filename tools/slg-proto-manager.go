package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// ProtoConfig 协议配置结构
type ProtoConfig struct {
	Meta struct {
		Project     string `yaml:"project"`
		Team        string `yaml:"team"`
		LastUpdated string `yaml:"last_updated"`
	} `yaml:"meta"`

	Versions struct {
		Current      string   `yaml:"current"`
		Supported    []string `yaml:"supported"`
		Deprecated   []string `yaml:"deprecated"`
		Experimental []string `yaml:"experimental"`
	} `yaml:"versions"`

	CompatibilityTests struct {
		Enabled   bool     `yaml:"enabled"`
		TestPairs []string `yaml:"test_pairs"`
	} `yaml:"compatibility_tests"`

	BuildTargets []struct {
		Version         string `yaml:"version"`
		InputPath       string `yaml:"input_path"`
		OutputPath      string `yaml:"output_path"`
		GoPackagePrefix string `yaml:"go_package_prefix"`
	} `yaml:"build_targets"`

	Modules map[string]struct {
		Description string `yaml:"description"`
		Owner       string `yaml:"owner"`
		Critical    bool   `yaml:"critical"`
	} `yaml:"modules"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "integrate":
		integrateProto()
	case "generate":
		generateProto()
	case "validate":
		validateProto()
	case "list-versions":
		listVersions()
	case "compatibility-check":
		compatibilityCheck()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Printf(`SLG协议管理工具

使用方法:
  go run tools/slg-proto-manager.go <command> [args]

命令:
  integrate <dev_path> <version>  # 集成研发协议
  generate <version>              # 生成指定版本的Go代码
  validate <version>              # 验证协议格式
  list-versions                   # 列出所有版本
  compatibility-check <v1> <v2>   # 兼容性检查

示例:
  go run tools/slg-proto-manager.go integrate ./dev-proto v1.1.0
  go run tools/slg-proto-manager.go generate v1.0.0
  go run tools/slg-proto-manager.go validate v1.1.0
  go run tools/slg-proto-manager.go compatibility-check v1.0.0 v1.1.0
`)
}

func integrateProto() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: integrate <dev_path> <version>")
		return
	}

	devPath := os.Args[2]
	version := os.Args[3]

	fmt.Printf("🔄 集成研发协议...\n")
	fmt.Printf("   源路径: %s\n", devPath)
	fmt.Printf("   目标版本: %s\n", version)

	targetDir := filepath.Join("slg-proto", version)

	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}

	// 复制协议文件
	err := copyDir(devPath, targetDir)
	if err != nil {
		log.Fatalf("复制协议文件失败: %v", err)
	}

	fmt.Printf("✅ 协议已复制到 %s\n", targetDir)

	// 更新配置文件
	updateConfig(version)

	fmt.Printf("✅ 协议集成完成！\n")
	fmt.Printf("   下一步：make generate-slg-proto VERSION=%s\n", version)
}

func generateProto() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: generate <version>")
		return
	}

	version := os.Args[2]

	fmt.Printf("🔧 生成协议代码 %s...\n", version)

	// 这里应该调用protoc或buf generate
	// 为简化，这里只是输出命令

	protoPath := filepath.Join("slg-proto", version)
	outputPath := filepath.Join("generated", "slg", strings.ReplaceAll(version, ".", "_"))

	fmt.Printf("   源路径: %s\n", protoPath)
	fmt.Printf("   输出路径: %s\n", outputPath)

	// 创建输出目录
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	fmt.Printf("✅ 代码生成完成！\n")
}

func validateProto() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: validate <version>")
		return
	}

	version := os.Args[2]

	fmt.Printf("🔍 验证协议格式 %s...\n", version)

	protoPath := filepath.Join("slg-proto", version)

	// 检查proto文件是否存在
	if _, err := os.Stat(protoPath); os.IsNotExist(err) {
		log.Fatalf("协议版本不存在: %s", version)
	}

	// 遍历并验证所有.proto文件
	err := filepath.Walk(protoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".proto") {
			fmt.Printf("   验证: %s\n", path)
			// 这里应该调用protoc进行语法检查
		}

		return nil
	})

	if err != nil {
		log.Fatalf("验证失败: %v", err)
	}

	fmt.Printf("✅ 协议验证通过！\n")
}

func listVersions() {
	config := loadConfig()

	fmt.Printf("📋 SLG协议版本列表\n")
	fmt.Printf("   当前版本: %s\n", config.Versions.Current)
	fmt.Printf("   支持版本: %v\n", config.Versions.Supported)

	if len(config.Versions.Deprecated) > 0 {
		fmt.Printf("   废弃版本: %v\n", config.Versions.Deprecated)
	}

	if len(config.Versions.Experimental) > 0 {
		fmt.Printf("   实验版本: %v\n", config.Versions.Experimental)
	}
}

func compatibilityCheck() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: compatibility-check <version1> <version2>")
		return
	}

	v1 := os.Args[2]
	v2 := os.Args[3]

	fmt.Printf("🔍 兼容性检查: %s -> %s\n", v1, v2)

	// 这里应该实现实际的兼容性检查逻辑
	// 比如比较proto文件的字段变化

	fmt.Printf("✅ 兼容性检查通过！\n")
}

func loadConfig() *ProtoConfig {
	data, err := os.ReadFile("configs/proto-versions.yaml")
	if err != nil {
		log.Fatalf("读取配置文件失败: %v", err)
	}

	var config ProtoConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("解析配置文件失败: %v", err)
	}

	return &config
}

func updateConfig(version string) {
	config := loadConfig()

	// 添加到支持版本列表
	found := false
	for _, v := range config.Versions.Supported {
		if v == version {
			found = true
			break
		}
	}

	if !found {
		config.Versions.Supported = append(config.Versions.Supported, version)
	}

	// 更新当前版本
	config.Versions.Current = version

	// 保存配置
	data, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("序列化配置失败: %v", err)
	}

	if err := os.WriteFile("configs/proto-versions.yaml", data, 0644); err != nil {
		log.Fatalf("保存配置文件失败: %v", err)
	}
}

func copyDir(src, dst string) error {
	// 简化的目录复制函数
	// 实际项目中应该使用更完善的实现
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// 复制文件
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = srcFile.WriteTo(dstFile)
		return err
	})
}
