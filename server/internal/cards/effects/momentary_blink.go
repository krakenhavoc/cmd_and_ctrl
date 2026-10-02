package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Momentary Blink — Instant {1}{W}:
//
//	"Exile target creature you control, then return it to the
//	 battlefield under its owner's control.
//	 Flashback {3}{U} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// An immediate blink (Flicker's one-go shape) that comes back under
// its OWNER's control rather than the caster's — the distinction
// matters for a creature you control but do not own, and is exactly
// why Flicker.Controller is left at its zero value here rather than
// set to the resolving item's controller. The flashback cost is a
// different color from the printed one, the same shape Unburial Rites
// and Deep Analysis already use — a cast PATH, not a second mana cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a3ec6b5d-08ec-4ae0-b1db-c4b87a1849c7",
		Name:             "Momentary Blink",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature you control", YouControl()),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{U}")},
		OnResolve:        flickerTheTargetToItsOwner,
	})
}
