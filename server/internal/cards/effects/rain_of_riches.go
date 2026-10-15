package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rain of Riches — Enchantment {3}{R}{R}:
//
//	"When this enchantment enters, create two Treasure tokens.
//	 The first spell you cast each turn that mana from a Treasure was
//	 spent to cast has cascade. (When you cast the spell, exile cards
//	 from the top of your library until you exile a nonland card that
//	 costs less. You may cast it without paying its mana cost. Put the
//	 exiled cards on the bottom in a random order.)"
//
// Cascade granted to a spell is the same trigger Maelstrom Nexus uses
// (GrantsCascade): it watches the cast, belongs to the enchantment,
// and cascades off the spell's mana value. What this card adds is the
// condition — the spell was paid for with mana from a Treasure, read
// off the cast's payment record — and the once-a-turn gate, which is
// this ability's own announced tally: the first qualifying spell puts
// the trigger on the stack and every later one finds it already used.
// A spell paid partly with Treasure mana counts, and a cascaded free
// cast, which spends no mana, never does.
//
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, and
// Treasure mana exists only as mana in the pool, so the record says
// exactly whether (and how much) Treasure mana paid.
func init() {
	Register(Spec{
		OracleID:       "ef1e2d3d-e977-4729-9430-eba6242e5dfe",
		Name:           "Rain of Riches",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Rain of Riches — create two Treasure tokens",
				Do(CreateToken{Template: TreasureToken(), N: 2})),
			GrantsCascade("Rain of Riches", rainOfRichesFirstTreasureSpell),
		},
	})
}

// rainOfRichesFirstTreasureSpell is the cascade grant's condition:
// Treasure mana paid for this spell, and no spell has earned the
// cascade yet this turn.
func rainOfRichesFirstTreasureSpell(spell game.Card, source *game.Card, g *game.Game) bool {
	if !g.StackItemPaidForEffect(spell.InstanceID).Spent().FromTreasure() {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, "Rain of Riches — cascade")
}
