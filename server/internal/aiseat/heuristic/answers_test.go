package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// answers_test.go is ADR 0142 decision 6: the bot prices the answers an
// activated row declares.

// answering gives a creature one untargeted activated row declaring
// answers, with an optional pump.
func answering(label string, pump *protocol.PumpView, answers ...string) cardOpt {
	return func(c *protocol.CardView) {
		a := append([]string(nil), answers...)
		c.ActivatedAbilities = append(c.ActivatedAbilities, protocol.ActivatedAbilityView{
			Index: len(c.ActivatedAbilities), Label: label, ManaCost: "{1}{G}",
			Purpose: &protocol.PurposeView{Answers: &a, Pump: pump},
		})
	}
}

func activate(t *testing.T, seat int, src, label string, extra map[string]any) legal.Move {
	p := map[string]any{"source_card_id": src, "ability_index": 0}
	for k, v := range extra {
		p[k] = v
	}
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: label, Source: uuid.MustParse(src), Params: mustJSON(t, p),
	}
}

// bolt is an opponent's spell on the stack that deals n damage to the
// target, declared on its slot.
func bolt(controller int, spellID, target string, n int) (protocol.CardView, protocol.StackItemView) {
	c := spell(spellID, controller, "Lightning Bolt", "{R}")
	c.Purpose = &protocol.PurposeView{Targets: &[]protocol.TargetPurposeView{{Slot: 0, Damage: n}}}
	return c, protocol.StackItemView{
		ID: "item-" + spellID, Kind: "spell", Controller: seatID(controller).String(),
		SourceCardID: spellID, Targets: []protocol.TargetRefView{{Kind: "card", ID: target}},
	}
}

// divination is an opponent's untargeted spell on the stack.
func divination(controller int, spellID string) (protocol.CardView, protocol.StackItemView) {
	c := spell(spellID, controller, "Divination", "{2}{U}")
	c.TypeLine = "Sorcery"
	return c, protocol.StackItemView{ID: "item-" + spellID, Kind: "spell", Controller: seatID(controller).String(), SourceCardID: spellID}
}

func respond(cards []protocol.CardView, stack protocol.CardView, item protocol.StackItemView, more ...protocol.StackItemView) protocol.GameView {
	return newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(cards...),
		withStack(stack), withStackItems(append([]protocol.StackItemView{item}, more...)...),
		withTurn(5, 1, "precombat_main"),
	)
}

func noAnswers() *heuristic.Policy {
	c := heuristic.DefaultConfig()
	c.PriceAnswers = false
	return heuristic.NewWithConfig(c)
}

func troll() protocol.CardView {
	return creature(cardID(1), 0, "Albino Troll", 3, 3, answering("{1}{G}: Regenerate this creature.", nil, "protect"))
}

// The ADR's `regenerate-the-targeted-troll`: an opponent's Murder
// targets the bot's Albino Troll, and the bot regenerates it. Without
// PriceAnswers the row is the flat ActivateBase and the Troll dies.
func TestRegenerateTheTargetedTroll(t *testing.T) {
	murder, item := removal(1, cardID(9), cardID(1))
	in := input(0, respond([]protocol.CardView{troll()}, murder, item), passMove(0), activate(t, 0, cardID(1), "Regenerate", nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Regenerate" {
		t.Errorf("the heuristic chose %q, want Regenerate", got)
	}
	if got := chose(t, in, decide(t, noAnswers(), in)); got != "Pass priority" {
		t.Errorf("without PriceAnswers it chose %q, want the old pass", got)
	}
	if got := chose(t, in, decide(t, baseline(), in)); got != "Pass priority" {
		t.Errorf("the baseline chose %q, want a pass", got)
	}
}

// The ADR's `do-not-regenerate-into-nothing`: Divination threatens
// nothing, so the shield is not worth the mana.
func TestDoNotRegenerateIntoNothing(t *testing.T) {
	div, item := divination(1, cardID(9))
	in := input(0, respond([]protocol.CardView{troll()}, div, item), passMove(0), activate(t, 0, cardID(1), "Regenerate", nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass", got)
	}
}

// A shield already on the stack above the Murder has answered it: the
// bot does not stack a second one.
func TestDoNotRegenerateTwice(t *testing.T) {
	murder, item := removal(1, cardID(9), cardID(1))
	shield := protocol.StackItemView{ID: "shield", Kind: "ability", Controller: seatID(0).String(), SourceCardID: cardID(1)}
	in := input(0, respond([]protocol.CardView{troll()}, murder, item, shield), passMove(0), activate(t, 0, cardID(1), "Regenerate", nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass", got)
	}
}

// A prevent row answers declared damage, and not a Murder.
func TestPreventAnswersDeclaredDamageOnly(t *testing.T) {
	frog := creature(cardID(1), 0, "Guardian", 3, 3, answering("{1}{G}: Prevent all damage that would be dealt to this creature this turn.", nil, "prevent"))
	b, bi := bolt(1, cardID(9), cardID(1), 3)
	in := input(0, respond([]protocol.CardView{frog}, b, bi), passMove(0), activate(t, 0, cardID(1), "Prevent", nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Prevent" {
		t.Errorf("against Bolt the heuristic chose %q, want Prevent", got)
	}
	murder, mi := removal(1, cardID(9), cardID(1))
	in = input(0, respond([]protocol.CardView{frog}, murder, mi), passMove(0), activate(t, 0, cardID(1), "Prevent", nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("against Murder the heuristic chose %q, want a pass", got)
	}
}

// A pump answers declared damage when one activation keeps the
// creature alive, and not when it falls short.
func TestPumpOutOfBurnRange(t *testing.T) {
	big := creature(cardID(1), 0, "Wall", 3, 2, answering("{1}{G}: +0/+2", &protocol.PumpView{Toughness: 2}, "pump"))
	small := creature(cardID(1), 0, "Wall", 3, 2, answering("{1}{G}: +1/+1", &protocol.PumpView{Power: 1, Toughness: 1}, "pump"))
	b, bi := bolt(1, cardID(9), cardID(1), 3)
	for _, tc := range []struct {
		card protocol.CardView
		want string
	}{{big, "Pump"}, {small, "Pass priority"}} {
		in := input(0, respond([]protocol.CardView{tc.card}, b, bi), passMove(0), activate(t, 0, cardID(1), "Pump", nil))
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
			t.Errorf("%s: the heuristic chose %q, want %q", tc.card.ActivatedAbilities[0].Label, got, tc.want)
		}
	}
}

// A row declared `value` waits for the stack, even one its purpose
// prices above the instant bar.
func TestValueRowWaitsForTheStack(t *testing.T) {
	well := creature(cardID(1), 0, "Library", 0, 4)
	a := []string{"value"}
	well.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Label: "{3}: Draw three cards.", ManaCost: "{3}",
		Purpose: &protocol.PurposeView{Draws: 3, Answers: &a}}}
	div, item := divination(1, cardID(9))
	in := input(0, respond([]protocol.CardView{well}, div, item), passMove(0), activate(t, 0, cardID(1), "Draw", nil))
	if got := chose(t, in, decide(t, noAnswers(), in)); got != "Draw" {
		t.Fatalf("without PriceAnswers it chose %q; the test needs a row priced above the bar", got)
	}
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass", got)
	}
}

// elder is Sakura-Tribe Elder: "Sacrifice this creature: search for a
// basic land", declared value.
func elder() protocol.CardView {
	c := creature(cardID(1), 0, "Sakura-Tribe Elder", 1, 1)
	a := []string{"value"}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Label: "Sacrifice this creature: Search for a basic land.",
		SacrificeSelf: true, Purpose: &protocol.PurposeView{Lands: 1, Answers: &a}}}
	return c
}

