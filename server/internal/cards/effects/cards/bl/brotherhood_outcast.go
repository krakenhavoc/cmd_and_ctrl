package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brotherhood Outcast — Creature — Human Soldier {2}{W}, 3/2 (Fallout):
//
//	"When this creature enters, choose one —
//	 • Return target Aura or Equipment card with mana value 3 or less from
//	   your graveyard to the battlefield.
//	 • Put a shield counter on target creature."
//
// A modal enters trigger (Adaptive Sporesinger's shape). The first bullet
// is One Last Job's one-card return (effects.R35OneLastJobReturn), so a returning Aura chooses
// what it enchants as it enters (CR 303.4f) and stays in the graveyard if
// nothing can be enchanted (CR 303.4g).
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Brotherhood Outcast — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Return target Aura or Equipment card with mana value 3 or less from your graveyard to the battlefield.",
			effects.TargetCardInGraveyard("target Aura or Equipment card with mana value 3 or less from your graveyard", effects.YouOwn(),
				effects.Or(effects.OfSubtype("Aura"), effects.OfSubtype("Equipment")), effects.ManaValueLE(3)),
			effects.R35OneLastJobReturn),
		effects.ModeDoing("Put a shield counter on target creature.",
			effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterShield, N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "79d0cdc1-2e27-48c5-8840-430490068f68",
		Name:         "Brotherhood Outcast",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
