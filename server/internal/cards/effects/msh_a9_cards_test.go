package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msh_a9_cards_test.go — Marvel Super Heroes, half A, ninth batch (#4455).
// Each test plays the real catalog entry and asserts what its text says.

const (
	mshA9USAgentOracle    = "2b6fa8b6-4865-4eb6-974b-00f6b16b6f0f"
	mshA9UltronOracle     = "a8f25e67-2112-4355-9298-363ed04a5cb8"
	mshA9LegionnaireOracl = "0edaf433-4ded-4841-bc86-136780c0767f"
	mshA9HideoutOracle    = "cd2888aa-71f3-47ee-ba33-7bb95d5bc836"
	mshA9VivOracle        = "f67a599c-200a-4831-9e73-126c7435a396"
	mshA9DroneFlockOracle = "282d6e61-3240-4f1e-9c2d-8ddb59b03af4"
	mshA9WarMachineOracle = "ae6d1aa3-4ceb-4fc1-a3a6-033afd4b3bce"
	mshA9WhiteTigerOracle = "e2084d73-5a4c-4b17-942a-93c1a1f7ac2a"
	mshA9WiccanOracle     = "fc17b7a3-7711-4449-80bd-692e8ed9087d"
	mshA9YellowjackOracle = "0a364879-31bd-4760-833d-165cde1e96c6"
)

func TestMshA9EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mshA9USAgentOracle:    "U.S.Agent, John Walker",
		mshA9UltronOracle:     "Ultron, Artificial Malevolence",
		mshA9LegionnaireOracl: "Unliving Legionnaire",
		mshA9HideoutOracle:    "Villainous Hideout",
		mshA9VivOracle:        "Viv Vision, Teen Synthezoid",
		mshA9DroneFlockOracle: "Wakandan Drone Flock",
		mshA9WarMachineOracle: "War Machine, Legacy of Iron",
		mshA9WhiteTigerOracle: "White Tiger, Ava Ayala",
		mshA9WiccanOracle:     "Wiccan, Rising Magician",
		mshA9YellowjackOracle: "Yellowjacket, Heartless Marauder",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// mshA9Live is a card as the layers leave it now, wherever it is.
func mshA9Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		out, found = g.LookupCardForEffect(id)
	})
	if !found {
		t.Fatalf("card %s is nowhere", id)
	}
	return out
}

// mshA9Named lists the battlefield permanents with a name.
func mshA9Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// mshA9CastFromHand casts a card with the floated mana and resolves just
// the spell, leaving whatever it triggered on the stack.
func mshA9CastFromHand(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, pay string) uuid.UUID {
	t.Helper()
	floatMana(t, g, p, pay)
	id := castFromHandForTest(t, g, p, name, typeLine, cost, "", game.CastSpellParams{})
	resolveTopOfStack(t, g)
	return id
}

func mshA9Ref(id uuid.UUID) *game.TargetRef { return &game.TargetRef{Kind: game.TargetCard, ID: id} }

// --- U.S.Agent, John Walker ---------------------------------------------------------------

func TestMshA9USAgentMakesAShieldAndAttachesItToHimself(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	agent := castPowerUpCreature(t, g, me, "U.S.Agent, John Walker", "Legendary Creature — Human Soldier Hero", "{3}{W/B}", "{C}{C}{C}{W}", mshA9USAgentOracle, 2)
	assistCardSettle(t, g, me.ID, nil, true)

	shields := mshA9Named(g, "Sturdy Shield")
	if len(shields) != 1 {
		t.Fatalf("%d Sturdy Shield tokens, want 1", len(shields))
	}
	shield := shields[0]
	if !shield.IsToken() || !shield.IsArtifact() || !shield.HasSubtype("Equipment") || len(shield.Colors) != 0 {
		t.Errorf("Sturdy Shield = %+v, want a colorless Equipment artifact token", shield)
	}
	if shield.AttachedTo.ID != agent {
		t.Errorf("the shield is attached to %s, want U.S.Agent", shield.AttachedTo.ID)
	}
	live := mshA9Live(t, g, agent)
	if live.CurrentPower() != 3 || live.CurrentToughness() != 4 {
		t.Errorf("U.S.Agent = %d/%d, want 3/4 (2/2 with +1/+2)", live.CurrentPower(), live.CurrentToughness())
	}
}

