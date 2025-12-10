module github.com/CuteReimu/YinYangJade

go 1.25.4

require (
	github.com/CuteReimu/bilibili/v2 v2.5.0
	github.com/CuteReimu/goutil v0.0.0-20250319114149-4d3965d6581c
	github.com/CuteReimu/neuquant v0.0.0-20240410080316-c01be0b1e2bb
	github.com/CuteReimu/onebot v0.0.0-20251201151712-c38ab9a341e1
	github.com/dgraph-io/badger/v4 v4.8.0
	github.com/dlclark/regexp2 v1.11.5
	github.com/go-resty/resty/v2 v2.17.0
	github.com/go-rod/rod v0.116.2
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	github.com/pkg/errors v0.9.1
	github.com/robfig/cron v1.2.0
	github.com/spf13/viper v1.21.0
	github.com/tidwall/gjson v1.18.0
	github.com/vicanso/go-charts/v2 v2.6.10
	github.com/wcharczuk/go-chart/v2 v2.1.2
	golang.org/x/sync v0.19.0
	golang.org/x/time v0.14.0
)

require (
	github.com/Baozisoftware/qrcode-terminal-go v0.0.0-20170407111555-c0650d8dff0f // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgraph-io/ristretto/v2 v2.3.0 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/google/flatbuffers v25.9.23+incompatible // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/lestrrat-go/strftime v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/match v1.2.0 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/ysmood/fetchup v0.3.0 // indirect
	github.com/ysmood/goob v0.4.0 // indirect
	github.com/ysmood/got v0.42.3 // indirect
	github.com/ysmood/gson v0.7.3 // indirect
	github.com/ysmood/leakless v0.9.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.39.0 // indirect
	go.opentelemetry.io/otel/metric v1.39.0 // indirect
	go.opentelemetry.io/otel/trace v1.39.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/image v0.34.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	google.golang.org/protobuf v1.36.10 // indirect
)

ignore (
	./assets
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
