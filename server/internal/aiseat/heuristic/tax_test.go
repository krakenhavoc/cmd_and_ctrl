package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// tax_test.go is ADR 0136's amendment of 2026-10-09: an opponent's tax
// against the plan (Config.PlanWeighTaxes). The board is seq 133's
// (TestPlanCastsTheRockFirst): the plan casts Arcane Signet, then
// Ornithopter of Paradise on the Signet's mana. An opponent's Rhystic
// Study asks for {1} while the Signet is on the stack.

const taxChoice = "choice-rhystic"

func taxSignet() protocol.CardView {
	return protocol.CardView{InstanceID: cardID(10), Name: "Arcane Signet", Owner: seatID(0).String(), Controller: seatID(0).String(),
		TypeLine: "Artifact", ManaCost: "{2}", KnownByYou: true,
		ManaAbilities: []protocol.ManaAbilityView{{TapCost: true, Produced: "{W|U|B|R|G}"}}}
}

func taxDork() protocol.CardView {
	dork := creature(cardID(11), 0, "Ornithopter of Paradise", 2, 1, keywords("flying"))
	dork.TypeLine, dork.ManaCost = "Artifact Creature — Thopter", "{2}"
	dork.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, Produced: "{W|U|B|R|G}"}}
	return dork
}

func taxBig() protocol.CardView {
	big := creature(cardID(12), 0, "Avenger of Zendikar", 5, 5)
	big.ManaCost = "{5}{G}{G}"
	return big
}

// taxStudy is an opponent's Rhystic Study, its triggered row declaring
// the card a decline gives (ADR 0126 §6). declares false leaves the row
// without a purpose.
func taxStudy(declares bool) protocol.CardView {
	study := protocol.CardView{InstanceID: cardID(40), Name: "Rhystic Study", Owner: seatID(1).String(), Controller: seatID(1).String(),
		TypeLine: "Enchantment", ManaCost: "{2}{U}"}
	row := protocol.AbilityRowView{Kind: "triggered", Label: "draw unless caster pays {1}"}
	if declares {
		row.Purpose = &protocol.PurposeView{Draws: 1}
	}
	study.AbilityRows = []protocol.AbilityRowView{row}
	return study
}

func taxLands(untapped int) []protocol.CardView {
	orchard := land(cardID(3), 0)
	orchard.Name, orchard.TypeLine = "Exotic Orchard", "Land"
	orchard.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, ColorOptions: [][]string{{"G", "U"}}}}
	field := land(cardID(2), 0)
	field.Name, field.TypeLine = "Field of the Dead", "Land"
	field.ManaAbilities = []protocol.ManaAbilityView{{TapCost: true, Produced: "{C}"}}
	lands := []protocol.CardView{planIsland(cardID(1)), field, orchard}
	for i := 0; i < untapped-1; i++ {
		lands = append(lands, planForest(cardID(4+i)))
	}
	// Two of them paid for the Signet.
	lands[1].Tapped, lands[2].Tapped = true, true
	return lands
}

func taxSeats(hand ...protocol.CardView) []protocol.PlayerView {
	return []protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1), newSeat(2), newSeat(3)}
}

// planWindow is seq 133 at a four-seat table, before the Signet is
// cast; extra adds untapped Forests.
func planWindow(t *testing.T, extra int) aiseat.Input {
	lands := taxLands(1 + extra)
	for i := range lands {
		lands[i].Tapped = false
	}
	v := newView(taxSeats(taxSignet(), taxDork(), taxBig()),
		withTurn(3, 0, "precombat_main"),
		withBattlefield(append(lands, taxStudy(true))...),
	)
	return input(0, v,
		passMove(0),
		stampedCast(t, cardID(11), "Cast Ornithopter of Paradise", "{2}"),
		stampedCast(t, cardID(10), "Cast Arcane Signet", "{2}"),
	)
}

