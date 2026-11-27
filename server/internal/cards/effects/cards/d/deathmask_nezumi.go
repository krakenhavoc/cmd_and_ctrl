package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathmask Nezumi — Creature — Rat Shaman {2}{B}, 2/2 (Saviors of Kamigawa,
// #5601):
//
//	"As long as you have seven or more cards in hand, this creature gets +2/+1
//	 and has fear. (It can't be blocked except by artifact creatures and/or
//	 black creatures.)"
//
// Akki Underling's seven-card gate with fear for its keyword: a layer 7c
// +2/+1 and a layer 6 fear, both read live against the size of your hand
// (DependsOnHandSize).
//
// No simplification.
func init() {
	const when = "As long as you have seven or more cards in hand, this creature "
	sevenCards := effects.OldB588YouHaveAtLeastCardsInHand(7)
	effects.Register(effects.Spec{
		OracleID:     "fede7174-261c-4724-b6b0-f2a71f8ff53f",
		Name:         "Deathmask Nezumi",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.OldB588WhileYouHavePump(sevenCards, 2, 1, when+"gets +2/+1."),
			effects.OldB588WhileYouHaveKeyword(sevenCards, "fear", when+"has fear."),
		},
	})
}
