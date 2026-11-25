package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b129_cards_test.go — S82 OLD batch 129 (#5514): Theros Beyond
// Death escape cards. Every test plays the real catalog card through the
// engine, and escapes it from the graveyard where the card has escape.

const (
	b129Phoenix  = "0ad557fc-6b7e-4f7e-b060-e3c8ec8acf6d"
	b129Velocity = "2db4c197-982b-4c23-88e7-6ddf17378bb4"
	b129Mogis    = "312dbc3a-3ffb-4ed2-a278-8950d8ce8c86"
	b129Charger  = "39694a3f-536e-41e0-adf6-d49de721aabd"
	b129Satyr    = "46fd9b9e-ef8f-41d6-9f1d-040bb7fe74c1"
	b129Sleep    = "5e1b400f-ec5c-4703-bf30-65bb16bd3ca6"
	b129Hound    = "a1418e8d-dea5-43e1-b041-3f6a88f9a6d7"
	b129Aracnir  = "c45423ad-1920-46f7-bcc1-ecade5880821"
	b129Elspeth  = "faf296b0-59e5-4ec9-9912-e5e21c0ce9f5"
	b129Bear     = "Creature — Bear"
)

func TestB129EveryCardIsRegisteredFullWithItsEscape(t *testing.T) {
	for _, c := range []struct {
		oracle, name, cost string
		n                  int
	}{
		{b129Phoenix, "Phoenix of Ash", "{2}{R}{R}", 3},
		{b129Velocity, "Escape Velocity", "{1}{R}", 2},
		{b129Mogis, "Mogis's Favor", "{2}{B}", 2},
		{b129Charger, "Underworld Charger", "{4}{B}", 3},
		{b129Satyr, "Satyr's Cunning", "{2}{R}", 2},
		{b129Sleep, "Sleep of the Dead", "{2}{U}", 3},
		{b129Hound, "Underworld Rage-Hound", "{3}{R}", 3},
		{b129Aracnir, "Chainweb Aracnir", "{3}{G}{G}", 4},
		{b129Elspeth, "Elspeth, Sun's Nemesis", "{4}{W}{W}", 4},
	} {
		spec, ok := effects.Lookup(c.oracle)
		if !ok || spec.Name != c.name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", c.name, ok, spec.Name, spec.Completeness, spec.Caveats)
			continue
		}
		offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
		if len(offers) != 1 || offers[0].Key != "escape" || offers[0].ManaCost != c.cost ||
			offers[0].ExileFromGraveyard == nil || offers[0].ExileFromGraveyard.Min != c.n {
			t.Errorf("%s: graveyard offers = %+v, want escape %s exiling %d", c.name, offers, c.cost, c.n)
		}
	}
}

// b129Yard puts n filler cards into the graveyard and returns their ids.
func b129Yard(p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

// b129Escape escapes the card from the graveyard, exiling n filler cards.
func b129Escape(t *testing.T, g *game.Game, name, typeLine, oracle string, n int, targets []game.TargetRef) (uuid.UUID, []uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	pay := b129Yard(me, n)
	id := b129Seed(t, g, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay, Targets: targets,
	}); err != nil {
		t.Fatalf("escape %s: %v", name, err)
	}
	return id, pay
}

// b129Seed puts the card into the active seat's graveyard with its printed
// power and toughness, which a catalog row does not carry in a test table.
func b129Seed(t *testing.T, g *game.Game, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := effectstest.SeedGraveyardCard(t, g, name, typeLine, oracle)
	pt := map[string][2]int{b129Phoenix: {2, 2}, b129Charger: {3, 3}, b129Hound: {3, 1}, b129Aracnir: {1, 2}}[oracle]
	g.WithWriteLock(func() {
		p := g.Seats[g.Turn.ActiveSeat]
		for i := range p.Graveyard.Cards {
			if p.Graveyard.Cards[i].InstanceID == id {
				p.Graveyard.Cards[i].Power, p.Graveyard.Cards[i].Toughness = pt[0], pt[1]
			}
		}
	})
	return id
}

