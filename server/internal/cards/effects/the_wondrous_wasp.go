package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Wondrous Wasp — {1}{U} Legendary Creature — Human Hero 2/1
// (Marvel Super Heroes, #4455):
//
//	"Flash
//	 Flying
//	 Wasp's Sting — When The Wondrous Wasp enters, tap up to one target
//	 creature. It loses all abilities for as long as The Wondrous Wasp
//	 remains on the battlefield."
//
// The tap is immediate. The ability loss is one layer-6 scoped effect that
// ends when the Wasp leaves (CR 611.2b, DurationWhileSourceRemains); a Wasp
// that left before the trigger resolved taps the creature and takes nothing
// from it, the way Kitesail Larcenist does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5ad8963b-8091-4787-9147-20a1fe7bf0fb",
		Name:            "The Wondrous Wasp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{Targeting(
			WhenThisEnters("The Wondrous Wasp — tap up to one target creature; it loses all abilities while she remains", wondrousWaspSting),
			TargetCreature("up to one target creature").WithCount(0, 1),
		)},
	})
}

func wondrousWaspSting(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	dur, ok := DurationWhileSourceRemains(ctx, ctx.Source())
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.LoseAllAbilitiesMod()},
		Duration: dur,
		Label:    "The Wondrous Wasp — loses all abilities while she remains",
	}.Apply(ctx)
}
