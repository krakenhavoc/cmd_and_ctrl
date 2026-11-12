package legal_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ghostly_prison_5271_test.go — #5271, a player report: "attacked me
// with Orcish Bowmasters and did not pay 2 mana" into their Ghostly
// Prison. The table, as reported: four seats, the real catalog cards,
// the Prison's controller across the table from the attacker, and the
// declaration sent the way the client sends it (`auto_tap: true`, one
// creature or the whole swing). The engine charged the tax there: what
// was missing was a log line saying so (protocol/attack_tax_log_test.go).
// This pins the charge, for a human's wire action and for the bot's
// enumerator.

const oracleOrcishBowmasters = "ea5103f5-27e0-4eb1-902c-7f34652d6bf3"

func untappedLandsOf(g *game.Game, p *game.Player) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p.ID && c.IsLand() && !c.Tapped {
			n++
		}
	}
	return n
}

func TestGhostlyPrisonChargesOrcishBowmastersAcrossTheTable(t *testing.T) {
	for _, bulk := range []bool{false, true} {
		for _, lands := range []int{0, 2} {
			g := newTable(t)
			atk := g.Seats[g.Turn.ActiveSeat]
			def := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
			clearHand(atk)
			bow := battlefieldCard(g, atk, game.Card{
				Name: "Orcish Bowmasters", TypeLine: "Creature — Orc Archer", ManaCost: "{1}{B}",
				Power: 1, Toughness: 1, OracleID: oracleOrcishBowmasters,
			})
			battlefieldCard(g, def, game.Card{
				Name: "Ghostly Prison", TypeLine: "Enchantment", ManaCost: "{2}{W}", OracleID: oracleGhostlyPrison,
			})
			for range lands {
				battlefieldCard(g, atk, basic("Swamp", "Swamp"))
			}
			advanceTo(t, g, game.StepDeclareAttackers)

			// The bot's half: offered only when payable, at {2}, with the
			// tapper's permission to pay it.
			offered := false
			for _, m := range legal.EnumerateFor(g, atk.ID) {
				if m.Kind != legal.KindAttack || decodeAttack(t, m).Target != def.ID.String() {
					continue
				}
				offered = true
				if m.Cost == nil || m.Cost.Mana != "{2}" || !decodeAttack(t, m).AutoTap {
					t.Errorf("bulk=%v lands=%d: offered %q at %#v", bulk, lands, m.Label, m.Cost)
				}
			}
			if offered != (lands == 2) {
				t.Errorf("bulk=%v lands=%d: attack into the Prison offered = %v", bulk, lands, offered)
			}

			// The human's half: the wire action the client sends.
			var a actions.Action
			if bulk {
				params, _ := json.Marshal(map[string]any{
					"attackers": []map[string]any{{"attacker": bow.String(), "target": def.ID.String()}},
					"auto_tap":  true,
				})
				a = actions.Action{Type: actions.TypeDeclareAttackers, Player: atk.ID, Caller: atk.ID, Params: params}
			} else {
				params, _ := json.Marshal(map[string]any{"attacker": bow.String(), "target": def.ID.String(), "auto_tap": true})
				a = actions.Action{Type: actions.TypeDeclareAttacker, Player: atk.ID, Caller: atk.ID, Params: params}
			}
			err := actions.Dispatch(g, a)
			var target uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.InstanceID == bow {
					target = c.AttackingTarget
				}
			}
			if lands == 0 {
				if !errors.Is(err, game.ErrAttackTaxUnpaid) || target != uuid.Nil {
					t.Errorf("bulk=%v: no lands, yet the attack into Ghostly Prison returned %v (attacking %v)", bulk, err, target)
				}
				continue
			}
			if err != nil || target != def.ID {
				t.Fatalf("bulk=%v: two lands, attack into Ghostly Prison: %v (attacking %v)", bulk, err, target)
			}
			if n := untappedLandsOf(g, atk); n != 0 {
				t.Errorf("bulk=%v: %d lands left untapped, want 0: the {2} is tapped for", bulk, n)
			}
		}
	}
}
