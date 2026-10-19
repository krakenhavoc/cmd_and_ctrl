package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roonb_cards2_test.go — #4128 slice b, second half: Lavinia of the
// Tenth and Reflector Mage.

// --- Lavinia of the Tenth -------------------------------------------

const roonbSolRingOracle = "6ad8011d-3471-4369-9d68-b264cc027487"

// roonbCastLavinia plays Lavinia for seat 0 and settles her trigger.
func roonbCastLavinia(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	id := castAndResolveCreature(t, g, "Lavinia of the Tenth", "Legendary Creature — Human Soldier", roonbLaviniaOracle)
	passPriorityAroundTable(t, g)
	return id
}

func roonbMV(g *game.Game, controller uuid.UUID, name, typeLine, cost string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: controller, Controller: controller,
	})
}

const roonbDetained = game.CantAttack | game.CantBlock | game.CantActivate | game.CantActivateMana

// Exactly the nonland permanents opponents control with mana value 4 or
// less are detained: a five-drop, a land and her controller's own
// permanents are not.
func TestRoonbLaviniaDetainsOnlyOpponentsCheapNonlands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	four := roonbMV(g, opp.ID, "Four", "Creature — Bear", "{2}{2}")
	five := roonbMV(g, opp.ID, "Five", "Creature — Bear", "{3}{2}")
	rock := roonbMV(g, g.Seats[2].ID, "Rock", "Artifact", "{1}")
	land := roonbMV(g, opp.ID, "Island", "Basic Land — Island", "")
	mine := roonbMV(g, me.ID, "Mine", "Creature — Bear", "{1}")
	roonbCastLavinia(t, g)

	for _, id := range []uuid.UUID{four, rock} {
		assertRestrictions(t, g, id, roonbDetained)
	}
	for _, id := range []uuid.UUID{five, land, mine} {
		assertRestrictions(t, g, id, 0)
	}
}

// The set is the permanents there as the trigger RESOLVES, and no
// others: one that arrives afterwards is not detained (CR 701.35a).
func TestRoonbLaviniaDoesNotDetainLaterArrivals(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	roonbCastLavinia(t, g)
	late := roonbMV(g, opp.ID, "Late", "Creature — Bear", "{1}")
	assertRestrictions(t, g, late, 0)
}

// A detained creature can't attack on its controller's turn, and then
// everything is back when Lavinia's controller's next turn begins.
func TestRoonbLaviniaDetainEndsAsHerControllersNextTurnBegins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := roonbMV(g, opp.ID, "Bear", "Creature — Bear", "{1}")
	roonbCastLavinia(t, g)

	advanceToStepOf(t, g, 1, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, me.ID); err == nil {
		t.Fatal("a detained creature attacked")
	}
	assertRestrictions(t, g, bear, roonbDetained)

	advanceToUpkeepOf(t, g, 0)
	assertRestrictions(t, g, bear, 0)
}

// A detained permanent's activated abilities, mana abilities included,
// can't be activated.
func TestRoonbLaviniaStopsActivatedAndManaAbilities(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[2]
	ring := pushCatalogPermanent(g, opp.ID, "Sol Ring", "Artifact", roonbSolRingOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ring {
				g.Battlefield.Cards[i].ManaCost = "{1}"
			}
		}
	})
	roonbCastLavinia(t, g)
	err := g.ActivateManaAbility(opp.ID, ring, 0, game.ManaAbilityParams{})
	if !errors.Is(err, game.ErrCantActivate) {
		t.Fatalf("a detained Sol Ring tapped for mana: %v", err)
	}
}

func TestRoonbLaviniaHasProtectionFromRed(t *testing.T) {
	g := newCatalogGame(t)
	id := roonbCastLavinia(t, g)
	if !hasEffectiveKeyword(t, g, id, "protection from red") {
		t.Error("Lavinia lost protection from red")
	}
}

// --- Reflector Mage -------------------------------------------------

func roonbCastReflectorMage(t *testing.T, g *game.Game, victim uuid.UUID) uuid.UUID {
	t.Helper()
	id := castAndResolveCreature(t, g, "Reflector Mage", "Creature — Human Wizard", roonbReflectOracle)
	pickCard(t, g, g.Seats[0].ID, victim)
	passPriorityAroundTable(t, g)
	return id
}

