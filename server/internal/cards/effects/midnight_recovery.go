package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Midnight Recovery — Sorcery {3}{B}:
//
//	"Return target creature card from your graveyard to your hand.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Regrowth's return narrowed to a creature card, and cipher, the
// engine's keyword (game/cipher.go, CR 702.99, #3867). A copy cast
// through the encoded creature targets a creature card in ITS caster's
// graveyard — "your" is whoever casts the copy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7904025c-7120-4362-abe9-a9f5fb64447f",
		Name:            "Midnight Recovery",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve:       returnTargetFromGraveyardToHand,
	})
}