func TestMshA9SturdyShieldEquipsForTwoAndMoves(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	castPowerUpCreature(t, g, me, "U.S.Agent, John Walker", "Legendary Creature — Human Soldier Hero", "{3}{W/B}", "{C}{C}{C}{W}", mshA9USAgentOracle, 2)
	assistCardSettle(t, g, me.ID, nil, true)
	shield := mshA9Named(g, "Sturdy Shield")[0].InstanceID
	bear := b12Creature(g, me.ID, "Their Bear", "Creature — Bear", 1, 1)

	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, shield, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("equip {2}: %v", err)
	}
	resolveTopOfStack(t, g)
	live := mshA9Live(t, g, bear)
	if live.CurrentPower() != 2 || live.CurrentToughness() != 3 {
		t.Errorf("equipped bear = %d/%d, want 2/3", live.CurrentPower(), live.CurrentToughness())
	}
}

// --- Ultron, Artificial Malevolence -------------------------------------------------------

func mshA9Ultron(g *game.Game, me *game.Player) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, "Ultron, Artificial Malevolence", "Legendary Artifact Creature — Robot Villain", mshA9UltronOracle, false)
}

func TestMshA9UltronCopiesANoncreatureArtifactAsARobotVillain(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mshA9Ultron(g, me)
	mshA9CastFromHand(t, g, me, "Test Gadget", "Artifact", "{1}", "{C}")
	passPriorityAroundTable(t, g)
	pay := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("Ultron did not offer the {2}")
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	gadgets := mshA9Named(g, "Test Gadget")
	if len(gadgets) != 2 {
		t.Fatalf("%d Test Gadgets, want the card and its copy", len(gadgets))
	}
	for _, c := range gadgets {
		if !c.IsToken() {
			continue
		}
		live := mshA9Live(t, g, c.InstanceID)
		if !live.IsCreature() || !live.IsArtifact() || !live.HasSubtype("Robot") || !live.HasSubtype("Villain") {
			t.Errorf("the copy is %q, want an artifact creature, Robot Villain in addition to its other types", live.TypeLine)
		}
		if live.CurrentPower() != 2 || live.CurrentToughness() != 2 {
			t.Errorf("the copy is %d/%d, want 2/2", live.CurrentPower(), live.CurrentToughness())
		}
		return
	}
	t.Fatal("no token copy of the artifact")
}

func TestMshA9UltronLeavesACreatureCopyAsItIs(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mshA9Ultron(g, me)
	castPowerUpCreature(t, g, me, "Test Golem", "Artifact Creature — Golem", "{1}", "{C}", "", 3)
	passPriorityAroundTable(t, g)
	pay := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("Ultron did not offer the {2} for an artifact creature")
	}
	floatMana(t, g, me, "{C}{C}")
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	for _, c := range mshA9Named(g, "Test Golem") {
		live := mshA9Live(t, g, c.InstanceID)
		if live.HasSubtype("Robot") || live.HasSubtype("Villain") || live.CurrentPower() != 3 {
			t.Errorf("%q %d/%d: a copy that is already a creature is not changed", live.TypeLine, live.CurrentPower(), live.CurrentToughness())
		}
	}
	if n := len(mshA9Named(g, "Test Golem")); n != 2 {
		t.Errorf("%d Test Golems, want 2", n)
	}
}

func TestMshA9UltronDecliningTheCostMakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mshA9Ultron(g, me)
	mshA9CastFromHand(t, g, me, "Test Gadget", "Artifact", "{1}", "{C}")
	passPriorityAroundTable(t, g)
	pay := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("Ultron did not offer the {2}")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if n := len(mshA9Named(g, "Test Gadget")); n != 1 {
		t.Errorf("%d Test Gadgets after declining, want 1", n)
	}
}

