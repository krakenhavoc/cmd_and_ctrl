package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morgue Burst — Sorcery {4}{B}{R} (Dragon's Maze, S82 OLD b406, #5564):
//
//	"Return target creature card from your graveyard to your hand. Morgue Burst
//	 deals damage to any target equal to the power of the card returned this
//	 way."
//
// Two clauses (CR 601.2c). The creature card's power is read where it sits in
// the graveyard, before it moves (a card outside the battlefield has its
// printed power, or its characteristic-defining one, CR 604.3), and the damage
// is dealt only if the card really reached the hand ("the card returned this
// way"); a card that a replacement sent elsewhere deals nothing. Each clause
// skips a target that is illegal as the spell resolves (CR 608.2b), so the
// damage still lands if the card left, but then nothing was returned and none
// is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee3b72ac-35b6-45d0-bb50-5a5899b291b1",
		Name:         "Morgue Burst",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			effects.TargetAny(),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			card, ok := ctx.ClauseTarget(0)
			if !ok || card.Kind != game.TargetCard {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(card.ID)
			if !found {
				return nil
			}
			power := ctx.Game.PowerAnywhereForEffect(c)
			if err := (effects.ReturnFromGraveyard{Target: card.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
			if z := ctx.Game.FindCardZoneForEffect(card.ID); z == nil || z.Kind != game.ZoneHand || power <= 0 {
				return nil
			}
			dmg, ok := ctx.ClauseTarget(1)
			if !ok {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: dmg.ID, Amount: power}.Apply(ctx)
		},
	})
}
