package testsg

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// herigast_erupting_nullkite_test.go — ADR 0135 PR 6 (#2416): Herigast,
// Erupting Nullkite gives every creature spell its controller casts emerge
// (CR 702.119a), "The emerge cost is equal to its mana cost", through the
// granted alternative-cost seam (ADR 0118 §3).

const herigastOracle = "76243b38-cab1-465f-aa7d-5bc617541753"

// herigastOnBattlefield puts Herigast on `owner`'s battlefield.
func herigastOnBattlefield(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{Name: "Herigast, Erupting Nullkite", OracleID: herigastOracle,
		TypeLine: "Legendary Creature — Eldrazi Dragon", ManaCost: "{9}", Power: 6, Toughness: 6})
}

// castGrantedEmerge casts `card` for Herigast's granted emerge cost
// sacrificing `sac`.
func castGrantedEmerge(g *game.Game, p *game.Player, card, sac uuid.UUID, strict bool, x int) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{
		AlternativeCost: effects.GrantedAltCostEmerge, AltCostIDs: []uuid.UUID{sac},
		Strict: strict, AutoTap: strict, XValue: x,
	})
}

// grantedEmergePrice is the engine's one pricer for `card` cast for the
// granted emerge cost over `sac`.
func grantedEmergePrice(t *testing.T, g *game.Game, p *game.Player, card, sac uuid.UUID, x int) string {
	t.Helper()
	return grantedEmergeTotal(t, g, p, card, sac, x).String()
}

// grantedEmergeTotal is the total cost that pricer charges.
func grantedEmergeTotal(t *testing.T, g *game.Game, p *game.Player, card, sac uuid.UUID, x int) game.ParsedCost {
	t.Helper()
	var out game.ParsedCost
	g.WithWriteLock(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			t.Fatal("the card is gone")
		}
		price, err := g.PriceCastForEffect(p.ID, c, game.CastSpellParams{
			AlternativeCost: effects.GrantedAltCostEmerge, AltCostIDs: []uuid.UUID{sac}, XValue: x})
		if err != nil {
			t.Fatalf("PriceCastForEffect: %v", err)
		}
		out = price.Total
	})
	return out
}

