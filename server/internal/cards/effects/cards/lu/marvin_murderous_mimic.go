package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marvin, Murderous Mimic — Legendary Artifact Creature — Toy {2}, 2/2:
//
//	"Marvin has all activated abilities of creatures you control that
//	 don't have the same name as this creature."
//
// ADR 0155's headline card, a layer-6 borrow (CR 113.10, CR 613.1f) over
// the battlefield. The lent rows are whatever each lender has right now,
// grants and removals included, so a Llanowar Elves gives Marvin its
// "{T}: Add {G}" and Marvin's controller activates it, taps Marvin for
// it and cannot use it while Marvin is summoning sick (CR 302.6).
//
// "Don't have the same name as this creature" is a name comparison with
// the source, which is why the lender predicate sees the source: a second
// Marvin lends nothing to the first.
//
// No simplification.
func init() {
	const clause = "Marvin has all activated abilities of creatures you control that don't have the same name as this creature."
	effects.Register(effects.Spec{
		OracleID:     "22f0dbbb-9d8d-42bc-8903-fc0b65e00952",
		Name:         "Marvin, Murderous Mimic",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label: clause,
			Lenders: effects.BattlefieldLenders(func(_ *game.Game, source *game.Card, c game.Card) bool {
				return c.IsCreature() && c.Controller == source.Controller && c.Name != source.Name
			}),
		})},
	})
}
