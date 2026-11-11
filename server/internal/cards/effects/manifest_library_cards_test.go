package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_library_cards_test.go — #5429: "manifest the top card of that
// player's library" (game/manifest_library.go) through Thieving Amalgam and
// Orochi Soul-Reaver.

const (
	thievingAmalgamOracle  = "f60cd1d8-eeb6-41a3-af32-3e2a54937ca2"
	orochiSoulReaverOracle = "6d9ab641-ba99-4b27-a84e-8c2ae57a2fc1"
)

// mlTop puts a card on top of p's library and returns its id.
func mlTop(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 4, Toughness: 4, Owner: p.ID, Controller: p.ID})
	return id
}

// mlManifested asserts id is a manifested 2/2 that `controller` controls,
// `owner` owns, and only `controller` may look at (CR 708.5).
func mlManifested(t *testing.T, g *game.Game, id uuid.UUID, controller, owner *game.Player) {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatal("the card is not on the battlefield")
	}
	if c.FaceDownKind != game.FaceDownManifested || !c.FaceDownIsPermanent() {
		t.Fatalf("face-down kind %q, want manifested", c.FaceDownKind)
	}
	if c.Controller != controller.ID || c.Owner != owner.ID {
		t.Fatalf("controller is the manifesting player: %v, owner is the library's: %v", c.Controller == controller.ID, c.Owner == owner.ID)
	}
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 || !c.IsCreature() {
		t.Errorf("a manifested permanent is a 2/2 creature, got %d/%d creature=%v", c.CurrentPower(), c.CurrentToughness(), c.IsCreature())
	}
	if !c.IsKnownTo(controller.ID) {
		t.Error("the controller cannot look at their own face-down permanent")
	}
	for _, p := range g.Seats {
		if p.ID != controller.ID && c.IsKnownTo(p.ID) {
			t.Errorf("seat %s can look at a face-down permanent another player controls (CR 708.5)", p.Name)
		}
	}
}

// Each opponent's upkeep manifests the top card of THAT player's library
// under the Amalgam's controller; the controller's own upkeep does not.
func TestThievingAmalgamManifestsTheTopCardOfEachOpponentsLibraryAtTheirUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Thieving Amalgam", "Creature — Ape Snake", thievingAmalgamOracle, false)
	for _, seat := range []int{1, 2, 3} {
		opp := g.Seats[seat]
		top := mlTop(opp, "Their Bear", "Creature — Bear", "{1}{G}")
		advanceToUpkeepOf(t, g, seat)
		passPriorityAroundTable(t, g)
		mlManifested(t, g, top, me, opp)
		if opp.Library.Contains(top) {
			t.Fatalf("seat %d's top card is still in their library", seat)
		}
	}
	mine := mlTop(me, "My Bear", "Creature — Bear", "{1}{G}")
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, mine); ok {
		t.Fatal("the Amalgam manifested on its controller's own upkeep")
	}
}

// The manifesting player, not the card's owner, may turn a manifested
// creature card face up for its mana cost (CR 701.40b), and a land card
// stays face down.
func TestThievingAmalgamManifestIsTheControllersToTurnFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Thieving Amalgam", "Creature — Ape Snake", thievingAmalgamOracle, false)
	top := mlTop(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	mlManifested(t, g, top, me, opp)

	floatForTest(g, opp, "CG")
	if err := g.PerformSpecialAction(opp.ID, top, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("the card's owner turned a permanent they do not control face up")
	}
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	advanceToMain(t, g)
	mkmMPTurnFaceUp(t, g, me, top, "C", "G")
	if c := findCardAnywhere(t, g, top); c.Controller != me.ID || c.Name != "Their Bear" {
		t.Fatalf("face up it is %q under my control: %v", c.Name, c.Controller == me.ID)
	}
}

// An empty library manifests nothing and is not a draw from it.
func TestThievingAmalgamManifestsNothingFromAnEmptyLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Thieving Amalgam", "Creature — Ape Snake", thievingAmalgamOracle, false)
	advanceToUpkeepOf(t, g, 1)
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	before := len(g.Battlefield.Cards)
	passPriorityAroundTable(t, g)
	if len(g.Battlefield.Cards) != before {
		t.Fatal("something entered from an empty library")
	}
	if opp.Eliminated {
		t.Fatal("manifesting from an empty library is not a draw")
	}
}

