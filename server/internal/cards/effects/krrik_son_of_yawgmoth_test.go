package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// krrik_son_of_yawgmoth_test.go — ADR 0131 (#2531), PRs 1 and 2: the third
// ability, "for each {B} in a cost, you may pay 2 life rather than pay
// that mana", for casts, activations, attack taxes and a mana ability, on the real
// catalog card. The engine's cases are game/life_for_mana_test.go.

const (
	krrikOracle = "cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd"
)

// krrikTable is a four-seat catalog game on the active seat's main
// phase with K'rrik on that seat's side, a creature to kill on the next
// seat's, and `life` on the active seat.
func krrikTable(t *testing.T) (g *game.Game, me *game.Player, krrik, victim uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	krrik = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "K'rrik, Son of Yawgmoth", OracleID: krrikOracle,
		TypeLine: "Legendary Creature — Phyrexian Horror Minion", ManaCost: "{4}{B/P}{B/P}{B/P}",
		Colors: []string{"B"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim = wpCreature(g, opp.ID, "Three Drop", "{2}{B}")
	return g, me, krrik, victim
}

// krrikHandSpell puts a catalog card with a printed cost into the active
// seat's hand.
func krrikHandSpell(me *game.Player, name, typeLine, oracle, cost string, colors ...string) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Colors: colors, Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	})
	return id
}

// The declaration is on the card: the catalog serves the static for
// K'rrik's key, and for nothing else.
func TestKrrikDeclaresTheLifeForManaStatic(t *testing.T) {
	got := game.CatalogLifeForMana(krrikOracle)
	if len(got) != 1 || got[0].Color != "B" {
		t.Fatalf("CatalogLifeForMana(K'rrik) = %+v, want one {B} grant", got)
	}
	if other := game.CatalogLifeForMana(deceitOracle); len(other) != 0 {
		t.Errorf("an unrelated card serves %+v", other)
	}
}

// Bloodchief's Thirst ({B}) cast off an empty pool: the {B} is paid
// with 2 life, the spell resolves, and K'rrik's own "whenever you cast a
// black spell" still sees a black spell (the colour and mana value do
// not change with the payment).
func TestKrrikPaysForABlackSpellWithLifeAndStillTriggers(t *testing.T) {
	g, me, krrik, _ := krrikTable(t)
	spell := krrikHandSpell(me, "Bloodchief's Thirst", "Sorcery", bloodchiefsThirstOracle, "{B}", "B")
	life := me.Life
	// The unkicked clause names mana value 2 or less, so aim at a
	// cheaper creature than the table's three-drop.
	cheap := wpCreature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Two Drop", "{1}{B}")

	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: twTarget(cheap)})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("{B} with an empty pool and no claim: %v, want insufficient mana", err)
	}
	if me.Life != life {
		t.Fatalf("the refused cast cost %d life", life-me.Life)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: twTarget(cheap), PhyrexianLife: 1}); err != nil {
		t.Fatalf("{B} paid with 2 life: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d", me.Life, life-2)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) {
		t.Error("the spell paid with life did not resolve")
	}
	if got := counterOn(g, krrik, "+1/+1"); got != 1 {
		t.Errorf("K'rrik has %d +1/+1 counters, want 1 — the black spell paid with life is still black", got)
	}
}

// The kicker's {B}s are costs too (the 2024 ruling: every mana payment):
// {B} + kicker {2}{B}, with two colourless in the pool, takes both
// {B} as 4 life.
func TestKrrikPaysAKickersBlackSymbolWithLife(t *testing.T) {
	g, me, _, victim := krrikTable(t)
	spell := krrikHandSpell(me, "Bloodchief's Thirst", "Sorcery", bloodchiefsThirstOracle, "{B}", "B")
	life := me.Life
	me.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, Targets: twTarget(victim), OptionalCosts: []int{0}, PhyrexianLife: 2,
	}); err != nil {
		t.Fatalf("kicked {B}+{2}{B} off two colourless and 4 life: %v", err)
	}
	if me.Life != life-4 {
		t.Errorf("life = %d, want %d", me.Life, life-4)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the kicked spell did not destroy the three-drop")
	}
}

