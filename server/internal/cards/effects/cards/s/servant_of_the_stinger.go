package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
		return effects.YouCommittedACrimeThisTurn(g, source.Controller)
	}
	effects.Register(effects.Spec{
		OracleID:        "0c9ec9a1-738a-4f69-b367-5dbf3d046bf1",
		Name:            "Servant of the Stinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(servantOfTheStingerCrime, effects.Optional(
				effects.On(game.EventDealDamage, effects.AllOf(effects.ThisDealtCombatDamageToAPlayer, crime),
					"Servant of the Stinger — sacrifice it, then search your library for a card",
					servantOfTheStingerSearch),
				"Sacrifice Servant of the Stinger to search your library for a card?")),
		},
	})
}

// servantOfTheStingerCrime is the intervening "if you've committed a
// crime this turn", read for the ability's controller.
var servantOfTheStingerCrime = effects.CondBoard{
	Name: "you've committed a crime this turn",
	Test: func(ctx *effects.Context) bool { return effects.YouCommittedACrimeThisTurn(ctx.Game, ctx.Controller()) },
}

// servantOfTheStingerSearch is "sacrifice this creature. If you do,
// search your library for a card, put it into your hand, then shuffle".
// An on-disk key.
var servantOfTheStingerSearch = effects.SacrificeThisThen("then/servant-of-the-stinger-search", func(ctx *effects.Context) error {
	return effects.SearchLibrary{
		Player:  ctx.Controller(),
		Dest:    game.ZoneHand,
		Limit:   1,
		Shuffle: true,
		Reason:  "Servant of the Stinger — search your library for a card",
	}.Apply(ctx)
})
