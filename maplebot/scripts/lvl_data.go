package scripts

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// BuildLvlData 从 viper 配置中读取等级数据并生成 lvl_data.json 文件
func BuildLvlData(data *viper.Viper) error {
	m := make(map[int]int64, 299)
	c := make(map[int]any, 301)
	for i := 1; i < 300; i++ {
		m[i] = data.GetInt64(fmt.Sprintf("data.%d", i))
	}
	c[0] = nil
	c[1] = int64(0)
	for i := 1; i < 300; i++ {
		c[i+1] = c[i].(int64) + m[i]
	}
	buf, err := json.MarshalIndent(map[string]any{
		"single":     m,
		"cumulative": c,
	}, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile("lvl_data.json", buf, 0600)
}
