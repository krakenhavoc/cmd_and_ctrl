package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	ab := powerUpAnswering(game.AnswerPump,
		"{5}{R/G}{R/G}: Put a +1/+1 counter on Abomination. He fights up to one target creature an opponent controls.",
		ManaCost("{5}{R/G}{R/G}"), func(g *game.Game, item *game.StackItem) error {
			if !b09SourceStillOnBattlefield(g, item) {
				return nil
			}
			if err := plusOneCountersOnThis(1)(g, item); err != nil {
				return err
			}
			return b30SourceFightsFirstLegalTarget(g, item)
		})
	ab.Targets = TargetCreature("up to one target creature an opponent controls", OpponentControls()).WithCount(0, 1)
	Register(Spec{
		OracleID:        "72101185-4336-480a-b4d7-f8ceb0007a51",
		Name:            "Abomination, Terrifying Titan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []ActivatedAbility{ab},
	})
}
