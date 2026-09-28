package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Detection Tower — Land (EDHREC rank 3612):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Until end of turn, your opponents and creatures your
//	 opponents control with hexproof can be the targets of spells and
//	 abilities you control as though they didn't have hexproof."
//
// The proof card for #1651's first half (ADR 0038's amendment of
// 2026-09-28, B1). It is Kaya, Bane of the Dead's static, with a
// duration, created by a resolving ability: a waiveHexproof data record
// under the live opponentsAndTheirCreatures scope, read at the
// targeting choke point after the battlefield statics.
//
//   - "You control": only the Tower's controller benefits. A third
//     player still cannot target the hexproof creature.
//   - The creatures keep hexproof. Nothing is removed in layer 6, so
//     their own controller's view still shows the badge.
//   - The set is live (a waiver changes no characteristic, CR 611.2c):
//     a hexproof creature an opponent casts after the ability resolves
//     is covered too.
//   - Shroud and protection are not waived (A2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93695c16-c441-492d-af12-b57df9739846",
		Name:         "Detection Tower",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}: Until end of turn, your opponents and creatures your opponents control with hexproof can be the targets of spells and abilities you control as though they didn't have hexproof.",
			Cost:  Plus(ManaCost("{1}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return WaiveHexproofUntilEOT{Label: "Detection Tower — hexproof waived"}.Apply(NewContext(g, item))
			},
		}},
	})
}
