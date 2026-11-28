package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 632 (Apocalypse: activated abilities): Living Airship,
// Diversionary Tactics, Raka Disciple, Helionaut, Zombie Boa, Vodalian Mystic,
// Necra Disciple. Unnatural Selection is already on develop.

const (
	oldB632Airship   = "3bcead90-f7be-4fba-84e7-3d7c788a8b64"
	oldB632Tactics   = "880367e4-50b9-49d0-b596-d462d9b42855"
	oldB632Raka      = "88c5a681-4095-4e37-9149-002bcbaf4c0b"
	oldB632Helionaut = "8d4de785-20e9-430e-8df1-f39dac0ef07d"
	oldB632Boa       = "c4949d62-c2f9-4425-a77e-444824955ecc"
	oldB632Mystic    = "c7feecf0-5229-4c12-806a-16c9ab38e147"
	oldB632Necra     = "ef85a5eb-9cde-47b7-97e4-daa9aa9d4676"
)

func TestOldB632EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB632Airship: "Living Airship", oldB632Tactics: "Diversionary Tactics", oldB632Raka: "Raka Disciple",
		oldB632Helionaut: "Helionaut", oldB632Boa: "Zombie Boa", oldB632Mystic: "Vodalian Mystic",
		oldB632Necra: "Necra Disciple",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func oldB632Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	g.WithWriteLock(func() { who.ManaPool.EmptyPool() })
	if mana != "" {
		effectstest.FloatForTest(g, who, mana)
	}
	params.Strict = true
	err := g.ActivateCatalogAbility(who.ID, id, idx, params)
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

// --- Living Airship --------------------------------------------------------------

func TestOldB632LivingAirshipFliesAndRegeneratesForTwoAndGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ship := effectstest.B12Push(g, me.ID, "Living Airship", "Creature — Metathran", oldB632Airship, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, ship, "flying") {
		t.Error("no flying")
	}
	effectstest.AdvanceToMain(t, g)
	if err := oldB632Activate(t, g, me, ship, 0, "RB", game.ActivateAbilityParams{}); err == nil {
		t.Error("regenerated without green mana")
	}
	if err := oldB632Activate(t, g, me, ship, 0, "RRG", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
	if n := effectstest.RegenShieldsOn(t, g, ship); n != 1 {
		t.Fatalf("regeneration shields %d, want 1", n)
	}
	effectstest.DestroyIt(t, g, ship)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(ship) {
		t.Error("the regenerated Airship was destroyed")
	}
}

// --- Diversionary Tactics -----------------------------------------------------------

func TestOldB632DiversionaryTacticsTapsTwoOfYoursToTapATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tac := effectstest.PushCatalogPermanent(g, me.ID, "Diversionary Tactics", "Enchantment", oldB632Tactics, false)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	if err := oldB632Activate(t, g, me, tac, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTargets(victim), TapIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("activated tapping one creature")
	}
	b := effectstest.PushCatalogPermanent(g, me.ID, "Bear B", "Creature — Bear", "", true) // summoning sick may pay
	if err := oldB632Activate(t, g, me, tac, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTargets(victim), TapIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !effectstest.IsTapped(g, victim) {
		t.Error("the target was not tapped")
	}
	if !effectstest.IsTapped(g, a) || !effectstest.IsTapped(g, b) {
		t.Error("the two creatures were not tapped as the cost")
	}
	// The opponent's creatures can't pay: only creatures you control.
	o1 := effectstest.B12Creature(g, opp.ID, "O1", "Creature — Bear", 1, 1)
	o2 := effectstest.B12Creature(g, opp.ID, "O2", "Creature — Bear", 1, 1)
	if err := oldB632Activate(t, g, me, tac, 0, "", game.ActivateAbilityParams{Targets: effectstest.CardTargets(victim), TapIDs: []uuid.UUID{o1, o2}}); err == nil {
		t.Error("tapped an opponent's creatures to pay")
	}
}

// --- Raka Disciple / Necra Disciple / Helionaut ----------------------------------------

func TestOldB632RakaDisciplePreventsOneDamageOrGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raka := effectstest.PushCatalogPermanent(g, me.ID, "Raka Disciple", "Creature — Minotaur Wizard", oldB632Raka, false)
	raka2 := effectstest.PushCatalogPermanent(g, me.ID, "Raka Disciple", "Creature — Minotaur Wizard", oldB632Raka, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 4)
	effectstest.AdvanceToMain(t, g)
	if err := oldB632Activate(t, g, me, raka, 0, "R", game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)}); err == nil {
		t.Error("the shield took red mana")
	}
	if err := oldB632Activate(t, g, me, raka, 0, "W", game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("shield: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 3) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("3 damage into a 1-shield: marked %d, want 2", got)
	}
	if err := oldB632Activate(t, g, me, raka2, 1, "U", game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("flying: %v", err)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target has no flying")
	}
	if err := oldB632Activate(t, g, me, raka, 1, "U", game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)}); err == nil {
		t.Error("a tapped Disciple activated again")
	}
}

