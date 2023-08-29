package core

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type ProgressBarInfiniteSmall struct {
	widget.ProgressBarInfinite
}

func NewProgressBarInfiniteSmall() *ProgressBarInfiniteSmall {
	p := &ProgressBarInfiniteSmall{}
	return p
}

func (p *ProgressBarInfiniteSmall) MinSize() fyne.Size {
	p.ExtendBaseWidget(p)
	return fyne.Size{Width: 100, Height: 5}
}
