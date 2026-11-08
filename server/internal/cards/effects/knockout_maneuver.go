package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knockout Maneuver — Sorcery {2}{G} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Put a +1/+1 counter on target creature you control, then it deals
//	 damage equal to its power to target creature an opponent controls."
//
// Bite Down with a counter first. Two target CLAUSES (#764), read
// positionally: slot 0 gets the counter and is then the damage source,
// slot 1 is the recipient. The counter is its own step, so the power the
// damage reads is the power with the counter on it (and a replacement
// choice on the counter settles first, CR 608.2c). A creature of yours
// that left in response gets no counter and deals nothing; a recipient
// that left takes nothing but your creature still gets its counter
// (CR 608.2b). Deathtouch and lifelink on the source carry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e4931f4-f518-4200-aad4-4ac8acf202e5",
		Name:         "Knockout Maneuver",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // one-sided fight, as Bite Down (#4734)
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: counterOnYourCreatureThenBite(),
	})
}

// counterOnYourCreatureThenBite is "Put a +1/+1 counter on target creature
// you control, then it deals damage equal to its power to target creature
// an opponent controls" over two target clauses: slot 0 gets the counter
// and then deals the damage, slot 1 takes it (Knockout Maneuver, Felling
// Blow). The counter is its own step, so the power read is the power with
// the counter on it.
func counterOnYourCreatureThenBite() func(*game.StackItem, *Context) error {
	return ResolveSteps(
		Instruction(func(ctx *Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			return AddCounter{Target: mine.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		}),
		Instruction(func(ctx *Context) error { return oneSidedBite(ctx.Item, ctx) }),
	)
}
