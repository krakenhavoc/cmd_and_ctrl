package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Damage Control Crew — Creature — Human Citizen {3}{G}, 3/3:
//
//	"When this creature enters, choose one —
//	 • Repair — Return target card with mana value 4 or greater from
//	   your graveyard to your hand.
//	 • Impound — Exile target artifact or enchantment."
//
// A modal enters trigger (Archive Arbiter's shape): the mode and its
// target are chosen as the ability goes on the stack (CR 603.3c). The
// Repair card is any card type with mana value 4 or greater, read
// through the engine's mana value. No simplification.
func init() {
	etb := WhenThisEnters("Damage Control Crew — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = ChooseOne(
		ModeDoing("Repair — Return target card with mana value 4 or greater from your graveyard to your hand.",
			TargetCardInGraveyard("target card with mana value 4 or greater in your graveyard", YouOwn(), ManaValueGE(4)),
			ReturnTheModesGraveyardTargetToHand),
		ModeDoing("Impound — Exile target artifact or enchantment.",
			TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			func(_ *game.StackItem, ctx *Context, occ int) error {
				if t, ok := ModeTarget(ctx, occ); ok {
					return ExileTarget{Target: t.ID}.Apply(ctx)
				}
				return nil
			}),
	)
	Register(Spec{
		OracleID:     "c0f9e604-a8bf-4c7d-8480-487e8076e1d8",
		Name:         "Damage Control Crew",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{etb},
	})
}
