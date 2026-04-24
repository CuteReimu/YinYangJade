module github.com/CuteReimu/YinYangJade

go 1.26.2

require (
	github.com/CuteReimu/bilibili/v2 v2.5.1
	github.com/CuteReimu/goutil v0.0.0-20260422123512-b447079b3afe
	github.com/CuteReimu/neuquant v0.0.0-20260422122435-5979c1b71f25
	github.com/CuteReimu/onebot v0.0.0-20260422123748-4b553a338f6d
	github.com/dgraph-io/badger/v4 v4.9.1
	github.com/dlclark/regexp2 v1.12.0
	github.com/go-ego/gse v1.0.2
	github.com/go-resty/resty/v2 v2.17.2
	github.com/go-rod/rod v0.116.2
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.21.0
	github.com/tidwall/gjson v1.18.0
	github.com/vicanso/go-charts/v2 v2.6.10
	github.com/wcharczuk/go-chart/v2 v2.1.2
	golang.org/x/sync v0.20.0
	golang.org/x/time v0.15.0
)

require (
	github.com/Baozisoftware/qrcode-terminal-go v0.0.0-20170407111555-c0650d8dff0f // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgraph-io/ristretto/v2 v2.4.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/flatbuffers v25.12.19+incompatible // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/lestrrat-go/strftime v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.21 // indirect
	github.com/pelletier/go-toml/v2 v2.3.0 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/vcaesar/cedar v0.30.0 // indirect
	github.com/ysmood/fetchup v0.3.0 // indirect
	github.com/ysmood/goob v0.4.0 // indirect
	github.com/ysmood/got v0.42.3 // indirect
	github.com/ysmood/gson v0.7.3 // indirect
	github.com/ysmood/leakless v0.9.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/image v0.39.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

ignore (
	./assets
	./maplebot // 屏蔽 maplebot ，它现在不用了
	__pycache__
	libs
	player_data
)

exclude ( // 有 breaking change，而依赖它的库 github.com/go-rod/rod 没有更新
	github.com/ysmood/fetchup v0.4.0
	github.com/ysmood/fetchup v0.5.0
	github.com/ysmood/fetchup v0.5.1
	github.com/ysmood/fetchup v0.5.2
	github.com/ysmood/fetchup v0.5.3
)
