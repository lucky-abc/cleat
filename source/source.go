package source

type Source interface {
	Start()
	Process()
	Stop()
}