// taxWindow is the Study's "pay {1}?" with the Signet on the stack.
func taxWindow(t *testing.T, extra int, declares bool) aiseat.Input {
	signet := taxSignet()
	v := newView(taxSeats(taxDork(), taxBig()),
		withTurn(3, 0, "precombat_main"),
		withBattlefield(append(taxLands(1+extra), taxStudy(declares))...),
		withStack(signet),
		withChoice(protocol.PendingChoiceView{
			ID: taxChoice, Kind: "pay_unless", Chooser: seatID(0).String(), FromPlayer: seatID(0).String(),
			Source: cardID(40), Reason: "Rhystic Study — pay {1}?", PayCost: "{1}",
		}),
	)
	v.StackItems = []protocol.StackItemView{{ID: "item-signet", Kind: "spell", Controller: seatID(0).String(), Owner: seatID(0).String(), SourceCardID: signet.InstanceID}}
	yes, no := true, false
	return input(0, v,
		choiceMove(t, 0, taxChoice, "pay", map[string]any{"apply": yes}),
		choiceMove(t, 0, taxChoice, "decline", map[string]any{"apply": no}),
	)
}

// planThenTax plays the plan window and then the tax window on one
// policy, as a seat does.
func planThenTax(t *testing.T, cfg heuristic.Config, extra int, declares bool) (aiseat.Decision, aiseat.Input) {
	t.Helper()
	p := heuristic.NewWithConfig(cfg)
	pin := planWindow(t, extra)
	if got := chose(t, pin, decide(t, p, pin)); got != "Cast Arcane Signet" {
		t.Fatalf("the plan window casts %q, want Arcane Signet", got)
	}
	in := taxWindow(t, extra, declares)
	return decide(t, p, in), in
}

func TestTaxIsDeclinedWhenItWouldStrandThePlan(t *testing.T) {
	d, in := planThenTax(t, heuristic.DefaultConfig(), 0, true)
	if got := chose(t, in, d); got != "decline" {
		t.Errorf("the bot chose %q (%s), want to decline: paying leaves one mana for the {2} Ornithopter", got, d.Reason)
	}
	if !strings.Contains(d.Reason, "Ornithopter of Paradise") {
		t.Errorf("reason %q does not name the plan member", d.Reason)
	}

	off := heuristic.DefaultConfig()
	off.PlanWeighTaxes = false
	if d, in := planThenTax(t, off, 0, true); chose(t, in, d) != "pay" {
		t.Errorf("with PlanWeighTaxes off the bot chose %q, want to pay as before", chose(t, in, d))
	}
}

func TestTaxIsPaidWhenThePlanStaysPayable(t *testing.T) {
	// One more Forest: after the {1}, the Island, the Forest and the
	// Signet's mana still pay the Ornithopter's {2}.
	if d, in := planThenTax(t, heuristic.DefaultConfig(), 1, true); chose(t, in, d) != "pay" {
		t.Errorf("the bot chose %q (%s), want to pay: the plan stays payable", chose(t, in, d), d.Reason)
	}
}

func TestTaxIsPaidWithoutADeclaredGiftOrAPlan(t *testing.T) {
	// The Study declares nothing: the bot cannot say what it prevents.
	if d, in := planThenTax(t, heuristic.DefaultConfig(), 0, false); chose(t, in, d) != "pay" {
		t.Errorf("with no declared gift the bot chose %q, want to pay", chose(t, in, d))
	}
	// A policy that chose no plan this turn pays.
	in := taxWindow(t, 0, true)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "pay" {
		t.Errorf("with no plan the bot chose %q, want to pay", got)
	}
}

func TestTaxIsPaidWhenTheGiftOutweighsTheMember(t *testing.T) {
	// With the opposition weighed at four times its price, a card for
	// an opponent is worth more than the Ornithopter.
	cfg := heuristic.DefaultConfig()
	cfg.Weights = heuristic.DefaultWeights()
	cfg.Weights.OpponentMean *= 4
	cfg.Weights.OpponentMax *= 4
	if d, in := planThenTax(t, cfg, 0, true); chose(t, in, d) != "pay" {
		t.Errorf("the bot chose %q (%s), want to pay", chose(t, in, d), d.Reason)
	}
}
