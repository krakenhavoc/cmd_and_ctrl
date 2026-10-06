package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cant_cause_discard_sacrifice_test.go — #2178: "spells and abilities
// your opponents control can't cause you to discard cards / sacrifice
// permanents" (Tajuru Preserver, Sigarda, Host of Herons, Tamiyo,
// Collector of Tales).

const (
	tajuruPreserverOracle = "d12d4b54-a13a-46ba-b176-3aaa453ce3e2"
	sigardaHostOracle     = "e55104e2-4900-48de-b288-d3e6abd5e09e"
	tamiyoCollectorOracle = "75d56a0a-2f64-4e80-b83e-85942d3e6dd7"
)

func pushTajuru(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Tajuru Preserver", "Creature — Elf Shaman", tajuruPreserverOracle, false)
}

// An opponent's edict skips the protected player and still asks the
// other opponents (and its own controller, for "each player").
func TestOpponentsEdictSkipsProtectedPlayerOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, shielded, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushTajuru(g, shielded.ID)
	pushCatalogPermanent(g, shielded.ID, "Shielded Bear", "Creature — Bear", "", false)
	otherBear := pushCatalogPermanent(g, other.ID, "Other Bear", "Creature — Bear", "", false)
	mine := pushCatalogPermanent(g, me.ID, "My Bear", "Creature — Bear", "", false)

	castCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", fleshbagMarauderOracle, nil)
	passPriorityAroundTable(t, g)

	if c := sacrificeChoiceFor(g, shielded.ID); c != nil {
		t.Fatalf("the protected player was asked to sacrifice: %+v", c)
	}
	if sacrificeChoiceFor(g, other.ID) == nil || sacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("the other opponent and the controller are still asked")
	}
	answerSacrifice(t, g, other.ID, otherBear)
	answerSacrifice(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if n := annCountOn(g, shielded.ID); n != 2 {
		t.Errorf("the protected player still controls %d permanents, want 2", n)
	}
}

// Your own edict still bites you: the clause covers OPPONENTS' spells.
func TestOwnEdictStillMakesYouSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTajuru(g, me.ID)
	castCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", fleshbagMarauderOracle, nil)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, me.ID) == nil {
		t.Fatal("Fleshbag's controller is protected from their OWN spell and ability")
	}
}

// The protection leaving the battlefield ends it.
func TestProtectionEndsWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	shielded := g.Seats[1]
	preserver := pushTajuru(g, shielded.ID)
	pushCatalogPermanent(g, shielded.ID, "Shielded Bear", "Creature — Bear", "", false)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(preserver) })
	passPriorityAroundTable(t, g)

	castCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", fleshbagMarauderOracle, nil)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, shielded.ID) == nil {
		t.Fatal("with the Preserver gone the player must be asked again")
	}
}

// A sacrifice paid as a cost is the player's own choice and is never
// blocked, even while an opponent's spell is the last thing that
// resolved.
func TestSacrificeCostIsNeverBlocked(t *testing.T) {
	g := newCatalogGame(t)
	shielded := g.Seats[1]
	pushTajuru(g, shielded.ID)
	bomb := pushCatalogPermanent(g, shielded.ID, "Goblin Bombardment", "Enchantment", goblinBombardmentOracle, false)
	fodder := pushCatalogPermanent(g, shielded.ID, "Fodder", "Creature — Goblin", "", false)

	// An opponent's spell resolves first.
	castCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", fleshbagMarauderOracle, nil)
	passPriorityAroundTable(t, g)

	if err := g.ActivateCatalogAbility(shielded.ID, bomb, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[0].ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice cost was blocked by the protection")
	}
}

// "Each opponent discards a card" from an opponent skips the player
// holding Tamiyo; the others are asked.
func TestOpponentsEachPlayerDiscardsSkipsProtectedPlayer(t *testing.T) {
	g := newCatalogGame(t)
	shielded := g.Seats[1]
	pushCatalogWalker(g, shielded.ID, "Tamiyo, Collector of Tales", tamiyoCollectorOracle, 5)

	castAndResolveCreature(t, g, "Burglar Rat", "Creature — Rat", "2f807301-37df-4724-871a-08e3512b07b3")
	passPriorityAroundTable(t, g)

	if n := discardOwed(g, shielded.ID); n != 0 {
		t.Errorf("the protected player owes %d discards", n)
	}
	for _, p := range g.Seats[2:] {
		if discardOwed(g, p.ID) != 1 {
			t.Errorf("seat %s should still owe one discard", p.ID)
		}
	}
}

// Annihilator from an opponent's attacker is stopped.
func TestAnnihilatorFromOpponentIsStopped(t *testing.T) {
	g := newCatalogGame(t)
	me, shielded := g.Seats[0], g.Seats[1]
	eldrazi := pushAnnihilatorCreature(g, me.ID, "annihilator 3")
	pushTajuru(g, shielded.ID)
	pushAnnFodder(g, shielded.ID, 5)

	declareAttack(t, g, shielded.ID, eldrazi)
	asked := annSettle(t, g, nil)
	if len(asked) != 0 {
		t.Fatalf("the protected defender was asked to sacrifice: %d prompts", len(asked))
	}
	if n := annCountOn(g, shielded.ID); n != 6 {
		t.Errorf("defender controls %d permanents, want all 6", n)
	}
}

