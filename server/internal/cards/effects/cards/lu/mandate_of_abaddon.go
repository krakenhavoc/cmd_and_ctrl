package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mandate of Abaddon — Sorcery {3}{B} (Warhammer 40,000 Commander, #5948):
//
//	"Choose target creature you control. Destroy all creatures with power
//	 less than that creature's power."
//
// The chosen creature's power is read as the spell resolves, through the
// layers; a target that left or stopped being a creature you control
// fizzles the spell (CR 608.2b). The chosen creature is never destroyed by
// its own power, and every creature with strictly less power dies, theirs
// and yours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20582c7b-edd3-4189-98ec-8fabd39034b0",
		Name:         "Mandate of Abaddon",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			chosen, ok := ctx.Game.LookupCardForEffect(item.Targets[0].ID)
			if !ok || !effects.OnBattlefield(ctx.Game, chosen.InstanceID) {
				return nil
			}
			power := chosen.CurrentPower()
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), PowerLessThan(power))}.Apply(ctx)
		},
	})
}

// PowerLessThan passes a creature whose current power is below n (Mandate of
// Abaddon's "power less than that creature's power").
func PowerLessThan(n int) effects.CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.CurrentPower() < n }
}
