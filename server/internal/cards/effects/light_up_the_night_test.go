package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// light_up_the_night_test.go — #5679: a flashback cost that removes X
// counters and sets the spell's X, and Light Up the Night, the card that
// waited on it.

const lightUpTheNightOracle = "4669926b-93f9-4d80-aee3-492fbf8e0a0a"

// lutnTable is a main phase on seat 0's turn with Light Up the Night in
// seat 0's graveyard and its flashback {3}{R} floating.
func lutnTable(t *testing.T) (g *game.Game, me, opp *game.Player, spell uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	spell = graveyardCardOf(g, me, "Light Up the Night", "Sorcery", "{X}{R}", lightUpTheNightOracle)
	payMana(t, g, me, "{C}{C}{C}{R}")
	return g, me, opp, spell
}

func lutnWalker(g *game.Game, owner uuid.UUID, name string, loyalty int) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{
		Name: name, TypeLine: "Legendary Planeswalker — Test", Counters: map[string]int{game.CounterLoyalty: loyalty},
	})
}

func lutnLoyalty(g *game.Game, id uuid.UUID) int {
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		return -1
	}
	return c.Counters[game.CounterLoyalty]
}

func lutnPlayer(p *game.Player) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: p.ID}}
}

func lutnCard(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// lutnFlashback casts the spell out of the graveyard for its flashback
// cost, announcing `x` and removing counts[i] loyalty counters off ids[i].
func lutnFlashback(g *game.Game, me *game.Player, spell uuid.UUID, x int, targets []game.TargetRef, ids []uuid.UUID, counts []int) error {
	return g.CastSpell(me.ID, spell, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Strict: true, XValue: x, Targets: targets,
		CounterSourceIDs: ids, CounterCounts: counts,
	})
}

// lutnHandCast casts the spell from hand for {X}{R} with X announced.
func lutnHandCast(t *testing.T, x int, target func(g *game.Game, me, opp *game.Player) []game.TargetRef) (*game.Game, *game.Player, *game.Player, []game.TargetRef) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Light Up the Night", TypeLine: "Sorcery",
		ManaCost: "{X}{R}", OracleID: lightUpTheNightOracle,
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	payMana(t, g, me, "{R}")
	for i := 0; i < x; i++ {
		payMana(t, g, me, "{C}")
	}
	targets := target(g, me, opp)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: x, Targets: targets}); err != nil {
		t.Fatalf("hand cast for X=%d: %v", x, err)
	}
	passPriorityAroundTable(t, g)
	if !inGraveyard(me, spell) {
		t.Fatal("a hand-cast Light Up the Night did not go to the graveyard")
	}
	return g, me, opp, targets
}

func TestLightUpTheNightDeclaresItsFlashback(t *testing.T) {
	spec, ok := Lookup(lightUpTheNightOracle)
	if !ok {
		t.Fatal("Light Up the Night is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Light Up the Night: %s %v, want Full with no caveats", spec.Completeness, spec.Caveats)
	}
	offers := game.AlternativeCostsOfferedFromZone(lightUpTheNightOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "flashback" {
		t.Fatalf("graveyard offers = %+v, want the one flashback", offers)
	}
	o := offers[0]
	if o.ManaCost != "{3}{R}" || !o.ExileOnLeavingStack || o.PaysCards() {
		t.Errorf("the flashback = %+v, want {3}{R}, exiled on leaving the stack, no card named", o)
	}
	rc := o.RemoveCounters
	if rc == nil || !rc.Variable || rc.N != 1 || !rc.Among || rc.Counter != game.CounterLoyalty || rc.From == nil || rc.OnePermanent {
		t.Fatalf("the flashback's counter removal = %+v, want X loyalty counters (at least 1) from among planeswalkers", rc)
	}
	if !o.CountersDefineX() || !o.PaymentDefinesX() {
		t.Error("the flashback is not an alternative cost that includes X")
	}
	if (game.CastCost{Printed: "{X}{R}", Paid: o.ManaCost, AltIncludesX: o.PaymentDefinesX()}).LocksXAtZero() {
		t.Error("CR 107.3b locks X at 0 under the flashback, which includes X")
	}
	if got := game.AlternativeCostsOfferedFromZone(lightUpTheNightOracle, game.ZoneHand); len(got) != 0 {
		t.Errorf("offers from hand = %+v, want none", got)
	}
}

// From hand it is {X}{R}: X damage to a player.
func TestLightUpTheNightDealsXToAPlayer(t *testing.T) {
	_, _, opp, _ := lutnHandCast(t, 3, func(_ *game.Game, _, opp *game.Player) []game.TargetRef { return lutnPlayer(opp) })
	if opp.Life != 37 {
		t.Errorf("opponent life = %d, want 37 (X = 3 to a player)", opp.Life)
	}
}

// X plus 1 to a creature: X = 3 kills a 4-toughness creature and leaves a
// 5-toughness one alive with 4 damage.
func TestLightUpTheNightDealsXPlusOneToACreature(t *testing.T) {
	g, _, _, targets := lutnHandCast(t, 3, func(g *game.Game, _, opp *game.Player) []game.TargetRef {
		return lutnCard(apaPush(g, opp.ID, opp.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4}))
	})
	if findBattlefieldCardForTest(g, targets[0].ID) != nil {
		t.Error("a 4-toughness creature survived X = 3 (4 damage)")
	}
	g, _, _, targets = lutnHandCast(t, 3, func(g *game.Game, _, opp *game.Player) []game.TargetRef {
		return lutnCard(apaPush(g, opp.ID, opp.ID, game.Card{Name: "Giant", TypeLine: "Creature — Giant", Power: 5, Toughness: 5}))
	})
	c := findBattlefieldCardForTest(g, targets[0].ID)
	if c == nil || c.DamageMarked != 4 {
		t.Errorf("the 5-toughness creature = %+v, want alive with 4 damage", c)
	}
}

