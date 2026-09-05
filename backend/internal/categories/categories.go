package categories

type Category struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func List() []Category {
	return []Category{
		{Slug: "web", Name: "Web 开发"},
		{Slug: "tool", Name: "工具脚本"},
		{Slug: "app", Name: "应用源码"},
		{Slug: "game", Name: "游戏源码"},
		{Slug: "other", Name: "其他"},
	}
}
