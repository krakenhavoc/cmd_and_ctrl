package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Old Flitterfang — Legendary Creature — Rat Faerie {4}{B}, 3/4 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 At the beginning of each end step, if a creature died this turn,
//	 create a Food token. (It's an artifact with "{2}, {T}, Sacrifice
//	 this token: You gain 3 life.")
//	 {2}{B}, Sacrifice another creature or artifact: Old Flitterfang gets
//	 +2/+2 until end of turn."
//
// The end step trigger fires at every player's end step, and "a
// creature died this turn" counts any creature, any controller (the
// turn's tally Zombie Ogre reads), as an intervening if (CR 603.4)
// checked as it triggers and as it resolves, written in Intermediate
// Chirography's CondBoard shape so the row declares its footprint. The Food is the
// catalog's token (CR 111.10b). The sacrifice is Compleated
// Huntmaster's "another creature or artifact" cost, so Old Flitterfang
// can't pay it with itself, and a creature sacrificed to it makes the
// end step's Food.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f95bdf0d-5bd1-4215-b143-0930973807ce",
		Name:            "Old Flitterfang",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "a creature died this turn",
				Test: func(ctx *Context) bool { return ctx.Game.TurnTally.CreaturesDied > 0 },
			}, On(game.EventBeginEndStep, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.TurnTally.CreaturesDied > 0
			}, "Old Flitterfang — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice another creature or artifact: Old Flitterfang gets +2/+2 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{2}{B}"),
				SacrificeAnotherN(1, "another creature or artifact", Or(Creature(), Artifact()))),
			Effect: thisGetsUntilEndOfTurn(2, 2, "Old Flitterfang — +2/+2"),
		}},
	})
}
