package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Immolating Gyre — Sorcery {4}{R}{R} (Jumpstart, #5510):
//
//	"Immolating Gyre deals X damage to each creature and planeswalker you
//	 don't control, where X is the number of instant and sorcery cards in
//	 your graveyard."
//
// Riveteers Confluence's sweep over the opponents' creatures and
// planeswalkers, dealt as one damage instance. X is counted as the spell
// resolves (CR 608.2h) over the caster's graveyard, the Gyre itself not yet
// in it (it is put there as the last step of resolution, CR 608.2n).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1888b473-e26a-4882-ae18-5a0a6caee410",
		Name:         "Immolating Gyre",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, AmountIsX: true, OpponentsOnly: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := effects.InstantAndSorceryCardsInGraveyard(ctx.Game, ctx.Controller())
			return effects.DamageEachMatching(ctx, effects.And(effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()), x)
		},
	})
}
