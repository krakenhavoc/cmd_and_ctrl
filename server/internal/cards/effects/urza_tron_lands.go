package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// urza_tron_lands.go — the Tron cycle (EDHREC ranks 840–856):
//
//	Urza's Tower        "{T}: Add {C}. If you control an Urza's Mine
//	                     and an Urza's Power-Plant, add {C}{C}{C} instead."
//	Urza's Mine         "{T}: Add {C}. If you control an Urza's
//	                     Power-Plant and an Urza's Tower, add {C}{C} instead."
//	Urza's Power Plant  "{T}: Add {C}. If you control an Urza's Mine
//	                     and an Urza's Tower, add {C}{C} instead."
//
// Three lands that are one card, so one table. The scaled amount is
// #356's ProducedFunc: the produced string is computed at activation
// from whether the controller has the other two pieces, read by
// land subtype off the post-layer characteristic (so a land that
// only became an Urza's Mine through a type-changing effect counts,
// as printed). Scryfall's type line is "Land — Urza's Power-Plant",
// one hyphenated subtype word.
//
// No simplification.
func init() {
	for _, l := range []struct{ oracleID, name, needA, needB, full string }{
		{"32fbb638-ab14-4e8b-a07a-d4c44e3496f2", "Urza's Tower", "Mine", "Power-Plant", "{C}{C}{C}"},
		{"33e85a8a-86df-4cdc-a9cc-8cbabe92c3c0", "Urza's Mine", "Power-Plant", "Tower", "{C}{C}"},
		{"e11966cd-2ee3-4df4-b099-abf42dcdf0db", "Urza's Power Plant", "Mine", "Tower", "{C}{C}"},
	} {
		hasA := ControlsAtLeast(1, MatchLandSubtype(l.needA))
		hasB := ControlsAtLeast(1, MatchLandSubtype(l.needB))
		full := l.full
		Register(Spec{
			OracleID:     l.oracleID,
			Name:         l.name,
			Completeness: CompletenessFull,
			ManaAbilities: []ManaAbility{{
				Cost: ManaAbilityCost{Tap: true},
				ProducedFunc: func(g *game.Game, controller, source uuid.UUID) string {
					if hasA(g, controller, source) && hasB(g, controller, source) {
						return full
					}
					return "{C}"
				},
				Label: "Add {C}, or " + l.full + " with an Urza's " + l.needA + " and an Urza's " + l.needB,
			}},
		})
	}
}
