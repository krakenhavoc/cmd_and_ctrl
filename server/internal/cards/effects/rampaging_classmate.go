package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rampaging Classmate — Creature — Lizard Berserker {2}{R}, 3/2
// (Marvel's Spider-Man, #4460):
//
//	"Whenever this creature attacks, it gets +1/+0 until end of turn for
//	 each other attacking creature."
//
// The count is taken as the trigger resolves (CR 608.2h), over every
// attacking creature on the battlefield other than the Classmate, whoever
// controls it: the printed text names no controller. The boost is
// pinned to the Classmate itself, so a Classmate that has left the
// battlefield by then gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bf3ff6ff-eb0e-4d0a-99cb-c377622ae640",
		Name:         "Rampaging Classmate",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Rampaging Classmate — +1/+0 until end of turn for each other attacking creature", rampagingClassmateBoost),
		},
	})
}

func rampagingClassmateBoost(g *game.Game, item *game.StackItem) error {
	self := item.SourceCardID
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID != self && c.AttackingTarget != uuid.Nil && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return BoostUntilEOT{
		Target: self,
		Power:  n,
		Label:  "Rampaging Classmate — +1/+0 for each other attacking creature",
	}.Apply(NewContext(g, item))
}
