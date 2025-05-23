// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/tview"
	"k8s.io/kubectl/pkg/cmd/get"
)

var fullRX = regexp.MustCompile(`^([\w\s%/-]+):?([\w\W]*?)\|?([NTWSLRH]{0,3})$`)

type colAttr byte

const (
	number     colAttr = 'N'
	age        colAttr = 'T'
	wide       colAttr = 'W'
	show       colAttr = 'S'
	alignLeft  colAttr = 'L'
	alignRight colAttr = 'R'
	hide       colAttr = 'H'
)

type colAttrs struct {
	align    int
	mx       bool
	mxc      bool
	mxm      bool
	time     bool
	wide     bool
	show     bool
	hide     bool
	capacity bool
}

func newColFlags(flags string) colAttrs {
	c := colAttrs{
		align: tview.AlignLeft,
		wide:  false,
	}
	for _, b := range []byte(flags) {
		switch colAttr(b) {
		case hide:
			c.hide = true
		case wide:
			c.wide, c.show = true, false
		case show:
			c.show, c.wide = true, false
		case alignLeft:
			c.align = tview.AlignLeft
		case alignRight:
			c.align = tview.AlignRight
		case age:
			c.time = true
		case number:
			c.capacity, c.align = true, tview.AlignRight
		default:
			slog.Warn("Unknown column attribute", slogs.Attr, b)
		}
	}

	return c
}

type colDef struct {
	colAttrs

	name string
	idx  int
	spec string
}

func parse(s string) (colDef, error) {
	mm := fullRX.FindStringSubmatch(s)
	if len(mm) == 4 {
		spec, err := get.RelaxedJSONPathExpression(mm[2])
		if err != nil {
			return colDef{idx: -1}, err
		}
		return colDef{
			name:     mm[1],
			idx:      -1,
			spec:     spec,
			colAttrs: newColFlags(mm[3]),
		}, nil
	}

	return colDef{idx: -1}, fmt.Errorf("invalid column definition %q", s)
}

func (c colDef) toHeaderCol() model1.HeaderColumn {
	return model1.HeaderColumn{
		Name: c.name,
		Attrs: model1.Attrs{
			Align:    c.align,
			Wide:     c.wide,
			Show:     c.show,
			Time:     c.time,
			MX:       c.mx,
			MXC:      c.mxc,
			MXM:      c.mxm,
			Hide:     c.hide,
			Capacity: c.capacity,
		},
	}
}
