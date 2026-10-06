package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sacrifice_test.go is ADR 0126 §7's second half (PR 8): a sacrifice
// costs the creature's value times the chance the bot would have kept
// it, and each death payoff the bot controls adds to every creature it
// sacrifices.

// viscera is a Viscera Seer: "Sacrifice a creature: Scry 1."
func viscera(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Viscera Seer", 1, 1, rows("activated"))
	c.ManaCost = "{B}"
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{Index: 0, SacrificeLabel: "a creature"}}
	return c
}

// payoff is a creature with a death_payoff row (Blood Artist).
func payoff(id string, controller int, name string) protocol.CardView {
	c := creature(id, controller, name, 0, 1)
	c.AbilityRows = []protocol.AbilityRowView{{Kind: "triggered", Label: "dies", Purpose: &protocol.PurposeView{DeathPayoff: true}}}
	return c
}

func sacMove(t *testing.T, seat int, src, label string, sac ...string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: label, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{"source_card_id": src, "ability_index": 0, "sacrifice_ids": sac}),
	}
}

func withStackItems(items ...protocol.StackItemView) viewOpt {
	return func(v *protocol.GameView) { v.StackItems = append(v.StackItems, items...) }
}

// removal is an opponent's spell on the stack targeting card id.
func removal(controller int, spellID, target string) (protocol.CardView, protocol.StackItemView) {
	c := spell(spellID, controller, "Murder", "{1}{B}{B}")
	return c, protocol.StackItemView{
		ID: "item-" + spellID, Kind: "spell", Controller: seatID(controller).String(),
		SourceCardID: spellID, Targets: []protocol.TargetRefView{{Kind: "card", ID: target}},
	}
}

func baseline() *heuristic.Policy { return heuristic.NewWithConfig(heuristic.BaselineConfig()) }

// The ADR's `sacrifice-the-creature-the-removal-targets`, by hand: an
// opponent's Murder targets the Bear, and the bot sacrifices it to the
// Seer in response for a Blood Artist trigger, rather than the Goblin
// nothing threatens. The baseline charges the whole Bear and passes.
func TestSacrificeTheCreatureTheRemovalTargets(t *testing.T) {
	seer := viscera(cardID(1), 0)
	artist := payoff(cardID(2), 0, "Blood Artist")
	bear := creature(cardID(3), 0, "Bear", 2, 2)
	goblin := creature(cardID(4), 0, "Goblin", 2, 2)
	murder, item := removal(1, cardID(9), cardID(3))
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(seer, artist, bear, goblin),
		withStack(murder), withStackItems(item),
		withTurn(5, 1, "precombat_main"),
	)
	in := input(0, v, passMove(0),
		sacMove(t, 0, cardID(1), "Sacrifice Goblin", cardID(4)),
		sacMove(t, 0, cardID(1), "Sacrifice Bear", cardID(3)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Sacrifice Bear" {
		t.Errorf("the heuristic chose %q, want Sacrifice Bear", got)
	}
	if got := chose(t, in, decide(t, baseline(), in)); got != "Pass priority" {
		t.Errorf("the baseline chose %q, want the pre-S66 pass", got)
	}
}

// With nothing threatening the Bear, the same board is a pass: the
// removal discount is what made the sacrifice cheap.
func TestDoNotSacrificeWhatNothingThreatens(t *testing.T) {
	seer := viscera(cardID(1), 0)
	artist := payoff(cardID(2), 0, "Blood Artist")
	bear := creature(cardID(3), 0, "Bear", 2, 2)
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(seer, artist, bear),
		withTurn(5, 1, "precombat_main"),
	)
	in := input(0, v, passMove(0), sacMove(t, 0, cardID(1), "Sacrifice Bear", cardID(3)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass", got)
	}
}

