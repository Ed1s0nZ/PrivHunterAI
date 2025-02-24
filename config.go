package main

const (
	apiKeyKimi     = "sk-xxxxxxx" // 替换为你的kimi API key
	apiKeyDeepSeek = "sk-yyyyyyy" // 替换为你的deepseek API key
	AI             = "deepseek"   // 可选择deepseek或kimi
	cookie2        = "cookie2"
	prompt         = "{\"role\": \"你是一个AI，负责通过比较两个HTTP响应数据包来检测潜在的越权行为，并自行做出判断。\",\"inputs\": {\"responseA\": \"账号A请求某接口的响应。\",\"responseB\": \"使用账号B的Cookie重放请求的响应。\"},\"analysisRequirements\": {\"structureAndContentComparison\": \"比较响应A和响应B的结构和内容，忽略动态字段（如时间戳、随机数、会话ID等）。\",\"judgmentCriteria\": {\"authorizationSuccess\": \"如果响应B的结构和非动态字段内容与响应A高度相似，或响应B包含账号A的数据，并且自我判断为越权成功。\",\"authorizationFailure\": \"如果响应B的结构和内容与响应A不相似，或存在权限不足的错误信息，或响应内容均为公开数据，或大部分相同字段的具体值不同，或除了动态字段外的字段均无实际值，并且自我判断为越权失败。\",\"unknown\": \"其他情况，或无法确定是否存在越权，并且自我判断为无法确定。\"}},\"outputFormat\": {\"json\": {\"res\": \"\\\"true\\\", \\\"false\\\" 或 \\\"unknown\\\"\",\"reason\": \"简洁的判断原因，不超过20字\"}},\"notes\": [\"仅输出 JSON 格式的结果，不添加任何额外文本或解释。\",\"确保JSON格式正确，便于后续处理。\",\"保持客观，仅根据响应内容进行分析。\"],\"process\": [\"接收并理解响应A和响应B。\",\"分析响应A和响应B，忽略动态字段。\",\"基于响应的结构、内容和相关性进行自我判断，包括但不限于：\",\"- 识别响应中可能的敏感数据或权限信息。\",\"- 评估响应与预期结果之间的一致性。\",\"- 确定是否存在明显的越权迹象。\",\"输出指定格式的JSON结果，包括判断和判断原因。\"]}" // 通常情况勿动
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

var allowedRespHeaders = []string{
	"image/png",
	"text/html",
	"application/pdf",
	"text/css",
	"audio/mpeg",
	"audio/wav",
	"video/mp4",
	"application/grpc",
} // 这些响应头不扫