func TestMshA9UltronIgnoresNonartifactsAndTokens(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mshA9Ultron(g, me)
	castPowerUpCreature(t, g, me, "Test Bear", "Creature — Bear", "{1}", "{C}", "", 2)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Error("a nonartifact creature offered a copy")
	}
	tmpl, ok := Tokens().TokenTemplate("Treasure")
	if !ok {
		t.Fatal("no Treasure template")
	}
	if _, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, tmpl, 1); err != nil {
		t.Fatalf("SpawnCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Error("an artifact TOKEN offered a copy")
	}
}

// --- Unliving Legionnaire -----------------------------------------------------------------

const mshA9LegionnaireType = "Creature — Vampire Villain"

func TestMshA9UnlivingLegionnairePowerUpReturnsACreatureCardAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	card := castPowerUpCreature(t, g, me, "Unliving Legionnaire", mshA9LegionnaireType, "{3}{B}", "{C}{C}{C}{B}", mshA9LegionnaireOracl, 2)
	dead := graveyardCardOf(g, me, "Fallen Bear", "Creature — Bear", "{1}{G}", "")
	if got := chargedPowerUp(t, g, me.ID, card); got != "{2}{B}" {
		t.Fatalf("the turn it entered: %q, want {2}{B} ({5}{B}{B} less {3}{B})", got)
	}
	floatMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Targets: cardRefs(dead)}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature card did not return to hand")
	}
	if got := plusOnes(g, card); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: err = %v, want ErrPowerUpUsed", err)
	}
}

func TestMshA9UnlivingLegionnaireWithNoTargetOnlyGrows(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	card := castPowerUpCreature(t, g, me, "Unliving Legionnaire", mshA9LegionnaireType, "{3}{B}", "{C}{C}{C}{B}", mshA9LegionnaireOracl, 2)
	floatMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up with up to one target declined: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, card); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
}

func TestMshA9UnlivingLegionnaireOnlyReturnsCreatureCardsFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	card := castPowerUpCreature(t, g, me, "Unliving Legionnaire", mshA9LegionnaireType, "{3}{B}", "{C}{C}{C}{B}", mshA9LegionnaireOracl, 2)
	sorcery := graveyardCardOf(g, me, "Old Spell", "Sorcery", "{1}", "")
	theirs := graveyardCardOf(g, opp, "Their Bear", "Creature — Bear", "{1}{G}", "")
	floatMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Targets: cardRefs(sorcery)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a sorcery card: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("an opponent's graveyard: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Villainous Hideout -------------------------------------------------------------------

func TestMshA9VillainousHideoutManaIsLockedToVillainSpellsAndAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Villainous Hideout", "Land", mshA9HideoutOracle, false)

	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility {C}: %v", err)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "C" || len(me.ManaPool.Units()[0].Restrictions) != 0 {
		t.Fatalf("pool %v, want one unrestricted {C}", got)
	}
	me.ManaPool = game.ManaPool{}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })

	if err := g.ActivateManaAbility(me.ID, land, 1, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("ActivateManaAbility any colour: %v", err)
	}
	cost, _ := game.ParseCost("{B}")
	villain := game.Card{Name: "Test Villain", TypeLine: "Creature — Human Villain", ManaCost: "{B}"}
	hero := game.Card{Name: "Test Hero", TypeLine: "Creature — Human Hero", ManaCost: "{B}"}
	rite := game.Card{Name: "Test Rite", TypeLine: "Sorcery", ManaCost: "{B}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(villain)) {
		t.Error("the mana was refused for a Villain spell")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(hero)) {
		t.Error("the mana paid for a Hero spell: stronger than printed")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(rite)) {
		t.Error("the mana paid for a sorcery: stronger than printed")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(villain)) {
		t.Error("the mana was refused for an ability of a Villain source")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(hero)) {
		t.Error("the mana paid for an ability of a non-Villain source: stronger than printed")
	}
}

