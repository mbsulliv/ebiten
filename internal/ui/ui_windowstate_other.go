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

//go:build android || ios || js || nintendosdk || playstation5

package ui

import "errors"

// These platforms have one screen and no windows of their own: the UserInterface's fields hold the state.

func (u *UserInterface) windowState() windowState { return nil }

func (u *UserInterface) currentContext() *context { return u.context }

// CurrentWindowID is 0: one screen, no windows.
func (u *UserInterface) CurrentWindowID() int { return 0 }

// OpenWindow is not available: there is one screen.
func (u *UserInterface) OpenWindow(game Game, opts *WindowOptions) (int, error) {
	return 0, errors.New("ui: windows cannot be opened on this platform")
}

func (u *UserInterface) CloseWindow(id int)          {}
func (u *UserInterface) FocusWindow(id int)          {}
func (u *UserInterface) WithWindow(id int, f func()) { f() }
func (u *UserInterface) IsWindowClosed(id int) bool  { return true }

func (u *UserInterface) NativeWindow(id int) uintptr { return 0 }
