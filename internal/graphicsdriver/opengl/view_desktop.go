// Copyright 2026 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !android && !ios && !js && !nintendosdk && !playstation5

package opengl

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/internal/graphicsdriver"
	"github.com/hajimehoshi/ebiten/v2/internal/graphicsdriver/opengl/gl"
)

// NewPresenterView registers a window: the first is the primary, whose context everything renders in and which
// SetPresenter would have set; the others are presented by blitting. Called on the main thread at window
// creation, before any frame of that window.
func (g *Graphics) NewPresenterView(p Presenter) (graphicsdriver.ViewID, error) {
	if g.views == nil {
		g.views = map[graphicsdriver.ViewID]*view{}
	}
	g.nextViewID++
	v := &view{id: g.nextViewID, presenter: p}
	if g.presenter == nil {
		g.presenter = p
		v.primary = true
	}
	g.viewsMu.Lock()
	g.views[v.id] = v
	g.viewsMu.Unlock()
	return v.id, nil
}

// NewView is the Viewer's constructor from a native window; OpenGL needs the window's context instead
// (NewPresenterView).
func (g *Graphics) NewView(nativeWindow uintptr) (graphicsdriver.ViewID, error) {
	return 0, errors.New("opengl: a view is made from a Presenter (NewPresenterView), not a native window")
}

// SetCurrentView names the view the frame's screen image and present belong to (render thread, before Begin).
func (g *Graphics) SetCurrentView(id graphicsdriver.ViewID) {
	g.viewsMu.Lock()
	g.view = g.views[id]
	g.viewsMu.Unlock()
}

// ReleaseView frees a closed window's framebuffer object in its context (render thread, outside a frame).
// Its screen image is disposed by its owner like any image.
func (g *Graphics) ReleaseView(id graphicsdriver.ViewID) {
	g.viewsMu.Lock()
	v := g.views[id]
	delete(g.views, id)
	g.viewsMu.Unlock()
	if v == nil {
		return
	}
	if g.view == v {
		g.view = nil
	}
	if v.primary || v.fbo == 0 {
		return
	}
	if err := v.presenter.MakeContextCurrent(); err == nil {
		g.context.ctx.DeleteFramebuffer(v.fbo)
		v.fbo = 0
	}
	_ = g.presenter.MakeContextCurrent()
}

// presentView shows a secondary view's screen texture in its window.
func (g *Graphics) presentView(v *view) error {
	s := v.screen
	if s == nil {
		return nil
	}
	if err := v.presenter.MakeContextCurrent(); err != nil {
		return err
	}
	defer func() { _ = g.presenter.MakeContextCurrent() }()
	ctx := g.context.ctx
	if v.fbo != 0 && v.fboTexture != s.texture {
		ctx.DeleteFramebuffer(v.fbo)
		v.fbo = 0
	}
	if v.fbo == 0 {
		f := ctx.CreateFramebuffer()
		if f == 0 {
			return errors.New("opengl: creating a window's framebuffer object failed")
		}
		ctx.BindFramebuffer(gl.FRAMEBUFFER, f)
		ctx.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, uint32(s.texture), 0)
		if st := ctx.CheckFramebufferStatus(gl.FRAMEBUFFER); st != gl.FRAMEBUFFER_COMPLETE {
			ctx.DeleteFramebuffer(f)
			return fmt.Errorf("opengl: a window's framebuffer object is incomplete: %v", st)
		}
		v.fbo = f
		v.fboTexture = s.texture
	}
	// The texture holds the image the way every offscreen does, the top row at texture row 0, which is the
	// bottom in the window's coordinates: the blit flips it, as the projection does for the default framebuffer.
	w, h := int32(s.width), int32(s.height)
	ctx.Disable(gl.SCISSOR_TEST)
	ctx.BindFramebuffer(gl.READ_FRAMEBUFFER, v.fbo)
	ctx.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	ctx.BlitFramebuffer(0, 0, w, h, 0, h, w, 0, gl.COLOR_BUFFER_BIT, gl.NEAREST)
	var interval int
	if g.vsync {
		interval = 1
	}
	if err := v.presenter.SwapInterval(interval); err != nil {
		return err
	}
	return v.presenter.SwapBuffers()
}
