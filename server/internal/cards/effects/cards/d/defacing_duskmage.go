package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defacing Duskmage // Vandal's Edit — a preparation card (CR 722,
// ADR 0090), Secrets of Strixhaven Commander (#4834):
//
//	Defacing Duskmage {W}{B}, Creature — Dog Warlock 2/2:
//	"Deathtouch
//	 Whenever an opponent draws their second card each turn, this
//	 creature becomes prepared."
//	Vandal's Edit {1}{W}{B}, Instant:
//	"Draw two cards. Each player loses 2 life."
//
// The trigger is Faerie Mastermind's: the drawing player's turn tally
// reads exactly two as the second card is drawn, so it fires on that
// card and no other, once per opponent per turn. The prepare spell is
// Risky Shortcut's body (b35DrawTwoThenEachPlayerLosesTwo).
//
// No simplification.
const DefacingDuskmageOracleID = "27c366cf-00d7-47d4-9577-736bfc4a3b81"

func init() {
	effects.Register(effects.Spec{
		OracleID:        DefacingDuskmageOracleID,
		Name:            "Defacing Duskmage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OpponentDrewTheirSecondCardThisTurn(ev, source, g)
			}, "Defacing Duskmage — becomes prepared", effects.SocB4ThisBecomesPrepared),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     DefacingDuskmageOracleID + "#1",
		Name:         "Vandal's Edit",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.B35DrawTwoThenEachPlayerLosesTwo,
	})
}
