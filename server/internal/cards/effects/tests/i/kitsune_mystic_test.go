package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const pacifismOracleForKitsune = "5f5e0b10-c8cf-450c-bfd3-bcb0528ec330"

func kitsuneMysticRow() cards.Card {
	return effectstest.FlipRow(cardsj.KitsuneMysticOracleID, "{3}{W}", []string{"W"}, nil,
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
	id := effectstest.ApaPush(g, p.ID, p.ID, game.Card{
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
	id := effectstest.FlipEntered(t, g, kitsuneMysticRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	return id
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
			g, me, _ := effectstest.FlipMain(t)
			id := effectstest.FlipEntered(t, g, kitsuneMysticRow(), me)
			for i := 0; i < tc.auras; i++ {
				kitsunePacifism(t, g, me, id)
			}
			effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
			effectstest.PassPriorityAroundTable(t, g)
			if tc.wantFlip {
				effectstest.WantFlipped(t, g, id, "Autumn-Tail, Kitsune Sage", []string{"W"}, 4)
			} else {
				effectstest.WantUnflipped(t, g, id, "Kitsune Mystic")
			}
		})
	}
}

// Autumn-Tail moves an Aura to the other creature the controller picks.
func TestAutumnTailMovesAnAura(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	tail := flippedAutumnTail(t, g, me)
	from := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bear", "oracle-bear", 2, 2))
	to := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Ox", "oracle-ox", 2, 4))
	aura := kitsunePacifism(t, g, me, from)

	effectstest.ApaMana(me, "W")
	effectstest.P7Activate(t, g, me, tail, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}}})
	effectstest.AnswerChooseCards(t, g, me.ID, to)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedToNow(t, g, aura, to) {
		t.Errorf("the Aura did not move to the picked creature")
	}
}

// CR 701.3b: the choice is offered only creatures the Aura can enchant,
// so a creature with protection from white is not among them.
func TestAutumnTailDoesNotOfferACreatureTheAuraCannotEnchant(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	tail := flippedAutumnTail(t, g, me)
	from := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bear", "oracle-bear", 2, 2))
	pro := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Pro-White Knight", OracleID: "oracle-pro-white",
		TypeLine: "Creature — Knight", Power: 2, Toughness: 2, Keywords: []string{"protection from white"}})
	ok := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Ox", "oracle-ox", 2, 4))
	aura := kitsunePacifism(t, g, me, from)

	effectstest.ApaMana(me, "W")
	effectstest.P7Activate(t, g, me, tail, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}}})
	choice := effectstest.ChooseCardsChoiceFor(g, me.ID)
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
	g, me, opp := effectstest.FlipMain(t)
	tail := flippedAutumnTail(t, g, me)
	from := effectstest.ApaPush(g, opp.ID, opp.ID, effectstest.ApaCreature("Bear", "oracle-bear", 2, 2))
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Pro-White Knight", OracleID: "oracle-pro-white",
		TypeLine: "Creature — Knight", Power: 2, Toughness: 2, Keywords: []string{"protection from white"}})
	// Autumn-Tail is a creature too; make it a place the Aura can't go.
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, tail)
		c.Keywords = append(c.Keywords, "protection from white")
		g.RecomputeLayersIfStaleLocked()
	})
	aura := kitsunePacifism(t, g, me, from)

	effectstest.ApaMana(me, "W")
	effectstest.P7Activate(t, g, me, tail, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: aura}}})
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("a choice was offered with no legal creature")
	}
	if !effectstest.AttachedToNow(t, g, aura, from) {
		t.Error("CR 701.3b: the Aura left its creature")
	}
}
