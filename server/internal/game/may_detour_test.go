package game

import (
	"testing"

	"github.com/google/uuid"
)

// may_detour_test.go — a test vehicle for the exit-pause machinery.
//
// Before ADR 0115 the CR 903.9 built-in paused every commander headed
// for a graveyard or exile on a "may" prompt, and a family of tests
// used a commander to reach the machinery behind that pause: a batch
// exit that waits for a paused leg, a continuation that runs once the
// leg lands, an undo into the open prompt, a prompt whose route was
// dropped. Since ADR 0115 a commander goes to a graveyard or exile at
// once (CR 903.9a asks afterwards), but the machinery is still reached
// by any optional exit replacement, so those tests keep it covered with
// the replacement below standing in for the commander's.
//
// mayDetourForTest is an optional ("may") replacement on ONE card:
// when it would be put into a graveyard or exile, its owner may put it
// into the command zone instead. Same question kind
// (optional_replacement), same chooser (the owner), same destination
// as the old built-in, on a card that is NOT a commander, so the
// CR 903.9a state-based action never asks about it.
func mayDetourForTest(only, owner uuid.UUID) ReplacementEffect {
	return ReplacementEffect{
		Watches:        []EventKind{EventZoneMove, EventDiscardCard},
		Optional:       true,
		PromptQuestion: "Send it to the command zone instead?",
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			if !isExitMove(ev.Kind) || ev.CardID != only {
				return false
			}
			return ev.NewZone == ZoneGraveyard || ev.NewZone == ZoneExile
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.NewZone = ZoneCommand
			ev.NewZoneOwner = owner
			return nil
		},
		Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return owner },
		Label:      "Test may-detour",
	}
}

// seatDetouredCard puts a plain legendary creature card (not a
// commander) into `zone`, owned and controlled by `owner`, with
// mayDetourForTest registered on it, and returns its ID. The drop-in
// for seatCommander in a test whose subject is the pause, not the
// commander.
func seatDetouredCard(t *testing.T, g *Game, zone *Zone, owner *Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	zone.PushTop(Card{
		InstanceID: id,
		Name:       "Atraxa",
		TypeLine:   "Legendary Creature — Angel",
		Power:      4,
		Toughness:  4,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	g.WithWriteLock(func() { g.RegisterReplacementForTest(mayDetourForTest(id, owner.ID)) })
	return id
}
