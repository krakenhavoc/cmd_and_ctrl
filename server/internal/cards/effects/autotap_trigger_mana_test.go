package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// autotap_trigger_mana_test.go — ADR 0074's amendment of 2026-10-30
// (#4716): the auto-tap planner counts what a triggered mana ability adds
// when it taps a source, judged by the real card's own condition. The
// engine rules are pinned in game/autotap_trigger_mana_test.go on stubs;
// these are the printed cards, through the real catalog.

// castAutoTappedFromHand casts a spell of `cost` from hand under strict
// payment with auto-tap, the path the client's cast button takes.
func castAutoTappedFromHand(t *testing.T, g *game.Game, me *game.Player, name, cost string) error {
	t.Helper()
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Creature — Horror", ManaCost: cost,
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		})
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AutoTap: true, FromZone: "hand"})
}

// plansFor reports whether the auto-tapper finds a plan for `cost`, and
// the sources it would spend.
func plansFor(t *testing.T, g *game.Game, me *game.Player, cost string) ([]uuid.UUID, bool) {
	t.Helper()
	parsed, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost(%s): %v", cost, err)
	}
	return g.AutoTapForCost(me.ID, parsed, 0)
}

// The #4716 regression, on the reported board as far as the report shows
// it: Crypt Ghast, five Swamps and a Treasure, paying Veinripper's
// {3}{B}{B}{B} (Veinripper is not catalogued, so the cost is the input).
// Under Crypt Ghast each Swamp taps for {B}{B}, so three Swamps pay six
// mana. Before the amendment the planner priced a Swamp at one {B}, tapped
// all five and cracked the Treasure.
func TestCryptGhastAutoPayTapsThreeSwampsForSixMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Crypt Ghast", "Creature — Spirit", cryptGhastOracle, false)
	swamps, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, game.Card{
		Name: "Swamp", TypeLine: "Basic Land — Swamp",
	}, 5)
	if err != nil || len(swamps) != 5 {
		t.Fatalf("spawn Swamps: %v (%d ids)", err, len(swamps))
	}
	tmpl, ok := Tokens().TokenTemplate("Treasure")
	if !ok {
		t.Fatal("no Treasure template")
	}
	treasures, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, tmpl, 1)
	if err != nil || len(treasures) != 1 {
		t.Fatalf("spawn Treasure: %v (%d ids)", err, len(treasures))
	}

	if err := castAutoTappedFromHand(t, g, me, "Veinripper stand-in", "{3}{B}{B}{B}"); err != nil {
		t.Fatalf("auto-paid {3}{B}{B}{B} under Crypt Ghast: %v", err)
	}
	tapped := 0
	for _, id := range swamps {
		if b31Tapped(t, g, id) {
			tapped++
		}
	}
	if tapped != 3 {
		t.Errorf("%d Swamps tapped, want exactly 3 (two untapped): each makes {B}{B} under Crypt Ghast", tapped)
	}
	if !g.Battlefield.Contains(treasures[0]) {
		t.Error("the Treasure was cracked, but three Swamps already pay six mana")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana floating after the payment, want none: %v", n, poolColors(me))
	}
}

// Mirari's Wake adds "one mana of any type that land produced", so a dual
// makes two of the colour it was tapped for: {R}{R} off a Taiga alone, and
// never {R}{G}, which would be stronger than the card.
func TestMirarisWakeOnADualPaysTwoOfOneColourOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedPermanentWithOracle(g, me.ID, "Mirari's Wake", "Enchantment", g3MirarisWakeOracle)
	taiga := seedPermanentWithOracle(g, me.ID, "Taiga", "Land — Mountain Forest", cagedSunTaigaOracle)

	if plan, ok := plansFor(t, g, me, "{R}{G}"); ok {
		t.Errorf("{R}{G} planned off one Taiga under Mirari's Wake: %v", plan)
	}
	if plan, ok := plansFor(t, g, me, "{R}{R}"); !ok || len(plan) != 1 || plan[0] != taiga {
		t.Fatalf("plan for {R}{R} = %v (ok %v), want the Taiga alone", plan, ok)
	}
	if err := castAutoTappedFromHand(t, g, me, "Two Red", "{R}{R}"); err != nil {
		t.Fatalf("auto-paid {R}{R}: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana floating, want none: %v", n, poolColors(me))
	}
}

