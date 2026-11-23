package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daily Bugle Reporters — Creature — Human Citizen {3}{W}, 2/3 (Marvel
// Super Heroes Commander):
//
//	"When this creature enters, choose one —
//	 • Puff Piece — Put a +1/+1 counter on each of up to two target creatures.
//	 • Investigative Journalism — Return target creature card with mana
//	   value 2 or less from your graveyard to your hand."
//
// A modal enters trigger (CR 700.2): each bullet owns its target group,
// read back by occurrence, and a target that went away is skipped on
// resolution (CR 608.2b).
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Daily Bugle Reporters — choose one", func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Puff Piece — Put a +1/+1 counter on each of up to two target creatures.",
			effects.TargetCreature("up to two target creatures").WithCount(0, 2),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				for _, t := range ctx.ModeTargets(occ) {
					if !ctx.IsTargetLegal(t) || t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		effects.ModeDoing("Investigative Journalism — Return target creature card with mana value 2 or less from your graveyard to your hand.",
			effects.TargetCardInGraveyard("target creature card with mana value 2 or less in your graveyard", effects.YouOwn(), effects.Creature(), effects.ManaValueLE(2)),
			effects.ReturnTheModesGraveyardTargetToHand),
	)
	effects.Register(effects.Spec{
		OracleID:     "2e677ded-7af3-4af0-81df-b70d6ee616ec",
		Name:         "Daily Bugle Reporters",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
