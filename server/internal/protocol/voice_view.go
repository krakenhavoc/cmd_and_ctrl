package protocol

// voice_view.go: a table's Discord voice room as the game view carries
// it (ADR 0159 §6).
//
// The room is not part of the game. The lobby keeps its state on the
// ws.Room (Room.SetVoice), every capture stamps it on, and nothing about
// it enters the engine, its snapshot or its undo stack, exactly like the
// host and the playmats. The full view holds every seat's access, which
// FilterViewFor reduces to the one viewer's: their own access, and the
// room's link only when they may use it.

// VoiceView is the voice room on the game view.
type VoiceView struct {
	// Enabled is the host's choice (games.voice_room).
	Enabled bool `json:"enabled"`
	// Status is one of the VoiceStatus* values.
	Status string `json:"status"`
	// You is this viewer's access, one of the VoiceAccess* values. Set
	// by FilterViewFor; empty on the full, unfiltered view.
	You string `json:"you,omitempty"`
	// URL opens the room in Discord. Only a viewer whose access is
	// VoiceAccessGranted receives it.
	URL string `json:"url,omitempty"`
	// Access maps each seat's player id to its access. The full view's
	// only: FilterViewFor drops it for every viewer but the admin.
	Access map[string]string `json:"access,omitempty"`
	// Connected maps a seat's player id to VoiceConnLive or VoiceConnMuted
	// for each seat connected to the room right now (#3311). A seat not
	// in the map is not connected. Only a seated viewer receives it.
	Connected map[string]string `json:"connected,omitempty"`
	// Granted and InRoom are the chip's "Voice · 3 of 4": how many seats
	// were let in, and how many of those are connected. Seated viewers
	// only.
	Granted int `json:"granted,omitempty"`
	InRoom  int `json:"in_room,omitempty"`
}

// A seat's connection to the room (#3311). Absent means not connected.
const (
	VoiceConnLive  = "live"
	VoiceConnMuted = "muted"
)

// Room states (ADR 0159 §6).
const (
	VoiceStatusOff         = "off"
	VoiceStatusPending     = "pending"
	VoiceStatusReady       = "ready"
	VoiceStatusIdle        = "idle"
	VoiceStatusFull        = "full"
	VoiceStatusUnavailable = "unavailable"
)

// A viewer's access to the room (ADR 0159 §3, §6).
const (
	// VoiceAccessGranted: let in; gets the link.
	VoiceAccessGranted = "granted"
	// VoiceAccessNotInServer: signed in with Discord, but not a member
	// of the community server, so there was nothing to grant.
	VoiceAccessNotInServer = "not_in_server"
	// VoiceAccessNoDiscord: a guest, with no Discord id to grant.
	VoiceAccessNoDiscord = "no_discord"
	// VoiceAccessSpectator: watching, not seated.
	VoiceAccessSpectator = "spectator"
	// VoiceAccessPending: a Discord seat whose membership has not been
	// checked yet.
	VoiceAccessPending = "pending"
)

// filterVoiceFor reduces the full view's voice room to what viewerID
// may see. The admin ("") sees it whole. A viewer not in the access map
// (a spectator, or a seat the stamp has not caught up with) is a
// spectator. The link goes only to a granted viewer.
func filterVoiceFor(v *VoiceView, viewerID string) *VoiceView {
	if v == nil {
		return nil
	}
	if viewerID == "" {
		return CopyVoice(v)
	}
	you, seated := v.Access[viewerID]
	if !seated || viewerID == SpectatorViewerID {
		you = VoiceAccessSpectator
	}
	out := &VoiceView{Enabled: v.Enabled, Status: v.Status, You: you}
	if seated && viewerID != SpectatorViewerID {
		// Voice connection state is shown to the table, not to people
		// watching it (ADR 0149).
		out.Connected = copyStrings(v.Connected)
		out.Granted, out.InRoom = v.Granted, v.InRoom
	}
	if you == VoiceAccessGranted && v.Status == VoiceStatusReady {
		out.URL = v.URL
	}
	return out
}

// CopyVoice deep-copies a voice view.
func CopyVoice(v *VoiceView) *VoiceView {
	if v == nil {
		return nil
	}
	out := *v
	out.Access = copyStrings(v.Access)
	out.Connected = copyStrings(v.Connected)
	return &out
}

func copyStrings(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, a := range m {
		out[k] = a
	}
	return out
}
