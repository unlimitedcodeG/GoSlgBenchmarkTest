package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// SLGProtocolConfig 协议配置结构
type SLGProtocolConfig struct {
	Versions map[string]struct {
		Description string `yaml:"description"`
		Owner       string `yaml:"owner"`
		Critical    bool   `yaml:"critical"`
		Modules     map[string]struct {
			Description string `yaml:"description"`
			Owner       string `yaml:"owner"`
			Critical    bool   `yaml:"critical"`
		} `yaml:"modules"`
	} `yaml:"versions"`
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
  go run tools/slg-proto-manager/main.go <command> [args]

命令:
  integrate <dev_path> <version>  # 集成研发协议
  generate <version>              # 生成指定版本的Go代码
  validate <version>              # 验证协议格式
  list-versions                   # 列出所有版本
  compatibility-check <v1> <v2>   # 兼容性检查

示例:
  go run tools/slg-proto-manager/main.go integrate ./dev-proto v1.1.0
  go run tools/slg-proto-manager/main.go generate v1.0.0
  go run tools/slg-proto-manager/main.go validate v1.1.0
  go run tools/slg-proto-manager/main.go compatibility-check v1.0.0 v1.1.0
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
	fmt.Printf("   使用命令: buf generate slg-proto/%s\n", version)
	fmt.Printf("   输出目录: generated/slg/%s\n", version)
}

func validateProto() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: validate <version>")
		return
	}

	version := os.Args[2]
	protoDir := filepath.Join("slg-proto", version)

	fmt.Printf("🔍 验证协议格式 %s...\n", version)
	fmt.Printf("   协议目录: %s\n", protoDir)

	if _, err := os.Stat(protoDir); os.IsNotExist(err) {
		log.Fatalf("协议目录不存在: %s", protoDir)
	}

	fmt.Printf("✅ 协议目录存在\n")
	fmt.Printf("   使用命令: buf lint slg-proto/%s\n", version)
}

func listVersions() {
	fmt.Printf("📋 可用的SLG协议版本:\n")

	slgProtoDir := "slg-proto"
	if _, err := os.Stat(slgProtoDir); os.IsNotExist(err) {
		fmt.Printf("   未找到协议目录: %s\n", slgProtoDir)
		return
	}

	entries, err := os.ReadDir(slgProtoDir)
	if err != nil {
		log.Fatalf("读取协议目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			fmt.Printf("   - %s\n", entry.Name())
		}
	}
}

func compatibilityCheck() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: compatibility-check <v1> <v2>")
		return
	}

	v1 := os.Args[2]
	v2 := os.Args[3]

	fmt.Printf("🔍 检查协议兼容性 %s -> %s...\n", v1, v2)
	fmt.Printf("   使用命令: buf breaking --against slg-proto/%s slg-proto/%s\n", v1, v2)
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}

func updateConfig(version string) {
	configFile := "configs/proto-versions.yaml"

	// 读取现有配置
	var config SLGProtocolConfig
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			yaml.Unmarshal(data, &config)
		}
	}

	// 初始化配置
	if config.Versions == nil {
		config.Versions = make(map[string]struct {
			Description string `yaml:"description"`
			Owner       string `yaml:"owner"`
			Critical    bool   `yaml:"critical"`
			Modules     map[string]struct {
				Description string `yaml:"description"`
				Owner       string `yaml:"owner"`
				Critical    bool   `yaml:"critical"`
			} `yaml:"modules"`
		})
	}

	// 添加新版本
	config.Versions[version] = struct {
		Description string `yaml:"description"`
		Owner       string `yaml:"owner"`
		Critical    bool   `yaml:"critical"`
		Modules     map[string]struct {
			Description string `yaml:"description"`
			Owner       string `yaml:"owner"`
			Critical    bool   `yaml:"critical"`
		} `yaml:"modules"`
	}{
		Description: fmt.Sprintf("SLG协议版本 %s", version),
		Owner:       "SLG团队",
		Critical:    true,
		Modules: make(map[string]struct {
			Description string `yaml:"description"`
			Owner       string `yaml:"owner"`
			Critical    bool   `yaml:"critical"`
		}),
	}

	// 写入配置
	data, err := yaml.Marshal(&config)
	if err != nil {
		log.Printf("序列化配置失败: %v", err)
		return
	}

	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		log.Printf("创建配置目录失败: %v", err)
		return
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		log.Printf("写入配置文件失败: %v", err)
		return
	}

	fmt.Printf("✅ 配置文件已更新: %s\n", configFile)
}