// X plus 1 to a planeswalker, and X = 0 is a legal hand cast that still
// deals 1 to one.
func TestLightUpTheNightDealsXPlusOneToAPlaneswalker(t *testing.T) {
	g, _, _, targets := lutnHandCast(t, 2, func(g *game.Game, _, opp *game.Player) []game.TargetRef {
		return lutnCard(lutnWalker(g, opp.ID, "Their Walker", 6))
	})
	if got := lutnLoyalty(g, targets[0].ID); got != 3 {
		t.Errorf("their planeswalker has %d loyalty, want 3 (X = 2, 3 damage)", got)
	}
	g, _, _, targets = lutnHandCast(t, 0, func(g *game.Game, _, opp *game.Player) []game.TargetRef {
		return lutnCard(lutnWalker(g, opp.ID, "Their Walker", 6))
	})
	if got := lutnLoyalty(g, targets[0].ID); got != 5 {
		t.Errorf("their planeswalker has %d loyalty, want 5 (X = 0, 1 damage)", got)
	}
}

// The flashback: {3}{R} and X loyalty counters split across two
// planeswalkers. The counters removed are the spell's X, the printed
// {X}{R} is not paid, and the card is exiled after it resolves.
func TestLightUpTheNightFlashbackRemovesXLoyaltyAndSetsX(t *testing.T) {
	g, me, opp, spell := lutnTable(t)
	a := lutnWalker(g, me.ID, "Walker A", 4)
	b := lutnWalker(g, me.ID, "Walker B", 3)
	giant := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Giant", TypeLine: "Creature — Giant", Power: 6, Toughness: 6})

	if err := lutnFlashback(g, me, spell, 3, lutnCard(giant), []uuid.UUID{a, b}, []int{2, 1}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	if lutnLoyalty(g, a) != 2 || lutnLoyalty(g, b) != 2 {
		t.Errorf("loyalty after paying = %d and %d, want 2 and 2", lutnLoyalty(g, a), lutnLoyalty(g, b))
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the {3}{R} was not paid: pool %v", me.ManaPool)
	}
	item := g.StackMeta[spell]
	if item == nil || item.XValue != 3 || item.Paid.CountersRemoved != 3 {
		t.Fatalf("stack item = %+v, want X = 3 and three counters removed", item)
	}
	passPriorityAroundTable(t, g)
	if c := findBattlefieldCardForTest(g, giant); c == nil || c.DamageMarked != 4 {
		t.Errorf("the creature = %+v, want 4 damage (X = 3, plus 1)", c)
	}
	if !inExile(g, spell) || inGraveyard(me, spell) {
		t.Error("a flashed-back Light Up the Night was not exiled")
	}
}

// A planeswalker left with no loyalty is put into the graveyard with the
// spell on the stack; the spell still resolves for the X announced.
func TestLightUpTheNightFlashbackMaySpendAPlaneswalker(t *testing.T) {
	g, me, opp, spell := lutnTable(t)
	a := lutnWalker(g, me.ID, "Walker A", 5)
	if err := lutnFlashback(g, me, spell, 5, lutnPlayer(opp), []uuid.UUID{a}, []int{5}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	g.RunStateChecksForTest()
	if findBattlefieldCardForTest(g, a) != nil {
		t.Error("a planeswalker with no loyalty is still on the battlefield")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != 35 {
		t.Errorf("opponent life = %d, want 35 (X = 5 to a player)", opp.Life)
	}
}

// Every way to get the payment wrong is refused, with no counter removed,
// no mana spent and the card still in the graveyard.
func TestLightUpTheNightFlashbackStrictRefusals(t *testing.T) {
	cases := []struct {
		name string
		pay  func(g *game.Game, me, opp *game.Player, spell, a, b, theirs, rock, bear uuid.UUID) error
	}{
		{"X of 0 with no counters", func(g *game.Game, me, opp *game.Player, spell, _, _, _, _, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 0, lutnPlayer(opp), nil, nil)
		}},
		{"X of 0 naming a planeswalker for none", func(g *game.Game, me, opp *game.Player, spell, a, _, _, _, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 0, lutnPlayer(opp), []uuid.UUID{a}, []int{0})
		}},
		{"X announced with no counters removed", func(g *game.Game, me, opp *game.Player, spell, _, _, _, _, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 3, lutnPlayer(opp), nil, nil)
		}},
		{"fewer counters than X", func(g *game.Game, me, opp *game.Player, spell, a, _, _, _, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 3, lutnPlayer(opp), []uuid.UUID{a}, []int{2})
		}},
		{"more counters than X", func(g *game.Game, me, opp *game.Player, spell, a, b, _, _, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 3, lutnPlayer(opp), []uuid.UUID{a, b}, []int{3, 1})
		}},
		{"counters removed with no X announced", func(g *game.Game, me, opp *game.Player, spell, a, _, _, _, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 0, lutnPlayer(opp), []uuid.UUID{a}, []int{2})
		}},
		{"more than a planeswalker holds", func(g *game.Game, me, opp *game.Player, spell, a, _, _, _, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 5, lutnPlayer(opp), []uuid.UUID{a}, []int{5})
		}},
		{"an opponent's planeswalker", func(g *game.Game, me, opp *game.Player, spell, _, _, theirs, _, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 2, lutnPlayer(opp), []uuid.UUID{theirs}, []int{2})
		}},
		{"a permanent that is not a planeswalker", func(g *game.Game, me, opp *game.Player, spell, _, _, _, rock, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 2, lutnPlayer(opp), []uuid.UUID{rock}, []int{2})
		}},
		{"a counter that is not loyalty", func(g *game.Game, me, opp *game.Player, spell, _, _, _, _, bear uuid.UUID) error {
			return g.CastSpell(me.ID, spell, game.CastSpellParams{
				FromZone: "graveyard", AlternativeCost: "flashback", Strict: true, XValue: 2, Targets: lutnPlayer(opp),
				CounterSourceIDs: []uuid.UUID{bear}, CounterCounts: []int{2}, CounterKind: game.CounterPlusOne,
			})
		}},
		{"one planeswalker named twice", func(g *game.Game, me, opp *game.Player, spell, a, _, _, _, _ uuid.UUID) error {
			return lutnFlashback(g, me, spell, 2, lutnPlayer(opp), []uuid.UUID{a, a}, []int{1, 1})
		}},
		{"the printed cost from the graveyard", func(g *game.Game, me, opp *game.Player, spell, _, _, _, _, _ uuid.UUID) error {
			return g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "graveyard", Strict: true, XValue: 2, Targets: lutnPlayer(opp)})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp, spell := lutnTable(t)
			a := lutnWalker(g, me.ID, "Walker A", 4)
			b := lutnWalker(g, me.ID, "Walker B", 3)
			theirs := lutnWalker(g, opp.ID, "Their Walker", 6)
			rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Odd Rock", TypeLine: "Artifact", Counters: map[string]int{game.CounterLoyalty: 5}})
			bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
				Counters: map[string]int{game.CounterPlusOne: 3}})
			if err := tc.pay(g, me, opp, spell, a, b, theirs, rock, bear); err == nil {
				t.Fatal("the cast was accepted")
			}
			if !inGraveyard(me, spell) || g.Stack.Size() != 0 || len(me.ManaPool) == 0 {
				t.Fatal("a refused cast moved the card or spent mana")
			}
			if lutnLoyalty(g, a) != 4 || lutnLoyalty(g, b) != 3 || lutnLoyalty(g, theirs) != 6 || lutnLoyalty(g, rock) != 5 {
				t.Fatal("a refused cast removed counters")
			}
		})
	}
}