// A declared sweep on the stack removes every creature it matches, so
// each one is free to sacrifice, payoff or none; an indestructible one
// survives a destroy sweep and is not.
func TestSacrificeIntoADeclaredSweep(t *testing.T) {
	seer := viscera(cardID(1), 0)
	bear := creature(cardID(3), 0, "Bear", 2, 2)
	god := creature(cardID(4), 0, "God", 4, 4, keywords("indestructible"))
	wrath := spell(cardID(9), 1, "Wrath of God", "{2}{W}{W}")
	wrath.TypeLine = "Sorcery"
	wrath.Purpose = &protocol.PurposeView{Sweep: &protocol.SweepView{Matches: "creatures", How: "destroy"}}
	item := protocol.StackItemView{ID: "w", Kind: "spell", Controller: seatID(1).String(), SourceCardID: cardID(9)}
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(seer, bear, god),
		withStack(wrath), withStackItems(item),
		withTurn(5, 1, "precombat_main"),
	)
	in := input(0, v, passMove(0),
		sacMove(t, 0, cardID(1), "Sacrifice God", cardID(4)),
		sacMove(t, 0, cardID(1), "Sacrifice Bear", cardID(3)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Sacrifice Bear" {
		t.Errorf("the heuristic chose %q, want Sacrifice Bear", got)
	}
	if got := chose(t, in, decide(t, baseline(), in)); got != "Pass priority" {
		t.Errorf("the baseline chose %q, want the pre-S66 pass", got)
	}
	// "creatures your opponents control", cast by the bot's opponent,
	// takes the bot's creatures; cast by the bot, it does not.
	wrath.Purpose.Sweep.OpponentsOnly = true
	item.Controller = seatID(0).String()
	v.Stack.Cards[0] = wrath
	v.StackItems[0] = item
	in.View = v
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("with the bot's own one-sided sweep the heuristic chose %q, want a pass", got)
	}
}

// Once blockers are declared, a chump blocker is dying anyway and the
// attacker stays blocked without it (CR 509.1h), so the bot sacrifices
// it. A blocker that trades is not: sacrificed, the trade is gone.
func TestSacrificeTheChumpBlockerAfterBlocks(t *testing.T) {
	seer := viscera(cardID(1), 0)
	wurm := creature(cardID(20), 1, "Wurm", 6, 6, attacking(0))
	chump := creature(cardID(3), 0, "Goblin", 1, 1, blocking(cardID(20)))
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(seer, wurm, chump),
		withTurn(5, 1, "declare_blockers"),
	)
	in := input(0, v, passMove(0), sacMove(t, 0, cardID(1), "Sacrifice Goblin", cardID(3)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Sacrifice Goblin" {
		t.Errorf("chumping, the heuristic chose %q, want Sacrifice Goblin", got)
	}

	bear := creature(cardID(21), 1, "Bear", 2, 2, attacking(0))
	trader := creature(cardID(3), 0, "Bear", 2, 2, blocking(cardID(21)))
	v = newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(seer, bear, trader),
		withTurn(5, 1, "declare_blockers"),
	)
	in = input(0, v, passMove(0), sacMove(t, 0, cardID(1), "Sacrifice Bear", cardID(3)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("trading, the heuristic chose %q, want a pass", got)
	}

	// A trampler's damage goes through if its blocker leaves (CR
	// 702.19d), so the chump in front of one is still doing its job.
	trampler := creature(cardID(22), 1, "Wurm", 6, 6, attacking(0), keywords("trample"))
	chump = creature(cardID(3), 0, "Goblin", 1, 1, blocking(cardID(22)))
	v = newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(seer, trampler, chump),
		withTurn(5, 1, "declare_blockers"),
	)
	in = input(0, v, passMove(0), sacMove(t, 0, cardID(1), "Sacrifice Goblin", cardID(3)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("chumping a trampler, the heuristic chose %q, want a pass", got)
	}
}

