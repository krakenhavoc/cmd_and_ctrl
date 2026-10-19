package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roon_slice_a_test.go — the Roon of the Hidden Realm deck request
// (#4128), slice a: Roon, Brago, Eldrazi Displacer, Venser, the
// Sojourner and Yorion, Sky Nomad.

const (
	roonaRoonOracle      = "fd336830-4a11-42b8-9fc7-d7526f569124"
	roonaBragoOracle     = "fcb0c7db-bd07-4c25-b7f8-b6c207e4f6be"
	roonaDisplacerOracle = "9b1f552a-bddc-4fcb-ac67-b4a65b2f48ba"
	roonaVenserOracle    = "a8bf8ff8-d924-4fd2-b5ed-05b38f55325a"
	roonaYorionOracle    = "e15504a7-2b67-4185-b916-172145f10b19"
)

func TestRoonaSliceIsRegisteredFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{roonaRoonOracle, roonaBragoOracle, roonaDisplacerOracle, roonaVenserOracle, roonaYorionOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// roonaStolen puts a creature `owner` owns onto the battlefield under
// `controller`'s control.
func roonaStolen(g *game.Game, owner, controller uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: owner, Controller: controller,
	})
}

// roonaOnly is the one battlefield card named `name`, or fails.
func roonaOnly(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	var found []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			found = append(found, c)
		}
	}
	if len(found) != 1 {
		t.Fatalf("%d permanents named %q on the battlefield, want 1", len(found), name)
	}
	return found[0]
}

func roonaCardRef(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Roon of the Hidden Realm --------------------------------------

// {2}, {T}: the creature leaves now and comes back at the next end
// step, as a new object, under its OWNER's control: a creature stolen
// from an opponent goes home.
func TestRoonaRoonBlinksUntilTheEndStepUnderItsOwnersControl(t *testing.T) {
	g, me, opp := spendTable(t)
	roon := pushCatalogPermanent(g, me.ID, "Roon of the Hidden Realm", "Legendary Creature — Rhino Soldier", roonaRoonOracle, false)
	stolen := roonaStolen(g, opp.ID, me.ID, "Stolen Bear")

	apaMana(me, "C", "C")
	p7Activate(t, g, me, roon, 0, game.ActivateAbilityParams{Targets: roonaCardRef(stolen), Strict: true})
	if onBattlefieldByName(g, "Stolen Bear") {
		t.Fatal("the creature is still on the battlefield before the end step")
	}
	if c, _ := battlefieldCard(g, roon); !c.Tapped {
		t.Error("Roon did not tap for the cost")
	}

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	back := roonaOnly(t, g, "Stolen Bear")
	if back.InstanceID == stolen {
		t.Error("the creature came back as the same object (CR 400.7)")
	}
	if back.Controller != opp.ID {
		t.Errorf("returned under %v's control, want its owner %v", back.Controller, opp.ID)
	}
}

// "Another target creature": Roon can't target itself.
func TestRoonaRoonCannotTargetItself(t *testing.T) {
	g, me, _ := spendTable(t)
	roon := pushCatalogPermanent(g, me.ID, "Roon of the Hidden Realm", "Legendary Creature — Rhino Soldier", roonaRoonOracle, false)
	apaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, roon, 0, game.ActivateAbilityParams{Targets: roonaCardRef(roon)}); err == nil {
		t.Fatal("Roon targeted itself")
	}
	if !onBattlefieldByID(g, roon) {
		t.Error("Roon left the battlefield")
	}
}

// --- Eldrazi Displacer ---------------------------------------------