// An alternative cost's {B} half is covered as well: Deceit's evoke
// {U/B}{U/B} is two hybrid symbols, each with a {B} half. K'rrik combines
// with it (it changes how the cost is paid, not which cost, CR 118.9a).
func TestKrrikPaysAnEvokeCostsBlackHalvesWithLife(t *testing.T) {
	g, me, _, _ := krrikTable(t)
	spell := krrikHandSpell(me, "Deceit", "Creature — Elemental Incarnation", deceitOracle, "{4}{U/B}{U/B}", "U", "B")
	life := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		Strict: true, AlternativeCost: "evoke", PhyrexianLife: 2,
	}); err != nil {
		t.Fatalf("evoke {U/B}{U/B} paid with 4 life: %v", err)
	}
	if me.Life != life-4 {
		t.Errorf("life = %d, want %d", me.Life, life-4)
	}
	if !g.Stack.Contains(spell) {
		t.Error("the evoked spell is not on the stack")
	}
}

// Without K'rrik on the battlefield the same {B} refuses the claim.
func TestWithoutKrrikABlackSymbolRefusesLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cheap := wpCreature(g, opp.ID, "Two Drop", "{1}{B}")
	spell := krrikHandSpell(me, "Bloodchief's Thirst", "Sorcery", bloodchiefsThirstOracle, "{B}", "B")
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: twTarget(cheap), PhyrexianLife: 1})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("claiming life for {B} with no K'rrik: %v, want ErrInvalidParam", err)
	}
}

// PR 2 closed the last gap: no caveat is left, and the card says it is
// complete.
func TestKrrikIsCompleteWithNoCaveat(t *testing.T) {
	spec, ok := Lookup(krrikOracle)
	if !ok {
		t.Fatal("K'rrik is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("completeness %v, caveats %v; want Full with none", spec.Completeness, spec.Caveats)
	}
}

// A mana ability's own cost (ADR 0131 §2): Fetid Heath's "{W/B}, {T}: Add
// {W}{W}, {W}{B}, or {B}{B}" is paid with 2 life for the {B} half of its
// hybrid symbol, with the real catalog card and the real K'rrik.
func TestKrrikPaysFetidHeathsFilterCostWithLife(t *testing.T) {
	g, me, _, _ := krrikTable(t)
	heath := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fetid Heath", OracleID: "42bf259d-4bb9-49c3-b4ec-223dca62f4d6",
		TypeLine: "Land", Owner: me.ID, Controller: me.ID,
	})
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, heath, 1, game.ManaAbilityParams{
		Colors: []string{"B", "B"}, PhyrexianLife: 1,
	}); err != nil {
		t.Fatalf("Fetid Heath's filter paid with 2 life: %v", err)
	}
	if me.Life != life-game.PhyrexianLifePerSymbol {
		t.Errorf("life = %d, want %d", me.Life, life-game.PhyrexianLifePerSymbol)
	}
	black := 0
	for _, tok := range me.ManaPool {
		if tok.Color == "B" {
			black++
		}
	}
	if black != 2 {
		t.Errorf("pool = %v, want {B}{B}", me.ManaPool)
	}
}

// Without a claim the filter still needs its mana — the auto-tapper never
// pays life (ADR 0131 §3).
func TestKrrikFetidHeathWithoutAClaimNeedsTheMana(t *testing.T) {
	g, me, _, _ := krrikTable(t)
	heath := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fetid Heath", OracleID: "42bf259d-4bb9-49c3-b4ec-223dca62f4d6",
		TypeLine: "Land", Owner: me.ID, Controller: me.ID,
	})
	life := me.Life
	err := g.ActivateManaAbility(me.ID, heath, 1, game.ManaAbilityParams{Colors: []string{"B", "B"}, AutoTap: true})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("filter with nothing to pay it: %v, want insufficient mana", err)
	}
	if me.Life != life {
		t.Errorf("a refused activation cost %d life", life-me.Life)
	}
}
