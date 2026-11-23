package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_library_cards_test.go — #5429: "manifest the top card of that
// player's library" (game/manifest_library.go) through Thieving Amalgam and
// Orochi Soul-Reaver.

const (
	thievingAmalgamOracle  = "f60cd1d8-eeb6-41a3-af32-3e2a54937ca2"
	orochiSoulReaverOracle = "6d9ab641-ba99-4b27-a84e-8c2ae57a2fc1"
)

// Each opponent's upkeep manifests the top card of THAT player's library
// under the Amalgam's controller; the controller's own upkeep does not.
func TestThievingAmalgamManifestsTheTopCardOfEachOpponentsLibraryAtTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Thieving Amalgam", "Creature — Ape Snake", thievingAmalgamOracle, false)
	for _, seat := range []int{1, 2, 3} {
		opp := g.Seats[seat]
		top := effectstest.MlTop(opp, "Their Bear", "Creature — Bear", "{1}{G}")
		effectstest.AdvanceToUpkeepOf(t, g, seat)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.MlManifested(t, g, top, me, opp)
		if opp.Library.Contains(top) {
			t.Fatalf("seat %d's top card is still in their library", seat)
		}
	}
	mine := effectstest.MlTop(me, "My Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mine); ok {
		t.Fatal("the Amalgam manifested on its controller's own upkeep")
	}
}

// The manifesting player, not the card's owner, may turn a manifested
// creature card face up for its mana cost (CR 701.40b), and a land card
// stays face down.
func TestThievingAmalgamManifestIsTheControllersToTurnFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Thieving Amalgam", "Creature — Ape Snake", thievingAmalgamOracle, false)
	top := effectstest.MlTop(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MlManifested(t, g, top, me, opp)

	effectstest.FloatForTest(g, opp, "CG")
	if err := g.PerformSpecialAction(opp.ID, top, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("the card's owner turned a permanent they do not control face up")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMain(t, g)
	effectstest.MkmMPTurnFaceUp(t, g, me, top, "C", "G")
	if c := effectstest.FindCardAnywhere(t, g, top); c.Controller != me.ID || c.Name != "Their Bear" {
		t.Fatalf("face up it is %q under my control: %v", c.Name, c.Controller == me.ID)
	}
}

// An empty library manifests nothing and is not a draw from it.
func TestThievingAmalgamManifestsNothingFromAnEmptyLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Thieving Amalgam", "Creature — Ape Snake", thievingAmalgamOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	before := len(g.Battlefield.Cards)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Thieving Amalgam", "Creature — Ape Snake", thievingAmalgamOracle, false)

	// Neither of these is "a creature you control but don't own".
	mineID := effectstest.ReqCreature(g, me.ID, "My Bear", "", 2, 2)
	theirsID := effectstest.ReqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	effectstest.KillCreature(t, g, me.ID, mineID)
	effectstest.KillCreature(t, g, opp.ID, theirsID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("life moved to %d/%d for creatures that are not mine-but-theirs", me.Life, opp.Life)
	}

	// A manifested instant of theirs is a creature I control but don't own.
	bolt := effectstest.MlTop(opp, "Lightning Bolt", "Instant", "{R}")
	g.WithWriteLock(func() {
		if _, err := g.ManifestTopOfLibraryForEffect(me.ID, opp.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	effectstest.MlManifested(t, g, bolt, me, opp)
	effectstest.KillCreature(t, g, opp.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+2 || opp.Life != oppLife-2 {
		t.Fatalf("after their manifested card died: me %d (want %d), them %d (want %d)", me.Life, myLife+2, opp.Life, oppLife-2)
	}

	// A token they own under my control: it has ceased to exist by the time
	// the trigger resolves, and its owner still loses the life.
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin",
		TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Owner: opp.ID, Controller: me.ID})
	effectstest.KillCreature(t, g, opp.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+4 || opp.Life != oppLife-4 {
		t.Fatalf("after their token died under my control: me %d (want %d), them %d (want %d)", me.Life, myLife+4, opp.Life, oppLife-4)
	}
}

// One trigger per damaged player however many creatures connect
// (CR 603.2c): a Treasure, and the top card of THAT player's library
// manifested under the Soul-Reaver's controller.
func TestOrochiSoulReaverMakesATreasureAndManifestsFromTheDamagedPlayersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	reaver := effectstest.PushCatalogPermanent(g, me.ID, "Orochi Soul-Reaver", "Creature — Snake Ninja Rogue", orochiSoulReaverOracle, false)
	bear := effectstest.ReqCreature(g, me.ID, "My Bear", "", 2, 2)
	under := effectstest.MlTop(opp, "Second From Top", "Creature — Bear", "{1}{G}")
	top := effectstest.MlTop(opp, "Their Sorcery", "Sorcery", "{2}{B}")
	otherTop := effectstest.MlTop(other, "Other Top", "Creature — Bear", "{1}{G}")

	// Two creatures connect with one opponent in one damage step: one trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: reaver, Target: opp.ID, Amount: 5, Combat: true})
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: bear, Target: opp.ID, Amount: 2, Combat: true})
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures after two creatures hit one player, want 1", n)
	}
	effectstest.MlManifested(t, g, top, me, opp)
	if _, ok := effectstest.BattlefieldCard(g, under); ok {
		t.Fatal("a second card was manifested for the same damage step")
	}
	if _, ok := effectstest.BattlefieldCard(g, otherTop); ok {
		t.Fatal("a player who was not damaged had a card manifested")
	}

	// A manifested sorcery card cannot be turned face up (CR 701.40b).
	effectstest.FloatForTest(g, me, "CCB")
	if err := g.PerformSpecialAction(me.ID, top, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("a manifested sorcery card was turned face up")
	}

	// Noncombat damage is not the trigger.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: reaver, Target: other.ID, Amount: 1})
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, otherTop); ok {
		t.Fatal("noncombat damage manifested a card")
	}
}

// An empty library still makes the Treasure: the two instructions are
// independent (CR 608.2c).
func TestOrochiSoulReaverMakesTheTreasureWhenTheLibraryIsEmpty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	reaver := effectstest.PushCatalogPermanent(g, me.ID, "Orochi Soul-Reaver", "Creature — Snake Ninja Rogue", orochiSoulReaverOracle, false)
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	effectstest.DealCombatDamageToPlayer(g, reaver, opp.ID, 5)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures, want 1", n)
	}
}

// Ninjutsu {3}{B} from hand, and the trigger from real combat damage: the
// Soul-Reaver arrives attacking and its own damage is the trigger.
func TestOrochiSoulReaverNinjutsuThenConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	ninja := effectstest.PushNinjaToHand(me, "Orochi Soul-Reaver", "Creature — Snake Ninja Rogue", orochiSoulReaverOracle, 5, 4)
	top := effectstest.MlTop(opp, "Their Bear", "Creature — Bear", "{1}{G}")
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)

	// {2}{B} floating cannot pay {3}{B}.
	effectstest.FloatForTest(g, me, "CCB")
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{attacker}}); err == nil {
		t.Fatal("ninjutsu {3}{B} was paid with three mana")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, ninja, 0, game.ActivateAbilityParams{Strict: true, ReturnIDs: []uuid.UUID{attacker}}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, ninja); !ok || !c.Tapped || c.AttackingTarget != opp.ID {
		t.Fatalf("the Soul-Reaver is not on the battlefield tapped and attacking: %v", ok)
	}
	life := opp.Life
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Fatalf("the defender is at %d, want %d", opp.Life, life-5)
	}
	effectstest.MlManifested(t, g, top, me, opp)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures, want 1", n)
	}
}
