package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bestow_view_test.go — ADR 0141 (#2862). The client's cast picker is
// read off this projection: a bestow card in hand arrives with the
// printed cost and a "Bestow {cost}" offer whose target clause is
// enchant creature (castTargetOverride takes it as the spell's clause),
// and a bestowed spell on the stack reads as the Aura it is.

const nighthowlerOracle = "57b3f7fc-1812-4134-a645-6cef48a8aa71"

func TestBestowOfferCarriesTheEnchantCreatureClause(t *testing.T) {
	g := buildActiveGame(t)
	castWindowOpen(t, g)
	me := g.Seats[0]
	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}

	bear := game.NewCard("Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Power, bear.Toughness = 2, 2
	bear.KnownBy = seen
	g.Battlefield.PushTop(bear)

	howler := game.Card{
		InstanceID: uuid.New(), OracleID: nighthowlerOracle, Name: "Nighthowler",
		TypeLine: "Enchantment Creature — Horror", ManaCost: "{1}{B}{B}",
		Owner: me.ID, Controller: me.ID, KnownBy: seen,
	}
	me.Hand.PushTop(howler)

	v := ViewOfGameFor(g, me.ID.String())
	var cv *CardView
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == howler.InstanceID.String() {
			cv = &v.Seats[0].Hand.Cards[i]
		}
	}
	if cv == nil {
		t.Fatal("hand view is missing Nighthowler")
	}
	if cv.AlternativeCostRequired {
		t.Error("the printed cost is claimable from hand too")
	}
	if len(cv.AlternativeCosts) != 1 {
		t.Fatalf("offers = %+v, want bestow alone", cv.AlternativeCosts)
	}
	offer := cv.AlternativeCosts[0]
	if offer.Key != game.BestowKey || offer.Label != "Bestow {2}{B}{B}" || offer.ManaCost != "{2}{B}{B}" {
		t.Errorf("offer = %+v, want Bestow {2}{B}{B}", offer)
	}
	if offer.TargetMode != "creature" || offer.LegalTargets == nil {
		t.Fatalf("the bestow offer carries no enchant creature clause: %+v", offer)
	}
	found := false
	for _, c := range offer.LegalTargets.Cards {
		if c == bear.InstanceID.String() {
			found = true
		}
	}
	if !found {
		t.Errorf("the bear is not among the bestowed Aura's legal targets: %+v", offer.LegalTargets)
	}
}

// CR 702.103b: on the stack, a bestowed spell's type line is the Aura's.
func TestBestowedSpellOnTheStackReadsAsAnAura(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	seen := map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
	spell := game.Card{
		InstanceID: uuid.New(), OracleID: nighthowlerOracle, Name: "Nighthowler",
		TypeLine: "Enchantment Creature — Horror", ManaCost: "{1}{B}{B}",
		Owner: me.ID, Controller: me.ID, KnownBy: seen, Bestowed: true,
	}
	g.Stack.PushTop(spell)

	v := ViewOfGameFor(g, me.ID.String())
	for _, c := range v.Stack.Cards {
		if c.InstanceID != spell.InstanceID.String() {
			continue
		}
		if c.TypeLine != "Enchantment — Aura" {
			t.Errorf("stack type line %q, want Enchantment — Aura", c.TypeLine)
		}
		return
	}
	t.Fatal("the stack view is missing the spell")
}