// "Whenever a creature you control but don't own dies, its owner loses 2
// life and you gain 2 life": a manifested card of theirs counts whatever
// card it is, and so does a token they created; my own creature and their
// creature under their control do not.
func TestThievingAmalgamDrainsTheOwnerOfACreatureYouControlButDontOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Thieving Amalgam", "Creature — Ape Snake", thievingAmalgamOracle, false)

	// Neither of these is "a creature you control but don't own".
	mineID := reqCreature(g, me.ID, "My Bear", "", 2, 2)
	theirsID := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	killCreature(t, g, me.ID, mineID)
	killCreature(t, g, opp.ID, theirsID)
	passPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("life moved to %d/%d for creatures that are not mine-but-theirs", me.Life, opp.Life)
	}

	// A manifested instant of theirs is a creature I control but don't own.
	bolt := mlTop(opp, "Lightning Bolt", "Instant", "{R}")
	g.WithWriteLock(func() {
		if _, err := g.ManifestTopOfLibraryForEffect(me.ID, opp.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	mlManifested(t, g, bolt, me, opp)
	killCreature(t, g, opp.ID, bolt)
	passPriorityAroundTable(t, g)
	if me.Life != myLife+2 || opp.Life != oppLife-2 {
		t.Fatalf("after their manifested card died: me %d (want %d), them %d (want %d)", me.Life, myLife+2, opp.Life, oppLife-2)
	}

	// A token they own under my control: it has ceased to exist by the time
	// the trigger resolves, and its owner still loses the life.
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin",
		TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Owner: opp.ID, Controller: me.ID})
	killCreature(t, g, opp.ID, token)
	passPriorityAroundTable(t, g)
	if me.Life != myLife+4 || opp.Life != oppLife-4 {
		t.Fatalf("after their token died under my control: me %d (want %d), them %d (want %d)", me.Life, myLife+4, opp.Life, oppLife-4)
	}
}

// One trigger per damaged player however many creatures connect
// (CR 603.2c): a Treasure, and the top card of THAT player's library
// manifested under the Soul-Reaver's controller.
func TestOrochiSoulReaverMakesATreasureAndManifestsFromTheDamagedPlayersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	advanceToMain(t, g)
	reaver := pushCatalogPermanent(g, me.ID, "Orochi Soul-Reaver", "Creature — Snake Ninja Rogue", orochiSoulReaverOracle, false)
	bear := reqCreature(g, me.ID, "My Bear", "", 2, 2)
	under := mlTop(opp, "Second From Top", "Creature — Bear", "{1}{G}")
	top := mlTop(opp, "Their Sorcery", "Sorcery", "{2}{B}")
	otherTop := mlTop(other, "Other Top", "Creature — Bear", "{1}{G}")

	// Two creatures connect with one opponent in one damage step: one trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: reaver, Target: opp.ID, Amount: 5, Combat: true})
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: bear, Target: opp.ID, Amount: 2, Combat: true})
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures after two creatures hit one player, want 1", n)
	}
	mlManifested(t, g, top, me, opp)
	if _, ok := battlefieldCard(g, under); ok {
		t.Fatal("a second card was manifested for the same damage step")
	}
	if _, ok := battlefieldCard(g, otherTop); ok {
		t.Fatal("a player who was not damaged had a card manifested")
	}

	// A manifested sorcery card cannot be turned face up (CR 701.40b).
	floatForTest(g, me, "CCB")
	if err := g.PerformSpecialAction(me.ID, top, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("a manifested sorcery card was turned face up")
	}

	// Noncombat damage is not the trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: reaver, Target: other.ID, Amount: 1})
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, otherTop); ok {
		t.Fatal("noncombat damage manifested a card")
	}
}

// An empty library still makes the Treasure: the two instructions are
// independent (CR 608.2c).
func TestOrochiSoulReaverMakesTheTreasureWhenTheLibraryIsEmpty(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	reaver := pushCatalogPermanent(g, me.ID, "Orochi Soul-Reaver", "Creature — Snake Ninja Rogue", orochiSoulReaverOracle, false)
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	dealCombatDamageToPlayer(g, reaver, opp.ID, 5)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures, want 1", n)
	}
}

// Ninjutsu {3}{B} from hand, and the trigger from real combat damage: the
// Soul-Reaver arrives attacking and its own damage is the trigger.
func TestOrochiSoulReaverNinjutsuThenConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja := pushNinjaToHand(me, "Orochi Soul-Reaver", "Creature — Snake Ninja Rogue", orochiSoulReaverOracle, 5, 4)
	top := mlTop(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)

	// {2}{B} floating cannot pay {3}{B}.
	floatForTest(g, me, "CCB")
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{attacker}}); err == nil {
		t.Fatal("ninjutsu {3}{B} was paid with three mana")
	}
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{attacker}}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c, ok := battlefieldCard(g, ninja); !ok || !c.Tapped || c.AttackingTarget != opp.ID {
		t.Fatalf("the Soul-Reaver is not on the battlefield tapped and attacking: %v", ok)
	}
	life := opp.Life
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Fatalf("the defender is at %d, want %d", opp.Life, life-5)
	}
	mlManifested(t, g, top, me, opp)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures, want 1", n)
	}
}
