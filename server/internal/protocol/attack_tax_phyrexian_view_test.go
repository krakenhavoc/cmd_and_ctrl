package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_tax_phyrexian_view_test.go — #4198: the wire half of paying a
// Phyrexian attack tax with life. `turn.attack_targets[].tax_phyrexian`
// is how many symbols one creature's tax prints, so the attack dock
// offers its stepper without parsing the `tax` string.

const phyrexianTaxViewOracle = "phyrexian-tax-view-test"

func stubPhyrexianAttackTax(t *testing.T, cost string) {
	t.Helper()
	prev := game.CatalogAttackTaxes
	t.Cleanup(func() { game.CatalogAttackTaxes = prev })
	game.CatalogAttackTaxes = func(key string) []game.AttackTax {
		if key != phyrexianTaxViewOracle {
			return nil
		}
		return []game.AttackTax{{
			Label:    "test tax",
			Scope:    game.AttackTaxOnPlayerOrPlaneswalkers,
			ManaCost: func(game.AttackTaxQuery) string { return cost },
		}}
	}
}

func taxRows(v GameView) map[string]AttackTargetView {
	out := map[string]AttackTargetView{}
	for _, row := range v.Turn.AttackTargets {
		out[row.ID] = row
	}
	return out
}

func TestAttackTargetsCarryThePhyrexianSymbolsOfTheTax(t *testing.T) {
	stubPhyrexianAttackTax(t, "{W/P}")
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	taxed := g.Seats[(g.Turn.ActiveSeat+1)%3]
	free := g.Seats[(g.Turn.ActiveSeat+2)%3]
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Taxer", TypeLine: "Artifact",
		OracleID: phyrexianTaxViewOracle, Owner: taxed.ID, Controller: taxed.ID,
	})
	pushAttacker(g, active.ID, uuid.Nil)

	rows := taxRows(ViewOfGame(g))
	if got := rows[taxed.ID.String()]; got.Tax != "{W/P}" || got.TaxPhyrexian != 1 || got.TaxPhyrexianGranted != 0 {
		t.Errorf("taxed seat row = %+v, want tax {W/P} with 1 Phyrexian symbol, none granted", got)
	}
	if got := rows[free.ID.String()]; got.Tax != "" || got.TaxPhyrexian != 0 {
		t.Errorf("untaxed seat row = %+v, want no tax and no symbols", got)
	}
}

func TestAMixedAttackTaxCountsOnlyItsPhyrexianSymbols(t *testing.T) {
	stubPhyrexianAttackTax(t, "{2}{W/P}")
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	taxed := g.Seats[(g.Turn.ActiveSeat+1)%3]
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Taxer", TypeLine: "Artifact",
		OracleID: phyrexianTaxViewOracle, Owner: taxed.ID, Controller: taxed.ID,
	})
	pushAttacker(g, active.ID, uuid.Nil)

	if got := taxRows(ViewOfGame(g))[taxed.ID.String()]; got.TaxPhyrexian != 1 {
		t.Errorf("{2}{W/P} reports %d Phyrexian symbols, want 1", got.TaxPhyrexian)
	}
}
