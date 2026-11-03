package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
const defacingDuskmageOracleID = "27c366cf-00d7-47d4-9577-736bfc4a3b81"

func init() {
	Register(Spec{
		OracleID:        defacingDuskmageOracleID,
		Name:            "Defacing Duskmage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentDrewTheirSecondCardThisTurn(ev, source, g)
			}, "Defacing Duskmage — becomes prepared", socB4ThisBecomesPrepared),
		},
	})
	Register(Spec{
		OracleID:     defacingDuskmageOracleID + "#1",
		Name:         "Vandal's Edit",
		Completeness: CompletenessFull,
		OnResolve:    b35DrawTwoThenEachPlayerLosesTwo,
	})
}
