package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fangkeeper's Familiar — Creature — Snake {1}{B}{G}{U}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flash
//	 When this creature enters, choose one —
//	 • You gain 3 life and surveil 3.
//	 • Destroy target enchantment.
//	 • Counter target creature spell."
//
// Apothecary Stomper's modal enters trigger: the mode, and its target,
// are chosen as the trigger goes on the stack and the target is re-checked
// at resolution (CR 608.2b). The Familiar was cast with flash, so a
// creature spell is a fine target while it resolves from the stack on top
// of it. The first mode gains the life, then surveils (the prompt is the
// last step, so the surveil happens after the life gain).
//
// No simplification.
func init() {
	trigger := WhenThisEnters("Fangkeeper's Familiar — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("You gain 3 life and surveil 3.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			if err := (GainLife{Player: item.Controller, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			return Surveil{Player: item.Controller, N: 3}.Apply(ctx)
		}),
		ModeDoing("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment()), DestroyTheModesTarget),
		ModeDoing("Counter target creature spell.", TargetSpell("target creature spell", Creature()), CounterTheModesTarget),
	)
	Register(Spec{
		OracleID:        "06343800-b555-4a13-ad4d-19e37d200138",
		Name:            "Fangkeeper's Familiar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
