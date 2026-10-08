package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// It of the Horrid Swarm — Creature — Eldrazi Insect {8}, 4/4:
//
//	"Emerge {6}{G} (You may cast this spell by sacrificing a creature and
//	 paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, create two 1/1 green Insect creature
//	 tokens."
//
// Emerge is the shared alternative cost (ADR 0135 §4). The tokens come
// from a cast trigger, so they are made above the spell, even if it is
// countered, however it was cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f4ad190a-6c82-4923-af54-b4cce8d6465f",
		Name:             "It of the Horrid Swarm",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Emerge("{6}{G}")},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell("It of the Horrid Swarm — create two 1/1 green Insect creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 green Insect"), N: 2})),
		},
	})
}