// {2}{C}: only colourless mana pays the {C}, and the creature returns
// at once, tapped, under its owner's control.
func TestRoonaDisplacerBlinksTappedForColorlessMana(t *testing.T) {
	g, me, opp := spendTable(t)
	displacer := pushCatalogPermanent(g, me.ID, "Eldrazi Displacer", "Creature — Eldrazi", roonaDisplacerOracle, true)
	bear := pushVanillaCreature(g, opp.ID, "Bear", 2, 2)

	apaMana(me, "W", "W", "W")
	err := g.ActivateCatalogAbility(me.ID, displacer, 0, game.ActivateAbilityParams{Targets: roonaCardRef(bear), Strict: true})
	if !errors.Is(err, game.ErrInsufficientMana) {
		t.Fatalf("{W}{W}{W} for {2}{C}: err = %v, want insufficient mana", err)
	}
	if !onBattlefieldByID(g, bear) || me.ManaPool.Total() != 3 {
		t.Fatal("a refused activation moved the Bear or spent mana")
	}

	apaMana(me, "C")
	// Summoning sickness does not matter: the cost has no {T}.
	p7Activate(t, g, me, displacer, 0, game.ActivateAbilityParams{Targets: roonaCardRef(bear), Strict: true})
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %d after paying {2}{C} from {W}{W}{W}{C}, want 1", me.ManaPool.Total())
	}
	back := roonaOnly(t, g, "Bear")
	if back.InstanceID == bear {
		t.Error("the Bear came back as the same object (CR 400.7)")
	}
	if !back.Tapped {
		t.Error("the Bear returned untapped")
	}
	if back.Controller != opp.ID {
		t.Errorf("returned under %v's control, want its owner %v", back.Controller, opp.ID)
	}
}

func TestRoonaDisplacerIsColorlessAndCannotTargetItself(t *testing.T) {
	g, me, _ := spendTable(t)
	displacer := pushCatalogPermanent(g, me.ID, "Eldrazi Displacer", "Creature — Eldrazi", roonaDisplacerOracle, false)
	apaMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, displacer, 0, game.ActivateAbilityParams{Targets: roonaCardRef(displacer)}); err == nil {
		t.Fatal("the Displacer targeted itself")
	}
	if got := effectiveColors(t, g, displacer); len(got) != 0 {
		t.Errorf("devoid: colours %v, want none", got)
	}
}

// --- Brago, King Eternal -------------------------------------------

// Combat damage to a player: the chosen nonland permanents leave and
// come back together, each as a new object, a stolen one to its owner;
// lands are not offered, and an untargeted permanent stays.
func TestRoonaBragoBlinksTheChosenPermanentsUnderTheirOwners(t *testing.T) {
	g, me, opp := spendTable(t)
	brago := pushCatalogPermanent(g, me.ID, "Brago, King Eternal", "Legendary Creature — Spirit Noble", roonaBragoOracle, false)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	stolen := roonaStolen(g, opp.ID, me.ID, "Stolen Bear")
	keep := pushVanillaCreature(g, me.ID, "Kept Bear", 2, 2)
	land := pushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	dealCombatDamageToPlayer(g, brago, opp.ID, 2)
	g.RunStateChecksForTest()
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt for Brago's trigger: %+v", g.PendingChoices)
	}
	if hasID(pick.PickTargetCards, land) || hasID(pick.PickTargetCards, theirs) {
		t.Error("a land or an opponent's permanent is offered")
	}
	for _, id := range []uuid.UUID{brago, rock, stolen, keep} {
		if !hasID(pick.PickTargetCards, id) {
			t.Errorf("%v (a nonland permanent you control) is not offered", id)
		}
	}
	pickTargetRefs(t, g, me.ID, nil,
		game.TargetRef{Kind: game.TargetCard, ID: rock},
		game.TargetRef{Kind: game.TargetCard, ID: stolen})
	passPriorityAroundTable(t, g)

	if onBattlefieldByID(g, rock) || onBattlefieldByID(g, stolen) {
		t.Fatal("a target is still the same object (CR 400.7)")
	}
	if c := roonaOnly(t, g, "Rock"); c.Controller != me.ID {
		t.Errorf("the Rock returned under %v, want %v", c.Controller, me.ID)
	}
	if c := roonaOnly(t, g, "Stolen Bear"); c.Controller != opp.ID {
		t.Errorf("the stolen Bear returned under %v, want its owner %v", c.Controller, opp.ID)
	}
	if !onBattlefieldByID(g, keep) || !onBattlefieldByID(g, brago) {
		t.Error("an untargeted permanent was blinked")
	}
}

