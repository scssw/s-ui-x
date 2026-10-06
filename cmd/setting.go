package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/deposist/s-ui-x/config"
	"github.com/deposist/s-ui-x/database"
	"github.com/deposist/s-ui-x/database/model"
	"github.com/deposist/s-ui-x/service"

	"github.com/shirou/gopsutil/v4/net"
)

func resetSetting() {
	err := database.InitDB(config.GetDBPath())
	if err != nil {
		fmt.Println(err)
		return
	}

	settingService := service.SettingService{}
	err = settingService.ResetSettings()
	if err != nil {
		fmt.Println("reset setting failed:", err)
	} else {
		fmt.Println("reset setting success")
	}
}

func clearWebDomain() {
	err := database.InitDB(config.GetDBPath())
	if err != nil {
		fmt.Println(err)
		return
	}

	settingService := service.SettingService{}
	if err := settingService.ClearWebDomainAndAddress(); err != nil {
		fmt.Println("clear panel domain and address failed:", err)
		return
	}
	fmt.Println("clear panel domain and address success")
	showSetting()
}

func updateSetting(port int, path string, subPort int, subPath string, domain string) {
	err := database.InitDB(config.GetDBPath())
	if err != nil {
		fmt.Println(err)
		return
	}

	settingService := service.SettingService{}

	if port > 0 {
		err := settingService.SetPort(port)
		if err != nil {
			fmt.Println("set port failed:", err)
		} else {
			fmt.Println("set port success")
		}
	}
	if path != "" {
		err := settingService.SetWebPath(path)
		if err != nil {
			fmt.Println("set path failed:", err)
		} else {
			fmt.Println("set path success")
		}
	}
	if subPort > 0 {
		err := settingService.SetSubPort(subPort)
		if err != nil {
			fmt.Println("set sub port failed:", err)
		} else {
			fmt.Println("set sub port success")
		}
	}
	if subPath != "" {
		err := settingService.SetSubPath(subPath)
		if err != nil {
			fmt.Println("set sub path failed:", err)
		} else {
			fmt.Println("set sub path success")
		}
	}
	if domain != "" {
		cert := "/root/cert/" + domain + "/fullchain.pem"
		key := "/root/cert/" + domain + "/privkey.pem"
		if path == "" {
			path, err = settingService.GetWebPath()
			if err != nil {
				fmt.Println("get panel path failed:", err)
				return
			}
		}
		if err := settingService.SetWebDomain(domain); err != nil {
			fmt.Println("set domain failed:", err)
			return
		}
		if err := settingService.SetWebURI("https://" + domain + path); err != nil {
			fmt.Println("set panel URI failed:", err)
			return
		}
		if err := settingService.SetWebCertFiles(cert, key); err != nil {
			fmt.Println("set domain certificate paths failed:", err)
			return
		}
		if err := ensureDomainTlsTemplate(domain, cert, key); err != nil {
			fmt.Println("create domain TLS template failed:", err)
			return
		}
		fmt.Println("set panel domain and certificate paths success")
	}
}

func ensureDomainTlsTemplate(domain string, certPath string, keyPath string) error {
	db := database.GetDB()
	var configs []model.Tls
	if err := db.Where("id > 0").Find(&configs).Error; err != nil {
		return err
	}
	for _, config := range configs {
		var server struct {
			ServerName      string `json:"server_name"`
			CertificatePath string `json:"certificate_path"`
			KeyPath         string `json:"key_path"`
		}
		if json.Unmarshal(config.Server, &server) == nil && server.ServerName == domain && server.CertificatePath != "" && server.KeyPath != "" {
			return nil
		}
	}

	prefix := strings.ToLower(strings.Split(strings.TrimSpace(domain), ".")[0])
	prefix = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, prefix)
	if prefix == "" {
		prefix = "tls"
	}
	name := prefix
	for suffix := 2; ; suffix++ {
		var count int64
		if err := db.Model(&model.Tls{}).Where("name = ?", name).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			break
		}
		name = fmt.Sprintf("%s-%d", prefix, suffix)
	}
	server, err := json.Marshal(map[string]any{
		"enabled": true, "server_name": domain,
		"certificate_path": certPath, "key_path": keyPath,
	})
	if err != nil {
		return err
	}
	return db.Create(&model.Tls{
		Name: name, Server: json.RawMessage(server), Client: json.RawMessage(`{}`),
	}).Error
}