// Death payoffs make a token worth sacrificing in the bot's own main
// phase once there are enough of them: three here (two are not quite:
// 0.5 − 0.3 − 1.45 + 2 × 0.6 is just below passing). The commander
// still is not worth sacrificing, and a payoff's own row does not count
// towards its own sacrifice.
func TestDeathPayoffsAndTheCommander(t *testing.T) {
	seer := viscera(cardID(1), 0)
	artist := payoff(cardID(2), 0, "Blood Artist")
	zulaport := payoff(cardID(5), 0, "Zulaport Cutthroat")
	bastion := payoff(cardID(6), 0, "Bastion of Remembrance")
	bastion.TypeLine, bastion.Power, bastion.Toughness = "Enchantment", 0, 0
	token := creature(cardID(3), 0, "Zombie", 1, 1)
	token.IsToken = true
	token.ManaCost = ""
	konrad := creature(cardID(4), 0, "Syr Konrad, the Grim", 5, 4, commander())
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(seer, artist, zulaport, bastion, token, konrad),
		withTurn(5, 0, "postcombat_main"),
	)
	in := input(0, v, passMove(0),
		sacMove(t, 0, cardID(1), "Sacrifice Syr Konrad, the Grim", cardID(4)),
		sacMove(t, 0, cardID(1), "Sacrifice Blood Artist", cardID(2)),
		sacMove(t, 0, cardID(1), "Sacrifice Zombie", cardID(3)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Sacrifice Zombie" {
		t.Errorf("the heuristic chose %q, want Sacrifice Zombie", got)
	}
	ranked := heuristic.New().Rank(t.Context(), in)
	val := map[string]float64{}
	for _, c := range ranked {
		val[in.Moves[c.Index].Label] = c.Value
	}
	if val["Sacrifice Syr Konrad, the Grim"] >= 0 {
		t.Errorf("sacrificing the commander is priced %.2f, want below passing", val["Sacrifice Syr Konrad, the Grim"])
	}
	// The Artist's sacrifice counts the other two rows, not its own.
	w := heuristic.DefaultWeights()
	want := 0.5 - 0.3 - w.CreatureValue(&artist) + 2*0.6
	if got := val["Sacrifice Blood Artist"]; !near(got, want) {
		t.Errorf("sacrificing Blood Artist is priced %.4f, want %.4f", got, want)
	}
	if got := chose(t, in, decide(t, baseline(), in)); got != "Pass priority" {
		t.Errorf("the baseline chose %q, want the pre-S66 pass", got)
	}
}

// A cast whose additional cost sacrifices a creature (Deadly Dispute)
// reads the same price: the Bear the removal targets costs a fifth of
// what the Goblin nothing threatens costs.
func TestASacrificeCastIsPricedTheSameWay(t *testing.T) {
	dispute := spell(cardID(10), 0, "Deadly Dispute", "{1}{B}")
	bear := creature(cardID(3), 0, "Bear", 2, 2)
	goblin := creature(cardID(4), 0, "Goblin", 2, 2)
	murder, item := removal(1, cardID(9), cardID(3))
	v := newView(
		[]protocol.PlayerView{newSeat(0, withHand(dispute)), newSeat(1)},
		withBattlefield(bear, goblin, land(cardID(30), 0), land(cardID(31), 0)),
		withStack(murder), withStackItems(item),
		withTurn(5, 1, "precombat_main"),
	)
	cast := func(label, sac string) legal.Move {
		return legal.Move{
			Type: legal.TypeCastSpell, Player: seatID(0), Kind: legal.KindCast, Label: label,
			Source: uuid.MustParse(cardID(10)),
			Params: mustJSON(t, map[string]any{"instance_id": cardID(10), "from_zone": "hand", "sacrifice_ids": []string{sac}}),
		}
	}
	in := input(0, v, passMove(0), cast("Goblin", cardID(4)), cast("Bear", cardID(3)))
	val := map[string]float64{}
	for _, c := range heuristic.New().Rank(t.Context(), in) {
		val[in.Moves[c.Index].Label] = c.Value
	}
	w := heuristic.DefaultWeights()
	if got, want := val["Bear"]-val["Goblin"], 0.8*w.CreatureValue(&bear); !near(got, want) {
		t.Errorf("the Bear's sacrifice is %.4f cheaper than the Goblin's, want %.4f", got, want)
	}
}
