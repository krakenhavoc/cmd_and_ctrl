package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle of Hoover Dam — Enchantment {3}{W} (Fallout):
//
//	"As this enchantment enters, choose NCR or Legion.
//	 • NCR — At the beginning of your end step, return target creature card
//	   with mana value 3 or less from your graveyard to the battlefield with
//	   a finality counter on it.
//	 • Legion — Whenever a creature you control dies, put two +1/+1
//	   counters on target creature you control."
//
// An anchor-word card, Citadel Siege's: each bullet is a trigger gated on
// the word chosen as it entered (ADR 0071). NCR returns the card under my
// control with a finality counter it enters with (Admiral Brass's return,
// CR 122.1). Legion's trigger targets as it goes on the stack, so the
// creature that died is no candidate; "a creature you control" includes
// the enchantment's controller's creatures only, read as it last existed
// (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91596da5-5b1a-430c-a878-7757ae366b6b",
		Name:         "Battle of Hoover Dam",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Battle of Hoover Dam", "NCR", "Legion"),
		Triggered: []game.TriggeredAbility{
			effects.WhenChosen("NCR", effects.Targeting(
				effects.AtYourEndStep("Battle of Hoover Dam — return target creature card with mana value 3 or less from your graveyard to the battlefield with a finality counter on it",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
							return nil
						}
						_, err := g.ReturnFromGraveyardWithCountersForEffect(id, item.Controller, false, map[string]int{game.CounterFinality: 1})
						return err
					}),
				effects.TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard", effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3)))),
			effects.WhenChosen("Legion", effects.Targeting(
				effects.WheneverACreatureYouControlDies("Battle of Hoover Dam — put two +1/+1 counters on target creature you control",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := ctx.ClauseTarget(0)
						if !ok || t.Kind != game.TargetCard {
							return nil
						}
						return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
					}),
				effects.TargetCreature("target creature you control", effects.YouControl()))),
		},
	})
}