func showSetting() {
	err := database.InitDB(config.GetDBPath())
	if err != nil {
		fmt.Println(err)
		return
	}
	settingService := service.SettingService{}
	allSetting, err := settingService.GetAllSetting()
	if err != nil {
		fmt.Println("get current port failed,error info:", err)
	}
	fmt.Println("Current panel settings:")
	fmt.Println("\tPanel port:\t", (*allSetting)["webPort"])
	fmt.Println("\tPanel path:\t", (*allSetting)["webPath"])
	if (*allSetting)["webListen"] != "" {
		fmt.Println("\tPanel IP:\t", (*allSetting)["webListen"])
	}
	if (*allSetting)["webDomain"] != "" {
		fmt.Println("\tPanel Domain:\t", (*allSetting)["webDomain"])
	}
	if (*allSetting)["webURI"] != "" {
		fmt.Println("\tPanel URI:\t", (*allSetting)["webURI"])
	}
	fmt.Println()
	fmt.Println("Current subscription settings:")
	fmt.Println("\tSub port:\t", (*allSetting)["subPort"])
	fmt.Println("\tSub path:\t", (*allSetting)["subPath"])
	if (*allSetting)["subListen"] != "" {
		fmt.Println("\tSub IP:\t", (*allSetting)["subListen"])
	}
	if (*allSetting)["subDomain"] != "" {
		fmt.Println("\tSub Domain:\t", (*allSetting)["subDomain"])
	}
	if (*allSetting)["subURI"] != "" {
		fmt.Println("\tSub URI:\t", (*allSetting)["subURI"])
	}
}

func getPublicIP() string {
	apis := []string{
		"https://api64.ipify.org",
		"https://ip.sb",
		"https://icanhazip.com",
		"https://ipinfo.io/ip",
		"https://checkip.amazonaws.com",
	}
	type result struct {
		ip  string
		err error
	}
	ch := make(chan result, len(apis))
	var wg sync.WaitGroup
	client := &http.Client{Timeout: 3 * time.Second}

	for _, api := range apis {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
			if err != nil {
				ch <- result{"", err}
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				ch <- result{"", err}
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
			if err != nil {
				ch <- result{"", err}
				return
			}
			ch <- result{string(body), nil}
		}(api)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for res := range ch {
		if res.err == nil && res.ip != "" {
			return strings.TrimSpace(res.ip)
		}
	}
	return ""
}

func getPanelURI() {
	err := database.InitDB(config.GetDBPath())
	if err != nil {
		fmt.Println(err)
		return
	}
	settingService := service.SettingService{}
	Port, _ := settingService.GetPort()
	BasePath, _ := settingService.GetWebPath()
	Listen, _ := settingService.GetListen()
	Domain, _ := settingService.GetWebDomain()
	KeyFile, _ := settingService.GetKeyFile()
	CertFile, _ := settingService.GetCertFile()
	TLS := false
	if KeyFile != "" && CertFile != "" {
		TLS = true
	}
	Proto := ""
	if TLS {
		Proto = "https://"
	} else {
		Proto = "http://"
	}
	PortText := fmt.Sprintf(":%d", Port)
	if (Port == 443 && TLS) || (Port == 80 && !TLS) {
		PortText = ""
	}
	if len(Domain) > 0 {
		fmt.Println(Proto + Domain + PortText + BasePath)
		return
	}
	if len(Listen) > 0 {
		fmt.Println(Proto + Listen + PortText + BasePath)
		return
	}
	fmt.Println("Local address:")
	netInterfaces, _ := net.Interfaces()
	for i := 0; i < len(netInterfaces); i++ {
		if len(netInterfaces[i].Flags) > 2 && netInterfaces[i].Flags[0] == "up" && netInterfaces[i].Flags[1] != "loopback" {
			addrs := netInterfaces[i].Addrs
			for _, address := range addrs {
				IP := strings.Split(address.Addr, "/")[0]
				if strings.Contains(address.Addr, ".") {
					fmt.Println(Proto + IP + PortText + BasePath)
				} else if address.Addr[0:6] != "fe80::" {
					fmt.Println(Proto + "[" + IP + "]" + PortText + BasePath)
				}
			}
		}
	}
	pubIP := getPublicIP()
	if pubIP != "" {
		fmt.Printf("\nGlobal address:\n%s%s%s\n", Proto, pubIP, PortText+BasePath)
	}
}
