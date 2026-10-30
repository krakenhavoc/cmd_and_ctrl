package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hyperion, Supreme Hero — {4}{W}{W}{W} Legendary Creature — Eternal Hero
// 7/7 (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 Flying
//	 If a source would deal damage to you or a Hero you control, prevent
//	 all but 1 of that damage."
//
// The prevention is a CR 615 static replacement read off the board (the
// shape of Ajani Steadfast's emblem, which shares preventAllButOneOfTheDamage):
// damage of 2 or more to its controller or to a Hero they control drops to
// 1; damage of 1 or less is left alone. "Hero" is a word of Hyperion's own
// text. It protects Hyperion himself (he is a Hero) and stops working when
// he leaves the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4f4a5966-3713-45c7-acc2-9ddb4ff73eab",
		Name:            "Hyperion, Supreme Hero",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Replacements: []game.ReplacementEffect{
			preventAllButOneOfTheDamage(
				"Hyperion, Supreme Hero — prevent all but 1 of that damage",
				"Prevent all but 1 of the damage dealt to you or a Hero you control.",
				func(target uuid.UUID, g *game.Game, src *game.Card) bool {
					if target == src.Controller {
						return true
					}
					c, ok := g.LookupCardForEffect(target)
					return ok && g.Battlefield.Contains(target) && c.Controller == src.Controller &&
						c.HasSubtype(src.TextSubtype("Hero"))
				}),
		},
	})
}
