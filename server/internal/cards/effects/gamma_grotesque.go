package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gamma Grotesque — Creature — Gamma Horror Villain {2}{G}, 3/3 (Marvel
// Super Heroes Commander, #4455):
//
//	"Vigilance
//	 Power-up — {4}{G}{G}: Put three +1/+1 counters on this creature. Then
//	 draw a card for each creature you control with a counter on it.
//	 (Activate each power-up ability only once. Reduce the cost by its mana
//	 cost if it entered this turn.)"
//
// Power-up is the engine's (CR 702.193). The count is read after the counters
// land, so the Grotesque itself is always one of the creatures counted while
// it is on the battlefield; a counter of any kind counts (stun, shield, ...).
//
// No simplification.
func init() {
	ab := PowerUp("{4}{G}{G}: Put three +1/+1 counters on this creature. Then draw a card for each creature you control with a counter on it.",
		ManaCost("{4}{G}{G}"), func(g *game.Game, item *game.StackItem) error {
			if err := plusOneCountersOnThis(3)(g, item); err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: len(creaturesWithACounter(g, item.Controller))}.Apply(NewContext(g, item))
		})
	ab.Purpose.Answers = game.AnswerPump
	Register(Spec{
		OracleID:        "b92b4894-0786-49a7-8003-d99393f080d4",
		Name:            "Gamma Grotesque",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []ActivatedAbility{ab},
	})
}

// creaturesWithACounter lists the creatures `controller` controls that have
// at least one counter of any kind.
func creaturesWithACounter(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsCreature() {
			continue
		}
		for _, n := range c.Counters {
			if n > 0 {
				out = append(out, c.InstanceID)
				break
			}
		}
	}
	return out
}
