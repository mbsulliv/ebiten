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

// The desktop's windows. The GLFW backend owns one window each (glfwBackend is the per-window struct: the
// GLFW handle, its input, its size and position bookkeeping, and, since the multi-window work, the game
// context, the desktop window settings and the per-window counters that used to live on UserInterface). The
// UserInterface keeps the list, and the primary window is the one the package-level API means; a backend of
// another kind (fbdev, a VM guest) has no windows of its own and the UserInterface's own fields serve it as
// before.

// windowState answers the per-window counters and settings for the primary window, nil when no GLFW window
// is running.
func (u *UserInterface) windowState() windowState {
	if b, ok := u.runningBackend().(*glfwBackend); ok {
		return b
	}
	return nil
}

// currentContext is the game context of the primary window, or the process-level one for a backend without
// windows of its own.
func (u *UserInterface) currentContext() *context {
	if b, ok := u.runningBackend().(*glfwBackend); ok {
		return b.context
	}
	return u.context
}

// primaryGLFW is the primary window when the running backend is the GLFW one.
func (u *UserInterface) primaryGLFW() *glfwBackend {
	b, _ := u.runningBackend().(*glfwBackend)
	return b
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
