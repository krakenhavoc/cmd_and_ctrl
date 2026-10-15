package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// landfall_leftovers_test.go — #3146, the three gaps #3072 left open:
// Ancient Greenwarden's doubling, a payoff's own cast, and fetch lands.

func greenwarden(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Ancient Greenwarden", 5, 7)
	c.TypeLine, c.ManaCost = "Creature — Elemental", "{4}{G}{G}"
	c.Purpose = &protocol.PurposeView{LandfallExtraTriggers: 1}
	c.AbilityRows = []protocol.AbilityRowView{{Kind: "static", Label: "landfall triggers twice"}}
	return c
}

func tirelessTracker(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Tireless Tracker", 3, 2)
	c.TypeLine, c.ManaCost = "Creature — Human Scout", "{2}{G}"
	return withLandfallPayoff(c, protocol.LandfallPayoffView{Tokens: 1})
}

func noLandfallSwitches(cfg heuristic.Config) heuristic.Config {
	cfg.PriceLandfallDoublers = false
	cfg.PriceLandfallCasts = false
	cfg.PriceFetchLandfall = false
	return cfg
}

// Greenwarden makes each of the Hart's lands pay Tatyova twice.
func TestGreenwardenDoublesTheLandfallPrice(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	once, _ := hartRank(t, cfg, tatyova(cardID(2), 0))
	twice, _ := hartRank(t, cfg, tatyova(cardID(2), 0), greenwarden(cardID(3), 0))
	alone, _ := hartRank(t, cfg)
	t.Logf("Hart: %.2f alone, %.2f with Tatyova, %.2f with Tatyova and Greenwarden", alone, once, twice)
	if got, want := twice-once, once-alone; got < want-1e-9 || got > want+1e-9 {
		t.Errorf("Greenwarden adds %.2f, want Tatyova's two lands again (%.2f)", got, want)
	}
	// An opponent's Greenwarden doubles the opponent's triggers.
	if theirs, _ := hartRank(t, cfg, tatyova(cardID(2), 0), greenwarden(cardID(3), 1)); theirs != once {
		t.Errorf("an opponent's Greenwarden moves the Hart from %.2f to %.2f", once, theirs)
	}
	// The switch.
	off := noLandfallSwitches(cfg)
	if a, b := hartRank2(t, off, tatyova(cardID(2), 0)), hartRank2(t, off, tatyova(cardID(2), 0), greenwarden(cardID(3), 0)); a != b {
		t.Errorf("PriceLandfallDoublers off: Greenwarden moves the Hart from %.2f to %.2f", a, b)
	}
}

func hartRank2(t *testing.T, cfg heuristic.Config, others ...protocol.CardView) float64 {
	v, _ := hartRank(t, cfg, others...)
	return v
}

// payoffCastRank prices casting `card` for seat 0 in its precombat main phase
// on turn 6, its land drop used, with six lands and `others` out, and
// `lands` lands in hand beside the card.
func payoffCastRank(t *testing.T, cfg heuristic.Config, card protocol.CardView, lands int, others ...protocol.CardView) float64 {
	t.Helper()
	bf := append([]protocol.CardView{}, others...)
	for i := range 6 {
		bf = append(bf, land(cardID(40+i), 0))
	}
	hand := []protocol.CardView{card}
	for i := range lands {
		hand = append(hand, land(cardID(60+i), 0))
	}
	drops := func(p *protocol.PlayerView) { p.LandDropsPerTurn, p.LandsPlayedThisTurn = 1, 1 }
	v := newView([]protocol.PlayerView{newSeat(0, withHand(hand...), drops), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(6, 0, "precombat_main"))
	in := input(0, v, passMove(0), castMove(t, 0, card.InstanceID, "Cast "+card.Name))
	vals := map[int]float64{}
	for _, c := range heuristic.NewWithConfig(cfg).Rank(context.Background(), in) {
		vals[c.Index] = c.Value
	}
	return vals[1]
}

// Tireless Tracker cast with two lands in hand is worth the two Clues
// those lands will make over the next turns; with none in hand, the
// lands it may draw.
func TestALandfallPayoffIsWorthTheLandsAhead(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	off := noLandfallSwitches(cfg)
	tracker := tirelessTracker(cardID(1), 0)
	on2, off2 := payoffCastRank(t, cfg, tracker, 2), payoffCastRank(t, off, tracker, 2)
	on0, off0 := payoffCastRank(t, cfg, tracker, 0), payoffCastRank(t, off, tracker, 0)
	t.Logf("Tracker: %.2f on / %.2f off with two lands in hand, %.2f / %.2f with none", on2, off2, on0, off0)
	if got, want := on2-off2, 2*cfg.TokenWeight; got < want-1e-9 || got > want+1e-9 {
		t.Errorf("two lands in hand add %.2f, want two Clues (%.2f)", got, want)
	}
	if d := on0 - off0; d <= 0 || d >= on2-off2 {
		t.Errorf("no lands in hand add %.2f, want something less than two lands' worth", d)
	}
	// A creature with no landfall row is priced as before.
	bear := creature(cardID(1), 0, "Grizzly Bears", 2, 2)
	if a, b := payoffCastRank(t, cfg, bear, 2), payoffCastRank(t, off, bear, 2); a != b {
		t.Errorf("a vanilla creature is %.2f on and %.2f off", a, b)
	}
}

// Greenwarden's cast is worth Tatyova's payoff again for each land ahead
// when she is out, and nothing beyond its body without a payoff.
func TestGreenwardenCastIsWorthTheDoubling(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	off := noLandfallSwitches(cfg)
	gw := greenwarden(cardID(1), 0)
	with, withOff := payoffCastRank(t, cfg, gw, 2, tatyova(cardID(2), 0)), payoffCastRank(t, off, gw, 2, tatyova(cardID(2), 0))
	if with-withOff < 2*1.2-1e-9 {
		t.Errorf("Greenwarden with Tatyova out: %.2f on, %.2f off — want at least two cards more", with, withOff)
	}
	if a, b := payoffCastRank(t, cfg, gw, 2), payoffCastRank(t, off, gw, 2); a != b {
		t.Errorf("Greenwarden with no payoff out: %.2f on, %.2f off — want the same", a, b)
	}
}

const mistyLabel = "{T}, Pay 1 life, Sacrifice this land: Search your library for a Forest or Island card, put it onto the battlefield, then shuffle."

func fetchLand(id string, lands int) protocol.CardView {
	c := land(id, 0)
	c.Name, c.TypeLine, c.ManaAbilities = "Misty Rainforest", "Land", nil
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: mistyLabel, SacrificeSelf: true, TapCost: true,
		Purpose: &protocol.PurposeView{Lands: lands, Answers: &[]string{"value"}},
	}}
	return c
}