func TestMshA9VillainousHideoutConnivesATargetVillainAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Villainous Hideout", "Land", mshA9HideoutOracle, false)
	villain := b12Creature(g, me.ID, "Test Villain", "Creature — Human Villain", 2, 2)
	hero := b12Creature(g, me.ID, "Test Hero", "Creature — Human Hero", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	spell := handCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	seedLibrary(me, "Drawn")

	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: cardRefs(hero)}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a non-Villain target: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: cardRefs(villain)}); err != nil {
		t.Fatalf("connive: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("the Villain did not connive: no discard prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	if got := plusOnes(g, villain); got != 1 {
		t.Errorf("Villain has %d +1/+1 counters, want 1 (it discarded a nonland card)", got)
	}
	if !b16Tapped(t, g, land) {
		t.Error("the {T} cost was not paid")
	}
}

func TestMshA9VillainousHideoutConniveIsSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	land := pushCatalogPermanent(g, me.ID, "Villainous Hideout", "Land", mshA9HideoutOracle, false)
	villain := b12Creature(g, me.ID, "Test Villain", "Creature — Human Villain", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	floatMana(t, g, me, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Targets: cardRefs(villain)}); err == nil {
		t.Error("the connive was activated outside a main phase")
	}
}

// --- Viv Vision, Teen Synthezoid ----------------------------------------------------------

func TestMshA9VivVisionDrawsOnlyWhilePowerIsFourOrMore(t *testing.T) {
	for _, tc := range []struct {
		counters int
		draws    int
	}{{0, 0}, {1, 0}, {2, 1}, {3, 1}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		viv := b12Push(g, me.ID, "Viv Vision, Teen Synthezoid", "Legendary Artifact Creature — Robot Hero", mshA9VivOracle, 2, 2)
		g.WithWriteLock(func() { _ = g.AddCounterForEffect(viv, game.CounterPlusOne, tc.counters) })
		seedLibrary(me, "A", "B")
		before := me.Hand.Size()
		declareAttack(t, g, opp.ID, viv)
		assistCardSettle(t, g, me.ID, nil, true)
		if got := me.Hand.Size() - before; got != tc.draws {
			t.Errorf("with %d counters (power %d): drew %d, want %d", tc.counters, 2+tc.counters, got, tc.draws)
		}
	}
}

func TestMshA9VivVisionPowerUpCostsSevenAndGivesTwoCounters(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	viv := castPowerUpCreature(t, g, me, "Viv Vision, Teen Synthezoid", "Legendary Artifact Creature — Robot Hero", "{3}", "{C}{C}{C}", mshA9VivOracle, 2)
	if got := chargedPowerUp(t, g, me.ID, viv); got != "{4}" {
		t.Fatalf("the turn she entered: %q, want {4} ({7} less {3})", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, viv, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, viv); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if mshA9Live(t, g, viv).CurrentPower() != 4 {
		t.Error("Viv Vision is not a 4/4 after the power-up")
	}
}

// --- Wakandan Drone Flock -----------------------------------------------------------------

func TestMshA9WakandanDroneFlockScriesTwoWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	seedLibrary(me, "Bottom", "Middle", "Top")
	castPowerUpCreature(t, g, me, "Wakandan Drone Flock", "Artifact Creature — Robot", "{3}{W}", "{C}{C}{C}{W}", mshA9DroneFlockOracle, 3)
	passPriorityAroundTable(t, g)
	scry := scryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("entering did not ask to scry")
	}
	if len(scry.ScryCards) != 2 {
		t.Errorf("scry looks at %d cards, want 2", len(scry.ScryCards))
	}
}

// --- War Machine, Legacy of Iron ----------------------------------------------------------

