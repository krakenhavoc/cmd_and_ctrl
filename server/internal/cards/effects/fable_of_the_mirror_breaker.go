package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fable of the Mirror-Breaker — an Enchantment — Saga:
//
//	I   — Create a 2/2 red Goblin Shaman creature token with
//	      "Whenever this creature attacks, create a Treasure token."
//	II  — You may discard up to two cards. If you do, draw that many
//	      cards.
//	III — Exile this Saga, then return it to the battlefield
//	      transformed under your control.
//
// The proof of ADR 0079's SECOND verb, and the reason there are two.
// Chapter III is not a transform: it is an exile and a return, so the
// permanent that comes back is a new object (CR 400.7) with a new
// InstanceID, no lore counters, a fresh CR 613.7 timestamp and
// summoning sickness. Writing it as an in-place transform would leave
// a Saga carrying three lore counters that had turned into a creature,
// which is wrong in a way that only shows up two turns later.
//
// The CR 714.4 sacrifice needs no special case, which is worth saying
// because it looks like it should. sagasReadyToSacrificeLocked filters
// on IsSaga (the EFFECTIVE subtypes) and on a final chapter derived
// from the card's declared chapter triggers; by the time the SBA next
// runs, chapter III has already exiled the Saga and what is on the
// battlefield is Reflection of Kiki-Jiki, which is neither a Saga nor
// a card with chapters. sagaHasChapterOnStackLocked is what keeps the
// door open until chapter III has actually resolved.
//
// The chapter I token carries its own attack trigger since ADR 0083
// (#1248): the Goblin Shaman is a catalog template under
// `token:goblin-shaman`, and `TriggersForCard` finds its ability
// through `game.CatalogKey`'s token-key fallback exactly as it finds
// a printed card's. Until then this card shipped as `caveats` with a
// plain 2/2, on the "Triggered and static abilities on non-copy
// tokens" seam.
func init() {
	Register(Spec{
		OracleID:     fableOfTheMirrorBreakerOracleID,
		Name:         "Fable of the Mirror-Breaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "Fable of the Mirror-Breaker — create a 2/2 red Goblin Shaman",
				Do(CreateToken{Template: fableGoblinShamanToken(), N: 1})),
			ChapterTrigger(2, "Fable of the Mirror-Breaker — discard up to two cards, then draw that many",
				func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(2, false,
						"Fable of the Mirror-Breaker — discard up to two cards, then draw that many",
						func(discarded int) int { return discarded })(NewContext(g, item))
				}),
			ChapterTrigger(3, "Fable of the Mirror-Breaker — exile it, then return it transformed",
				ChapterExileAndReturnTransformed),
		},
	})
}

// fableOfTheMirrorBreakerOracleID is shared with the back face, which
// registers under it plus "#1" (game.CatalogKey).
const fableOfTheMirrorBreakerOracleID = "c0957e5e-c71b-439c-931c-9f55d2f76ace"

// fableGoblinShamanToken is chapter I's 2/2 red Goblin Shaman with
// "Whenever this creature attacks, create a Treasure token."
func fableGoblinShamanToken() game.Card { return tokenFromCatalog(printedFableGoblinShamanToken) }

// printedFableGoblinShamanToken is that Goblin Shaman as PRINTED.
//
// The Treasure it makes is the ordinary catalog Treasure, so a token
// whose ability creates another token is two catalog entries and no
// new machinery: the attack trigger resolves and calls the same
// CreateToken every printed card calls.
func printedFableGoblinShamanToken() tokenTemplate {
	return tokenTemplate{
		Slug: "goblin-shaman",
		Card: game.Card{
			Name:      "Goblin Shaman",
			TypeLine:  "Token Creature — Goblin Shaman",
			Power:     2,
			Toughness: 2,
			Colors:    []string{"R"},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Goblin Shaman — create a Treasure",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
		Text: "Whenever this creature attacks, create a Treasure token.",
	}
}
