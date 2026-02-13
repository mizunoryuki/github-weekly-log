package github

var LANGUAGE_MAP = map[string]string{
	".go":           "Go",
	".js":           "JavaScript",
	".jsx":          "JavaScript",
	".ts":           "TypeScript",
	".tsx":          "TypeScript",
	".py":           "Python",
	".java":         "Java",
	".rb":           "Ruby",
	".php":          "PHP",
	".c":            "C",
	".cpp":          "C++",
	".cc":           "C++",
	".cxx":          "C++",
	".cs":           "C#",
	".rs":           "Rust",
	".swift":        "Swift",
	".kt":           "Kotlin",
	".kts":          "Kotlin",
	".scala":        "Scala",
	".dart":         "Dart",
	".html":         "HTML",
	".htm":          "HTML",
	".css":          "CSS",
	".scss":         "SCSS",
	".sass":         "Sass",
	".less":         "Less",
	".vue":          "Vue",
	".svelte":       "Svelte",
	".sql":          "SQL",
	".sh":           "Shell",
	".bash":         "Shell",
	".zsh":          "Shell",
	".json":         "JSON",
	".yaml":         "YAML",
	".yml":          "YAML",
	".xml":          "XML",
	".md":           "Markdown",
	".txt":          "Text",
	".toml":         "TOML",
	".ini":          "INI",
	".conf":         "Config",
	".env":          "Environment",
	".lock":         "Lock File",
	".sum":          "Checksum",
	".mod":          "Module",
	".config":       "Config",
	".eslintrc":     "ESLint",
	".prettierrc":   "Prettier",
	".editorconfig": "EditorConfig",
	".gitignore":    "Git",
	".log":          "Log",
	".dockerfile":   "Docker",
	".makefile":     "Makefile",
}

var SPECIAL_LANGUAGE_MAP = map[string]string{
	"Dockerfile":         "Docker",
	"Makefile":           "Makefile",
	"Gemfile":            "Ruby",
	"Rakefile":           "Ruby",
	"package.json":       "JavaScript",
	"package-lock.json":  "JavaScript",
	"pnpm-lock.yaml":     "JavaScript",
	"yarn.lock":          "JavaScript",
	"go.mod":             "Go",
	"go.sum":             "Go",
	"requirements.txt":   "Python",
	"Pipfile":            "Python",
	"Cargo.toml":         "Rust",
	"Cargo.lock":         "Rust",
	"pom.xml":            "Java",
	"build.gradle":       "Java",
	".gitignore":         "Git",
	".dockerignore":      "Docker",
	"tsconfig.json":      "TypeScript",
	"tsconfig.app.json":  "TypeScript",
	"tsconfig.node.json": "TypeScript",
	"vite.config.ts":     "TypeScript",
	"vite.config.js":     "JavaScript",
	".eslintrc.js":       "JavaScript",
	".prettierrc.js":     "JavaScript",
	"jest.config.js":     "JavaScript",
	"tailwind.config.js": "JavaScript",
	"postcss.config.js":  "JavaScript",
	"vercel.json":        "Config",
	"netlify.toml":       "Config",
}

// 主要言語の定数定義
const (
	LangGo         = "Go"
	LangJavaScript = "JavaScript"
	LangTypeScript = "TypeScript"
	LangPython     = "Python"
	LangRust       = "Rust"
	LangJava       = "Java"
	LangCpp        = "C++"
	LangCSharp     = "C#"
	LangRuby       = "Ruby"
	LangPHP        = "PHP"
	LangSwift      = "Swift"
	LangKotlin     = "Kotlin"
	LangDart       = "Dart"
)

// 基本的な言語のスライス
var MAIN_LANGUAGES = []string{
	LangGo,
	LangTypeScript,
	LangJavaScript,
	LangPython,
	LangRust,
	LangJava,
	LangCpp,
	LangCSharp,
	LangRuby,
	LangPHP,
	LangSwift,
	LangKotlin,
	LangDart,
}

// フィルタリング用の高速参照マップ
var MAIN_LANGUAGES_SET = func() map[string]bool {
	set := make(map[string]bool)
	for _, lang := range MAIN_LANGUAGES {
		set[lang] = true
	}
	return set
}()

// 除外リポジトリ（集計対象外）
var EXCLUDED_REPOSITORIES = []string{
	"obsidian-vault",
}
