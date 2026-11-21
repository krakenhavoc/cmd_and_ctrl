package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "bd9d761f-79cc-4082-8375-51af582f1472",
		Name:            "Kozilek's Unsealing",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Tokens: 2, CreatureTokens: 2, Draws: 3},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(And(Creature(), ManaValueGE(4), ManaValueLE(6)),
				"Kozilek's Unsealing — create two 0/1 colorless Eldrazi Spawn",
				Do(CreateToken{Template: EldraziSpawnToken(), N: 2})),
			WheneverYouCast(And(Creature(), ManaValueGE(7)),
				"Kozilek's Unsealing — draw three cards", Do(DrawCards{N: 3})),
		},
	})
}
