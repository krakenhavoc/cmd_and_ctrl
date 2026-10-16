package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// target_tax_view_test.go — #3860, CR 601.2f. A permanent that charges
// life for a spell that targets it (Terror of the Peaks) is named beside
// that candidate in the spell's legal-target clause, priced by the
// accessor CastSpell charges, so the picker can say "3 life" before the
// click -- and a candidate the viewer cannot pay for is named as such.

func TestLegalTargetsCarryTheLifeTax(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const (
		removal = "test-tax-view-removal"
		dragon  = "test-tax-view-dragon"
	)

	prevSpec, prevTax := game.CatalogTargetSpec, game.CatalogTargetTaxes
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != removal {
			return nil
		}
		return &game.TargetSpec{
			Mode: "creature", Label: "target creature",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1, Max: 1,
		}
	}
	game.CatalogTargetTaxes = func(id string) []game.TargetTax {
		if id != dragon {
			return nil
		}
		return []game.TargetTax{{
			Label: "Spells your opponents cast that target this creature cost an additional 3 life to cast.",
			Life:  3,
			AppliesTo: func(q game.TargetTaxQuery) bool {
				return q.Target.InstanceID == q.Source.InstanceID && q.Caster != q.Source.Controller
			},
		}}
	}
	t.Cleanup(func() { game.CatalogTargetSpec, game.CatalogTargetTaxes = prevSpec, prevTax })

	taxed := game.NewCard("Taxing Dragon", opp.ID)
	taxed.TypeLine = "Creature — Dragon"
	taxed.OracleID = dragon
	g.Battlefield.PushTop(taxed)
	free := game.NewCard("Their Bear", opp.ID)
	free.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(free)
	spell := game.NewCard("Smash", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = removal
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	find := func(t *testing.T) *LegalTargetsView {
		t.Helper()
		view := ViewOfGameFor(g, me.ID.String())
		for i := range view.Seats[0].Hand.Cards {
			c := &view.Seats[0].Hand.Cards[i]
			if c.InstanceID == spell.InstanceID.String() {
				if c.LegalTargets == nil {
					t.Fatal("the spell ships its legal targets")
				}
				return c.LegalTargets
			}
		}
		t.Fatal("the spell is in the viewer's hand")
		return nil
	}

	me.Life = 20
	lt := find(t)
	if got := lt.LifeTaxes[taxed.InstanceID.String()]; got != 3 {
		t.Errorf("the Dragon's tax = %d, want 3 (%+v)", got, lt.LifeTaxes)
	}
	if _, ok := lt.LifeTaxes[free.InstanceID.String()]; ok {
		t.Errorf("an untaxed candidate carries no tax: %+v", lt.LifeTaxes)
	}
	if len(lt.LifeTaxesUnpayable) != 0 {
		t.Errorf("at 20 life the tax is payable: %v", lt.LifeTaxesUnpayable)
	}

	me.Life = 2
	lt = find(t)
	if len(lt.LifeTaxesUnpayable) != 1 || lt.LifeTaxesUnpayable[0] != taxed.InstanceID.String() {
		t.Errorf("at 2 life the Dragon's tax is unpayable: %v", lt.LifeTaxesUnpayable)
	}
	found := false
	for _, id := range lt.Cards {
		if id == taxed.InstanceID.String() {
			found = true
		}
	}
	if !found {
		t.Error("an unpayable candidate stays a legal target on the wire (the picker explains it); the move list is what drops the cast")
	}

	// Absent when nothing is taxed.
	clearTaxed := func() {
		g.Battlefield.Cards = g.Battlefield.Cards[:0]
		g.Battlefield.PushTop(free)
	}
	clearTaxed()
	raw, err := json.Marshal(find(t))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "life_taxes") {
		t.Errorf("`life_taxes` is omitted when no candidate is taxed: %s", raw)
	}
}