// Noncombat damage does not trigger Brago.
func TestRoonaBragoIgnoresNoncombatDamage(t *testing.T) {
	g, me, opp := spendTable(t)
	brago := pushCatalogPermanent(g, me.ID, "Brago, King Eternal", "Legendary Creature — Spirit Noble", roonaBragoOracle, false)
	dealNoncombatDamage(g, me.ID, brago, opp.ID)
	g.RunStateChecksForTest()
	if latestPickTarget(g, me.ID) != nil || triggerOnStack(g, brago) != nil {
		t.Error("noncombat damage triggered Brago")
	}
}

// --- Venser, the Sojourner -----------------------------------------

func roonaVenser(t *testing.T, g *game.Game, me *game.Player, loyalty int) uuid.UUID {
	t.Helper()
	v := pushCatalogPermanent(g, me.ID, "Venser, the Sojourner", "Legendary Planeswalker — Venser", roonaVenserOracle, false)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(v, game.CounterLoyalty, loyalty) })
	return v
}

// +2: a permanent you own leaves until the next end step and comes back
// under your control. An opponent's permanent is not a legal target.
func TestRoonaVenserPlusTwoBlinksAPermanentYouOwnUntilTheEndStep(t *testing.T) {
	g, me, opp := spendTable(t)
	v := roonaVenser(t, g, me, 3)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	if err := g.ActivateCatalogAbility(me.ID, v, 0, game.ActivateAbilityParams{Targets: roonaCardRef(theirs)}); err == nil {
		t.Fatal("+2 targeted a permanent you don't own")
	}
	b16Activate(t, g, me.ID, v, 0, game.ActivateAbilityParams{Targets: roonaCardRef(mine)})
	passPriorityAroundTable(t, g)
	if onBattlefieldByName(g, "My Bear") {
		t.Fatal("the Bear is still on the battlefield before the end step")
	}
	if loyaltyOf(g, v) != 5 {
		t.Errorf("loyalty %d after +2, want 5", loyaltyOf(g, v))
	}

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	back := roonaOnly(t, g, "My Bear")
	if back.InstanceID == mine || back.Controller != me.ID {
		t.Errorf("returned as %v under %v, want a new object under %v", back.InstanceID, back.Controller, me.ID)
	}
}

// +2 on Venser itself: it comes back at the end step as a new object
// with its printed loyalty (CR 306.5b), not the 5 it left with.
func TestRoonaVenserPlusTwoOnItselfReturnsWithPrintedLoyalty(t *testing.T) {
	g, me, _ := spendTable(t)
	v := roonaVenser(t, g, me, 3)
	b16Activate(t, g, me.ID, v, 0, game.ActivateAbilityParams{Targets: roonaCardRef(v)})
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, v) {
		t.Fatal("Venser is still on the battlefield")
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	back := roonaOnly(t, g, "Venser, the Sojourner")
	if got := loyaltyOf(g, back.InstanceID); got != 3 {
		t.Errorf("Venser returned with loyalty %d, want 3", got)
	}
}

// −1: every creature, every player's, can't be blocked this turn, one
// that enters after the ability resolved included (CR 611.2c).
func TestRoonaVenserMinusOneMakesEveryCreatureUnblockable(t *testing.T) {
	g, me, opp := spendTable(t)
	v := roonaVenser(t, g, me, 3)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	b16Activate(t, g, me.ID, v, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	late := roonaStolen(g, opp.ID, opp.ID, "Late Bear") // enters through the zone-move listener
	for _, id := range []uuid.UUID{mine, theirs, late} {
		if restrictionsOf(t, g, id)&game.CantBeBlocked == 0 {
			t.Errorf("%v can be blocked", id)
		}
	}
	if loyaltyOf(g, v) != 2 {
		t.Errorf("loyalty %d after −1, want 2", loyaltyOf(g, v))
	}
}

// −8: the emblem exiles a target permanent whenever its owner casts a
// spell, and the trigger targets as it goes on the stack.
func TestRoonaVenserEmblemExilesATargetPermanentOnEachSpell(t *testing.T) {
	g, me, opp := spendTable(t)
	v := roonaVenser(t, g, me, 8)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	b16Activate(t, g, me.ID, v, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("emblems: %d, want 1", me.Emblems.Size())
	}

	castCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	g.RunStateChecksForTest()
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick_target prompt for the emblem: %+v", g.PendingChoices)
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if onBattlefieldByID(g, theirs) || !g.Exile.Contains(theirs) {
		t.Error("the emblem did not exile the target")
	}
}

// An opponent's spell does not trigger the emblem.
func TestRoonaVenserEmblemIgnoresOpponentsSpells(t *testing.T) {
	g, me, opp := spendTable(t)
	v := roonaVenser(t, g, me, 8)
	pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	b16Activate(t, g, me.ID, v, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)

	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventCast, Actor: opp.ID, CardID: uuid.New()})
	})
	g.RunStateChecksForTest()
	if latestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's spell triggered the emblem")
	}
}

