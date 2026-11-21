package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Token rows for GAP batch 73, added from this file's package-level
// initializer so the shared table's column alignment is not rewritten and the
// rows exist before any card file's init() asks for one (init functions run in
// file-name order; package variables are initialised first). The keys read like
// the printed text, as the table's do. Conclave Cavalier's Elf Knight is
// batch 79's row in tokens_table.go.
var _ = registerGapB73Tokens()

func registerGapB73Tokens() bool {
	for key, card := range map[string]game.Card{
		// Tolsimir Wolfblood.
		"2/2 green and white Legendary Wolf named Voja": {Name: "Voja", TypeLine: "Token Legendary Creature — Wolf", Power: 2, Toughness: 2, Colors: []string{"G", "W"}},
	} {
		if _, dup := TokenTable[key]; dup {
			panic("s82_gap_b73_helpers: token key already in the table: " + key)
		}
		TokenTable[key] = card
	}
	return true
}
