package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Park Heights Pegasus — Creature — Pegasus {G}{W}, 2/1 (Streets of New
// Capenna):
//
//	"Flying, trample
//	 Whenever this creature deals combat damage to a player, draw a card if
//	 you had two or more creatures enter the battlefield under your control
//	 this turn."
//
// The condition is not an intervening "if": it is read as the trigger
// resolves, off the turn's entry tally for creatures (Spider-UK's), tokens
// and the Pegasus itself included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fd92b588-f273-4206-9887-34e6877fc14f",
		Name:            "Park Heights Pegasus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Park Heights Pegasus — draw a card if two or more creatures entered under your control this turn",
				Do(If{Cond: parkHeightsPegasusCond, Then: []Applier{DrawCards{N: 1}}})),
		},
	})
}

var parkHeightsPegasusCond = CondBoard{
	Name: "two or more creatures entered the battlefield under your control this turn",
	Test: func(ctx *Context) bool {
		return ctx.Game.EnteredWithCardTypeThisTurn(ctx.Controller(), "creature") >= 2
	},
}
