package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corpse Explosion — Sorcery {1}{B}{R} (Streets of New Capenna, #5491):
//
//	"As an additional cost to cast this spell, exile a creature card from
//	 your graveyard.
//	 Corpse Explosion deals damage equal to the exiled card's power to each
//	 creature and each planeswalker."
//
// The cost is ExileFromGraveyardCost with a creature-card match (Draconic
// Intervention's shape): the caster names the card as exile_ids, it is exiled
// with the spell on the stack, and ctx.Exiled() hands it back at resolution.
// The power is read from the card where the cost left it, in exile (its
// printed power; a power of 0 or less deals nothing, CR 120.8). The damage is
// one simultaneous instance to every creature and planeswalker, the spell
// being the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab715312-0103-4681-b559-520026abd350",
		Name:         "Corpse Explosion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true}},
		AdditionalCost: effects.ExileFromGraveyardCost(1, "a creature card", func(c game.Card) bool {
			return c.IsCreature()
		}),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			power := 0
			for _, id := range ctx.Exiled() {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok {
					power = c.CurrentPower()
				}
			}
			return effects.DamageEachMatching(ctx, effects.Or(effects.Creature(), effects.Planeswalker()), power)
		},
	})
}
