package ws

// voice.go: the table's Discord voice room on the game view (ADR 0159
// §6).
//
// Like the host and the playmats, the room is not part of the game: the
// lobby sets it here and every captured view is stamped with it. Nothing
// about it enters the engine, its snapshot or its undo stack. The stamp
// is the FULL state, every seat's access included;
// protocol.FilterViewFor reduces it per viewer.

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// SetVoice replaces the room's voice state. nil clears it, which is what
// a server with voice rooms off leaves every room at. The value is
// copied, so the caller may keep and change its own. Like SetHost it
// takes only its own mutex, so the lobby may call it from inside an
// ApplyExternal fn.
func (r *Room) SetVoice(v *protocol.VoiceView) {
	r.voiceMu.Lock()
	defer r.voiceMu.Unlock()
	r.voice = copyVoice(v)
}

// Voice returns a copy of the room's voice state, nil when unset.
func (r *Room) Voice() *protocol.VoiceView {
	r.voiceMu.Lock()
	defer r.voiceMu.Unlock()
	return copyVoice(r.voice)
}

// stampVoiceLocked writes the voice state onto a freshly captured view.
// Caller holds r.mu.
func (r *Room) stampVoiceLocked(view *protocol.GameView) {
	r.voiceMu.Lock()
	defer r.voiceMu.Unlock()
	view.Voice = copyVoice(r.voice)
}

func copyVoice(v *protocol.VoiceView) *protocol.VoiceView { return protocol.CopyVoice(v) }
