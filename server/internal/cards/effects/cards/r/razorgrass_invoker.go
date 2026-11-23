package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razorgrass Invoker — Creature — Elf Scout {3}{G}, 4/3 (Foundations
// Jumpstart, #5464):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	 {8}: This creature and up to one other target creature each get
//	 +3/+3 until end of turn."
//
// "Other" is object identity (Another): the Invoker can't name itself.
// With no target chosen only the Invoker grows. With one chosen that is
// illegal by resolution, the ability has no legal target left and does
// nothing at all (CR 608.2b), the Invoker included. An Invoker that left
// the battlefield in response gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ea2a6bee-8634-429d-a5ca-e5f9f7eb9350",
		Name:            "Razorgrass Invoker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{8}: This creature and up to one other target creature each get +3/+3 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{8}"),
			Targets: effects.Another(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
			Effect:  razorgrassInvokerPump,
		}},
	})
}

// razorgrassInvokerPump gives the Invoker, while it is still the same
// permanent, and the target, while still legal, +3/+3 until end of turn
// each.
func razorgrassInvokerPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	const label = "Razorgrass Invoker — +3/+3"
	if effects.B09SourceStillOnBattlefield(g, item) {
		if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 3, Toughness: 3, Label: label}).Apply(ctx); err != nil {
			return err
		}
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: label}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
