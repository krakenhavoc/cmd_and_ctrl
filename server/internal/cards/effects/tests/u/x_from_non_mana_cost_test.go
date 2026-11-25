package testsu

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// x_from_non_mana_cost_test.go — #6122: X fixed by a non-mana cost, or
// read by a target. Three shapes, each driven through the real cast or
// activation path on a test-registered spec:
//
//   - "As an additional cost to cast this spell, return X Swamps you
//     control to their owner's hand" (Infernal Harvest): the count is the
//     announced X (CR 107.3a, 601.2b).
//   - "Flashback—{3}{U}{R}, Exile X cards from your graveyard" (Summons of
//     Saruman): an alternative cost that includes X (CR 107.3a/b,
//     702.34a), the cards exiled being the X.
//   - "{X}, {T}: … target creature with power X" (Ent-Draught Basin): X is
//     announced before targets (CR 601.2b-c via 602.2b).

const (
	xncReturnOracle    = "00000000-6122-4000-8000-000000000001"
	xncFlashbackOracle = "00000000-6122-4000-8000-000000000002"
	xncBasinOracle     = "00000000-6122-4000-8000-000000000003"
)

// gainX is the test spells' effect: "You gain X life", so the X the
// stack item carries is visible on the life total.
func gainX(_ *game.StackItem, ctx *effects.Context) error {
	return effects.GainLife{Amount: ctx.X()}.Apply(ctx)
}

func registerXncReturn(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:       xncReturnOracle,
		Name:           "Test Harvest",
		Completeness:   effects.CompletenessFull,
		XMatters:       true,
		AdditionalCost: effects.ReturnXToHandCost("X Swamps you control", effects.HasSubtype("Swamp")),
		OnResolve:      gainX,
	})
}

func xncHand(g *game.Game, me *game.Player, name, oracle, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, OracleID: oracle, TypeLine: "Sorcery",
			ManaCost: cost, Owner: me.ID, Controller: me.ID})
	})
	return id
}

func xncLand(g *game.Game, me *game.Player, typeLine string) uuid.UUID {
	return effectstest.PushPermanent(g, me.ID, game.Card{Name: typeLine, TypeLine: typeLine})
}

func xncInHand(me *game.Player, id uuid.UUID) bool {
	for _, c := range me.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

// X = 2 returns exactly the two named Swamps to their owner's hand as the
// cost is paid, and the spell reads X = 2.
func TestReturnXSwampsReturnsXAndTheSpellReadsX(t *testing.T) {
	registerXncReturn(t)
	g, me, _ := effectstest.P7Table(t)
	a, b := xncLand(g, me, "Basic Land — Swamp"), xncLand(g, me, "Basic Land — Swamp")
	keep := xncLand(g, me, "Basic Land — Swamp")
	spell := xncHand(g, me, "Test Harvest", xncReturnOracle, "{B}")
	effectstest.FloatMana(t, g, me, "{B}")
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, ReturnIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("cast X=2: %v", err)
	}
	if !xncInHand(me, a) || !xncInHand(me, b) || xncInHand(me, keep) {
		t.Fatal("the two named Swamps were not returned, or the third was")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("life %d -> %d, want +2 (X = 2)", life, me.Life)
	}
}

