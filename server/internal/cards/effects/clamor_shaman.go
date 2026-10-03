package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// clamorShamanLabel is the attack trigger's stack label.
const clamorShamanLabel = "Clamor Shaman — target creature an opponent controls can't block this turn"

// Clamor Shaman — Creature — Goblin Shaman {2}{R}, 1/1:
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 Whenever this creature attacks, target creature an opponent
//	 controls can't block this turn."
//
// Riot is the engine's keyword (ADR 0109 §10). The attack trigger
// targets as it goes on the stack (CR 603.3d) and gives the target the
// CantBlock restriction until the turn ends, if it is still a legal
// target as the trigger resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7fd2dcb2-760d-461d-8e5b-693aacf548cf",
		Name:            "Clamor Shaman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks(clamorShamanLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return RestrictUntilEOT{Target: id, Restrictions: game.CantBlock, Label: clamorShamanLabel}.Apply(ctx)
			}), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
