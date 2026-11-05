package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Test of Courage — {G} Instant (Star Trek, #4973):
//
//	"Choose one —
//	 • Alter the Past — Exile target card from a graveyard. You gain 2
//	   life.
//	 • Obliterate the Transwarp Network — Destroy target artifact or
//	   enchantment with mana value 3 or less."
//
// The mode names are flavour words (CR 700.2h). Each mode has one target;
// a target that is illegal on resolution makes the spell do nothing (CR
// 608.2b), so a card that left the graveyard in response gains no life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "101b2387-58a1-4ef9-8395-0024713e217b",
		Name:         "Test of Courage",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Alter the Past — Exile target card from a graveyard. You gain 2 life.",
				TargetCardInGraveyard("target card from a graveyard"),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
				}),
			ModeDoing("Obliterate the Transwarp Network — Destroy target artifact or enchantment with mana value 3 or less.",
				TargetPermanent("target artifact or enchantment with mana value 3 or less", Or(Artifact(), Enchantment()), ManaValueLE(3)),
				DestroyTheModesTarget),
		),
	})
}
