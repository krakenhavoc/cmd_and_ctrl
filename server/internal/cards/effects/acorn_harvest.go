package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acorn Harvest — Sorcery {3}{G} (Torment, #5612):
//
//	"Create two 1/1 green Squirrel creature tokens.
//	 Flashback—{1}{G}, Pay 3 life. (You may cast this card from your
//	 graveyard for its flashback cost. Then exile it.)"
//
// Chatter of the Squirrel at two tokens: the existing 1/1 green Squirrel
// row, created together. The flashback price is mana and life
// (torB1FlashbackPayingLife): a caster with less than 3 life can't claim
// it (CR 118.3), and the card is exiled as it leaves the stack (CR
// 702.34a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "6aeba615-2d37-4eb8-a377-41a517e95aa3",
		Name:             "Acorn Harvest",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Tokens: 2, CreatureTokens: 2},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{torB1FlashbackPayingLife("{1}{G}", 3)},
		OnResolve:        ResolveSteps(CreateToken{Template: TokenCard("1/1 green Squirrel"), N: 2}),
	})
}
