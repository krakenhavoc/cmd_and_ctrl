package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// DuringCombatAfterBlockers is the CastCondition of "Cast this spell only
// during combat after blockers are declared" (Chaotic Strike, Curtain of
// Light): nobody has priority in the declare blockers step until every
// blocker is declared (#1501), so each step from there to the end of combat
// is "after blockers are declared".
func DuringCombatAfterBlockers(g *game.Game, _ uuid.UUID, _ game.Card) bool {
	switch g.Turn.Step {
	case game.StepDeclareBlockers, game.StepCombatDamage, game.StepFirstStrikeDamage, game.StepEndCombat:
		return true
	}
	return false
}

// AnotherPermanentSharesName reports whether a permanent other than `target`
// on the battlefield has the same name (Winnow).
func AnotherPermanentSharesName(g *game.Game, target game.Card) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID != target.InstanceID && c.Name == target.Name {
			return true
		}
	}
	return false
}
