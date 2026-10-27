package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trait Doctoring — Sorcery {U}:
//
//	"Change the text of target permanent by replacing all instances of
//	 one color word with another or one basic land type with another
//	 until end of turn.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Whim of Volrath's change (CR 612.2, ADR 0160's 2026-10-15 amendment,
// one word of either class until end of turn) without the buyback, and
// cipher (CR 702.99, #3867), the engine's keyword: once both words are
// chosen the caster may encode it on a creature they control, and each
// time that creature deals combat damage to a player its controller may
// cast a copy, which makes a fresh change of its own.
//
// Graded Full on the same footing as the cards it joins.
func init() {
	Register(Spec{
		OracleID:        "3b5b926b-ee8f-4c7a-8737-1de7d18ca60d",
		Name:            "Trait Doctoring",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetPermanent("target permanent"),
		OnResolve:       changeTextOfTarget(textEitherWords, true, nothingAfterTheChange),
	})
}
