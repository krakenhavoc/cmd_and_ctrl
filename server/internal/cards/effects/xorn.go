package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Xorn — Creature — Elemental {2}{R}, 3/2 (EDHREC top 1,000, #3654):
//
//	"If you would create one or more Treasure tokens, instead create
//	 those tokens plus an additional Treasure token."
//
// A CR 614 token-creation replacement like Academy Manufactor, but it
// adds a group rather than rewriting one: the instruction's own Treasures
// are left alone and one more is appended, once per instruction. It
// applies only when a Treasure is among the tokens being created under
// Xorn's controller, and a mixed instruction (a Treasure and a Soldier)
// still gets the one extra Treasure. Two Xorns each add one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "11d8fab8-93af-4291-80bd-cf9436e99f4b",
		Name:         "Xorn",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens &&
					ev.TokenController == src.Controller &&
					ev.TokenTemplatesMatch(func(t game.Card) bool { return t.HasSubtype("Treasure") })
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.TokenGroups = append(ev.TokenGroups, game.TokenGroup{Template: ruleDefinedToken(TreasureToken()), Count: 1})
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Xorn: an additional Treasure",
		}},
	})
}
