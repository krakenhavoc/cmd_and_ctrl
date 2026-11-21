package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kozilek's Unsealing — Enchantment {2}{U} (Modern Horizons 3):
//
//	"Devoid (This card has no color.)
//	 Whenever you cast a creature spell with mana value 4, 5, or 6, create
//	 two 0/1 colorless Eldrazi Spawn creature tokens with "Sacrifice this
//	 token: Add {C}."
//	 Whenever you cast a creature spell with mana value 7 or greater, draw
//	 three cards."
//
// Two cast triggers (Helga's shape: the spell's mana value read on the stack,
// {X} counted, CR 202.3e), each resolving above the spell, so they happen
// even if the creature spell is countered. A spell is in one band or the
// other, never both. The Spawn is the shared Eldrazi Spawn token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bd9d761f-79cc-4082-8375-51af582f1472",
		Name:            "Kozilek's Unsealing",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2, Draws: 3},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.Creature(), effects.ManaValueGE(4), effects.ManaValueLE(6)),
				"Kozilek's Unsealing — create two 0/1 colorless Eldrazi Spawn",
				effects.Do(effects.CreateToken{Template: effects.EldraziSpawnToken(), N: 2})),
			effects.WheneverYouCast(effects.And(effects.Creature(), effects.ManaValueGE(7)),
				"Kozilek's Unsealing — draw three cards", effects.Do(effects.DrawCards{N: 3})),
		},
	})
}
