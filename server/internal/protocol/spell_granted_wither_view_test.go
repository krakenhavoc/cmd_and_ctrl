package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spell_granted_wither_view_test.go — #4693: a keyword an effect gives a
// spell on the stack (Spinerock Tyrant's "those spells gain wither", ADR
// 0107 §3) is on that spell's card in the stack view, and not on another
// spell beside it.
func TestStackViewShowsWitherGivenToASpell(t *testing.T) {
	g := buildActiveGame(t)
	given := pushStackSpell(g, "Bolt", "plain", game.PaidCost{}, game.StackItemSpell)
	other := pushStackSpell(g, "Bolt 2", "plain", game.PaidCost{}, game.StackItemSpell)
	g.WithWriteLock(func() {
		if !g.GrantKeywordsToSpellForEffect(uuid.New(), given, []string{game.KeywordWither},
			game.IndefiniteDuration(), "gains wither") {
			t.Fatal("the grant was refused")
		}
	})
	abilities := map[string][]string{}
	for _, c := range ViewOfGame(g).Stack.Cards {
		abilities[c.InstanceID] = c.Abilities
	}
	if !slices.Contains(abilities[given.String()], game.KeywordWither) {
		t.Errorf("the spell given wither shows %v", abilities[given.String()])
	}
	if slices.Contains(abilities[other.String()], game.KeywordWither) {
		t.Errorf("the other spell shows %v", abilities[other.String()])
	}
}
