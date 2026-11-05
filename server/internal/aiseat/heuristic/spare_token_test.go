package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spare_token_test.go is #4991: with PriceSpareTokens on, Sai, Master
// Thopterist trades two spare Thopters for a card, keeps two, and never
// trades a body it has no copies of. The switch ships off (the price lost
// in the arena, docs/bot-testing.md), so the default policy passes.

// sai is Sai, Master Thopterist with her "{1}{U}, Sacrifice two
// artifacts: Draw a card." row as the view projects it.
func sai(id string) protocol.CardView {
	c := creature(id, 0, "Sai, Master Thopterist", 1, 4)
	c.TypeLine = "Legendary Creature — Human Artificer"
	c.ManaCost = "{2}{U}"
	c.AbilityRows = []protocol.AbilityRowView{{Kind: "triggered", Label: "cast an artifact"}, {Kind: "activated", Label: "draw"}}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: "{1}{U}, Sacrifice two artifacts: Draw a card.", ManaCost: "{1}{U}",
		SacrificeLabel: "two artifacts",
		Purpose:        &protocol.PurposeView{Draws: 1, Answers: &[]string{"value"}},
	}}
	return c
}

// thopter is a 1/1 colorless Thopter artifact creature token with flying.
func thopter(id string, opts ...cardOpt) protocol.CardView {
	c := creature(id, 0, "Thopter", 1, 1, keywords("flying"))
	c.TypeLine = "Token Artifact Creature — Thopter"
	c.ManaCost = ""
	c.IsToken = true
	for _, o := range opts {
		o(&c)
	}
	return c
}

// saiDraw is Sai's activation paying with the two permanents named, as
// the enumerator offers it (tokens first).
func saiDraw(t *testing.T, src string, sac ...string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(0), Kind: legal.KindActivate,
		Label: "Sai: draw a card", Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{"source_card_id": src, "ability_index": 0, "sacrifice_ids": sac}),
	}
}

// saiBoard is Sai, the Thopters and two Islands in the bot's step, with
// Sai's activation paying with the first two Thopters.
func saiBoard(t *testing.T, step string, thopters ...protocol.CardView) aiseat.Input {
	t.Helper()
	cards := []protocol.CardView{sai(cardID(1)), island(cardID(90), 0), island(cardID(91), 0)}
	cards = append(cards, thopters...)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(cards...), withTurn(6, 0, step))
	return input(0, v, passMove(0), saiDraw(t, cardID(1), thopters[0].InstanceID, thopters[1].InstanceID))
}

// spareTokens is today's heuristic with the spare-token price on, as the
// arena's `heuristic-sparetokens` seats it.
func spareTokens() *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.PriceSpareTokens = true
	return heuristic.NewWithConfig(cfg)
}

// The issue's position: four Thopters after combat. With the price on,
// two are spare, so Sai trades them for a card and keeps two. With it off
// (the default), the two cost their bodies (2.95 each against the card's
// 1.20) and the bot passes, as it did in every arena game of #4991.
func TestSaiTradesSpareThoptersForACard(t *testing.T) {
	in := saiBoard(t, "postcombat_main", thopter(cardID(2)), thopter(cardID(3)), thopter(cardID(4)), thopter(cardID(5)))
	if got := chose(t, in, decide(t, spareTokens(), in)); got != "Sai: draw a card" {
		t.Errorf("with spare tokens priced, the heuristic chose %q, want Sai's draw with two spare Thopters", got)
	}
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Errorf("the default heuristic chose %q, want the pass it makes with the switch off", got)
	}
	if got := chose(t, in, decide(t, baseline(), in)); got != "Pass priority" {
		t.Errorf("the baseline chose %q, want a pass", got)
	}
}

// With three Thopters the pair would leave one: neither is spare, and
// the bot keeps them.
func TestSaiKeepsTwoThopters(t *testing.T) {
	in := saiBoard(t, "postcombat_main", thopter(cardID(2)), thopter(cardID(3)), thopter(cardID(4)))
	if got := chose(t, in, decide(t, spareTokens(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q, want a pass: the trade would leave one Thopter", got)
	}
}

// Before combat on the bot's own turn a Thopter that can attack also
// costs its attack, so the trade waits for the second main phase.
func TestSaiTradesThoptersAfterTheyAttack(t *testing.T) {
	in := saiBoard(t, "precombat_main", thopter(cardID(2)), thopter(cardID(3)), thopter(cardID(4)), thopter(cardID(5)))
	if got := chose(t, in, decide(t, spareTokens(), in)); got != "Pass priority" {
		t.Errorf("the heuristic chose %q before combat, want the attack kept", got)
	}
}

// Copies are the same body: a Thopter with a +1/+1 counter on it is not
// one of four Thopters, a creature card is never spare, and the copies
// kept must be the bot's own.
func TestSpareTokensNeedKeptCopiesOfTheSameBody(t *testing.T) {
	grown := func(c *protocol.CardView) { c.Counters = map[string]int{"+1/+1": 1}; c.Power, c.Toughness = 2, 2 }
	theirs := func(c *protocol.CardView) { c.Controller, c.Owner = seatID(1).String(), seatID(1).String() }
	card := func(c *protocol.CardView) {
		c.IsToken = false
		c.TypeLine = "Artifact Creature — Thopter"
		c.ManaCost = "{0}"
	}
	for name, kept := range map[string][]protocol.CardView{
		"a grown copy": {thopter(cardID(4)), thopter(cardID(5), grown)},
		"their copies": {thopter(cardID(4), theirs), thopter(cardID(5), theirs)},
		"a Wurm token": {thopter(cardID(4)), func() protocol.CardView {
			c := thopter(cardID(5))
			c.Name, c.Power, c.Toughness, c.Abilities = "Wurm", 3, 3, nil
			return c
		}()},
	} {
		t.Run(name, func(t *testing.T) {
			in := saiBoard(t, "postcombat_main", append([]protocol.CardView{thopter(cardID(2)), thopter(cardID(3))}, kept...)...)
			if got := chose(t, in, decide(t, spareTokens(), in)); got != "Pass priority" {
				t.Errorf("the heuristic chose %q, want a pass: only one kept copy of the Thopter", got)
			}
		})
	}
	t.Run("artifact creature cards", func(t *testing.T) {
		in := saiBoard(t, "postcombat_main", thopter(cardID(2), card), thopter(cardID(3), card), thopter(cardID(4), card), thopter(cardID(5), card))
		if got := chose(t, in, decide(t, spareTokens(), in)); got != "Pass priority" {
			t.Errorf("the heuristic chose %q, want a pass: a card is never spare", got)
		}
	})
}

// The switch is off by default, keeping two copies when it is turned on,
// and off in the baseline.
func TestPriceSpareTokensDefaults(t *testing.T) {
	if cfg := heuristic.DefaultConfig(); cfg.PriceSpareTokens || cfg.SpareTokenKeep != 2 {
		t.Errorf("DefaultConfig: PriceSpareTokens %v, SpareTokenKeep %d; want off, keeping two", cfg.PriceSpareTokens, cfg.SpareTokenKeep)
	}
	if cfg := heuristic.BaselineConfig(); cfg.PriceSpareTokens || cfg.SpareTokenKeep != 0 {
		t.Errorf("BaselineConfig: PriceSpareTokens %v, SpareTokenKeep %d; want both zero", cfg.PriceSpareTokens, cfg.SpareTokenKeep)
	}
}
