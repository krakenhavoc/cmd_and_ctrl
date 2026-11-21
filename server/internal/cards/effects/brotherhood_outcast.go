package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brotherhood Outcast — Creature — Human Soldier {2}{W}, 3/2 (Fallout):
//
//	"When this creature enters, choose one —
//	 • Return target Aura or Equipment card with mana value 3 or less from
//	   your graveyard to the battlefield.
//	 • Put a shield counter on target creature."
//
// A modal enters trigger (Adaptive Sporesinger's shape). The first bullet
// is One Last Job's one-card return (oneLastJobReturn), so a returning Aura chooses
// what it enchants as it enters (CR 303.4f) and stays in the graveyard if
// nothing can be enchanted (CR 303.4g).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Brotherhood Outcast — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeDoing("Return target Aura or Equipment card with mana value 3 or less from your graveyard to the battlefield.",
			TargetCardInGraveyard("target Aura or Equipment card with mana value 3 or less from your graveyard", YouOwn(),
				Or(OfSubtype("Aura"), OfSubtype("Equipment")), ManaValueLE(3)),
			oneLastJobReturn),
		ModeDoing("Put a shield counter on target creature.",
			TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				t, ok := ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				return AddCounter{Target: t.ID, Kind: game.CounterShield, N: 1}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "79d0cdc1-2e27-48c5-8840-430490068f68",
		Name:         "Brotherhood Outcast",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