// A hand cast pays {X}{R} and no counters: a counter payment sent with it
// is a stray and refused.
func TestLightUpTheNightHandCastTakesNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	a := lutnWalker(g, me.ID, "Walker A", 4)
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Light Up the Night", TypeLine: "Sorcery",
		ManaCost: "{X}{R}", OracleID: lightUpTheNightOracle,
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	payMana(t, g, me, "{C}{C}{R}")
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, Targets: lutnPlayer(opp),
		CounterSourceIDs: []uuid.UUID{a}, CounterCounts: []int{2}})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("stray counter payment on a hand cast: %v, want ErrInvalidParam", err)
	}
	if lutnLoyalty(g, a) != 4 {
		t.Error("a refused hand cast removed loyalty")
	}
}

// With no loyalty counter to remove the flashback can't be paid (CR
// 118.3): the view and the enumerator withhold it.
func TestLightUpTheNightFlashbackIsWithheldWithNoLoyalty(t *testing.T) {
	g, me, opp, spell := lutnTable(t)
	lutnWalker(g, opp.ID, "Their Walker", 6)
	apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Counters: map[string]int{game.CounterPlusOne: 3}})
	v := graveyardViewOf(t, g, me.ID, spell)
	for _, o := range v.AlternativeCosts {
		if o.Key == "flashback" {
			t.Errorf("the view offers the flashback with no loyalty to remove: %+v", o)
		}
	}
	if moves := retraceMoves(g, me.ID, spell); len(moves) != 0 {
		t.Errorf("the enumerator offers %d graveyard casts with no loyalty to remove", len(moves))
	}
}

