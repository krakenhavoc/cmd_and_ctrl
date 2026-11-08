package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rictus Robber — Creature — Zombie Rogue {3}{B}, 4/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, if a creature died this turn, create a
//	 2/2 blue and black Zombie Rogue creature token.
//	 Plot {2}{B} (You may pay {2}{B} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// "If a creature died this turn" is an intervening if (CR 603.4), Old
// Flitterfang's CondBoard over the table-wide tally (any creature, any
// controller, tokens included): checked as the creature enters and again
// as the trigger resolves. The token is Geralf, the Fleshwright's 2/2
// blue and black Zombie Rogue. Plot is the ordinary special action (CR
// 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "91e49e8e-575f-43c8-9263-a8568697e7f4",
		Name:           "Rictus Robber",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "a creature died this turn",
				Test: func(ctx *Context) bool { return ctx.Game.TurnTally.CreaturesDied > 0 },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && g.TurnTally.CreaturesDied > 0
			}, "Rictus Robber — create a 2/2 Zombie Rogue", Do(CreateToken{Template: TokenCard("2/2 blue and black Zombie Rogue"), N: 1}))),
		},
	})
}