// Sacrifice-all sweeps skip the protected player's permanents only.
func TestSacrificeAllSkipsProtectedPlayersPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, shielded, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushTajuru(g, shielded.ID)
	green := func(owner uuid.UUID) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
			Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		})
	}
	theirs, others, mine := green(shielded.ID), green(other.ID), green(me.ID)

	castCatalogSpell(t, g, "All Is Dust", "Kindred Sorcery — Eldrazi", b08AllIsDustOracle, nil)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(theirs) {
		t.Error("the protected player's permanent was sacrificed")
	}
	if g.Battlefield.Contains(others) || g.Battlefield.Contains(mine) {
		t.Error("everyone else's coloured permanents are still sacrificed")
	}
}

// --- one test per card ---------------------------------------------

func TestTajuruPreserverIsRegisteredAndComplete(t *testing.T) {
	spec, ok := Lookup(tajuruPreserverOracle)
	if !ok || spec.Completeness != CompletenessFull || !spec.OpponentEffectProtections[0].Sacrifice || spec.OpponentEffectProtections[0].Discard {
		t.Fatalf("Tajuru Preserver: %+v ok=%v", spec, ok)
	}
}

func TestSigardaHostOfHeronsProtectsFromSacrificeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, shielded := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, shielded.ID, "Sigarda, Host of Herons", "Legendary Creature — Angel", sigardaHostOracle, false)
	_ = me
	castCatalogSpell(t, g, "Fleshbag Marauder", "Creature — Zombie Warrior", fleshbagMarauderOracle, nil)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, shielded.ID) != nil {
		t.Error("Sigarda's controller was asked to sacrifice")
	}
	spec, _ := Lookup(sigardaHostOracle)
	if spec.Completeness != CompletenessFull {
		t.Errorf("Sigarda is %v, the clause is live so it ships complete", spec.Completeness)
	}
	// Sigarda says nothing about discarding.
	if spec.OpponentEffectProtections[0].Discard {
		t.Error("Sigarda protects from discard; it prints only sacrifice")
	}
}

// Tamiyo +1: name a card, reveal four, matching nonland cards to hand,
// the rest (including lands named in error) to the graveyard.
func TestTamiyoPlusOneNamesRevealsAndTakes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tamiyo := pushCatalogWalker(g, me.ID, "Tamiyo, Collector of Tales", tamiyoCollectorOracle, 5)
	deep := pushLibraryCardForTest(me, game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	// Pushed last = top of the library.
	for _, c := range []game.Card{
		{Name: "Sol Ring", TypeLine: "Artifact"},
		{Name: "Grizzly Bears", TypeLine: "Creature — Bear"},
		{Name: "Sol Ring", TypeLine: "Artifact"},
		{Name: "Forest", TypeLine: "Basic Land — Forest"},
	} {
		c.Owner, c.Controller, c.InstanceID = me.ID, me.ID, uuid.New()
		me.Library.PushTop(c)
	}
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	hand := len(me.Hand.Cards)
	grave := len(me.Graveyard.Cards)

	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	actNameFor(t, g, me.ID, "sol ring")
	passPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards) - hand; got != 2 {
		t.Errorf("two Sol Rings in the top four go to hand, got %d", got)
	}
	if got := len(me.Graveyard.Cards) - grave; got != 2 {
		t.Errorf("the other two of the four go to the graveyard, got %d", got)
	}
	if me.Library.Size() == 0 || me.Library.Cards[0].InstanceID != deep {
		t.Error("the fifth card is not revealed and stays in the library")
	}
	if loyaltyCount(g, tamiyo) != 6 {
		t.Errorf("loyalty = %d, want 6", loyaltyCount(g, tamiyo))
	}
}

// Naming a land takes nothing: the clause says nonland.
func TestTamiyoPlusOneNamingALandTakesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tamiyo := pushCatalogWalker(g, me.ID, "Tamiyo, Collector of Tales", tamiyoCollectorOracle, 5)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	hand := len(me.Hand.Cards)
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	actNameFor(t, g, me.ID, "Forest")
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("a land was taken to hand by naming it")
	}
}

// Tamiyo −3: return target card from your graveyard to your hand.
func TestTamiyoMinusThreeReturnsACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tamiyo := pushCatalogWalker(g, me.ID, "Tamiyo, Collector of Tales", tamiyoCollectorOracle, 5)
	card := pushCatalogGraveyardCard(me, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	advanceToMainOf(t, g, g.Turn.ActiveSeat)
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: card}},
	}); err != nil {
		t.Fatalf("−3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) || me.Graveyard.Contains(card) {
		t.Error("the targeted card did not return to hand")
	}
	if loyaltyCount(g, tamiyo) != 2 {
		t.Errorf("loyalty = %d, want 2", loyaltyCount(g, tamiyo))
	}
}

// Tamiyo's static covers discard and sacrifice, against opponents only.
func TestTamiyoStaticDeclaresBothClauses(t *testing.T) {
	spec, ok := Lookup(tamiyoCollectorOracle)
	if !ok || spec.Completeness != CompletenessFull {
		t.Fatalf("Tamiyo: ok=%v %+v", ok, spec.Completeness)
	}
	p := spec.OpponentEffectProtections[0]
	if !p.Discard || !p.Sacrifice {
		t.Errorf("Tamiyo's clause = %+v, want discard and sacrifice", p)
	}
}
