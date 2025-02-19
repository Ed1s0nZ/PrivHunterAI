package main

const (
	apiKeyKimi     = "sk-xxxxxxx" // 替换为你的kimi API key
	apiKeyDeepSeek = "sk-yyyyyyy"                 // 替换为你的deepseek API key
	AI             = "deepseek"                                            // 可选择deepseek或kimi
	cookie2        = "cookie2"
)

var suffixes = []string{
	// 静态资源文件
	".js", ".ico", ".png", ".jpg", ".jpeg", ".gif", ".bmp", ".svg", ".ttf", ".woff", ".woff2", ".eot", ".otf",
	".mp3", ".wav", ".ogg", ".mp4", ".webm", ".avi", ".css", ".scss", ".less",

	// 临时文件和缓存文件
	".tmp", ".temp", ".cache", ".swp",

	// 日志文件
	".log", ".err",

	// 配置文件
	".env", ".json", ".yml", ".yaml", ".xml", ".ini",

	// 编译生成文件
	".class", ".dll", ".so", ".zip", ".tar", ".gz",

	// 其他常见文件
	".txt", ".md", ".doc", ".docx", ".pdf", ".csv", ".xlsx", ".xls", ".sh", ".bat",

	// 特殊用途文件
	".gitignore", ".lock", ".bak",
} //这些后缀不扫
