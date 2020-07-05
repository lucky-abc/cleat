package tunnel

import (
	"github.com/lucky-abc/cleat/output"
	"github.com/lucky-abc/cleat/source"
)

type Tunnel interface {
	Start()
	Transfer()
	Stop()
}

type TunnelModel struct {
	Source source.Source
	Output output.Output
}

func (t *TunnelModel) startSource() {
	t.Source.Start()
}

func (t *TunnelModel) closeSource() {
	t.Source.Stop()
}

func (t *TunnelModel) startOutput() {
	t.Output.Start()
}

func (t *TunnelModel) closeOutput() {
	t.Output.Stop()
}
