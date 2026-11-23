package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trostani Discordant — Legendary Creature — Dryad {3}{G}{W}, 1/4 (Ravnica:
// Clue Edition):
//
//	"Other creatures you control get +1/+1.
//	 When Trostani enters, create two 1/1 white Soldier creature tokens with
//	 lifelink.
//	 At the beginning of your end step, each player gains control of all
//	 creatures they own."
//
// The anthem is a layer-7 static. The end-step ability registers, for every
// creature someone else controls, an indefinite control effect for its owner
// (CR 611.2a, 613.1b), the shape of Alicia Masters' Sense the Good.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a2fdc0e-e0de-47fc-95e4-04e725ddf060",
		Name:         "Trostani Discordant",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Trostani Discordant — create two 1/1 white Soldier creature tokens with lifelink",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier with lifelink"), N: 2})),
			effects.AtYourEndStep("Trostani Discordant — each player gains control of all creatures they own",
				effects.Do(effects.Instruction(effects.BnsB22EachPlayerRegainsTheirCreatures("Trostani Discordant — its owner gains control of it")))),
		},
	})
}