func TestMshA9WarMachineGivesAnotherCreatureHisPowerAtBeginningOfCombat(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	wm := pushCatalogPermanent(g, me.ID, "War Machine, Legacy of Iron", "Legendary Artifact Creature — Human Hero", mshA9WarMachineOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(wm, game.CounterPlusOne, 2) })

	advanceTo(t, g, game.StepBeginCombat)
	assistCardSettle(t, g, me.ID, mshA9Ref(bear), true)
	live := mshA9Live(t, g, bear)
	if live.CurrentPower() != 5 || live.CurrentToughness() != 2 {
		t.Errorf("bear = %d/%d, want 5/2 (+3/+0, War Machine's power)", live.CurrentPower(), live.CurrentToughness())
	}
	if got := mshA9Live(t, g, wm).CurrentPower(); got != 3 {
		t.Errorf("War Machine's power = %d, want 3 (he does not pump himself)", got)
	}
}

func TestMshA9WarMachineCannotTargetHimselfOrAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	wm := pushCatalogPermanent(g, me.ID, "War Machine, Legacy of Iron", "Legendary Artifact Creature — Human Hero", mshA9WarMachineOracle, false)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	advanceTo(t, g, game.StepBeginCombat)
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt at the beginning of combat")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, *mshA9Ref(wm)); err == nil {
		t.Error("War Machine was accepted as his own target")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, *mshA9Ref(theirs)); err == nil {
		t.Error("an opponent's creature was accepted as the target")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, *mshA9Ref(mine)); err != nil {
		t.Fatalf("my other creature: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
}

// --- White Tiger, Ava Ayala ---------------------------------------------------------------

func TestMshA9WhiteTigerPowerUpGrowsAndMakesTheTigerGod(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	tiger := castPowerUpCreature(t, g, me, "White Tiger, Ava Ayala", "Legendary Creature — Human Hero", "{1}{G}", "{C}{G}", mshA9WhiteTigerOracle, 2)
	if got := chargedPowerUp(t, g, me.ID, tiger); got != "{4}" {
		t.Fatalf("the turn she entered: %q, want {4} ({5}{G} less {1}{G})", got)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tiger, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	if got := plusOnes(g, tiger); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	gods := mshA9Named(g, "The Tiger God")
	if len(gods) != 1 {
		t.Fatalf("%d The Tiger God tokens, want 1", len(gods))
	}
	god := mshA9Live(t, g, gods[0].InstanceID)
	if !god.IsToken() || !god.IsLegendary() || !god.HasSubtype("Cat") || !god.HasSubtype("God") ||
		len(god.Colors) != 1 || god.Colors[0] != "G" || god.CurrentPower() != 4 || god.CurrentToughness() != 4 {
		t.Errorf("The Tiger God = %+v, want a legendary 4/4 green Cat God token", god)
	}
}

func TestMshA9TheTigerGodCantBeBlockedByMoreThanOneCreature(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	tiger := castPowerUpCreature(t, g, me, "White Tiger, Ava Ayala", "Legendary Creature — Human Hero", "{1}{G}", "{C}{G}", mshA9WhiteTigerOracle, 2)
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tiger, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power up: %v", err)
	}
	resolveTopOfStack(t, g)
	god := mshA9Named(g, "The Tiger God")[0].InstanceID
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)

	powerUpPassTo(t, g, 0)
	brAttack(t, g, god)
	brRefusal(t, g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: a, Attacker: god}, {Blocker: b, Attacker: god},
	}), game.BlockReasonTooManyBlockers)
	if err := g.DeclareBlocker(a, god); err != nil {
		t.Fatalf("one blocker is legal: %v", err)
	}
}

// --- Wiccan, Rising Magician --------------------------------------------------------------

const mshA9WiccanType = "Legendary Creature — Mutant Warlock Hero"

