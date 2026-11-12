package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_tax_log_test.go — #5271, the log half of the CR 508.1h attack
// tax. A player reported that an opponent attacked through their
// Ghostly Prison without paying: the engine had charged the {2} by
// tapping the attacker's lands, and the log said only "<creature>
// attacks <player>". The payment is now its own line.

const attackTaxLogOracle = "attack-tax-log-test"

func TestAPaidAttackTaxIsALogLine(t *testing.T) {
	prev := game.CatalogAttackTaxes
	t.Cleanup(func() { game.CatalogAttackTaxes = prev })
	game.CatalogAttackTaxes = func(key string) []game.AttackTax {
		if key != attackTaxLogOracle {
			return nil
		}
		return []game.AttackTax{{
			Label:    "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you.",
			ManaCost: func(game.AttackTaxQuery) string { return "{2}" },
		}}
	}
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	// The Prison is two seats from the attacker, not the next one, as
	// at the reporter's table.
	defender := g.Seats[(g.Turn.ActiveSeat+2)%3]
	prison := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: prison, Name: "Ghostly Prison", TypeLine: "Enchantment",
		OracleID: attackTaxLogOracle, Owner: defender.ID, Controller: defender.ID,
	})
	bear := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: bear, Name: "Orcish Bowmasters", TypeLine: "Creature — Orc Archer",
		Power: 1, Toughness: 1, Owner: active.ID, Controller: active.ID,
	})
	active.ManaPool.AddMana(game.ManaToken{Color: "C"})
	active.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.DeclareAttacker(bear, defender.ID); err != nil {
		t.Fatalf("attacking into Ghostly Prison with {2}: %v", err)
	}

	var line *LogEvent
	v := ViewOfGame(g)
	for i := range v.Log {
		if v.Log[i].Kind == LogAttackTax {
			if line != nil {
				t.Fatalf("two attack_tax lines for one payment: %+v", v.Log)
			}
			line = &v.Log[i]
		}
	}
	if line == nil {
		t.Fatalf("no attack_tax line in the log: %+v", v.Log)
	}
	if line.CardID != prison.String() || line.Label != "{2}" {
		t.Errorf("card_id %q label %q, want the Prison and {2}", line.CardID, line.Label)
	}
	if line.TargetSeat == nil || g.Seats[*line.TargetSeat].ID != defender.ID || g.Seats[line.Seat].ID != active.ID {
		t.Errorf("seat %d target_seat %v, want the attacker paying to attack the Prison's controller", line.Seat, line.TargetSeat)
	}
	want := active.Name + " paid {2} to attack " + defender.Name + " (Ghostly Prison)"
	if line.Text != want {
		t.Errorf("text = %q, want %q", line.Text, want)
	}
}

// The projection and the sentence, for a two-creature payment: the
// price is what that one permanent charged, concatenated per creature.
func TestTheAttackTaxLineNamesThePriceAndTheCard(t *testing.T) {
	turn, step := 1, ""
	var sacrificed uuid.UUID
	src, payer, defender := uuid.New(), uuid.New(), uuid.New()
	e, ok := projectEvent(game.Event{Kind: game.EventAttackTaxPaid, Actor: payer, CardID: src, Source: src, Target: defender, Label: "{2}{2}"},
		func(id uuid.UUID) int {
			switch id {
			case payer:
				return 0
			case defender:
				return 1
			}
			return NoSeat
		}, &turn, &step, &sacrificed)
	if !ok || e.Kind != LogAttackTax || e.CardID != src.String() || e.Label != "{2}{2}" || e.TargetSeat == nil || *e.TargetSeat != 1 {
		t.Fatalf("projected %+v, %v", e, ok)
	}
	e.actorName, e.targetSeatName = "P1", "P2"
	if got := renderLogText(e, "Propaganda", ""); got != "P1 paid {2}{2} to attack P2 (Propaganda)" {
		t.Errorf("line = %q", got)
	}
}