// The ADR's `sacrifice-the-elder-it-targets`: an opponent's removal
// targets the Elder, so it is sacrificed for its land in response. With
// nothing aimed at it, it waits.
func TestSacrificeTheElderItTargets(t *testing.T) {
	murder, item := removal(1, cardID(9), cardID(1))
	in := input(0, respond([]protocol.CardView{elder()}, murder, item), passMove(0), activate(t, 0, cardID(1), "Sacrifice the Elder", nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Sacrifice the Elder" {
		t.Errorf("the heuristic chose %q, want the sacrifice", got)
	}
	div, di := divination(1, cardID(9))
	in = input(0, respond([]protocol.CardView{elder()}, div, di), passMove(0), activate(t, 0, cardID(1), "Sacrifice the Elder", nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("against Divination the heuristic chose %q, want a pass", got)
	}
}

// A self-sacrifice row with no priced purpose costs its source. Before
// PriceAnswers it cost nothing past the blocker it taps.
func TestSelfSacrificeIsNotFree(t *testing.T) {
	c := creature(cardID(1), 0, "Bear", 2, 2)
	a := []string{"protect"}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Label: "Sacrifice this creature: Regenerate target creature.",
		SacrificeSelf: true, Purpose: &protocol.PurposeView{Answers: &a}}}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(c), withTurn(5, 0, "precombat_main"))
	in := input(0, v, passMove(0), activate(t, 0, cardID(1), "Sacrifice", nil))
	price := func(p *heuristic.Policy) float64 {
		for _, cand := range p.Rank(context.Background(), in) {
			if cand.Index == 1 {
				return cand.Value
			}
		}
		t.Fatal("the sacrifice was not ranked")
		return 0
	}
	if with, without := price(heuristic.New()), price(noAnswers()); with >= without-1 {
		t.Errorf("the sacrifice is priced %.2f with PriceAnswers and %.2f without; want it to pay for the Bear", with, without)
	}
}

// A `sac_outlet` row whose cost exiles a creature (The Soul Stone)
// spends a creature the removal targets at the dying-anyway price.
func TestExileOutletSpendsTheTargetedCreature(t *testing.T) {
	stone := creature(cardID(1), 0, "Outlet", 0, 3)
	a := []string{"sac_outlet"}
	stone.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, Label: "Exile a creature you control: Draw a card.",
		ExilePermanentLabel: "a creature you control", Purpose: &protocol.PurposeView{Draws: 1, Answers: &a}}}
	bear := creature(cardID(2), 0, "Bear", 2, 2)
	murder, item := removal(1, cardID(9), cardID(2))
	in := input(0, respond([]protocol.CardView{stone, bear}, murder, item), passMove(0),
		activate(t, 0, cardID(1), "Exile the Bear", map[string]any{"exile_permanent_ids": []string{cardID(2)}}))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Exile the Bear" {
		t.Errorf("the heuristic chose %q, want the exile", got)
	}
	if got := chose(t, in, decide(t, noAnswers(), in)); got != "Pass priority" {
		t.Errorf("without PriceAnswers it chose %q, want the old pass", got)
	}
}
