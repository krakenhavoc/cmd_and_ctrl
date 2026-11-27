package game

import "github.com/google/uuid"

// AttachAtETB makes the entering permanent arrive attached to `host`, as a
// replacement effect on its entry (CR 614.1c, CR 614.12): Grifter's Blade's
// "choose a creature you control it could be attached to. If you do, it
// enters attached to that creature". Only meaningful on a RepEventMove to the
// battlefield.
//
// It is the host an Aura put onto the battlefield names (auraHost, CR
// 303.4f), carried through the same pause and resume, attached between the
// zone move and EventETB by attachEnteringAuraLocked, so an enters trigger
// already sees the attachment. The host is named, not asked: the settle step
// only asks for an Aura. AttachForEffect is the one attach verb, so a host
// that has left, or one the Equipment can't be attached to, leaves it
// unattached (CR 301.5c, CR 701.3b).
func (ev *ReplacementEvent) AttachAtETB(host uuid.UUID) {
	if ev == nil || host == uuid.Nil || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield {
		return
	}
	ev.auraHost = TargetRef{Kind: TargetCard, ID: host}
	ev.auraHostFixed = true
}
