// Copyright 2026 The Ebitengine Authors
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

package ui

import (
	"errors"

	"github.com/hajimehoshi/ebiten/v2/internal/clock"
	"github.com/hajimehoshi/ebiten/v2/internal/glfw"
	"github.com/hajimehoshi/ebiten/v2/internal/graphicscommand"
	"github.com/hajimehoshi/ebiten/v2/internal/graphicsdriver"
)

// The desktop's windows. The GLFW backend owns one window each (glfwBackend is the per-window struct: the
// GLFW handle, its input, its size and position bookkeeping, and, since the multi-window work, the game
// context, the desktop window settings and the per-window counters that used to live on UserInterface). The
// UserInterface keeps the list, and the primary window is the one the package-level API means; a backend of
// another kind (fbdev, a VM guest) has no windows of its own and the UserInterface's own fields serve it as
// before.

// windowState answers the per-window counters and settings for the current window (the one whose step is
// running), else the primary; nil when no GLFW window is running.
func (u *UserInterface) windowState() windowState {
	if b := u.currentGLFW(); b != nil {
		return b
	}
	return nil
}

// currentContext is the game context of the current or primary window, or the process-level one for a backend
// without windows of its own.
func (u *UserInterface) currentContext() *context {
	if b := u.currentGLFW(); b != nil {
		return b.context
	}
	return u.context
}

// currentGLFW is the window every package-level call resolves to: the one whose step is running (set by the
// loop around layout, input, Update, Draw and the forced frame), else the primary, when the running backend is
// the GLFW one. runningBackend follows the same rule.
func (u *UserInterface) currentGLFW() *glfwBackend {
	if c := u.current.Load(); c != nil {
		return c
	}
	b, _ := u.primaryBackend().(*glfwBackend)
	return b
}

// primaryGLFW is the primary window when the running backend is the GLFW one.
func (u *UserInterface) primaryGLFW() *glfwBackend {
	b, _ := u.primaryBackend().(*glfwBackend)
	return b
}

// setCurrentWindow names the window whose step is running; nil between steps.
func (u *UserInterface) setCurrentWindow(b *glfwBackend) {
	u.current.Store(b)
}

// CurrentWindowID identifies the window every package-level call resolves to right now: the stepping window,
// else the primary; 0 while no window runs. The input states and the per-window screen sizes key on it.
func (u *UserInterface) CurrentWindowID() int {
	if b := u.currentGLFW(); b != nil {
		return b.id
	}
	return 0
}

// WithWindow runs f with the package-level calls resolving to the window id (a handle's method), restoring the
// current window after. Meant for the game goroutine; from elsewhere it races the loop's own current window.
func (u *UserInterface) WithWindow(id int, f func()) {
	b := u.windowByID(id)
	if b == nil {
		return
	}
	prev := u.current.Swap(b)
	defer u.current.Store(prev)
	f()
}

func (u *UserInterface) windowByID(id int) *glfwBackend {
	for _, w := range u.snapshotWindows() {
		if w.id == id {
			return w
		}
	}
	return nil
}

// IsWindowClosed reports whether a window id no longer names a live window.
func (u *UserInterface) IsWindowClosed(id int) bool {
	b := u.windowByID(id)
	return b == nil || b.closed.Load()
}

