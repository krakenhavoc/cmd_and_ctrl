package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// double_major_test.go — CR 608.3f / CR 111.13, end to end. The copy
// of a permanent spell must become a TOKEN, not a second card, and
// it must be a token in every sense the rules use the word: a
// doubler doubles it, its enters-the-battlefield abilities fire, and
// it cannot come back as a card.

const (
	OracleDoubleMajor    = "ece44a82-dcf0-4439-bdd9-a09c99a6f159"
	OracleDoublingSeason = "01546b7d-a233-4176-8843-d732074dc5b6"
	OracleThrabenInsp    = "caa02547-66e3-4e27-a2d3-5e94f3e7a069"
)

// castCreatureSpellForCopy seeds a creature card in the active
// seat's hand with real printed stats and casts it, leaving it on
// the stack for something to copy.
func CastCreatureSpellForCopy(t *testing.T, g *game.Game, name, typeLine, oracleID string, p, tough int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		Power:      p,
		Toughness:  tough,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// copiesOnBattlefield splits the battlefield cards named `name` into
// the tokens and the non-tokens.
func CopiesOnBattlefield(g *game.Game, name string) (tokens, cards []game.Card) {
	for _, c := range g.Battlefield.Cards {
		if c.Name != name {
			continue
		}
		if c.IsToken() {
			tokens = append(tokens, c)
		} else {
			cards = append(cards, c)
		}
	}
	return tokens, cards
}
