package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// token_payoff_test.go — #3165: Hordeling Outburst was cast in 9 of the
// 37 games it was offered in, and Krenko, Mob Boss's activation taken in
// 8 of 30, in a deck where every creature that enters is damage to each
// opponent. A creature token is now worth what the bot's own
// creature-enter payoffs pay for it.

// withEnterPayoff gives a permanent a triggered row that pays that for
// each creature entering under its controller's control.
func withEnterPayoff(c protocol.CardView, l protocol.CreatureEnterPayoffView) protocol.CardView {
	c.AbilityRows = append(c.AbilityRows, protocol.AbilityRowView{
		Kind: "triggered", Label: "creature enters",
		Purpose: &protocol.PurposeView{CreatureEnterPayoff: &l},
	})
	return c
}

func purphoros(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Purphoros, God of the Forge", 6, 5)
	c.TypeLine, c.ManaCost = "Legendary Enchantment Creature — God", "{3}{R}"
	return withEnterPayoff(c, protocol.CreatureEnterPayoffView{DamageEachOpponent: 2})
}

func hordelingOutburst(id string) protocol.CardView {
	c := spell(id, 0, "Hordeling Outburst", "{1}{R}{R}")
	c.TypeLine = "Sorcery"
	c.Purpose = &protocol.PurposeView{Tokens: 3, CreatureTokens: 3}
	return c
}

func goblins(from, n int, controller int) []protocol.CardView {
	var out []protocol.CardView
	for i := range n {
		g := creature(cardID(from+i), controller, "Goblin token", 1, 1)
		g.TypeLine, g.IsToken = "Token Creature — Goblin", true
		out = append(out, g)
	}
	return out
}

// castRank prices casting c for seat 0 on turn 5 with three lands, with
// `others` on the battlefield, and returns that and passing.
func castRank(t *testing.T, cfg heuristic.Config, c protocol.CardView, others ...protocol.CardView) (cast, pass float64) {
	t.Helper()
	bf := append([]protocol.CardView{}, others...)
	for i := range 3 {
		bf = append(bf, land(cardID(40+i), 0))
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHand(c)), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(5, 0, "precombat_main"))
	in := input(0, v, passMove(0), castMove(t, 0, c.InstanceID, "Cast "+c.Name))
	vals := map[int]float64{}
	for _, r := range heuristic.NewWithConfig(cfg).Rank(context.Background(), in) {
		vals[r.Index] = r.Value
	}
	return vals[1], vals[0]
}

// With Purphoros out, each of Hordeling Outburst's three Goblins is 2
// damage to each of three opponents.
func TestHordelingOutburstIsWorthMoreWithPurphorosOut(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	h := hordelingOutburst(cardID(3))
	without, _ := castRank(t, cfg, h)
	with, pass := castRank(t, cfg, h, purphoros(cardID(2), 0))
	t.Logf("Hordeling Outburst: %.2f without Purphoros, %.2f with, pass %.2f", without, with, pass)
	if want := 3 * 3 * 2 * cfg.DamageToOpponent; with-without < want-1e-9 {
		t.Errorf("Purphoros adds %.2f, want at least three Goblins x 2 damage x three opponents (%.2f)", with-without, want)
	}
	if with <= cfg.PassThreshold || with <= pass {
		t.Errorf("with Purphoros out: cast %.2f against pass %.2f — want it cast", with, pass)
	}
	// Three bodies are worth more than two.
	fodder := hordelingOutburst(cardID(4))
	fodder.Name, fodder.Purpose = "Dragon Fodder", &protocol.PurposeView{Tokens: 2, CreatureTokens: 2}
	if f, _ := castRank(t, cfg, fodder, purphoros(cardID(2), 0)); f >= with {
		t.Errorf("Dragon Fodder %.2f, Hordeling Outburst %.2f with Purphoros — want three Goblins above two", f, with)
	}
}