// Caged Sun adds one more only when the land made the chosen colour: a
// Taiga under red pays {R}{R} alone and not {G}{G}.
func TestCagedSunIsCountedOnlyForTheChosenColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pushChosenColorPermanent(t, g, me.ID, "Caged Sun", "Artifact", cagedSunOracle, "R")
	taiga := seedPermanentWithOracle(g, me.ID, "Taiga", "Land — Mountain Forest", cagedSunTaigaOracle)

	if plan, ok := plansFor(t, g, me, "{G}{G}"); ok {
		t.Errorf("{G}{G} planned off one Taiga under a red Caged Sun: %v", plan)
	}
	if plan, ok := plansFor(t, g, me, "{R}{R}"); !ok || len(plan) != 1 || plan[0] != taiga {
		t.Fatalf("plan for {R}{R} = %v (ok %v), want the Taiga alone", plan, ok)
	}
	if err := castAutoTappedFromHand(t, g, me, "Two Red", "{R}{R}"); err != nil {
		t.Fatalf("auto-paid {R}{R}: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana floating, want none: %v", n, poolColors(me))
	}
}

// Forsaken Monument: "Whenever you tap a permanent for {C}, add an
// additional {C}" is one {C} for Sol Ring's whole production, so the Ring
// pays {3} alone and not {4}.
func TestForsakenMonumentAddsOneColorlessToSolRingsPlan(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	seedPermanentWithOracle(g, me.ID, "Forsaken Monument", "Legendary Artifact", s58ForsakenMonumentOracle)
	ring := pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", solRingOracle, false)

	if plan, ok := plansFor(t, g, me, "{4}"); ok {
		t.Errorf("{4} planned off one Sol Ring under Forsaken Monument: %v", plan)
	}
	if plan, ok := plansFor(t, g, me, "{3}"); !ok || len(plan) != 1 || plan[0] != ring {
		t.Fatalf("plan for {3} = %v (ok %v), want the Sol Ring alone", plan, ok)
	}
}

// Utopia Sprawl adds "one mana of the chosen color", which is nothing
// before a colour is chosen, so the planner counts nothing for it; once a
// colour is chosen it counts that colour.
func TestUtopiaSprawlIsCountedOnlyOnceAColourIsChosen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	forest := b31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	sprawl := pushAuraOnLand(t, g, me.ID, "Utopia Sprawl", utopiaSprawlOracle, forest)

	if plan, ok := plansFor(t, g, me, "{G}{G}"); ok {
		t.Errorf("{G}{G} planned off one Forest under a Utopia Sprawl with no colour: %v", plan)
	}
	g.WithWriteLock(func() { g.QueueColorChoiceForEffect(me.ID, sprawl, "Utopia Sprawl", nil, game.ColorForMana) })
	answerColor(t, g, me.ID, "R")
	if plan, ok := plansFor(t, g, me, "{G}{R}"); !ok || len(plan) != 1 || plan[0] != forest {
		t.Fatalf("plan for {G}{R} = %v (ok %v), want the Forest under a red Sprawl", plan, ok)
	}
}

// Sasaya's Essence adds one mana per OTHER land with the tapped land's
// name, read off the battlefield by the real card: with three Forests,
// one Forest is a three-mana source.
func TestSasayasEssenceIsCountedPerSameNameLand(t *testing.T) {
	g, me, _ := flipMain(t)
	flippedSasaya(t, g, me)
	forests := []uuid.UUID{
		pushNamedLand(g, me, "Forest", "Basic Land — Forest"),
		pushNamedLand(g, me, "Forest", "Basic Land — Forest"),
		pushNamedLand(g, me, "Forest", "Basic Land — Forest"),
	}

	plan, ok := plansFor(t, g, me, "{G}{G}{G}")
	if !ok || len(plan) != 1 {
		t.Fatalf("plan for {G}{G}{G} = %v (ok %v), want one Forest", plan, ok)
	}
	if plan[0] != forests[0] && plan[0] != forests[1] && plan[0] != forests[2] {
		t.Errorf("plan %v names no Forest", plan)
	}
}
