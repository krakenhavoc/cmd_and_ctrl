package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Delver — 3/2 Creature — Phyrexian Zombie for {3}{B}{B}
// (EDHREC rank 3968):
//
//	"When this creature enters, return target creature card from your
//	 graveyard to the battlefield. You lose life equal to that card's
//	 mana value."
//
// Five mana for a 3/2 and a Zombify stapled together, with the life
// loss as the price. It is in the batch because the rider is the one
// place in the reanimation family where the ORDER of two reads
// matters: the life lost is "that CARD's mana value", read off the
// card as it sat in the graveyard, not off the permanent that just
// arrived.
//
// Those differ more often than they look. A creature card with {X} in
// its cost has mana value 0 everywhere but the stack (CR 202.3b), so
// reanimating Walking Ballista costs no life at all. A card the
// permanent side of which has been altered — one the layer engine has
// turned into something else on entry, a face-down side, a copy
// effect — still charges its own printed value. And a card whose
// return was replaced (a Rest in Peace-style exile replacement, or a
// creature that never reached the battlefield) leaves nothing to
// charge for.
//
// reanimateSingleTarget hands the pre-move card back for exactly this
// reason, so the value is read while it is still guaranteed findable.
//
// "Your graveyard" is the narrow half of the family: the creature
// returns to its owner, and an opponent's pile is never a legal
// target.
//
// The life loss is LOSS, not damage: prevention and protection do
// nothing about it, and a Delver reanimating a ten-drop with five
// life left kills its own controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a13cbac0-4c76-4970-b61e-5f4e020ee95c",
		Name:         "Phyrexian Delver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Phyrexian Delver — reanimate, then lose life equal to its mana value",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					card, ok := reanimateSingleTarget(ctx, item.Controller)
					if !ok {
						return nil
					}
					mv := card.ManaValue()
					if mv <= 0 {
						return nil
					}
					return g.ChangePlayerLifeForEffect(ctx.Source(), item.Controller, -mv)
				}), targetCreatureInYourGraveyard()),
		},
	})
}
