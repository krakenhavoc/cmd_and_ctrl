package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Piper Wright, Publick Reporter — Legendary Creature — Human Detective
// {1}{U}, 1/2 (Fallout, Commander Masters era reprint set GAP):
//
//	"Whenever Piper Wright deals combat damage to a player, investigate that
//	 many times.
//	 Whenever you sacrifice a Clue, put a +1/+1 counter on target creature
//	 you control."
//
// "That many" is the combat damage the trigger saw (April, Reporter of the
// Weird's reading), one investigate per point (CR 701.16). The second
// ability is Tireless Tracker's Clue read with a target: EventSacrifice
// fires before the zone move, so the Clue is still there to be read.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ff9598a-0dac-4437-9450-fc8134fa3740",
		Name:         "Piper Wright, Publick Reporter",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Piper Wright, Publick Reporter — investigate that many times",
				func(g *game.Game, item *game.StackItem) error {
					n := item.Trigger.Event.Amount
					if n <= 0 {
						return nil
					}
					return Investigate{Player: item.Controller, N: n}.Apply(NewContext(g, item))
				}),
			Targeting(On(game.EventSacrifice, lonisYouSacrificedAClue,
				"Piper Wright, Publick Reporter — put a +1/+1 counter on target creature you control",
				putPlusOneCounterOnEachLegalTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
