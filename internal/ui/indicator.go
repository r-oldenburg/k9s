// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tview"
)

// StatusIndicator represents a status indicator when main header is collapsed.
type StatusIndicator struct {
	*tview.TextView

	app       *App
	styles    *config.Styles
	permanent string
	cancel    context.CancelFunc
}

// NewStatusIndicator returns a new status indicator.
func NewStatusIndicator(app *App, styles *config.Styles) *StatusIndicator {
	s := StatusIndicator{
		TextView: tview.NewTextView(),
		app:      app,
		styles:   styles,
	}
	s.SetTextAlign(tview.AlignCenter)
	s.SetTextColor(styles.FgColor())
	s.SetBackgroundColor(styles.BgColor())
	s.SetDynamicColors(true)
	styles.AddListener(&s)

	return &s
}

// StylesChanged notifies the skins changed.
func (s *StatusIndicator) StylesChanged(styles *config.Styles) {
	s.styles = styles
	s.SetBackgroundColor(styles.BgColor())
	s.SetTextColor(styles.FgColor())
}

const statusIndicatorFmt = "[%s::b]K9s [%s::]%s [%s::]%s:%s:%s [%s::]%s[%s::]::[%s::]%s"

// ClusterInfoUpdated notifies the cluster meta was updated.
func (s *StatusIndicator) ClusterInfoUpdated(data *model.ClusterMeta) {
	s.app.QueueUpdateDraw(func() {
		s.SetPermanent(fmt.Sprintf(
			statusIndicatorFmt,
			s.styles.Body().LogoColor.String(),
			s.styles.K9s.Info.K9sRevColor.String(),
			data.K9sVer,
			s.styles.K9s.Info.FgColor.String(),
			data.Context,
			data.Cluster,
			data.K8sVer,
			s.styles.K9s.Info.CPUColor.String(),
			render.PrintPerc(data.Cpu),
			s.styles.Body().FgColor.String(),
			s.styles.K9s.Info.MEMColor.String(),
			render.PrintPerc(data.Mem),
		))
	})
}

// ClusterInfoChanged notifies the cluster meta was changed.
func (s *StatusIndicator) ClusterInfoChanged(prev, cur *model.ClusterMeta) {
	if !s.app.IsRunning() {
		return
	}
	s.app.QueueUpdateDraw(func() {
		s.SetPermanent(fmt.Sprintf(
			statusIndicatorFmt,
			s.styles.Body().LogoColor.String(),
			s.styles.K9s.Info.K9sRevColor.String(),
			cur.K9sVer,
			s.styles.K9s.Info.FgColor.String(),
			cur.Context,
			cur.Cluster,
			cur.K8sVer,
			s.styles.K9s.Info.CPUColor.String(),
			AsPercDelta(prev.Cpu, cur.Cpu),
			s.styles.Body().FgColor.String(),
			s.styles.K9s.Info.MEMColor.String(),
			AsPercDelta(prev.Cpu, cur.Mem),
		))
	})
}

// SetPermanent sets permanent title to be reset to after updates.
func (s *StatusIndicator) SetPermanent(info string) {
	s.permanent = info
	s.SetText(info)
}

// Reset clears out the logo view and resets colors.
func (s *StatusIndicator) Reset() {
	s.Clear()
	s.SetPermanent(s.permanent)
}

// Err displays a log error state.
func (s *StatusIndicator) Err(msg string) {
	s.update(msg, "orangered")
}

// Warn displays a log warning state.
func (s *StatusIndicator) Warn(msg string) {
	s.update(msg, "mediumvioletred")
}

// Info displays a log info state.
func (s *StatusIndicator) Info(msg string) {
	s.update(msg, "lawngreen")
}

func (s *StatusIndicator) update(msg, c string) {
	s.setText(fmt.Sprintf("[%s::b] <%s> ", c, msg))
}

func (s *StatusIndicator) setText(msg string) {
	if s.cancel != nil {
		s.cancel()
	}
	s.SetText(msg)

	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			s.app.QueueUpdateDraw(func() {
				s.Reset()
			})
		}
	}(ctx)
}

// Helpers...

// AsPercDelta represents a percentage with a delta indicator.
func AsPercDelta(ov, nv int) string {
	prev, cur := render.IntToStr(ov), render.IntToStr(nv)
	return cur + "%" + Deltas(prev, cur)
}
