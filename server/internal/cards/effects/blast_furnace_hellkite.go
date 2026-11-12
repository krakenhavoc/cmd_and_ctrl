package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blast-Furnace Hellkite — Creature — Dragon {7}{R}{R}, 5/5 (The
// Brothers' War Commander, #5657):
//
//	"Artifact offering (You may cast this spell as though it had flash by
//	 sacrificing an artifact and paying the difference in mana costs
//	 between this and the sacrificed artifact. Mana cost includes color.)
//	 Flying, double strike
//	 Creatures attacking your opponents have double strike."
//
// Artifact offering is Offering (offering.go, CR 702.48): an optional
// additional cost that sacrifices one artifact, takes that artifact's
// mana cost off the total cost symbol by symbol, and lets the cast that
// announces it be made any time its controller could cast an instant. A
// Hellkite cast without the offering costs {7}{R}{R} at sorcery speed.
//
// The static is a layer 6 grant to EVERY creature attacking a player who
// is one of the Hellkite's controller's opponents, whoever controls it:
// an opponent's creature attacking another opponent has double strike
// too, and a creature attacking you, a planeswalker or a battle does not
// (Oviya, Automech Artisan's "attacking one of your opponents"). It is
// declared DependsOnAttackingStatus, so a declaration, a creature put
// onto the battlefield attacking, a removal from combat and the end of
// combat all rebuild the cached characteristics.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "c13079d5-6749-43e5-a8ee-ef3af31ff679",
		Name:          "Blast-Furnace Hellkite",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offering("artifact", Artifact())},
		Static: []game.StaticAbility{{
			Layer:                    game.Layer6Ability,
			Label:                    "Creatures attacking your opponents have double strike.",
			DependsOnAttackingStatus: true,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && socB8AttackingAnOpponentOf(g, source.Controller, *target)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				appendKeywordsTo(c, []string{"double strike"})
			},
		}},
	})
}