func fetchMove(t *testing.T, src string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label: "Misty Rainforest: " + mistyLabel, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{"source_card_id": src, "ability_index": 0}),
	}
}

func fetchRank(t *testing.T, cfg heuristic.Config, fetch protocol.CardView, others ...protocol.CardView) float64 {
	t.Helper()
	bf := append([]protocol.CardView{fetch}, others...)
	for i := range 5 {
		bf = append(bf, land(cardID(40+i), 0))
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(6, 0, "precombat_main"))
	in := input(0, v, passMove(0), fetchMove(t, fetch.InstanceID))
	vals := map[int]float64{}
	for _, c := range heuristic.NewWithConfig(cfg).Rank(context.Background(), in) {
		vals[c.Index] = c.Value
	}
	return vals[1]
}

// A fetch with Tatyova out is worth a card and a life more than one
// without; without a payoff it keeps the flat price it always had, so it
// is still cracked.
func TestAFetchPaysTheLandfallPayoffs(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	off := noLandfallSwitches(cfg)
	misty := fetchLand(cardID(1), 1)
	bare, bareOff := fetchRank(t, cfg, misty), fetchRank(t, off, misty)
	with, withOff := fetchRank(t, cfg, misty, tatyova(cardID(2), 0)), fetchRank(t, off, misty, tatyova(cardID(2), 0))
	t.Logf("Misty: %.2f / %.2f bare (on / off), %.2f / %.2f with Tatyova", bare, bareOff, with, withOff)
	if bare != bareOff || bare != cfg.ActivateBase {
		t.Errorf("a fetch with no payoff is %.2f on and %.2f off, want the flat %.2f", bare, bareOff, cfg.ActivateBase)
	}
	if bare <= cfg.PassThreshold {
		t.Errorf("a fetch with no payoff %.2f is under PassThreshold: it would stop being cracked", bare)
	}
	if withOff != bareOff {
		t.Errorf("PriceFetchLandfall off: Tatyova moves the fetch from %.2f to %.2f", bareOff, withOff)
	}
	if with-bare < 1.2-1e-9 {
		t.Errorf("Tatyova adds %.2f to the fetch, want at least a card", with-bare)
	}
	// Myriad Landscape's second land is ramp.
	if two := fetchRank(t, cfg, fetchLand(cardID(1), 2)); two <= bare {
		t.Errorf("a two-land fetch %.2f is not above a one-land fetch %.2f", two, bare)
	}
	// Each new switch is off in the baseline and on by default.
	b, d := heuristic.BaselineConfig(), heuristic.DefaultConfig()
	if b.PriceLandfallDoublers || b.PriceLandfallCasts || b.PriceFetchLandfall {
		t.Error("BaselineConfig turns on a #3146 switch")
	}
	if !d.PriceLandfallDoublers || !d.PriceLandfallCasts || !d.PriceFetchLandfall {
		t.Error("DefaultConfig leaves a #3146 switch off")
	}
}
