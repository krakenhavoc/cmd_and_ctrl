package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sauron, Dino Devotee — {3}{G}{G} Legendary Creature — Dinosaur Vampire
// Villain, 4/4 (Marvel Super Heroes, #4455):
//
//	"Flying
//	 Whenever Sauron enters or attacks, choose one —
//	 • Cure Cancer — You gain 3 life.
//	 • Turn People into Dinosaurs — Put a saurian counter on another target
//	   creature. It's a green Dinosaur with base power and toughness 5/5 for
//	   as long as it has a saurian counter on it."
//
// One trigger on two events (Avengers Quinjet's shape) with two bullets. The
// second is Aven Mimeomancer's counter-held duration (CR 611.2b): the counter
// goes on, then a continuous effect that lasts exactly while a saurian counter
// is on it. "Is a green Dinosaur" replaces the creature's creature types and
// its colors (CR 205.1a, 105.3) and sets base 5/5 (layer 7b), so +1/+1
// counters and pumps still apply on top.
//
// No simplification.
func init() {
	trigger := WhenThisEntersOrAttacks("Sauron — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Cure Cancer — You gain 3 life.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Amount: 3}.Apply(ctx)
			}),
		ModeDoing("Turn People into Dinosaurs — Put a saurian counter on another target creature. It's a green Dinosaur with base power and toughness 5/5 for as long as it has a saurian counter on it.",
			Another(TargetCreature("another target creature", Creature())),
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CounterThenWhileItHasIt(ctx, FirstLegalBattlefieldTarget(ctx), "saurian",
					"Sauron — a green Dinosaur with base power and toughness 5/5 while it has a saurian counter",
					game.SetCreatureTypesMod(ctx.TextSubtypes("Dinosaur")...), game.SetColorsMod(ctx.TextColors("G")...),
					game.SetBasePowerMod(5), game.SetBaseToughnessMod(5))
			}),
	)
	Register(Spec{
		OracleID:        "f5295bfe-7e73-44ad-9cfd-200e31c5d52b",
		Name:            "Sauron, Dino Devotee",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
