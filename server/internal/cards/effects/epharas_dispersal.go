package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ephara's Dispersal — Instant {2}{U}:
//
//	"This spell costs {2} less to cast if it targets an attacking
//	 creature.
//	 Return target creature to its owner's hand. Surveil 2."
//
// The reduction reads the announced target (CR 601.2f), as Ajani's
// Response does, with the attacking-creature predicate. The surveil
// is the bounce's continuation, as Unwind History's is: returning a
// commander can pause for the command-zone question.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e0bb57a-f5c7-45e9-b5b9-5bb3ec930108",
		Name:         "Ephara's Dispersal",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets an attacking creature.", AttackingCreature()),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return BounceToHand{Target: t.ID, ThenTail: ctx.Tail(epharasDispersalSurveil)}.Apply(ctx)
			}
			return nil
		},
	})
}

// epharasDispersalSurveil is "Surveil 2" after the bounce, whether or not
// the creature reached a hand. An on-disk key: never renamed, never reused.
var epharasDispersalSurveil = CardTail("then/epharas-dispersal-surveil", func(ctx *Context, _ game.TailOutcome) error {
	return Surveil{Player: ctx.Controller(), N: 2}.Apply(ctx)
})
