package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valkyrie Harbinger — Creature — Angel Cleric {4}{W}{W}, 4/5 (EDHREC
// rank 2834):
//
//	"Flying
//	 Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)
//	 At the beginning of each end step, if you gained 4 or more life
//	 this turn, create a 4/4 white Angel creature token with flying
//	 and vigilance."
//
// The lifegain deck's Angel that makes more Angels — its own lifelink
// hit is four, so connecting is enough. "Each end step" is any
// player's (The Gaffer's shape), and "if you gained 4 or more life
// this turn" is an intervening-if (CR 603.4): checked as the end step
// begins, and again as the trigger resolves, both off the per-turn
// tally's LifeGained cell (b15LifeGainedThisTurn
// — a lifelink hit, a GainLife primitive and a drain's gain half all
// count, a loss does not). The token is Parhelion II's Angel.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "18a5e71e-eb75-4ac7-bedd-2220aaa24f78",
		Name:            "Valkyrie Harbinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b26EachEndStepAndYouGainedAtLeastThisTurn(ev, source, g, 4)
			}, "Valkyrie Harbinger — create a 4/4 Angel with flying and vigilance", func(g *game.Game, item *game.StackItem) error {
				if b15LifeGainedThisTurn(g, item.Controller) < 4 {
					return nil
				}
				return CreateToken{Controller: item.Controller, Template: TokenCard("4/4 colorless Angel with flying and vigilance"), N: 1}.Apply(NewContext(g, item))
			}),
		},
	})
}