// --- Yorion, Sky Nomad ---------------------------------------------

// The ETB offers only other nonland permanents you both own and
// control, exiles the picks, and returns them at the next end step.
func TestRoonaYorionExilesChosenPermanentsUntilTheEndStep(t *testing.T) {
	g, me, opp := spendTable(t)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	keep := pushVanillaCreature(g, me.ID, "Kept Bear", 2, 2)
	land := pushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	stolen := roonaStolen(g, opp.ID, me.ID, "Stolen Bear")
	lent := roonaStolen(g, me.ID, opp.ID, "Lent Bear")

	yorion := castCatalogSpell(t, g, "Yorion, Sky Nomad", "Legendary Creature — Bird Serpent", roonaYorionOracle, nil)
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no choose-cards prompt: %+v", g.PendingChoices)
	}
	if c.PickThen == "" {
		t.Error("the prompt holds a closure, not a keyed continuation")
	}
	for _, id := range []uuid.UUID{land, yorion, stolen, lent} {
		if hasID(c.ChooseCards, id) {
			t.Errorf("%v is offered (a land, Yorion, or not both owned and controlled)", id)
		}
	}
	for _, id := range []uuid.UUID{bear, rock, keep} {
		if !hasID(c.ChooseCards, id) {
			t.Errorf("%v is not offered", id)
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear, rock}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !g.Exile.Contains(bear) || !g.Exile.Contains(rock) {
		t.Fatal("the picks were not exiled")
	}
	if !onBattlefieldByID(g, keep) {
		t.Error("an unpicked permanent was exiled")
	}

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	for _, name := range []string{"My Bear", "Rock"} {
		back := roonaOnly(t, g, name)
		if back.Controller != me.ID {
			t.Errorf("%s returned under %v, want %v", name, back.Controller, me.ID)
		}
	}
}

// A pick that stopped being yours between the question and the answer
// stays where it is (the keyed continuation re-checks).
func TestRoonaYorionSkipsAPickThatChangedControl(t *testing.T) {
	g, me, opp := spendTable(t)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	castCatalogSpell(t, g, "Yorion, Sky Nomad", "Legendary Creature — Bird Serpent", roonaYorionOracle, nil)
	passPriorityAroundTable(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no choose-cards prompt: %+v", g.PendingChoices)
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == bear {
			g.Battlefield.Cards[i].Controller = opp.ID
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !onBattlefieldByID(g, bear) {
		t.Error("a permanent you no longer control was exiled")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("a return was scheduled with nothing exiled: %+v", g.DelayedTriggers)
	}
}

// With no other nonland permanent of yours, Yorion asks nothing.
func TestRoonaYorionWithNothingToExileAsksNothing(t *testing.T) {
	g, me, _ := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	castCatalogSpell(t, g, "Yorion, Sky Nomad", "Legendary Creature — Bird Serpent", roonaYorionOracle, nil)
	passPriorityAroundTable(t, g)
	if c := chooseCardsChoiceFor(g, me.ID); c != nil {
		t.Errorf("a prompt with nothing to choose: %+v", c)
	}
}