// roonbTryCast has `p` try to cast a card of that name and type from
// hand and returns the cast's error.
func roonbTryCast(g *game.Game, p *game.Player, name, typeLine string) error {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return g.CastSpell(p.ID, id, game.CastSpellParams{})
}

func TestRoonbReflectorMageBouncesAndBansTheNameUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := roonbCreature(g, opp.ID, "Runeclaw Bear", 2, 2)
	roonbCastReflectorMage(t, g, victim)

	if g.Battlefield.Contains(victim) || !opp.Hand.Contains(victim) {
		t.Fatal("the creature did not return to its owner's hand")
	}
	// An instant keeps the timing gate out of it, so only the name can
	// be the refusal.
	if err := roonbTryCast(g, opp, "Runeclaw Bear", "Instant"); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("casting a spell named like the bounced creature: %v, want ErrCantCast", err)
	}
	if err := roonbTryCast(g, opp, "Some Other Spell", "Instant"); err != nil {
		t.Fatalf("a spell with another name is not banned: %v", err)
	}
	// The creature itself, back in hand, has the same name.
	advanceToMainOf(t, g, 1)
	if err := g.CastSpell(opp.ID, victim, game.CastSpellParams{}); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("recasting the bounced creature: %v, want ErrCantCast", err)
	}

	// Until Reflector Mage's controller's next turn: it ends as that
	// turn begins.
	advanceToUpkeepOf(t, g, 0)
	if err := roonbTryCast(g, opp, "Runeclaw Bear", "Instant"); err != nil {
		t.Fatalf("the ban should be over once the next turn has begun: %v", err)
	}
}

// The ban falls on the creature's OWNER, not its controller.
func TestRoonbReflectorMageBansTheOwnerNotTheController(t *testing.T) {
	g := newCatalogGame(t)
	owner, thief := g.Seats[1], g.Seats[2]
	stolen := roonbCreature(g, owner.ID, "Stolen Bear", 2, 2)
	g.WithWriteLock(func() {
		g.GainControlForEffect(uuid.Nil, stolen, thief.ID, game.IndefiniteDuration(), "stolen")
	})
	roonbCastReflectorMage(t, g, stolen)

	if err := roonbTryCast(g, owner, "Stolen Bear", "Instant"); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("the owner should be banned: %v", err)
	}
	if err := roonbTryCast(g, thief, "Stolen Bear", "Instant"); err != nil {
		t.Fatalf("the controller is not the one banned: %v", err)
	}
}

// The name is the creature's CURRENT name: a Clone copying a Bear bans
// "Grizzly Bears", not "Clone".
func TestRoonbReflectorMageBansTheCopiedName(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	seedCopyableCreature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	advanceToMainOf(t, g, 1)
	cloneID := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: cloneID, Name: "Clone", TypeLine: "Creature — Shapeshifter",
		OracleID: oracleClone, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, cloneID, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	var bears uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Grizzly Bears" {
			bears = c.InstanceID
		}
	}
	resolveWithCopyChoice(t, g, bears)
	if got := copyBattlefieldCard(t, g, cloneID); got.Name != "Grizzly Bears" {
		t.Fatalf("the Clone is %q, want it to be a copy of Grizzly Bears", got.Name)
	}

	advanceToMainOf(t, g, 0)
	roonbCastReflectorMage(t, g, cloneID)
	if err := roonbTryCast(g, opp, "Grizzly Bears", "Instant"); !errors.Is(err, game.ErrCantCast) {
		t.Fatalf("the copied name should be banned: %v", err)
	}
	if err := roonbTryCast(g, opp, "Clone", "Instant"); err != nil {
		t.Fatalf("the Clone's printed name is not the creature's name: %v", err)
	}
}

// A target that left in response means the whole ability does nothing,
// ban included (CR 608.2b).
func TestRoonbReflectorMageWithAnIllegalTargetBansNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	victim := roonbCreature(g, opp.ID, "Runeclaw Bear", 2, 2)
	castAndResolveCreature(t, g, "Reflector Mage", "Creature — Human Wizard", roonbReflectOracle)
	pickCard(t, g, g.Seats[0].ID, victim)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(victim); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if err := roonbTryCast(g, opp, "Runeclaw Bear", "Instant"); err != nil {
		t.Fatalf("the ability fizzled; nothing should be banned: %v", err)
	}
}