// The count must be the announced X, each a Swamp the caster controls;
// a refused cast returns nothing. X = 0 returns nothing and is legal.
func TestReturnXSwampsRefusesAMismatchAndAllowsZero(t *testing.T) {
	registerXncReturn(t)
	g, me, opp := effectstest.P7Table(t)
	a, b := xncLand(g, me, "Basic Land — Swamp"), xncLand(g, me, "Basic Land — Swamp")
	forest := xncLand(g, me, "Basic Land — Forest")
	theirs := xncLand(g, opp, "Basic Land — Swamp")
	spell := xncHand(g, me, "Test Harvest", xncReturnOracle, "{B}")
	effectstest.FloatMana(t, g, me, "{B}")
	for name, p := range map[string]game.CastSpellParams{
		"one Swamp for X=2":       {XValue: 2, ReturnIDs: []uuid.UUID{a}},
		"two Swamps for X=1":      {XValue: 1, ReturnIDs: []uuid.UUID{a, b}},
		"a Forest":                {XValue: 1, ReturnIDs: []uuid.UUID{forest}},
		"an opponent's Swamp":     {XValue: 1, ReturnIDs: []uuid.UUID{theirs}},
		"the same Swamp twice":    {XValue: 2, ReturnIDs: []uuid.UUID{a, a}},
		"more Swamps than it has": {XValue: 3, ReturnIDs: []uuid.UUID{a, b}},
	} {
		p.Strict = true
		if err := g.CastSpell(me.ID, spell, p); err == nil {
			t.Fatalf("%s: cast accepted", name)
		}
		if !xncInHand(me, spell) || xncInHand(me, a) || xncInHand(me, b) {
			t.Fatalf("%s: a refused cast moved something", name)
		}
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 0}); err != nil {
		t.Fatalf("X=0 with nothing returned: %v", err)
	}
}

func registerXncFlashback(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:         xncFlashbackOracle,
		Name:             "Test Summons",
		Completeness:     effects.CompletenessFull,
		XMatters:         true,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.FlashbackExileX("{U}")},
		OnResolve:        gainX,
	})
}

func xncGraveyard(g *game.Game, me *game.Player, name, oracle, cost string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, OracleID: oracle, TypeLine: "Sorcery",
			ManaCost: cost, Owner: me.ID, Controller: me.ID})
	})
	return id
}

// Flashback with X = 2: exactly two other graveyard cards are exiled, the
// flashback mana has no {X} and is all that is paid, and the spell reads
// X = 2 (CR 107.3a: an alternative cost that includes X is not locked at
// 0, CR 107.3b).
func TestFlashbackExileXCardsSetsX(t *testing.T) {
	registerXncFlashback(t)
	g, me, _ := effectstest.P7Table(t)
	a := xncGraveyard(g, me, "Fuel A", "", "{1}")
	b := xncGraveyard(g, me, "Fuel B", "", "{1}")
	c := xncGraveyard(g, me, "Fuel C", "", "{1}")
	spell := xncGraveyard(g, me, "Test Summons", xncFlashbackOracle, "{X}{U}{R}")
	effectstest.FloatMana(t, g, me, "{U}")
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "flashback", XValue: 2, AltCostIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("flashback X=2: %v", err)
	}
	if me.Graveyard.Contains(a) || me.Graveyard.Contains(b) || !me.Graveyard.Contains(c) {
		t.Fatal("the two named cards were not exiled, or the third was")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("life %d -> %d, want +2 (X = 2)", life, me.Life)
	}
	if me.Graveyard.Contains(spell) {
		t.Fatal("a flashed-back spell went to the graveyard instead of exile (CR 702.34a)")
	}
}

// The number of cards must be the announced X; X = 0 exiles nothing.
func TestFlashbackExileXRefusesAMismatchAndAllowsZero(t *testing.T) {
	registerXncFlashback(t)
	g, me, _ := effectstest.P7Table(t)
	a := xncGraveyard(g, me, "Fuel A", "", "{1}")
	spell := xncGraveyard(g, me, "Test Summons", xncFlashbackOracle, "{X}{U}{R}")
	effectstest.FloatMana(t, g, me, "{U}")
	for name, p := range map[string]game.CastSpellParams{
		"one card for X=2": {XValue: 2, AltCostIDs: []uuid.UUID{a}},
		"one card for X=0": {XValue: 0, AltCostIDs: []uuid.UUID{a}},
		"the spell itself": {XValue: 1, AltCostIDs: []uuid.UUID{spell}},
	} {
		p.Strict, p.FromZone, p.AlternativeCost = true, "graveyard", "flashback"
		if err := g.CastSpell(me.ID, spell, p); err == nil {
			t.Fatalf("%s: cast accepted", name)
		}
		if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(spell) {
			t.Fatalf("%s: a refused cast moved something", name)
		}
	}
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback X=0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || !me.Graveyard.Contains(a) {
		t.Fatal("X = 0 gained life or exiled a card")
	}
}

