package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Collector's Cage — Artifact {1}{W} (The Big Score, #4468):
//
//	"Hideaway 5 (When this artifact enters, look at the top five cards of
//	 your library, exile one face down, then put the rest on the bottom in
//	 a random order.)
//	 {1}, {T}: Put a +1/+1 counter on target creature you control. Then if
//	 you control three or more creatures with different powers, you may
//	 play the exiled card without paying its mana cost."
//
// Clive's Hideaway's shape: the counter goes on first and the condition is
// read afterwards (CR 608.2c), through the layered powers, so the counter
// itself can make the third distinct power. The counter waits for any
// prompt it asks (a CR 616 order) before the condition is read.
//
// No simplification.
func init() {
	const name = "Collector's Cage"
	Register(Spec{
		OracleID:     "8f5b71fd-469f-4370-a806-9bfefbae03a2",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Hideaway(name, 5)},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Put a +1/+1 counter on target creature you control. Then if you control three or more creatures with different powers, you may play the exiled card without paying its mana cost.",
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: Do(
				Instruction(func(ctx *Context) error { return PlusOneCounterOnChosenTarget(ctx.Game, ctx.Item) }),
				Instruction(func(ctx *Context) error {
					if rfCreatureCDifferentPowers(ctx.Game, ctx.Item.Controller, ctx.Item.SourceCardID) < 3 {
						return nil
					}
					return PlayHiddenCard{
						Source: HiddenRefOfActivation(ctx.Item),
						Label:  name + " — play the exiled card without paying its mana cost",
					}.Apply(ctx)
				}),
			),
		}},
	})
}
