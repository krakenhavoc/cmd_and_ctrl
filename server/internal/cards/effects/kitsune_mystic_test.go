package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const pacifismOracleForKitsune = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"

func kitsuneMysticRow() cards.Card {
	return flipRow(kitsuneMysticOracleID, "{3}{W}", []string{"W"}, nil,
		cards.CardFace{Name: "Kitsune Mystic", TypeLine: "Creature — Fox Wizard",
			OracleText: "At the beginning of the end step, if this creature is enchanted by two or more Auras, flip it.",
			Power:      "2", Toughness: "3"},
		cards.CardFace{Name: "Autumn-Tail, Kitsune Sage", TypeLine: "Legendary Creature — Fox Wizard",
			OracleText: "{1}: Attach target Aura attached to a creature to another creature.",
			Power:      "4", Toughness: "5"})
}

// kitsunePacifism puts a white Pacifism on the battlefield attached to host.
func kitsunePacifism(t *testing.T, g *game.Game, p *game.Player, host uuid.UUID) uuid.UUID {
	t.Helper()
	id := apaPush(g, p.ID, p.ID, game.Card{
		Name: "Pacifism", OracleID: pacifismOracleForKitsune,
		TypeLine: "Enchantment — Aura", Colors: []string{"W"},
	})
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatal(err)
		}
	})
	return id
}

func flippedAutumnTail(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := flipEntered(t, g, kitsuneMysticRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	return id
}

func attachedToNow(t *testing.T, g *game.Game, aura, host uuid.UUID) bool {
	t.Helper()
	c := findBattlefieldCardForTest(g, aura)
	return c != nil && c.IsAttachedTo(host)
}

// Kitsune Mystic flips at the beginning of the end step if it is
// enchanted by two or more Auras, and not with one.
func TestKitsuneMysticFlipsWithTwoAuras(t *testing.T) {
	for _, tc := range []struct {
		name     string
		auras    int
		wantFlip bool
	}{{"one aura", 1, false}, {"two auras", 2, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := flipMain(t)
			id := flipEntered(t, g, kitsuneMysticRow(), me)
			for i := 0; i < tc.auras; i++ {
				kitsunePacifism(t, g, me, id)
			}
			advanceToEndStepOf(t, g, g.Turn.ActiveSeat)
			passPriorityAroundTable(t, g)
			if tc.wantFlip {
				wantFlipped(t, g, id, "Autumn-Tail, Kitsune Sage", []string{"W"}, 4)
			} else {
				wantUnflipped(t, g, id, "Kitsune Mystic")
			}
		})
	}
}

// Autumn-Tail moves an Aura to the other creature the controller picks.
func TestAutumnTailMovesAnAura(t *testing.T) {
	g, me, opp := flipMain(t)
	tail := flippedAutumnTail(t, g, me)
	from := apaPush(g, opp.ID, opp.ID, apaCreature("Bear", "oracle-bear", 2, 2))
	to := apaPush(g, opp.ID, opp.ID, apaCreature("Ox", "oracle-ox", 2, 4))
	aura := kitsunePacifism(t, g, me, from)

	apaMana(me, "W")
	p7Activate(t, g, me, tail, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}}})
	answerChooseCards(t, g, me.ID, to)
	passPriorityAroundTable(t, g)
	if !attachedToNow(t, g, aura, to) {
		t.Errorf("the Aura did not move to the picked creature")
	}
}

// CR 701.3b: the choice is offered only creatures the Aura can enchant,
// so a creature with protection from white is not among them.
func TestAutumnTailDoesNotOfferACreatureTheAuraCannotEnchant(t *testing.T) {
	g, me, opp := flipMain(t)
	tail := flippedAutumnTail(t, g, me)
	from := apaPush(g, opp.ID, opp.ID, apaCreature("Bear", "oracle-bear", 2, 2))
	pro := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Pro-White Knight", OracleID: "oracle-pro-white",
		TypeLine: "Creature — Knight", Power: 2, Toughness: 2, Keywords: []string{"protection from white"}})
	ok := apaPush(g, opp.ID, opp.ID, apaCreature("Ox", "oracle-ox", 2, 4))
	aura := kitsunePacifism(t, g, me, from)

	apaMana(me, "W")
	p7Activate(t, g, me, tail, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}}})
	choice := chooseCardsChoiceFor(g, me.ID)
	if choice == nil {
		t.Fatal("no choice was offered")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range choice.ChooseCards {
		offered[id] = true
	}
	if offered[pro] || offered[from] || !offered[ok] {
		t.Errorf("offered pro-white %v, current host %v, legal Ox %v; want false, false, true", offered[pro], offered[from], offered[ok])
	}
}

// With no creature the Aura can enchant, the Aura stays where it is
// instead of being put into the graveyard.
func TestAutumnTailLeavesAnAuraWithNowhereLegalToGo(t *testing.T) {
	g, me, opp := flipMain(t)
	tail := flippedAutumnTail(t, g, me)
	from := apaPush(g, opp.ID, opp.ID, apaCreature("Bear", "oracle-bear", 2, 2))
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Pro-White Knight", OracleID: "oracle-pro-white",
		TypeLine: "Creature — Knight", Power: 2, Toughness: 2, Keywords: []string{"protection from white"}})
	// Autumn-Tail is a creature too; make it a place the Aura can't go.
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, tail)
		c.Keywords = append(c.Keywords, "protection from white")
		g.RecomputeLayersIfStaleLocked()
	})
	aura := kitsunePacifism(t, g, me, from)

	apaMana(me, "W")
	p7Activate(t, g, me, tail, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}}})
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a choice was offered with no legal creature")
	}
	if !attachedToNow(t, g, aura, from) {
		t.Error("CR 701.3b: the Aura left its creature")
	}
}
