package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gloom Surgeon — Creature — Spirit {1}{B}, 2/1:
//
//	"If combat damage would be dealt to this creature, prevent that damage and exile that many cards from the top of your library."
//
// ADR 0108 §8 (#1906): a prevention static with an additional effect, one
// application per recipient in a damage instance (PreventDamageDealtTo).
// "That many" is the damage it was applied to, so damage that can't be
// prevented is dealt and still exiles that many cards (CR 615.12; the
// ruling), and the damage is prevented even with too few cards left.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "48b8ab2d-9f3f-42b2-9c5d-7497bceae43e",
		Name:         "Gloom Surgeon",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:     ToThisCreature,
				Damage: CombatDamage,
				Then:   gloomSurgeonExileBody,
				Label:  "Gloom Surgeon — prevent combat damage to it and exile that many cards from the top of your library",
			}),
		},
	})
}

// The additional effect: exile "that many" from the top of your library.
var gloomSurgeonExileBody = game.DelayedBody("gloom-surgeon/exile-that-many", gloomSurgeonExile)

func gloomSurgeonExile(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
	return preventedExileLibraryTop(g, item, game.EffectParams{Amount: thatDamage(item)})
}
