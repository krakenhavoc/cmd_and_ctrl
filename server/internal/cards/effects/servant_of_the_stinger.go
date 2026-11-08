package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Servant of the Stinger — Creature — Human Warlock {1}{B}, 1/3 (Outlaws
// of Thunder Junction, #4468):
//
//	"Deathtouch
//	 Whenever this creature deals combat damage to a player, if you've
//	 committed a crime this turn, you may sacrifice this creature. If you
//	 do, search your library for a card, put it into your hand, then
//	 shuffle."
//
// The combat-damage trigger carries a CR 603.4 intervening "if": the
// crime (YouCommittedACrimeThisTurn, this turn's event log) is checked
// as it triggers and again before the "you may" is asked. On yes the
// Servant is sacrificed as the ability resolves, and the search runs
// only if it really went (SacrificeThisThen): a Servant that left the
// battlefield in response is not sacrificed, so nothing is found. The
// search names no quality, so it is unrevealed.
//
// No simplification.
func init() {
	crime := func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return YouCommittedACrimeThisTurn(g, source.Controller)
	}
	Register(Spec{
		OracleID:        "0c9ec9a1-738a-4f69-b367-5dbf3d046bf1",
		Name:            "Servant of the Stinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(servantOfTheStingerCrime, Optional(
				On(game.EventDealDamage, AllOf(ThisDealtCombatDamageToAPlayer, crime),
					"Servant of the Stinger — sacrifice it, then search your library for a card",
					servantOfTheStingerSearch),
				"Sacrifice Servant of the Stinger to search your library for a card?")),
		},
	})
}

// servantOfTheStingerCrime is the intervening "if you've committed a
// crime this turn", read for the ability's controller.
var servantOfTheStingerCrime = CondBoard{
	Name: "you've committed a crime this turn",
	Test: func(ctx *Context) bool { return YouCommittedACrimeThisTurn(ctx.Game, ctx.Controller()) },
}

// servantOfTheStingerSearch is "sacrifice this creature. If you do,
// search your library for a card, put it into your hand, then shuffle".
// An on-disk key.
var servantOfTheStingerSearch = SacrificeThisThen("then/servant-of-the-stinger-search", func(ctx *Context) error {
	return SearchLibrary{
		Player:  ctx.Controller(),
		Dest:    game.ZoneHand,
		Limit:   1,
		Shuffle: true,
		Reason:  "Servant of the Stinger — search your library for a card",
	}.Apply(ctx)
})
