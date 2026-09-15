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

//go:build !playstation5

package opengl

import (
	"github.com/hajimehoshi/ebiten/v2/internal/graphics"
	"github.com/hajimehoshi/ebiten/v2/internal/graphicsdriver"
)

// Several windows on OpenGL (graphicsdriver.Viewer). Every window is a GLFW window with its own GL context,
// created sharing the primary window's context, so textures, buffers, shaders and programs are one set. All
// rendering stays in the primary context, on the render thread, where the state caches, the vertex array and
// every image's framebuffer object live (those are not shared between contexts). A secondary window's screen
// image is therefore not the default framebuffer but an ordinary texture, drawn like any offscreen; presenting
// it makes the window's own context current, blits the texture to that context's default framebuffer through a
// framebuffer object made there, swaps the window's buffers and makes the primary context current again. One
// context switch pair per presented frame, and the single-window path is untouched.

// view is one window the driver presents into.
type view struct {
	id        graphicsdriver.ViewID
	presenter Presenter // the window: its context, swap interval and buffer swap; nil for the primary
	primary   bool

	// screen is the texture-backed screen image of a secondary view, and fbo the framebuffer object over that
	// texture in the view's own context (fboTexture says which texture; 0 until made, remade after a resize).
	screen     *Image
	fbo        uint32
	fboTexture textureNative
}

// secondaryView is the current view when it is not the primary window: its screen is a texture.
func (g *Graphics) secondaryView() *view {
	if g.view == nil || g.view.primary {
		return nil
	}
	return g.view
}

// newViewScreenImage makes a secondary view's screen: a texture the size of the window, drawn in the primary
// context and blitted to the window at present.
func (g *Graphics) newViewScreenImage(v *view, width, height int) (graphicsdriver.Image, error) {
	// the internal (power-of-two) size like any offscreen: the projection every draw uses assumes it, and the
	// blit takes the window-sized region
	w := graphics.InternalImageSize(width)
	h := graphics.InternalImageSize(height)
	g.checkSize(w, h)
	t, err := g.context.newTexture(w, h)
	if err != nil {
		return nil, err
	}
	i := &Image{
		id:       g.genNextImageID(),
		graphics: g,
		width:    width,
		height:   height,
		texture:  t,
		view:     v,
	}
	g.addImage(i)
	v.screen = i
	return i, nil
}