// The view ships the flashback with the counter picker's block, variable
// with a floor of 1 and a ceiling of every loyalty counter the caster's
// planeswalkers hold, and leaves X open.
func TestLightUpTheNightViewShipsTheCounterPicker(t *testing.T) {
	g, me, opp, spell := lutnTable(t)
	a := lutnWalker(g, me.ID, "Walker A", 4)
	b := lutnWalker(g, me.ID, "Walker B", 3)
	lutnWalker(g, opp.ID, "Their Walker", 6)

	v := graveyardViewOf(t, g, me.ID, spell)
	var found bool
	for _, o := range v.AlternativeCosts {
		if o.Key != "flashback" {
			continue
		}
		found = true
		if o.CounterCostN != 1 || !o.CounterCostVariable || !o.CounterCostAmong || o.CounterCostKind != game.CounterLoyalty ||
			o.CounterCostSelf || o.CounterCostLabel != "planeswalkers you control" || o.CounterCostMax != 7 {
			t.Errorf("counter block = %+v", o)
		}
		if o.ManaCost != "{3}{R}" || o.XLockedAtZero || o.PayOptions != nil {
			t.Errorf("the offer = %+v, want {3}{R}, X open and no card picker", o)
		}
		ids := map[string]int{}
		for _, opt := range o.CounterCostOptions {
			for _, k := range opt.Kinds {
				ids[opt.CardID] += k.Count
			}
		}
		if len(ids) != 2 || ids[a.String()] != 4 || ids[b.String()] != 3 {
			t.Errorf("counter options = %+v, want my two planeswalkers", o.CounterCostOptions)
		}
	}
	if !found {
		t.Fatalf("the graveyard view does not offer the flashback: %+v", v.AlternativeCosts)
	}
}

type lutnMove struct {
	FromZone         string   `json:"from_zone"`
	AlternativeCost  string   `json:"alternative_cost"`
	XValue           int      `json:"x_value"`
	CounterSourceIDs []string `json:"counter_source_ids"`
	CounterCounts    []int    `json:"counter_counts"`
	Targets          []struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
	} `json:"targets"`
}