func TestHerigastDeclaresItsEmergeAndTheGrant(t *testing.T) {
	spec, ok := effects.Lookup(herigastOracle)
	if !ok {
		t.Fatal("Herigast is not registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("completeness %v, want Full", spec.Completeness)
	}
	offers := game.AlternativeCostsFor(herigastOracle)
	if len(offers) != 1 || offers[0].Key != effects.AltCostKeyEmerge || offers[0].Label != "Emerge {6}{R}{R}" {
		t.Errorf("own offers %+v, want Emerge {6}{R}{R}", offers)
	}
	if len(spec.GrantedAlternativeCosts) != 1 || spec.GrantedAlternativeCosts[0].Offer.Key != effects.GrantedAltCostEmerge ||
		!spec.GrantedAlternativeCosts[0].PricedAtManaCost || !spec.GrantedAlternativeCosts[0].Spells.CreatureOnly {
		t.Errorf("granted %+v, want the creature-spell emerge priced at the mana cost", spec.GrantedAlternativeCosts)
	}
}

// The grant's price is the spell's own mana cost, reduced by the
// sacrificed creature's mana value on the generic part only (CR 702.119a,
// 118.7a); it reaches creature spells and nothing else, and only the
// spells of Herigast's controller.
func TestHerigastGivesCreatureSpellsEmergeAtTheirManaCost(t *testing.T) {
	g, me, opp := effectstest.EmergeTable(t)
	herigastOnBattlefield(g, me.ID)
	bear := herigastBear(g, me)
	bolt := effectstest.HandCardOf(g, me, "Bolt", "Instant", "{4}{R}", "")
	four := effectstest.EmergeFodder(g, me.ID, "Four", "{4}", 4, 4)
	seven := effectstest.EmergeFodder(g, me.ID, "Seven", "{7}", 7, 7)
	spawn := effectstest.ApaPush(g, me.ID, me.ID, effects.EldraziSpawnToken())

	offers := effectstest.HandOffers(t, g, me, bear)
	o, ok := offers[effects.GrantedAltCostEmerge]
	if !ok {
		t.Fatalf("no granted emerge offer on a creature spell: %+v", offers)
	}
	if o.Label != "Emerge {4}{G}{G} (Herigast, Erupting Nullkite)" || !o.ReducesByManaValue || o.SacrificeOptions == nil {
		t.Errorf("offer %+v, want a flagged sacrifice offer labelled with its price and source", o)
	}
	if got := o.SacrificePrices[four.String()]; got.Price != "{G}{G}" {
		t.Errorf("wire price over a four-drop: %+v, want {G}{G}", got)
	}
	if _, ok := effectstest.HandOffers(t, g, me, bolt)[effects.GrantedAltCostEmerge]; ok {
		t.Error("an instant was given emerge")
	}

	for _, c := range []struct {
		name string
		sac  uuid.UUID
		want string
	}{
		{"a four-drop", four, "{G}{G}"},
		{"a seven-drop", seven, "{G}{G}"},
		{"a token", spawn, "{4}{G}{G}"},
	} {
		if got := grantedEmergePrice(t, g, me, bear, c.sac, 0); got != c.want {
			t.Errorf("%s: price %s, want %s", c.name, got, c.want)
		}
	}

	theirs := effectstest.HandCardOf(g, opp, "Their Bear", "Creature — Bear", "{4}{G}{G}", "")
	if _, ok := effectstest.HandOffers(t, g, opp, theirs)[effects.GrantedAltCostEmerge]; ok {
		t.Error("an opponent's creature spell was given emerge")
	}
	if err := castGrantedEmerge(g, me, bolt, four, false, 0); err == nil {
		t.Error("an instant was cast for a granted emerge cost")
	}
}

// End to end under the strict gate, sacrificing Herigast itself (the
// ruling: the grant is checked as the cost is chosen). Its mana value 9
// takes all of {4}{G}{G}'s generic, so two Forests pay {G}{G}.
func TestHerigastMayBeSacrificedToTheEmergeItGrants(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	herigast := herigastOnBattlefield(g, me.ID)
	bear := herigastBear(g, me)
	forests := effectstest.EmergeLands(g, me.ID, 1, "Forest")
	if err := castGrantedEmerge(g, me, bear, herigast, true, 0); err == nil {
		t.Fatal("a granted emerge for {G}{G} was paid with one Forest")
	}
	if effectstest.FindBattlefieldCardForTest(g, herigast) == nil || !me.Hand.Contains(bear) {
		t.Fatal("a refused cast paid something")
	}
	forests = append(forests, effectstest.EmergeLands(g, me.ID, 1, "Forest")...)
	if err := castGrantedEmerge(g, me, bear, herigast, true, 0); err != nil {
		t.Fatalf("Big Bear over Herigast for {G}{G}: %v", err)
	}
	if !me.Graveyard.Contains(herigast) {
		t.Error("Herigast was not sacrificed")
	}
	for _, id := range forests {
		if !effectstest.TappedForTest(t, g, id) {
			t.Error("a Forest was left untapped paying {G}{G}")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, bear) == nil {
		t.Error("the creature did not resolve")
	}
}

// With Herigast gone, the grant is gone (ADR 0118 §3: the offer lasts as
// long as the source).
func TestHerigastGrantEndsWhenItLeaves(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	herigast := herigastOnBattlefield(g, me.ID)
	bear := herigastBear(g, me)
	four := effectstest.EmergeFodder(g, me.ID, "Four", "{4}", 4, 4)
	g.WithWriteLock(func() { g.BounceCardsToHandForEffect([]uuid.UUID{herigast}) })
	if _, ok := effectstest.HandOffers(t, g, me, bear)[effects.GrantedAltCostEmerge]; ok {
		t.Error("the offer outlived Herigast")
	}
	if err := castGrantedEmerge(g, me, bear, four, false, 0); err == nil {
		t.Error("a granted emerge was claimed with no Herigast")
	}
}

// "The emerge cost is equal to its mana cost": an {X} in the mana cost
// stays in the emerge cost, and the caster chooses X (CR 107.3a). The
// sacrificed creature's mana value is a generic reduction of the total
// cost (CR 702.119a), and the total counts X at its announced value (CR
// 601.2f), so once the {X}{G} has no printed generic left the reduction
// comes off the mana announced for X (#2701). X on the stack stays 3.
func TestHerigastEmergeKeepsTheSpellsX(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	herigastOnBattlefield(g, me.ID)
	hydra := withHandPT(g, me, effectstest.HandCardOf(g, me, "Hydra", "Creature — Hydra", "{X}{G}", ""), 0, 0)
	spawn := effectstest.ApaPush(g, me.ID, me.ID, effects.EldraziSpawnToken())
	two := effectstest.EmergeFodder(g, me.ID, "Two", "{1}{G}", 2, 2)
	if got := grantedEmergeSettled(t, g, me, hydra, spawn, 3); got != "{3}{G}" {
		t.Errorf("X = 3 over a token: %s, want {3}{G}", got)
	}
	if got := grantedEmergeSettled(t, g, me, hydra, two, 3); got != "{1}{G}" {
		t.Errorf("X = 3 over a two-drop: %s, want {1}{G}", got)
	}
	if got := grantedEmergeSettled(t, g, me, hydra, two, 1); got != "{G}" {
		t.Errorf("X = 1 over a two-drop: %s, want {G} (never below the X)", got)
	}
	// Two lands pay {1}{G}: the four-mana cast is paid with two.
	effectstest.EmergeLands(g, me.ID, 2, "Forest")
	if err := castGrantedEmerge(g, me, hydra, two, true, 3); err != nil {
		t.Fatalf("granted emerge with X = 3 over a two-drop, two Forests: %v", err)
	}
	var x int
	g.ReadSnapshot(func() {
		if item := g.StackMeta[hydra]; item != nil {
			x = item.XValue
		}
	})
	if x != 3 {
		t.Errorf("X on the stack = %d, want 3", x)
	}
}

// grantedEmergeSettled is grantedEmergePrice with X settled at the
// announced x, which is what the payment charges.
func grantedEmergeSettled(t *testing.T, g *game.Game, p *game.Player, card, sac uuid.UUID, x int) string {
	t.Helper()
	return grantedEmergeTotal(t, g, p, card, sac, x).SettleX(x).String()
}

// A creature that prints emerge keeps its own beside Herigast's (the
// ruling: "you may choose to pay it for any one of those emerge costs").
func TestHerigastGivesASecondEmergeToACreatureWithOne(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	herigastOnBattlefield(g, me.ID)
	gryff := effectstest.HandCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", effectstest.WretchedGryffOracle)
	two := effectstest.EmergeFodder(g, me.ID, "Two", "{2}", 2, 2)
	offers := effectstest.HandOffers(t, g, me, gryff)
	if _, ok := offers[effects.AltCostKeyEmerge]; !ok {
		t.Error("the Gryff lost its own emerge")
	}
	if o, ok := offers[effects.GrantedAltCostEmerge]; !ok || o.Label != "Emerge {7} (Herigast, Erupting Nullkite)" {
		t.Errorf("granted offer %+v, want Emerge {7}", o)
	}
	if got := grantedEmergePrice(t, g, me, gryff, two, 0); got != "{5}" {
		t.Errorf("the granted emerge over a two-drop: %s, want {5}", got)
	}
	if got := effectstest.EmergePrice(t, g, me, gryff, two); got != "{3}{U}" {
		t.Errorf("the printed emerge over a two-drop: %s, want {3}{U}", got)
	}
}

// Adipose Offspring's "if this creature's emerge cost was paid" counts the
// emerge Herigast gave it.
func TestAdiposeOffspringCountsHerigastsEmerge(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	herigastOnBattlefield(g, me.ID)
	wall := effectstest.EmergeFodder(g, me.ID, "Wall", "{3}", 0, 3)
	off := effectstest.HandCardOf(g, me, "Adipose Offspring", "Creature — Alien", "{3}{W}", effectstest.AdiposeOffspringOracle)
	if err := castGrantedEmerge(g, me, off, wall, false, 0); err != nil {
		t.Fatalf("granted emerge: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Alien"); n != 3 {
		t.Errorf("%d Aliens, want 3 (the sacrificed Wall's toughness)", n)
	}
}

// The bot is offered the granted emerge, priced per payment (owner
// decision 5): only the creatures that make it affordable.
func TestHerigastEmergeIsEnumerated(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	herigastOnBattlefield(g, me.ID)
	bear := herigastBear(g, me)
	effectstest.EmergeLands(g, me.ID, 2, "Forest")
	two := effectstest.EmergeFodder(g, me.ID, "Two", "{2}", 2, 2)
	four := effectstest.EmergeFodder(g, me.ID, "Four", "{4}", 4, 4)
	offered := map[uuid.UUID]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != bear {
			continue
		}
		var p struct {
			AlternativeCost string      `json:"alternative_cost"`
			AltCostIDs      []uuid.UUID `json:"alt_cost_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.AlternativeCost != effects.GrantedAltCostEmerge || len(p.AltCostIDs) != 1 {
			continue
		}
		offered[p.AltCostIDs[0]] = true
	}
	if !offered[four] {
		t.Errorf("offered %v, want the four-drop", offered)
	}
	if offered[two] {
		t.Errorf("offered %v, which includes a payment two Forests can't afford", offered)
	}
}

// "When you cast this spell, you may exile your hand. If you do, draw
// three cards." Yes exiles the hand and draws three, even from an empty
// hand; No keeps it.
func TestHerigastCastTriggerExilesTheHandAndDrawsThree(t *testing.T) {
	for _, c := range []struct {
		name   string
		hand   int
		accept bool
		want   int
	}{
		{"yes", 2, true, 3},
		{"yes with an empty hand", 0, true, 3},
		{"no", 2, false, 2},
	} {
		t.Run(c.name, func(t *testing.T) {
			g, me, _ := effectstest.EmergeTable(t)
			herigast := withHandPT(g, me, effectstest.HandCardOf(g, me, "Herigast, Erupting Nullkite", "Legendary Creature — Eldrazi Dragon", "{9}", herigastOracle), 6, 6)
			var kept []uuid.UUID
			for i := 0; i < c.hand; i++ {
				kept = append(kept, effectstest.HandCardOf(g, me, "Filler", "Sorcery", "{1}", ""))
			}
			if err := g.CastSpell(me.ID, herigast, game.CastSpellParams{}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			choice := effectstest.PassUntilConfirmFor(t, g, me.ID)
			if choice == nil {
				t.Fatal("the cast trigger asked nothing")
			}
			if err := g.ResolveConfirm(choice.ID, me.ID, c.accept); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			if me.Hand.Size() != c.want {
				t.Errorf("hand %d, want %d", me.Hand.Size(), c.want)
			}
			for _, id := range kept {
				if c.accept != g.Exile.Contains(id) {
					t.Errorf("a hand card exiled = %v, want %v", g.Exile.Contains(id), c.accept)
				}
			}
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.FindBattlefieldCardForTest(g, herigast) == nil {
				t.Error("Herigast did not resolve")
			}
		})
	}
}

// effects.Register holds a granted offer to a price and a label, plus
// emerge's one-creature sacrifice; a priced-at-mana-cost grant names no
// mana cost of its own.
func TestRegisterRefusesAMisshapenGrantedEmerge(t *testing.T) {
	withCost := effects.EmergeForCreatureSpellsYouCast()
	withCost.Offer.ManaCost = "{2}"
	two := effects.EmergeForCreatureSpellsYouCast()
	two.Offer.Sacrifice = two.Offer.Sacrifice.WithCount(2, 2)
	for name, c := range map[string]struct {
		gr   game.GrantedAlternativeCost
		want string
	}{
		"a mana cost": {withCost, "priced at the spell's mana cost"},
		"two":         {two, "CR 702.119a"},
	} {
		msg := effectstest.RegisterPanics(effects.Spec{OracleID: "test-0135-granted-emerge-" + name, Name: "Test " + name,
			GrantedAlternativeCosts: []game.GrantedAlternativeCost{c.gr}})
		if !strings.Contains(msg, c.want) {
			t.Errorf("%s: panic %q, want %q", name, msg, c.want)
		}
	}
}

// herigastBear puts a vanilla 6/6 creature card costing {4}{G}{G} in
// `p`'s hand.
func herigastBear(g *game.Game, p *game.Player) uuid.UUID {
	return withHandPT(g, p, effectstest.HandCardOf(g, p, "Big Bear", "Creature — Bear", "{4}{G}{G}", ""), 6, 6)
}

// withHandPT gives the hand card `id` its printed power and toughness.
func withHandPT(g *game.Game, p *game.Player, id uuid.UUID, power, toughness int) uuid.UUID {
	g.WithWriteLock(func() {
		for i := range p.Hand.Cards {
			if p.Hand.Cards[i].InstanceID == id {
				p.Hand.Cards[i].Power, p.Hand.Cards[i].Toughness = power, toughness
			}
		}
	})
	return id
}
