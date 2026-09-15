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

package ebiten

import (
	"github.com/hajimehoshi/ebiten/v2/internal/ui"
)

// Several windows.
//
// While RunGame runs, OpenWindow opens another window with a Game of its own. Every window is stepped in turn by
// the one game loop: its Layout, Update and Draw run with the package-level functions (the window ones, the input
// ones, ScreenSize, IsFocused, ScheduleFrame...) resolving to that window, so a game written for one window needs
// no change. Outside a step they resolve to the primary window, the one RunGame opened; a handle's Do runs a
// function with them resolving to its window instead. ScheduleFrame wakes the loop for every window.
//
// Closing a window (its close button, Close, or Termination from its Update) tears that window down and the
// others go on; RunGame returns when the last one closes. The windows share the graphics device: the driver must
// support several views (Metal does; the OpenGL and DirectX drivers present into one window, and OpenWindow
// reports that).

// WindowOptions are a new window's settings; the zero value takes the defaults the process started with.
type WindowOptions struct {
	// Title is the window's title.
	Title string

	// Width and Height are the window's client size in device-independent pixels; 0 keeps the default.
	Width, Height int

	// X and Y position the window on its monitor when PositionSet is true; otherwise it is centred.
	X, Y        int
	PositionSet bool

	// Undecorated, Floating, Hidden and Maximized are the window attributes at creation.
	Undecorated bool
	Floating    bool
	Hidden      bool
	Maximized   bool

	// ResizingMode is the window's resizing mode; the zero value is WindowResizingModeDisabled.
	ResizingMode WindowResizingModeType

	// MinWidth, MinHeight, MaxWidth and MaxHeight limit the window size when SizeLimitsSet is true (-1 = no
	// limit on that side).
	MinWidth, MinHeight, MaxWidth, MaxHeight int
	SizeLimitsSet                            bool
}

// Window is a handle to a window opened with OpenWindow, or to the primary one (PrimaryWindow).
type Window struct {
	id int
}

// OpenWindow opens another window running game and returns its handle. It may be called from a Game's Update or
// from any goroutine once RunGame is running.
func OpenWindow(game Game, options *WindowOptions) (*Window, error) {
	op := &ui.WindowOptions{Decorated: true, Visible: true}
	if options != nil {
		op.Title = options.Title
		op.WidthInDIP, op.HeightInDIP = options.Width, options.Height
		op.XInDIP, op.YInDIP, op.PositionSet = options.X, options.Y, options.PositionSet
		op.Decorated = !options.Undecorated
		op.Visible = !options.Hidden
		op.Floating, op.Maximized = options.Floating, options.Maximized
		op.ResizingMode = ui.WindowResizingMode(options.ResizingMode)
		op.MinWidthInDIP, op.MinHeightInDIP, op.MaxWidthInDIP, op.MaxHeightInDIP = options.MinWidth, options.MinHeight, options.MaxWidth, options.MaxHeight
		op.SizeLimitsSet = options.SizeLimitsSet
	}
	g := newGameForUI(game, screenTransparent.Load())
	id, err := ui.Get().OpenWindow(g, op)
	if err != nil {
		return nil, err
	}
	g.windowID = id
	return &Window{id: id}, nil
}

// CurrentWindow is the window the package-level functions resolve to right now: the one whose Update or Draw is
// running, else the primary window. It is nil before the game starts.
func CurrentWindow() *Window {
	id := ui.Get().CurrentWindowID()
	if id == 0 {
		return nil
	}
	return &Window{id: id}
}

// ID identifies the window for the life of the process.
func (w *Window) ID() int { return w.id }

// Do runs f with the package-level functions resolving to this window: SetWindowTitle, SetWindowSize,
// WindowPosition, IsFocused, the input functions and the rest act on it. Meant for the game goroutine (a Game's
// Update or Draw); from another goroutine it races the loop's own resolution.
func (w *Window) Do(f func()) { ui.Get().WithWindow(w.id, f) }

// Close asks the window to close; it is torn down at its next step. Closing the last window ends RunGame.
func (w *Window) Close() { ui.Get().CloseWindow(w.id) }

// IsClosed reports whether the window has been closed.
func (w *Window) IsClosed() bool { return ui.Get().IsWindowClosed(w.id) }

// The common window settings as methods, for convenience; Do covers the rest.

func (w *Window) SetTitle(title string) { w.Do(func() { SetWindowTitle(title) }) }

func (w *Window) SetSize(width, height int) { w.Do(func() { SetWindowSize(width, height) }) }

func (w *Window) Size() (width, height int) {
	w.Do(func() { width, height = WindowSize() })
	return
}

func (w *Window) SetPosition(x, y int) { w.Do(func() { SetWindowPosition(x, y) }) }

func (w *Window) Position() (x, y int) {
	w.Do(func() { x, y = WindowPosition() })
	return
}

func (w *Window) SetSizeLimits(minw, minh, maxw, maxh int) {
	w.Do(func() { SetWindowSizeLimits(minw, minh, maxw, maxh) })
}

func (w *Window) SetResizingMode(mode WindowResizingModeType) {
	w.Do(func() { SetWindowResizingMode(mode) })
}

func (w *Window) IsFocused() (focused bool) {
	w.Do(func() { focused = IsFocused() })
	return
}

func (w *Window) IsMinimized() (minimized bool) {
	w.Do(func() { minimized = IsWindowMinimized() })
	return
}

func (w *Window) IsFullscreen() (fullscreen bool) {
	w.Do(func() { fullscreen = IsFullscreen() })
	return
}

func (w *Window) SetFullscreen(fullscreen bool) { w.Do(func() { SetFullscreen(fullscreen) }) }

func (w *Window) RequestAttention() { w.Do(func() { RequestAttention() }) }
