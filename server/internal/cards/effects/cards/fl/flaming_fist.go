package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flaming Fist — Legendary Enchantment — Background {2}{W} (EDHREC rank
// 5066):
//
//	"Commander creatures you own have "Whenever this creature attacks,
//	 it gains double strike until end of turn.""
//
// An ADR 0093 grant to each commander creature you own, under anyone's
// control (choose_a_background.go). The trigger is the creature's own:
// any attack, on a player, a planeswalker or a battle, and the double
// strike is on the creature that attacked.
//
// No simplification.
const flamingFistGrant = "flaming-fist/double-strike"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "cdaefc96-4560-43bb-9514-bf87657bd481",
		Name:         "Flaming Fist",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: flamingFistGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisAttacks("Flaming Fist — double strike until end of turn", func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GrantKeywordUntilEOT{
						Target:   ctx.Source(),
						Keywords: []string{"double strike"},
						Label:    "Flaming Fist — double strike",
					}.Apply(ctx)
				}),
			},
			Text: "Whenever this creature attacks, it gains double strike until end of turn.",
		}},
		Static: []game.StaticAbility{effects.GrantToCommanderCreaturesYouOwn(flamingFistGrant)},
	})
}
