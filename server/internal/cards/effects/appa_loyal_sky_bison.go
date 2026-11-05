package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Appa, Loyal Sky Bison — Legendary Creature — Bison Ally {4}{W}{W}, 4/4
// (Avatar: The Last Airbender, #4459):
//
//	"Flying
//	 Whenever Appa enters or attacks, choose one —
//	 • Target creature you control gains flying until end of turn.
//	 • Airbend another target nonland permanent you control. (Exile it.
//	   While it's exiled, its owner may cast it for {2} rather than its
//	   mana cost.)"
//
// One ability with two conditions (WhenThisEntersOrAttacks), modal, each
// bullet with its own target clause. The second bullet's "another" is
// object identity (Another), so Appa never airbends himself, and the
// airbend is the shared keyword action (Airbend through
// AirbendOtherTarget): the card's OWNER may cast it from exile for {2}
// while it stays there, a token simply ceases to exist (CR 111.7), and an
// airbent commander's owner is asked about the command zone first
// (CR 903.9).
//
// No simplification.
func init() {
	trigger := WhenThisEntersOrAttacks("Appa, Loyal Sky Bison — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeWithPurpose(ModeDoing("Target creature you control gains flying until end of turn.",
			TargetCreature("target creature you control", YouControl()),
			func(item *game.StackItem, ctx *Context, _ int) error {
				return firstTargetGainsUntilEndOfTurn(0, "Appa, Loyal Sky Bison — flying until end of turn", "flying")(ctx.Game, item)
			}), ForTargets(BenefitUntilEOT(0, 0, 0, "flying"))),
		ModeDoing("Airbend another target nonland permanent you control.",
			Another(TargetPermanent("another target nonland permanent you control", Nonland(), YouControl())),
			func(item *game.StackItem, ctx *Context, _ int) error {
				return AirbendOtherTarget(ctx.Game, item)
			}),
	)
	Register(Spec{
		OracleID:        "c6e28cf0-81c7-4f52-a29b-988cc6fd31ee",
		Name:            "Appa, Loyal Sky Bison",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