func TestMshA9WiccanBlinksAnotherPermanentUntilTheNextEndStep(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wiccan, Rising Magician", mshA9WiccanType, mshA9WiccanOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	floatMana(t, g, me, "{C}")
	castFromHandForTest(t, g, me, "Test Sorcery", "Sorcery", "{1}", "", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, mshA9Ref(theirs), true)
	if z := b12ZoneOf(g, theirs); z != game.ZoneExile {
		t.Fatalf("the target is in %v, want exile until the next end step", z)
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	back := mshA9Named(g, "Their Bear")
	if len(back) != 1 {
		t.Fatalf("%d Their Bears on the battlefield, want the one that returned at the end step", len(back))
	}
	if back[0].Controller != opp.ID {
		t.Error("the card did not return under its owner's control")
	}
}

func TestMshA9WiccanOnlyTargetsOtherNonlandNontokenPermanents(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	wiccan := pushCatalogPermanent(g, me.ID, "Wiccan, Rising Magician", mshA9WiccanType, mshA9WiccanOracle, false)
	land := pushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	token := b12Creature(g, opp.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	floatMana(t, g, me, "{C}")
	castFromHandForTest(t, g, me, "Test Sorcery", "Sorcery", "{1}", "", game.CastSpellParams{})
	prompt := latestPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatal("no target prompt for the cast trigger")
	}
	for name, id := range map[string]uuid.UUID{"Wiccan himself": wiccan, "a land": land, "a token": token} {
		if err := g.ResolvePickTarget(prompt.ID, me.ID, *mshA9Ref(id)); err == nil {
			t.Errorf("%s was accepted as the target", name)
		}
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, *mshA9Ref(theirs)); err != nil {
		t.Fatalf("a nonland, nontoken permanent: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
}

func TestMshA9WiccanDoesNotTriggerOnACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wiccan, Rising Magician", mshA9WiccanType, mshA9WiccanOracle, false)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castPowerUpCreature(t, g, me, "Test Bear", "Creature — Bear", "{1}", "{C}", "", 2)
	if latestPickTarget(g, me.ID) != nil || !g.Battlefield.Contains(theirs) {
		t.Error("a creature spell triggered Wiccan")
	}
}

// --- Yellowjacket, Heartless Marauder -----------------------------------------------------

func TestMshA9YellowjacketGrowsAndGainsLifelinkForEachAnotherVillain(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	yj := pushCatalogPermanent(g, me.ID, "Yellowjacket, Heartless Marauder", "Legendary Creature — Human Rogue Villain", mshA9YellowjackOracle, false)
	mshA9Cast := func(name, typeLine string) {
		castPowerUpCreature(t, g, me, name, typeLine, "{1}", "{C}", "", 2)
		assistCardSettle(t, g, me.ID, nil, true)
	}
	mshA9Cast("Test Hero", "Creature — Human Hero")
	if live := mshA9Live(t, g, yj); live.CurrentPower() != 1 || game.HasKeyword(&live, "lifelink") {
		t.Fatalf("a Hero entering changed Yellowjacket: %d power, lifelink %v", live.CurrentPower(), game.HasKeyword(&live, "lifelink"))
	}
	mshA9Cast("Test Villain A", "Creature — Human Villain")
	live := mshA9Live(t, g, yj)
	if live.CurrentPower() != 2 || live.CurrentToughness() != 1 || !game.HasKeyword(&live, "lifelink") {
		t.Errorf("after one Villain: %d/%d lifelink=%v, want power 2, toughness 1, lifelink", live.CurrentPower(), live.CurrentToughness(), game.HasKeyword(&live, "lifelink"))
	}
	mshA9Cast("Test Villain B", "Creature — Human Villain")
	if got := mshA9Live(t, g, yj).CurrentPower(); got != 3 {
		t.Errorf("after two Villains: power %d, want 3", got)
	}
	powerUpPassTo(t, g, 0)
	live = mshA9Live(t, g, yj)
	if live.CurrentPower() != 1 || game.HasKeyword(&live, "lifelink") {
		t.Errorf("a turn later: power %d lifelink %v, want the bonus gone", live.CurrentPower(), game.HasKeyword(&live, "lifelink"))
	}
}
