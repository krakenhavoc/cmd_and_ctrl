package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Company Commander — Creature — Human Soldier {2}{W}{B}, 2/4:
//
//	"Command Section — When this creature enters, create a number of 1/1
//	 white Soldier creature tokens equal to the number of opponents you have.
//	 Bring it Down! — Whenever this creature attacks, creatures you control
//	 gain deathtouch until end of turn."
//
// The count is the opponents still in the game as the trigger resolves
// (Chittering Witch's shape). The deathtouch goes to every creature you
// control as the trigger resolves; one that arrives later does not get it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53643dd7-183b-4062-b7a3-e7723d23bdb0",
		Name:         "Company Commander",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Company Commander — create a number of 1/1 white Soldier tokens equal to the number of opponents you have",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier"), N: len(ctx.Opponents())}.Apply(ctx)
				}),
			effects.WheneverThisAttacks("Company Commander — creatures you control gain deathtouch until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"deathtouch"},
						Label:    "Company Commander — deathtouch",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
