module github.com/aulaga/cloud

go 1.19

replace github.com/emersion/go-webdav => ./lib/go-webdav

replace github.com/aulaga/webdav => ./lib/webdav/

require (
	github.com/aulaga/webdav v0.0.0-00010101000000-000000000000
	github.com/beyondstorage/go-service-fs/v3 v3.5.0
	github.com/beyondstorage/go-service-memory v0.3.0
	github.com/beyondstorage/go-storage/v4 v4.7.0
	github.com/go-chi/chi/v5 v5.0.10
	github.com/google/uuid v1.4.0
)

require (
	github.com/Xuanwo/templateutils v0.1.0 // indirect
	github.com/dave/dst v0.26.2 // indirect
	github.com/kevinburke/go-bindata v3.22.0+incompatible // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/qingstor/go-mime v0.1.0 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/tools v0.1.1 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
)
