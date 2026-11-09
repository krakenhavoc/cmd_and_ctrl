package game

import "github.com/google/uuid"

// planes_b6_attackers.go — the declared attackers of one player (Valor's
// Reach, ADR 0216 slice 4, #4870).

// DeclaredAttackersOfPlayerForEffect are the creatures `player` controls that
// were declared as attackers this combat, in battlefield order (CR 508.1: a
// creature put onto the battlefield attacking was never declared, CR 506.3c,
// and is not among them). "Whenever your team attacks with exactly two
// creatures" counts this list.
//
// Caller must hold g.mu.
func (g *Game) DeclaredAttackersOfPlayerForEffect(player uuid.UUID) []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.IsCreature() && c.AttackingTarget != uuid.Nil && g.announcedAttacks[c.InstanceID] {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
