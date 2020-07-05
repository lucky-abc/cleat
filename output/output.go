package output

type Output interface {
	Start()
	Process()
	Stop()
}
