package testscr

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// discard_by_opponent_wire_test.go — ADR 0153 §5 (#3168): what the bot is
// told about a card an opponent's discard does something with.

func handView(v protocol.GameView, seat string, name string) (protocol.CardView, bool) {
	for _, s := range v.Seats {
		if s.ID != seat {
			continue
		}
		for _, c := range s.Hand.Cards {
			if c.Name == name {
				return c, true
			}
		}
	}
	return protocol.CardView{}, false
}

// The catalog's declaration is on the hand card, and only there: the
// Baloth says battlefield, Metrognome says trigger, and with a Nephalia
// Academy out every other card in the controller's hand says library.
func TestHandCardsCarryTheirDiscardByOpponentDeclaration(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.HandCard(d.Victim, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", effectstest.ObstinateBalothOracle, 4, 4)
	d.HandCard(d.Victim, "Metrognome", "Artifact", "{4}", effectstest.MetrognomeOracle, 0, 0)
	d.HandCard(d.Victim, "Plain Card", "Sorcery", "{1}", "", 0, 0)
	d.HandCard(d.Caster, "Plain Card", "Sorcery", "{1}", "", 0, 0)

	v := protocol.ViewOfGameFor(d.G, d.Victim.ID.String())
	for name, want := range map[string]string{"Obstinate Baloth": "battlefield", "Metrognome": "trigger", "Plain Card": ""} {
		c, ok := handView(v, d.Victim.ID.String(), name)
		if !ok {
			t.Fatalf("%s is not in the victim's hand view", name)
		}
		if c.DiscardByOpponent != want {
			t.Errorf("%s: discard_by_opponent = %q, want %q", name, c.DiscardByOpponent, want)
		}
	}

	d.Permanent(d.Victim, "Nephalia Academy", "Land", effectstest.NephaliaAcademyOracle)
	v = protocol.ViewOfGameFor(d.G, d.Victim.ID.String())
	for name, want := range map[string]string{"Obstinate Baloth": "battlefield", "Plain Card": "library"} {
		c, _ := handView(v, d.Victim.ID.String(), name)
		if c.DiscardByOpponent != want {
			t.Errorf("with the Academy, %s: discard_by_opponent = %q, want %q", name, c.DiscardByOpponent, want)
		}
	}

	// It is the Academy's controller's hand only.
	mine := protocol.ViewOfGameFor(d.G, d.Caster.ID.String())
	if c, _ := handView(mine, d.Caster.ID.String(), "Plain Card"); c.DiscardByOpponent != "" {
		t.Errorf("an opponent's Academy marked my card %q", c.DiscardByOpponent)
	}
}

// Nobody else sees it: another seat's view of the hand carries no
// declaration, and the field is omitted from the JSON when empty.
func TestDiscardByOpponentIsNotSentForAnotherSeatsHand(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.HandCard(d.Victim, "Obstinate Baloth", "Creature — Beast", "{2}{G}{G}", effectstest.ObstinateBalothOracle, 4, 4)

	v := protocol.ViewOfGameFor(d.G, d.Caster.ID.String())
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "discard_by_opponent") {
		t.Error("another seat's view carries discard_by_opponent")
	}
}

// While the opponent's Mind Rot resolves, the prompt names its source and
// the view shows who controls that spell, which is what the bot reads. A
// resolving spell's stack item is gone by the time its prompt is open, so
// the spell is placed by its card.
func TestAnOpponentsDiscardPromptNamesASourceTheViewCanPlace(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	d.HandCard(d.Victim, "Card A", "Sorcery", "{1}", "", 0, 0)
	d.HandCard(d.Victim, "Card B", "Sorcery", "{1}", "", 0, 0)
	effectstest.CastCatalogSpell(t, d.G, "Mind Rot", "Sorcery", effectstest.MindRotOracleID,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: d.Victim.ID}})
	effectstest.PassPriorityAroundTable(t, d.G)

	v := protocol.ViewOfGameFor(d.G, d.Victim.ID.String())
	var source string
	for _, c := range v.PendingChoices {
		if c.Chooser == d.Victim.ID.String() {
			source = c.Source
		}
	}
	if source == "" {
		t.Fatalf("the discard prompt names no source: %+v", v.PendingChoices)
	}
	for _, it := range v.StackItems {
		if it.SourceCardID == source {
			if it.Controller != d.Caster.ID.String() {
				t.Errorf("the source's stack item is controlled by %s, want the caster", it.Controller)
			}
			return
		}
	}
	// A spell that is resolving has left the stack but is not yet in the
	// graveyard's view as anything but its card: the bot places it there.
	for _, s := range v.Seats {
		for _, c := range s.Graveyard.Cards {
			if c.InstanceID == source {
				if c.Controller != d.Caster.ID.String() {
					t.Errorf("the source card is controlled by %s, want the caster", c.Controller)
				}
				return
			}
		}
	}
	t.Errorf("the view places the prompt's source nowhere: stack items %+v", v.StackItems)
}
