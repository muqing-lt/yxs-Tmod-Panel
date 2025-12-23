package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"terra-panel/pkg/process" // 核心逻辑包
	"terra-panel/pkg/ws"      // WebSocket包

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3" 
)

// --- 嵌入前端静态文件 ---
//go:embed all:frontend/dist
var frontendAssets embed.FS

const ConfigFile = "config.yaml"

// --- 配置结构 ---
type AppConfig struct {
	User *UserConfig `yaml:"user,omitempty" json:"user,omitempty"`
}

type UserConfig struct {
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
}

// --- 配置管理器 ---
type ConfigManager struct {
	mu     sync.RWMutex
	Config AppConfig
}

var cm = &ConfigManager{Config: AppConfig{}}

// 加载配置
func (c *ConfigManager) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := os.ReadFile(ConfigFile)
	if os.IsNotExist(err) {
		return nil
	}
	return yaml.Unmarshal(data, &c.Config)
}

// 保存配置
func (c *ConfigManager) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	data, err := yaml.Marshal(&c.Config)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile, data, 0644)
}

func main() {
	// 1. 初始化核心模块
	process.Init()
	if err := cm.Load(); err != nil {
		log.Printf("配置文件加载警告: %v", err)
	}

	r := gin.Default()

	// 2. 开放 API (无鉴权)
	r.GET("/api/ws", ws.Handler) // WebSocket 日志流

	// 初始化检查 (前端判断是否显示Setup页面)
	r.GET("/api/init_check", func(c *gin.Context) {
		cm.mu.RLock()
		initialized := cm.Config.User != nil
		cm.mu.RUnlock()
		c.JSON(200, gin.H{"initialized": initialized})
	})

	// 首次设置 (注册管理员)
	r.POST("/api/setup", func(c *gin.Context) {
		var user UserConfig
		if err := c.BindJSON(&user); err != nil {
			c.JSON(400, gin.H{"error": "参数无效"})
			return
		}
		if user.Username == "" || user.Password == "" {
			c.JSON(400, gin.H{"error": "用户名或密码不能为空"})
			return
		}

		cm.mu.Lock()
		cm.Config.User = &user
		if err := cm.Save(); err != nil {
			cm.mu.Unlock()
			c.JSON(500, gin.H{"error": "配置文件保存失败"})
			return
		}
		cm.mu.Unlock()

		c.JSON(200, gin.H{"message": "初始化完成，请登录"})
	})

	// 登录接口
	r.POST("/api/login", func(c *gin.Context) {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数无效"})
			return
		}

		cm.mu.RLock()
		user := cm.Config.User
		cm.mu.RUnlock()

		if user == nil {
			c.JSON(403, gin.H{"error": "系统未初始化"})
			return
		}

		if req.Username == user.Username && req.Password == user.Password {
			// 生成简易 Token (实际生产建议用 JWT)
			token := fmt.Sprintf("tk-%d-%s", time.Now().Unix(), req.Username)
			c.JSON(200, gin.H{
				"token":    token,
				"username": user.Username,
			})
		} else {
			c.JSON(401, gin.H{"error": "用户名或密码错误"})
		}
	})

	// 3. 受保护 API (需要 Token)
	auth := r.Group("/api")
	auth.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "未登录"})
			return
		}
		c.Next()
	})

	{
		// --- 基础服务器控制 ---
		auth.GET("/stats", func(c *gin.Context) {
			c.JSON(200, process.GetSystemStats())
		})
		auth.POST("/start", func(c *gin.Context) {
			if err := process.StartServer(); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
			} else {
				c.JSON(200, gin.H{"message": "启动指令已发送"})
			}
		})
		auth.POST("/stop", func(c *gin.Context) {
			if err := process.StopServer(); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
			} else {
				c.JSON(200, gin.H{"message": "停止指令已发送"})
			}
		})
		auth.POST("/command", func(c *gin.Context) {
			var req struct {
				Cmd string `json:"cmd"`
			}
			c.BindJSON(&req)
			process.SendCommand(req.Cmd)
			c.Status(200)
		})

		// --- 服务器参数与配置 ---
		auth.POST("/server_params", func(c *gin.Context) {
			var p process.ServerParams
			if err := c.BindJSON(&p); err == nil {
				process.SaveParams(p)
				c.Status(200)
			} else {
				c.JSON(400, gin.H{"error": "参数格式错误"})
			}
		})
		auth.GET("/server_config", func(c *gin.Context) {
			file := c.Query("file")
			if file == "" {
				file = "serverconfig.txt"
			}
			data, err := process.GetServerConfigFile(file)
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
			} else {
				c.JSON(200, data)
			}
		})
		auth.POST("/server_config", func(c *gin.Context) {
			var req struct {
				Filename string            `json:"filename"`
				Data     map[string]string `json:"data"`
			}
			if err := c.BindJSON(&req); err == nil {
				process.SaveServerConfigFile(req.Filename, req.Data)
				c.Status(200)
			}
		})

		// --- 权限管理 (黑名单) ---
		auth.GET("/banlist", func(c *gin.Context) {
			list, err := process.GetBanList()
			if err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
			} else {
				c.JSON(200, list)
			}
		})
		auth.POST("/banlist", func(c *gin.Context) {
			var entry process.BanEntry
			if err := c.BindJSON(&entry); err == nil {
				if err := process.AddBan(entry); err != nil {
					c.JSON(500, gin.H{"error": err.Error()})
				} else {
					c.Status(200)
				}
			}
		})
		auth.DELETE("/banlist", func(c *gin.Context) {
			target := c.Query("target")
			if err := process.RemoveBan(target); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
			} else {
				c.Status(200)
			}
		})

		// --- Hero's Mod 管理 ---
		auth.GET("/heros/status", func(c *gin.Context) {
			c.JSON(200, gin.H{"installed": process.CheckHerosMod()})
		})
		auth.POST("/heros/install", func(c *gin.Context) {
			if err := process.InstallHerosMod(); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
			} else {
				c.JSON(200, gin.H{"status": "installed"})
			}
		})
	}

	// 4. 前端静态文件托管
	// 必须先在 frontend 目录下执行 npm run build，并把 dist 复制到 backend/frontend/dist
	subFS, err := fs.Sub(frontendAssets, "frontend/dist")
	if err != nil {
		log.Fatalf("前端文件加载失败 (请确保 backend/frontend/dist 存在): %v", err)
	}
	r.NoRoute(gin.WrapH(http.FileServer(http.FS(subFS))))

	// 启动服务
	fmt.Println("面板已启动: http://localhost:8080")
	r.Run(":8080")
}