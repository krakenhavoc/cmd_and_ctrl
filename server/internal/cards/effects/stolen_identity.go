package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stolen Identity — Sorcery {4}{U}{U}:
//
//	"Create a token that's a copy of target artifact or creature.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// Cackling Counterpart's token copy (TokenCopyOfSingleTarget, CR
// 707.2: the copiable values) over any artifact or creature, and cipher
// (game/cipher.go, CR 702.99, #3867).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "30b533a3-afc3-4500-a8f0-54e84301988f",
		Name:            "Stolen Identity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve:       TokenCopyOfSingleTarget,
	})
}