// OpenWindow opens another window with its own game while the game runs, and returns its id. The window shares
// the graphics device (the driver must present into several views); it starts from the given settings and
// the process-wide defaults, and gets its frames from the next pass on. It may be called from a window's Update
// or from any goroutine once RunGame is running.
func (u *UserInterface) OpenWindow(game Game, opts *WindowOptions) (int, error) {
	first := u.primaryGLFW()
	if first == nil || u.runOptions == nil {
		return 0, errors.New("ui: OpenWindow needs a running game")
	}
	if _, ok := u.graphicsDriver.(graphicsdriver.Viewer); !ok {
		return 0, errors.New("ui: the graphics driver presents into one window only")
	}
	b := newGLFWBackend(u)
	b.context = newContext(game, u.runOptions.ScreenTransparent)
	// the window's own settings, over the process defaults it started from
	if opts == nil {
		opts = &WindowOptions{Decorated: true, Visible: true, WidthInDIP: 640, HeightInDIP: 480}
	}
	w := &b.desktopWindow
	w.SetTitle(opts.Title)
	if opts.WidthInDIP > 0 && opts.HeightInDIP > 0 {
		w.setInitWindowSizeInDIP(opts.WidthInDIP, opts.HeightInDIP)
	}
	if opts.PositionSet {
		w.setInitWindowPositionInDIP(opts.XInDIP, opts.YInDIP)
	} else if from := u.currentGLFW(); from != nil && from.created.Load() {
		// no position asked for: cascade from the window that opened it, rather than centring on top of it
		x, y := from.desktopWindow.Position()
		w.setInitWindowPositionInDIP(x+30, y+30)
	}
	w.setInitWindowDecorated(opts.Decorated)
	w.setInitWindowVisible(opts.Visible)
	w.setInitWindowFloating(opts.Floating)
	w.setInitWindowMaximized(opts.Maximized)
	w.windowResizingMode.Store(int32(opts.ResizingMode))
	if opts.SizeLimitsSet {
		w.setWindowSizeLimitsInDIP(opts.MinWidthInDIP, opts.MinHeightInDIP, opts.MaxWidthInDIP, opts.MaxHeightInDIP)
	}
	ro := *u.runOptions
	ro.InitWindowWidthInDIP, ro.InitWindowHeightInDIP = w.getInitWindowSizeInDIP()
	ro.WindowPositionSet = w.initWindowPositionInDIP.Load() != nil
	var err error
	u.mainThread.Call(func() {
		err = b.initOnMainThread(&ro)
	})
	if err != nil {
		return 0, err
	}
	// the window's first frame comes from the next pass; wake the pump in case it is waiting for an event
	_ = glfw.PostEmptyEvent()
	return b.id, nil
}

// CloseWindow asks a window to close: its next step ends it, and the loop tears it down.
func (u *UserInterface) CloseWindow(id int) {
	b := u.windowByID(id)
	if b == nil {
		return
	}
	u.mainThread.Call(func() {
		if b.closed.Load() || !b.created.Load() {
			return
		}
		_ = b.window.SetShouldClose(true)
	})
	_ = glfw.PostEmptyEvent()
}

// publishWindow makes a freshly created window the primary when none is, else adds it to the list.
func (u *UserInterface) publishWindow(b *glfwBackend) {
	if u.primaryBackend() == nil {
		u.setRunningBackend(b)
		return
	}
	u.addWindow(b)
}

// setPrimaryWindow moves the primary to another window (the primary closed while others live).
func (u *UserInterface) setPrimaryWindow(b *glfwBackend) {
	var ub uiBackend = b
	u.backend.Store(&ub)
	if b.context != nil {
		clock.SetPrimary(b.context.clock)
	}
}

// closeWindow tears one window down while others live: its view on the render thread, its GLFW window on the
// main thread (behind the closed flag, so a closure still queued against it returns), then the bookkeeping.
func (u *UserInterface) closeWindow(b *glfwBackend) {
	graphicscommand.ReleaseView(u.graphicsDriver, b.viewID)
	var wasFocused bool
	u.mainThread.Call(func() {
		b.closed.Store(true)
		if b.created.Load() {
			if a, err := b.window.GetAttrib(glfw.Focused); err == nil {
				wasFocused = a == glfw.True
			}
			_ = b.window.Destroy()
		}
	})
	u.removeWindow(b)
	if b.context != nil {
		clock.Unregister(b.context.clock)
	}
	ws := u.snapshotWindows()
	if u.primaryGLFW() == b && len(ws) > 0 {
		u.setPrimaryWindow(ws[0])
	}
	// The focus went with the window; hand it to the newest of the rest, so the keyboard keeps working
	// without a click.
	if wasFocused && len(ws) > 0 {
		next := ws[len(ws)-1]
		u.mainThread.Call(func() {
			if next.created.Load() && !next.closed.Load() {
				_ = next.window.Focus()
			}
		})
	}
}

// addWindow and removeWindow keep the list of the GLFW backend's windows; snapshotWindows copies it.
func (u *UserInterface) addWindow(b *glfwBackend) {
	u.windowsMu.Lock()
	defer u.windowsMu.Unlock()
	for _, w := range u.windows {
		if w == b {
			return
		}
	}
	u.windows = append(u.windows, b)
}

func (u *UserInterface) removeWindow(b *glfwBackend) {
	u.windowsMu.Lock()
	defer u.windowsMu.Unlock()
	for i, w := range u.windows {
		if w == b {
			u.windows = append(u.windows[:i], u.windows[i+1:]...)
			return
		}
	}
}

func (u *UserInterface) snapshotWindows() []*glfwBackend {
	u.windowsMu.Lock()
	defer u.windowsMu.Unlock()
	return append([]*glfwBackend(nil), u.windows...)
}
