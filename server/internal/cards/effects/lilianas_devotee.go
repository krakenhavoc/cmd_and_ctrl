package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const lilianasDevoteeLabel = "Liliana's Devotee — you may pay {1}{B}; if you do, create a 2/2 black Zombie"

// Liliana's Devotee — Creature — Human Warlock {2}{B}, 2/3:
//
//	"Zombies you control get +1/+0.
//	 At the beginning of your end step, if a creature died this turn, you
//	 may pay {1}{B}. If you do, create a 2/2 black Zombie creature token."
//
// The anthem is Layer 7c over every Zombie you control, read through
// the layered view; it is not "other", though the Devotee is no Zombie.
// The end-step trigger has an intervening "if" (CR 603.4): checked as
// the step begins and again as the ability resolves. The fact is the
// turn's, whoever's creature died and tokens included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3b8c7d0-46b9-4519-808a-154e503a58b6",
		Name:         "Liliana's Devotee",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Zombie"}, YoursOnly: true}, 1, 0),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou,
				func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b11CreaturesDiedThisTurn(g) > 0
				}), lilianasDevoteeLabel, lilianasDevoteeOffer),
		},
	})
}

func lilianasDevoteeOffer(g *game.Game, item *game.StackItem) error {
	if b11CreaturesDiedThisTurn(g) == 0 {
		return nil
	}
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}{B}",
		Question: "Liliana's Devotee — pay {1}{B} to create a 2/2 black Zombie?",
		Then:     lilianasDevoteePaid,
	}.Apply(NewContext(g, item))
}

var lilianasDevoteePaid = PayUnlessBranches("pay-unless/lilianas-devotee-zombie",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CreateToken{Controller: ctx.Controller(), Template: TokenCard("2/2 black Zombie"), N: 1}.Apply(ctx)
	}, nil)