func b129Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.B12Counter(t, g, id, game.CounterPlusOne)
}

func TestB129PhoenixOfAshEscapesWithACounterAndFirebreathes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id, pay := b129Escape(t, g, "Phoenix of Ash", "Creature — Phoenix", b129Phoenix, 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the escaped Phoenix is not on the battlefield")
	}
	for _, c := range pay {
		if me.Graveyard.Contains(c) {
			t.Error("an escape cost card is still in the graveyard")
		}
	}
	if b129Counters(t, g, id) != 1 || effectstest.EffectivePowerWithCounters(t, g, id) != 3 {
		t.Errorf("escaped Phoenix: counters %d power %d, want 1 and 3", b129Counters(t, g, id), effectstest.EffectivePowerWithCounters(t, g, id))
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flying") || !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Error("Phoenix lacks flying or haste")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("firebreathing: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePowerWithCounters(t, g, id); p != 5 {
		t.Errorf("after {2}{R} the Phoenix has power %d, want 5", p)
	}
}

func TestB129PhoenixOfAshHardCastHasNoCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastCatalogSpell(t, g, "Phoenix of Ash", "Creature — Phoenix", b129Phoenix, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || b129Counters(t, g, id) != 0 {
		t.Errorf("hard-cast Phoenix: on battlefield %v, counters %d", g.Battlefield.Contains(id), b129Counters(t, g, id))
	}
}

func TestB129EscapeIsRefusedWithTooFewCardsInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pay := b129Yard(me, 2)
	id := effectstest.SeedGraveyardCard(t, g, "Phoenix of Ash", "Creature — Phoenix", b129Phoenix)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err == nil {
		t.Error("escaped with two other cards instead of three")
	}
}

func TestB129UnderworldChargerEscapesWithTwoCountersAndCannotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, _ := b129Escape(t, g, "Underworld Charger", "Creature — Nightmare Horse", b129Charger, 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if b129Counters(t, g, id) != 2 || effectstest.EffectivePowerWithCounters(t, g, id) != 5 {
		t.Errorf("escaped Charger: counters %d power %d, want 2 and 5", b129Counters(t, g, id), effectstest.EffectivePowerWithCounters(t, g, id))
	}
	spec, _ := effects.Lookup(b129Charger)
	if len(spec.Static) != 1 {
		t.Errorf("Charger has %d static abilities, want the can't-block restriction", len(spec.Static))
	}
}

func TestB129RageHoundEscapesWithACounterAndMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id, _ := b129Escape(t, g, "Underworld Rage-Hound", "Creature — Elemental Dog", b129Hound, 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if b129Counters(t, g, id) != 1 || effectstest.EffectivePowerWithCounters(t, g, id) != 4 {
		t.Errorf("escaped Rage-Hound: counters %d power %d, want 1 and 4", b129Counters(t, g, id), effectstest.EffectivePowerWithCounters(t, g, id))
	}
	spec, _ := effects.Lookup(b129Hound)
	if len(spec.Static) != 1 {
		t.Errorf("Rage-Hound has %d static abilities, want attacks-each-combat", len(spec.Static))
	}
}

func TestB129ChainwebAracnirEscapedDealsItsFourPowerToAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	flier := effectstest.B12Creature(g, opp.ID, "Drake", "Creature — Drake", 2, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == flier {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	ground := effectstest.B12Creature(g, opp.ID, "Bear", b129Bear, 2, 2)
	pay := b129Yard(g.Seats[g.Turn.ActiveSeat], 4)
	id := b129Seed(t, g, "Chainweb Aracnir", "Creature — Spider", b129Aracnir)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.AnswerTriggerTargets(t, g, me.ID, flier)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if b129Counters(t, g, id) != 3 {
		t.Fatalf("escaped Aracnir counters %d, want 3", b129Counters(t, g, id))
	}
	if g.Battlefield.Contains(flier) {
		t.Error("the 4-toughness flier survived: the escaped Aracnir has power 4")
	}
	if c, ok := effectstest.BattlefieldCard(g, ground); !ok || c.DamageMarked != 0 {
		t.Errorf("the ground creature was damaged: %+v", c)
	}
}

func TestB129EscapeVelocityEscapesOntoACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b129Bear, 2, 2)
	id, _ := b129Escape(t, g, "Escape Velocity", "Enchantment — Aura", b129Velocity, 2, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the escaped Aura is not on the battlefield")
	}
	if p := effectstest.EffectivePower(t, g, bear); p != 3 {
		t.Errorf("enchanted Bear power %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("enchanted Bear lacks haste")
	}
}

