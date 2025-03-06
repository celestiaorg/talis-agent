package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type IPResponse struct {
	IP string `json:"ip"`
}

func IPHandler(c *fiber.Ctx) error {
	// Try to get IP from ipify.org
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return c.JSON(fiber.Map{
			"ips": []string{},
		})
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(fiber.Map{
			"ips": []string{},
		})
	}

	var ipResp IPResponse
	if err := json.Unmarshal(body, &ipResp); err != nil {
		return c.JSON(fiber.Map{
			"ips": []string{},
		})
	}

	return c.JSON(fiber.Map{
		"ips": []string{ipResp.IP},
	})
}