func lutnBotMoves(t *testing.T, g *game.Game, seat, spell uuid.UUID) []lutnMove {
	t.Helper()
	var out []lutnMove
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeCastSpell || m.Source != spell {
			continue
		}
		var mp lutnMove
		if err := json.Unmarshal(m.Params, &mp); err != nil {
			t.Fatal(err)
		}
		out = append(out, mp)
	}
	return out
}

// The bot is offered the flashback with one payment that leaves a loyalty
// counter on each planeswalker, announcing the counters removed as X, and
// the engine accepts the move as written.
func TestLightUpTheNightBotPathKeepsItsPlaneswalkers(t *testing.T) {
	g, me, _, spell := lutnTable(t)
	a := lutnWalker(g, me.ID, "Walker A", 4)
	b := lutnWalker(g, me.ID, "Walker B", 3)
	lutnWalker(g, me.ID, "Walker C", 1)

	moves := lutnBotMoves(t, g, me.ID, spell)
	if len(moves) == 0 {
		t.Fatal("the enumerator offers no flashback")
	}
	for _, p := range moves {
		if p.FromZone != "graveyard" || p.AlternativeCost != "flashback" {
			t.Fatalf("move = %+v, want a flashback from the graveyard", p)
		}
		total := 0
		for _, n := range p.CounterCounts {
			total += n
		}
		if total != 5 || p.XValue != 5 || len(p.CounterSourceIDs) != 2 {
			t.Fatalf("move = %+v, want X = 5 from two planeswalkers", p)
		}
	}
	p := moves[0]
	params := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: p.AlternativeCost, Strict: true,
		XValue: p.XValue, CounterCounts: p.CounterCounts}
	for _, raw := range p.CounterSourceIDs {
		params.CounterSourceIDs = append(params.CounterSourceIDs, uuid.MustParse(raw))
	}
	for _, tr := range p.Targets {
		params.Targets = append(params.Targets, game.TargetRef{Kind: game.TargetRefKind(tr.Kind), ID: uuid.MustParse(tr.ID)})
	}
	if err := g.CastSpell(me.ID, spell, params); err != nil {
		t.Fatalf("the enumerated flashback was refused: %v", err)
	}
	if lutnLoyalty(g, a) != 1 || lutnLoyalty(g, b) != 1 {
		t.Errorf("loyalty left = %d and %d, want 1 and 1", lutnLoyalty(g, a), lutnLoyalty(g, b))
	}
}

// With only one loyalty counter in all, the one payment there is spends
// it: X = 1.
func TestLightUpTheNightBotPathWithOneLoyaltyCounter(t *testing.T) {
	g, me, _, spell := lutnTable(t)
	lutnWalker(g, me.ID, "Walker A", 1)
	moves := lutnBotMoves(t, g, me.ID, spell)
	if len(moves) == 0 {
		t.Fatal("the enumerator offers no flashback")
	}
	for _, p := range moves {
		if p.XValue != 1 || len(p.CounterCounts) != 1 || p.CounterCounts[0] != 1 {
			t.Fatalf("move = %+v, want X = 1 from the one counter", p)
		}
	}
}

// The X a flashback announced through its counters survives a restore
// with the spell on the stack: it resolves for X plus 1 and is exiled.
func TestLightUpTheNightFlashbackSurvivesARestore(t *testing.T) {
	g, me, opp, spell := lutnTable(t)
	a := lutnWalker(g, me.ID, "Walker A", 4)
	giant := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Giant", TypeLine: "Creature — Giant", Power: 6, Toughness: 6})
	if err := lutnFlashback(g, me, spell, 3, lutnCard(giant), []uuid.UUID{a}, []int{3}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	restored := restoreRoundTrip(t, g, true)
	if item := restored.StackMeta[spell]; item == nil || item.XValue != 3 || item.Paid.CountersRemoved != 3 {
		t.Fatalf("restored stack item = %+v, want X = 3 and three counters removed", item)
	}
	passPriorityAroundTable(t, restored)
	if c := findBattlefieldCardForTest(restored, giant); c == nil || c.DamageMarked != 4 {
		t.Errorf("restored: the creature = %+v, want 4 damage", c)
	}
	if !inExile(restored, spell) {
		t.Error("restored: the flashed-back spell was not exiled")
	}
	if lutnLoyalty(restored, a) != 1 {
		t.Errorf("restored: loyalty = %d, want 1", lutnLoyalty(restored, a))
	}
}