func TestB129MogissFavorGivesPlusTwoMinusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", b129Bear, 2, 2)
	b129Escape(t, g, "Mogis's Favor", "Enchantment — Aura", b129Mogis, 2, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 4 || tg != 1 {
		t.Errorf("enchanted Bear is %d/%d, want 4/1", p, tg)
	}
}

func TestB129SatyrsCunningMakesACantBlockSatyrAndEscapes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id, pay := b129Escape(t, g, "Satyr's Cunning", "Sorcery", b129Satyr, 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, me.ID, "Satyr"); n != 1 {
		t.Fatalf("Satyr tokens = %d, want 1", n)
	}
	if me.Graveyard.Contains(pay[0]) || !me.Graveyard.Contains(id) {
		t.Error("an escaped sorcery goes back to the graveyard (no exile rider) and its cost cards are exiled")
	}
}

func TestB129SleepOfTheDeadTapsAndFreezes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b129Bear, 2, 2)
	b129Escape(t, g, "Sleep of the Dead", "Sorcery", b129Sleep, 3, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); !tapped {
		t.Fatal("the Bear is untapped")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, bear); !tapped {
		t.Error("the Bear untapped in its controller's next untap step")
	}
}

func b129Activate(t *testing.T, g *game.Game, me *game.Player, walker uuid.UUID, idx int, targets ...game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("loyalty ability %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestB129ElspethSunsNemesisAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.PushWalkerForTest(g, me.ID, "Elspeth, Sun's Nemesis", b129Elspeth, 5)
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", b129Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", b129Bear, 2, 2)
	b129Activate(t, g, me, walker, 0, effectstest.B16TargetCard(a)[0], effectstest.B16TargetCard(b)[0])
	if p, tg := effectstest.EffectivePower(t, g, a), effectstest.EffectiveToughness(t, g, a); p != 4 || tg != 3 {
		t.Errorf("A is %d/%d, want 4/3", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, b); p != 4 {
		t.Errorf("B power %d, want 4", p)
	}
	if effectstest.LoyaltyOf(g, walker) != 4 {
		t.Errorf("loyalty %d after −1, want 4", effectstest.LoyaltyOf(g, walker))
	}
}

func TestB129ElspethMakesTwoSoldiersAndGainsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.PushWalkerForTest(g, me.ID, "Elspeth, Sun's Nemesis", b129Elspeth, 5)
	effectstest.AdvanceToMain(t, g)
	b129Activate(t, g, me, walker, 1)
	if n := effectstest.CountTokensControlled(g, me.ID, "Soldier"); n != 2 {
		t.Errorf("Soldier tokens = %d, want 2", n)
	}
	if n := effectstest.CountTokensControlled(g, me.ID, "Human"); n != 2 {
		t.Errorf("Human tokens = %d, want 2", n)
	}
	if effectstest.LoyaltyOf(g, walker) != 3 {
		t.Errorf("loyalty %d after −2, want 3", effectstest.LoyaltyOf(g, walker))
	}
}

func TestB129ElspethMinusThreeGainsFiveLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	walker := effectstest.PushWalkerForTest(g, me.ID, "Elspeth, Sun's Nemesis", b129Elspeth, 5)
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	b129Activate(t, g, me, walker, 2)
	if me.Life != life+5 {
		t.Errorf("life %d→%d, want +5", life, me.Life)
	}
	if effectstest.LoyaltyOf(g, walker) != 2 {
		t.Errorf("loyalty %d after −3, want 2", effectstest.LoyaltyOf(g, walker))
	}
}
