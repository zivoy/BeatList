package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

type AdaptiveColumn struct {
	vertical bool
	size     fyne.Size
}

func (a *AdaptiveColumn) Layout(objs []fyne.CanvasObject, cont fyne.Size) {
	a.size = cont
	pos := fyne.NewPos(0, 0)
	for i, o := range objs {
		adjust := theme.Padding() / 2
		if i == len(objs)-1 {
			adjust = 0
		}
		if a.vertical {
			s := o.MinSize().Height
			o.Resize(fyne.NewSize(cont.Width, s-adjust))
			o.Move(pos)
			pos = pos.Add(fyne.NewPos(0, s))
		} else {
			s := cont.Width / float32(len(objs))
			o.Resize(fyne.NewSize(s-adjust, cont.Height))
			o.Move(pos)
			pos = pos.Add(fyne.NewPos(s, 0))
		}
	}
}
func (a *AdaptiveColumn) MinSize(objs []fyne.CanvasObject) fyne.Size {
	var w, h, sumHeight float32
	for _, O := range objs {
		s := O.MinSize()
		sumHeight += s.Height
		w = fyne.Max(w, s.Width)
		h = fyne.Max(h, s.Height)
	}
	a.vertical = w > (a.size.Width-theme.Padding()*float32(len(objs)-1))/float32(len(objs))
	if a.vertical {
		h = sumHeight
	}
	return fyne.NewSize(w, h)
}

func NewShrinkingColumns(obj ...fyne.CanvasObject) *fyne.Container {
	return container.New(&AdaptiveColumn{size: fyne.NewSize(10, 10)}, obj...)
}

// ---

const (
	posLeading = iota
	posCenter
	posTrailing
)

type Center struct {
	VCenter, HCenter int
}

func (c Center) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return objects[0].MinSize()
}

func (c Center) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	s := c.MinSize(objects)
	var xPos, yPos float32
	switch c.HCenter {
	case posLeading:
		xPos = 0
	case posCenter:
		xPos = containerSize.Width/2 - s.Width/2
	case posTrailing:
		xPos = containerSize.Width - s.Width
	}

	switch c.VCenter {
	case posLeading:
		yPos = 0
	case posCenter:
		yPos = containerSize.Height/2 - s.Height/2
	case posTrailing:
		yPos = containerSize.Height - s.Height
	}

	pos := fyne.NewPos(xPos, yPos)
	o1 := objects[0]
	o1.Resize(s)
	o1.Move(pos)
}

func NewCenter(obj fyne.CanvasObject, vPos, hPos int) *fyne.Container {
	return container.New(&Center{vPos, hPos}, obj)
}

// ---

type MinSize struct {
	w, h float32
}

func (m MinSize) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w, h := m.w, m.h
	o1 := objects[0]
	if w == 0 {
		w = o1.MinSize().Width
	}
	if h == 0 {
		h = o1.MinSize().Height
	}
	return fyne.NewSize(w, h)
}

func (m MinSize) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	o1 := objects[0]
	o1.Resize(containerSize)
}

func NewSetMinSize(obj fyne.CanvasObject, width, height float32) *fyne.Container {
	return container.New(&MinSize{w: width, h: height}, obj)
}

// --

type SizeSwitcher struct {
	switchSize float32
	size       fyne.Size
}

func (a *SizeSwitcher) Layout(objs []fyne.CanvasObject, cont fyne.Size) {
	a.size = cont
	o1 := objs[0]
	o2 := objs[1]
	if a.size.Width > a.switchSize {
		o1.Show()
		o2.Hide()
		o1.Resize(cont)
	} else {
		o1.Hide()
		o2.Show()
		o2.Resize(cont)
	}
}
func (a SizeSwitcher) MinSize(objs []fyne.CanvasObject) fyne.Size {
	if a.size.Width > a.switchSize {
		return objs[0].MinSize()
	}
	return objs[1].MinSize()
}

func NewSizeSwitcher(switchSize float32, o1, o2 fyne.CanvasObject) *fyne.Container {
	return container.New(&SizeSwitcher{size: fyne.NewSize(10, 10), switchSize: switchSize}, o1, o2)
}
