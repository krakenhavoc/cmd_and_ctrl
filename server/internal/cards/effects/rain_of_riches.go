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
// Declared weaker than printed: the payment record exists only when
// mana is paid through the strict-mana path. A cast whose payment the
// engine waived (the permissive human default) is "unknown", and
// unknown is the weaker answer.
func init() {
	Register(Spec{
		OracleID:     "ef1e2d3d-e977-4729-9430-eba6242e5dfe",
		Name:         "Rain of Riches",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so Rain of Riches never gives a spell its bonus."},
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