func TestOldB632NecraDisciplePreventsOneDamageAndFiltersManaIntoAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	n1 := effectstest.PushCatalogPermanent(g, me.ID, "Necra Disciple", "Creature — Human Wizard", oldB632Necra, false)
	n2 := effectstest.PushCatalogPermanent(g, me.ID, "Necra Disciple", "Creature — Human Wizard", oldB632Necra, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 4)
	effectstest.AdvanceToMain(t, g)
	if err := oldB632Activate(t, g, me, n1, 0, "W", game.ActivateAbilityParams{Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("shield: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(uuid.New(), bear, 1) })
	if got := effectstest.DamageMarkedOn(g, bear); got != 0 {
		t.Errorf("1 damage into a 1-shield: marked %d, want 0", got)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "G")
	effectstest.ActivateManaFor(t, g, me.ID, n2, 0, game.ManaAbilityParams{Colors: []string{"R"}})
	if me.ManaPool.CountColor("R") != 1 || me.ManaPool.CountColor("G") != 0 {
		t.Errorf("pool has %d red and %d green, want {G} turned into one {R}", me.ManaPool.CountColor("R"), me.ManaPool.CountColor("G"))
	}
}

func TestOldB632HelionautFliesAndFiltersOneManaIntoAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	h := effectstest.PushCatalogPermanent(g, me.ID, "Helionaut", "Creature — Human Soldier", oldB632Helionaut, false)
	if !effectstest.HasEffectiveKeyword(t, g, h, "flying") {
		t.Error("no flying")
	}
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "B")
	effectstest.ActivateManaFor(t, g, me.ID, h, 0, game.ManaAbilityParams{Colors: []string{"U"}})
	if me.ManaPool.CountColor("U") != 1 || me.ManaPool.CountColor("B") != 0 {
		t.Errorf("pool has %d blue and %d black, want {B} paid and one {U} added", me.ManaPool.CountColor("U"), me.ManaPool.CountColor("B"))
	}
	if !effectstest.IsTapped(g, h) {
		t.Error("the Helionaut did not tap")
	}
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Helionaut", "Creature — Human Soldier", oldB632Helionaut, true)
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateManaAbility(me.ID, sick, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err == nil {
		t.Error("a summoning-sick Helionaut paid {T}")
	}
}

// --- Zombie Boa --------------------------------------------------------------------

func oldB632BoaTurn(t *testing.T, color string) (g *game.Game, boa, red, white uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boa = effectstest.B12Push(g, me.ID, "Zombie Boa", "Creature — Zombie Snake", oldB632Boa, 3, 3)
	red = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Blocker", TypeLine: "Creature — Bear",
		Power: 1, Toughness: 4, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	white = effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "White Blocker", TypeLine: "Creature — Bear",
		Power: 1, Toughness: 4, Colors: []string{"W"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	if err := oldB632Activate(t, g, me, boa, 0, "RB", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.AnswerColor(t, g, me.ID, color)
	effectstest.PassPriorityAroundTable(t, g)
	return g, boa, red, white
}

func TestOldB632ZombieBoaDestroysOnlyBlockersOfTheChosenColour(t *testing.T) {
	g, boa, red, white := oldB632BoaTurn(t, "R")
	effectstest.Pr7bAttack(t, g, g.Seats[1].ID, boa)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{red: boa, white: boa})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(red) {
		t.Error("the red blocker survived blocking a Boa that chose red")
	}
	if !g.Battlefield.Contains(white) {
		t.Error("the white blocker was destroyed")
	}
}

func TestOldB632ZombieBoaDestroysEveryBlockerOfThatColour(t *testing.T) {
	g, boa, red, white := oldB632BoaTurn(t, "W")
	effectstest.Pr7bAttack(t, g, g.Seats[1].ID, boa)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{red: boa, white: boa})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) {
		t.Error("the white blocker survived blocking a Boa that chose white")
	}
	if !g.Battlefield.Contains(red) {
		t.Error("the red blocker was destroyed")
	}
}

func TestOldB632ZombieBoaIsSorcerySpeedAndLastsOnlyThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	boa := effectstest.B12Push(g, me.ID, "Zombie Boa", "Creature — Zombie Snake", oldB632Boa, 3, 3)
	effectstest.AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	if err := oldB632Activate(t, g, me, boa, 0, "RB", game.ActivateAbilityParams{}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Vodalian Mystic ------------------------------------------------------------------

func TestOldB632VodalianMysticRecoloursASpellAndNoOtherTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	mystic := effectstest.PushCatalogPermanent(g, me.ID, "Vodalian Mystic", "Creature — Merfolk Wizard", oldB632Mystic, false)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Warded Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Keywords: []string{"protection from blue"}, Owner: opp.ID, Controller: opp.ID})
	bolt := effectstest.CastFor(t, g, me, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		game.CastSpellParams{Targets: effectstest.RtCardTarget(bear)})
	// A permanent is not a legal target: only an instant or sorcery spell is.
	if err := g.ActivateCatalogAbility(me.ID, mystic, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)}); err == nil {
		t.Fatal("targeted a permanent")
	}
	if err := g.ActivateCatalogAbility(me.ID, mystic, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bolt)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 8 && effectstest.PendingOfKind(g, game.PendingChoiceColor) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	effectstest.AnswerColor(t, g, me.ID, "U")
	colors := []string(nil)
	for _, c := range g.Stack.Cards {
		if c.InstanceID == bolt {
			colors = c.EffectiveColors()
		}
	}
	if len(colors) != 1 || colors[0] != "U" {
		t.Fatalf("the Bolt on the stack is %v, want blue", colors)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the blue Bolt hit a creature with protection from blue")
	}
	if !effectstest.IsTapped(g, mystic) {
		t.Error("the Mystic did not tap")
	}
}
