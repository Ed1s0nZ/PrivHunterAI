package domain

import "time"

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Disabled bool   `json:"disabled"`
}

type Finding struct {
	ID         int64     `json:"id"`
	Method     string    `json:"method"`
	URL        string    `json:"url"`
	RequestA   string    `json:"requestA,omitempty"`
	RequestB   string    `json:"requestB,omitempty"`
	ResponseA  string    `json:"responseA,omitempty"`
	ResponseB  string    `json:"responseB,omitempty"`
	Result     string    `json:"result"`
	Reason     string    `json:"reason"`
	Confidence string    `json:"confidence"`
	Review     string    `json:"review"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Settings struct {
	Revision     int64             `json:"revision"`
	Paused       bool              `json:"paused"`
	Domains      []string          `json:"domains"`
	Headers      map[string]string `json:"-"`
	Endpoint     string            `json:"endpoint"`
	Model        string            `json:"model"`
	APIKey       string            `json:"-"`
	SkipSuffixes []string          `json:"skipSuffixes"`
	DenyKeywords []string          `json:"denyKeywords"`
}

type Audit struct {
	ID        int64  `json:"id"`
	Actor     string `json:"actor"`
	Action    string `json:"action"`
	Target    string `json:"target"`
	CreatedAt string `json:"createdAt"`
}
