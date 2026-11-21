package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abomination, Terrifying Titan — {3}{R/G} Legendary Creature — Gamma
// Villain 4/4 (Marvel Super Heroes, #4455):
//
//	"Trample
//	 Power-up — {5}{R/G}{R/G}: Put a +1/+1 counter on Abomination. He
//	 fights up to one target creature an opponent controls. (Activate
//	 each power-up ability only once. Reduce the cost by his mana cost
//	 if he entered this turn.)"
//
// The counter goes on first, then he fights the announced target if one
// was chosen and is still legal (CR 701.14); if he has left the
// battlefield neither half does anything. Power-up is the engine's
// (CR 702.193).
//
// No simplification.
func init() {
	ab := effects.PowerUp(
		"{5}{R/G}{R/G}: Put a +1/+1 counter on Abomination. He fights up to one target creature an opponent controls.",
		effects.ManaCost("{5}{R/G}{R/G}"), func(g *game.Game, item *game.StackItem) error {
			if !effects.B09SourceStillOnBattlefield(g, item) {
				return nil
			}
			if err := effects.PlusOneCountersOnThis(1)(g, item); err != nil {
				return err
			}
			return effects.B30SourceFightsFirstLegalTarget(g, item)
		})
	ab.Purpose.Answers = game.AnswerPump
	ab.Targets = effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:        "72101185-4336-480a-b4d7-f8ceb0007a51",
		Name:            "Abomination, Terrifying Titan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{ab},
	})
}
