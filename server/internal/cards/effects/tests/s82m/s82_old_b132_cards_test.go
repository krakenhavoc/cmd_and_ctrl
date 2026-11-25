package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b132_cards_test.go — S82 OLD batch 132 (#5514): Theros Beyond
// Death constellation creatures. Every test plays the real catalog card
// through the engine and enters a real enchantment.

const (
	b132Thaumaturge = "148331cc-24de-481f-9131-c3089d1ff7ce"
	b132Eutropia    = "4de1319f-b3c7-42d4-9342-3e9c4f526ee0"
	b132Sage        = "62c949c8-0011-4ab7-8388-e643be79df83"
	b132Skirmisher  = "789d2858-f342-422c-9187-c6ace9485ed5"
	b132Waverider   = "7ee66c03-3684-404f-8f2c-dc91e99249bc"
	b132Wardens     = "a2652158-e63f-477b-b328-cae7ef2263bd"
	b132Kraken      = "be700da5-123e-461e-8192-88f9d86925b9"
)

func TestB132EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b132Thaumaturge: "Protean Thaumaturge", b132Eutropia: "Eutropia the Twice-Favored", b132Sage: "Sage of Mysteries",
		b132Skirmisher: "Setessan Skirmisher", b132Waverider: "Triton Waverider", b132Wardens: "Nexus Wardens",
		b132Kraken: "Shoal Kraken",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// b132Table puts the catalog creature on my battlefield.
func b132Table(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogPermanent(g, me.ID, name, typeLine, oracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = power, toughness
			}
		}
	})
	return id
}

// b132Enchant casts a plain enchantment from my hand and resolves it.
func b132Enchant(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := effectstest.CastCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestB132EutropiaPutsACounterAndGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eutropia := b132Table(t, g, "Eutropia the Twice-Favored", "Legendary Creature — Human Wizard", b132Eutropia, 2, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B12Counter(t, g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters on the Bear: %d, want 1", n)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
	if effectstest.HasEffectiveKeyword(t, g, eutropia, "flying") {
		t.Error("Eutropia gained flying though the Bear was the target")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("flying outlasted the turn")
	}
}

func TestB132ConstellationIgnoresAnOpponentsEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b132Table(t, g, "Nexus Wardens", "Creature — Satyr Archer", b132Wardens, 1, 4)
	life := me.Life
	effectstest.B12Permanent(g, opp.ID, "Their Enchantment", "Enchantment")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("an opponent's enchantment triggered constellation: life %d → %d", life, me.Life)
	}
	// The same table does trigger for an enchantment of mine.
	b132Enchant(t, g)
	if me.Life != life+2 {
		t.Errorf("my own enchantment: life %d → %d, want +2", life, me.Life)
	}
}

func TestB132SageOfMysteriesMillsTheTargetPlayerTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b132Table(t, g, "Sage of Mysteries", "Creature — Human Wizard", b132Sage, 0, 2)
	effectstest.SeedLibrary(opp, "A", "B", "C", "D")
	before := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick the opponent: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != before-2 {
		t.Errorf("opponent's library %d → %d, want two milled", before, got)
	}
}

func TestB132SetessanSkirmisherGetsPlusOnePlusOneUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	skirm := b132Table(t, g, "Setessan Skirmisher", "Creature — Human Warrior", b132Skirmisher, 2, 1)
	b132Enchant(t, g)
	if p, tg := effectstest.EffectivePower(t, g, skirm), effectstest.EffectiveToughness(t, g, skirm); p != 3 || tg != 2 {
		t.Errorf("Skirmisher is %d/%d, want 3/2", p, tg)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.EffectivePower(t, g, skirm); p != 2 {
		t.Errorf("the bonus outlasted the turn: power %d", p)
	}
}

func TestB132TritonWaveriderGainsFlyingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	wave := b132Table(t, g, "Triton Waverider", "Creature — Merfolk Wizard", b132Waverider, 3, 3)
	if effectstest.HasEffectiveKeyword(t, g, wave, "flying") {
		t.Fatal("the Waverider flies before constellation")
	}
	b132Enchant(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, wave, "flying") {
		t.Error("the Waverider did not gain flying")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, wave, "flying") {
		t.Error("flying outlasted the turn")
	}
}

func TestB132NexusWardensGainsTwoLifeAndHasReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	wardens := b132Table(t, g, "Nexus Wardens", "Creature — Satyr Archer", b132Wardens, 1, 4)
	life := me.Life
	b132Enchant(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if !effectstest.HasEffectiveKeyword(t, g, wardens, "reach") {
		t.Error("Nexus Wardens lacks reach")
	}
}

func TestB132ShoalKrakenMayLoot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b132Table(t, g, "Shoal Kraken", "Creature — Kraken", b132Kraken, 3, 5)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	handAfterCast := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handAfterCast+1 {
		t.Fatalf("hand %d → %d, want one card drawn", handAfterCast, got)
	}
	if effectstest.DiscardOwed(g, me.ID) != 1 {
		t.Fatalf("discards owed %d, want 1", effectstest.DiscardOwed(g, me.ID))
	}
	c := effectstest.DiscardChoiceFor(g, me.ID)
	effectstest.AnswerDiscard(t, g, me.ID, c.ChooseCards[0])
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handAfterCast {
		t.Errorf("after the discard the hand is %d, want %d", got, handAfterCast)
	}
}

func TestB132ShoalKrakenDeclinedDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b132Table(t, g, "Shoal Kraken", "Creature — Kraken", b132Kraken, 3, 5)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	handAfterCast := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handAfterCast {
		t.Errorf("declined: hand %d → %d", handAfterCast, got)
	}
}

func TestB132ProteanThaumaturgeBecomesACopyAndKeepsItsAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thau := b132Table(t, g, "Protean Thaumaturge", "Creature — Human Wizard", b132Thaumaturge, 1, 1)
	target := effectstest.B12Creature(g, g.Seats[1].ID, "Big Drake", "Creature — Drake", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, target)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, thau)
	if c.Name != "Big Drake" {
		t.Fatalf("the Thaumaturge is %q, want a copy of Big Drake", c.Name)
	}
	if p := effectstest.EffectivePower(t, g, thau); p != 5 {
		t.Errorf("copy power %d, want 5", p)
	}
	// It has this ability: another enchantment asks again.
	effectstest.CastCatalogSpell(t, g, "Second Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) == nil && effectstest.LatestTriggerPrompt(g, me.ID) == nil {
		t.Error("the copy lost the constellation ability")
	}
}

func TestB132ProteanThaumaturgeDeclinedStaysItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thau := b132Table(t, g, "Protean Thaumaturge", "Creature — Human Wizard", b132Thaumaturge, 1, 1)
	target := effectstest.B12Creature(g, g.Seats[1].ID, "Big Drake", "Creature — Drake", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Test Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, target)
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, thau); c.Name != "Protean Thaumaturge" {
		t.Errorf("declined, but it is now %q", c.Name)
	}
}
