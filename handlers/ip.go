package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// IPResponse represents the response from the IP service
type IPResponse struct {
	IP string `json:"ip"`
}

// IPHandler handles IP address requests
func IPHandler(c *fiber.Ctx) error {
	// Try to get IP from ipify.org
	resp, err := http.Get("https://api.ipify.org?format=json")
	if err != nil {
		return c.JSON(fiber.Map{
			"ips": []string{},
		})
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

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