func registerXncBasin(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:     xncBasinOracle,
		Name:         "Test Basin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{X}, {T}: Put a +1/+1 counter on target creature with power X. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{X}"), effects.TapCost()),
			Targets:      effects.TargetCreature("target creature with power X").WithPowerEqualsX(),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutPlusOneCounterOnTheTarget(effects.NewContext(g, item))
			},
		}},
	})
}

// "Target creature with power X": X = 2 admits a 2/2 and refuses a 3/3
// (an exact match), at announce; the 2/2 gets the counter.
func TestPowerXTargetMatchesTheAnnouncedXExactly(t *testing.T) {
	registerXncBasin(t)
	g, me, _ := effectstest.P7Table(t)
	basin := effectstest.PushCatalogPermanent(g, me.ID, "Test Basin", "Artifact", xncBasinOracle, false)
	two := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	three := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3})
	effectstest.FloatMana(t, g, me, "{C}{C}")
	err := g.ActivateCatalogAbility(me.ID, basin, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: effectstest.B16TargetCard(three)})
	if !errors.Is(err, game.ErrIllegalTarget) && !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("X=2 at a 3/3: err %v, want an illegal-target refusal", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, basin, 0, game.ActivateAbilityParams{Strict: true, XValue: 2, Targets: effectstest.B16TargetCard(two)}); err != nil {
		t.Fatalf("X=2 at a 2/2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, two); c.Counters[game.CounterPlusOne] != 1 {
		t.Fatalf("the 2/2 has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
	}
}

// xncCastMoves is every cast move the enumerator offers for `card`, with
// the fields these tests read.
type xncCastMove struct {
	XValue     int      `json:"x_value"`
	ReturnIDs  []string `json:"return_ids"`
	AltCostIDs []string `json:"alt_cost_ids"`
	AltCost    string   `json:"alternative_cost"`
}

func xncCastMoves(t *testing.T, g *game.Game, seat, card uuid.UUID) []xncCastMove {
	t.Helper()
	var out []xncCastMove
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeCastSpell || m.Source != card {
			continue
		}
		var p xncCastMove
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	return out
}

// The enumerator's "return X" and "exile X cards" moves name exactly X
// permanents or cards, so each is one the engine accepts (#544).
func TestXFromNonManaCostMovesNameExactlyX(t *testing.T) {
	registerXncReturn(t)
	registerXncFlashback(t)
	g, me, _ := effectstest.P7Table(t)
	xncLand(g, me, "Basic Land — Swamp")
	xncLand(g, me, "Basic Land — Swamp")
	harvest := xncHand(g, me, "Test Harvest", xncReturnOracle, "{B}")
	xncGraveyard(g, me, "Fuel A", "", "{1}")
	xncGraveyard(g, me, "Fuel B", "", "{1}")
	summons := xncGraveyard(g, me, "Test Summons", xncFlashbackOracle, "{X}{U}{R}")
	effectstest.FloatMana(t, g, me, "{B}{U}")
	moves := xncCastMoves(t, g, me.ID, harvest)
	if len(moves) == 0 {
		t.Fatal("no cast move for the return-X spell")
	}
	for _, m := range moves {
		if len(m.ReturnIDs) != m.XValue {
			t.Errorf("return-X move announces X=%d but returns %d", m.XValue, len(m.ReturnIDs))
		}
	}
	flash := 0
	for _, m := range xncCastMoves(t, g, me.ID, summons) {
		if m.AltCost != "flashback" {
			continue
		}
		flash++
		if len(m.AltCostIDs) != m.XValue {
			t.Errorf("flashback move announces X=%d but exiles %d", m.XValue, len(m.AltCostIDs))
		}
	}
	if flash == 0 {
		t.Fatal("no flashback move for the exile-X spell")
	}
}
