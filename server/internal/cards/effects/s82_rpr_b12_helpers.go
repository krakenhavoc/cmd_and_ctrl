package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// NonartifactSpell passes on a spell that is not an artifact spell —
// Lodestone Golem's "Nonartifact spells cost {1} more to cast".
func NonartifactSpell() CostPredicate {
	return func(q game.CostQuery) bool { return !q.Card.IsArtifact() }
}