// An opponent's Purphoros pays the opponent, a token that is not a
// creature does not trigger it, and the switch turns it all off.
func TestTokenPayoffNeedsMineAndCreaturesAndTheSwitch(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	h := hordelingOutburst(cardID(3))
	base, _ := castRank(t, cfg, h)

	if got, _ := castRank(t, cfg, h, purphoros(cardID(2), 1)); got != base {
		t.Errorf("an opponent's Purphoros moves the price %.2f -> %.2f", base, got)
	}

	treasures := hordelingOutburst(cardID(5))
	treasures.Name, treasures.Purpose = "Big Score", &protocol.PurposeView{Tokens: 2}
	tb, _ := castRank(t, cfg, treasures)
	if got, _ := castRank(t, cfg, treasures, purphoros(cardID(2), 0)); got != tb {
		t.Errorf("Treasures move the price %.2f -> %.2f with Purphoros out", tb, got)
	}

	off := cfg
	off.PriceTokenPayoffs = false
	offBase, _ := castRank(t, off, h)
	if got, _ := castRank(t, off, h, purphoros(cardID(2), 0)); got != offBase {
		t.Errorf("PriceTokenPayoffs off: Purphoros moves the price %.2f -> %.2f", offBase, got)
	}
	if heuristic.BaselineConfig().PriceTokenPayoffs {
		t.Error("BaselineConfig prices token payoffs; want it off")
	}
	if !heuristic.DefaultConfig().PriceTokenPayoffs {
		t.Error("DefaultConfig does not price token payoffs; want it on")
	}
}

func krenkoMob(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Krenko, Mob Boss", 3, 3)
	c.TypeLine, c.ManaCost = "Legendary Creature — Goblin Warrior", "{2}{R}{R}"
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: krenkoLabel,
		Purpose: &protocol.PurposeView{TokensPerControlled: "goblin", Answers: &[]string{"makes_blocker"}},
	}}
	return c
}

const krenkoLabel = "{T}: Create X 1/1 Goblins, where X is the number of Goblins you control"

// krenkoRank prices Krenko's activation and passing in seat 0's second
// main phase, with `n` other Goblins and `others` beside Krenko.
func krenkoRank(t *testing.T, cfg heuristic.Config, n int, others ...protocol.CardView) (act, pass float64) {
	t.Helper()
	bf := append([]protocol.CardView{krenkoMob(cardID(1), 0)}, goblins(60, n, 0)...)
	bf = append(bf, others...)
	for i := range 4 {
		bf = append(bf, land(cardID(40+i), 0))
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(6, 0, "postcombat_main"))
	m := legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label: "Krenko, Mob Boss: " + krenkoLabel, Source: uuid.MustParse(cardID(1)),
		Params: mustJSON(t, map[string]any{"source_card_id": cardID(1), "ability_index": 0}),
	}
	in := input(0, v, passMove(0), m)
	vals := map[int]float64{}
	for _, r := range heuristic.NewWithConfig(cfg).Rank(context.Background(), in) {
		vals[r.Index] = r.Value
	}
	return vals[1], vals[0]
}

// Krenko's activation is priced by the Goblins it makes, X counted from
// the board, and by what each of them triggers.
func TestKrenkoActivationIsPricedByItsGoblins(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	a1, _ := krenkoRank(t, cfg, 0)
	a4, _ := krenkoRank(t, cfg, 3)
	if d, want := a4-a1, 3*cfg.TokenWeight; d < want-1e-9 {
		t.Errorf("three more Goblins add %.2f to Krenko's activation, want %.2f", d, want)
	}
	withP, pass := krenkoRank(t, cfg, 3, purphoros(cardID(2), 0))
	if want := 4 * 3 * 2 * cfg.DamageToOpponent; withP-a4 < want-1e-9 {
		t.Errorf("Purphoros adds %.2f to four Goblins, want at least %.2f", withP-a4, want)
	}
	if withP <= cfg.PassThreshold || withP <= pass {
		t.Errorf("Krenko with four Goblins and Purphoros: %.2f against pass %.2f — want it activated", withP, pass)
	}

	off := cfg
	off.PriceTokenPayoffs = false
	o1, _ := krenkoRank(t, off, 0)
	o4, _ := krenkoRank(t, off, 3)
	if o1 != o4 {
		t.Errorf("PriceTokenPayoffs off: Krenko's activation depends on the Goblins (%.2f vs %.2f)", o1, o4)
	}
}
