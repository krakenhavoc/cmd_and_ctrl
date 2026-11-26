package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Novice Knight — Creature — Human Knight {W}, 2/3 (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"Defender
//	 As long as this creature is enchanted or equipped, it can attack as
//	 though it didn't have defender."
//
// Backstreet Bruiser's conditional defender waiver (CR 702.3b): the keyword
// stays and the waiver is read live at declaration, true while some Aura or
// Equipment, whoever controls it, is attached to the Knight (Halvar's
// "enchanted or equipped").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "916d2a0d-1c3e-4102-958d-524b6b4df509",
		Name:            "Novice Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{effects.AvB18CanAttackAsThoughNoDefenderWhile(
			"As long as this creature is enchanted or equipped, it can attack as though it didn't have defender.",
			func(g *game.Game, source *game.Card) bool {
				return effects.CreatureIsEnchantedOrEquipped(g, source)
			})},
	})
}
