package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damage_dealers_cards_test.go — #2149: the cards that read "creatures
// that dealt (combat) damage to you this turn".

const (
	witchKingOfAngmarOracle = "2eb1b429-8d11-42f3-8815-5a60b6a4e2d5"
	reciprocateOracle       = "ebdd29c0-2c33-4410-a05c-80ced58c7b81"
	retaliateOracle         = "37f5d022-2989-4b44-9bca-dc6b606afe10"
	spearOfHeliodOracle     = "fd66aa66-75a2-41b6-8d57-dc4b9c221ccb"
)

func dealCombatDamage(g *game.Game, owner, source, victim uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: owner, Source: source, Target: victim, Amount: 2, Combat: true})
	})
}

func dealNoncombatDamage(g *game.Game, owner, source, victim uuid.UUID) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: owner, Source: source, Target: victim, Amount: 2})
	})
}

func hasCard(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() { found = g.Battlefield.Contains(id) })
	return found
}

func TestWitchKingMakesEachOpponentSacrificeACreatureThatConnected(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	pushCatalogPermanent(g, me.ID, "Witch-king of Angmar", "Legendary Creature — Wraith Noble", witchKingOfAngmarOracle, false)
	hit1 := b12Creature(g, opp1.ID, "Hit Bear", "Creature — Bear", 2, 2)
	hit1b := b12Creature(g, opp1.ID, "Hit Ox", "Creature — Ox", 2, 2)
	idle := b12Creature(g, opp1.ID, "Idle Bear", "Creature — Bear", 2, 2)
	hit2 := b12Creature(g, opp2.ID, "Other Hit Bear", "Creature — Bear", 2, 2)
	idle2 := b12Creature(g, opp2.ID, "Other Idle Bear", "Creature — Bear", 2, 2)
	_ = b12Creature(g, opp3.ID, "Bystander", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	// Two of opponent one's creatures and one of opponent two's connect
	// in the same damage step: ONE trigger (CR 603.2c), so each
	// opponent sacrifices exactly one creature.
	dealCombatDamage(g, opp1.ID, hit1, me.ID)
	dealCombatDamage(g, opp1.ID, hit1b, me.ID)
	dealCombatDamage(g, opp2.ID, hit2, me.ID)
	passPriorityAroundTable(t, g)

	c1 := sacrificeChoiceFor(g, opp1.ID)
	if c1 == nil {
		t.Fatal("opponent one was not asked to sacrifice")
	}
	if len(c1.SacrificeOptions) != 2 {
		t.Errorf("opponent one is offered %d creatures, want only the two that connected", len(c1.SacrificeOptions))
	}
	if sacrificeChoiceFor(g, opp3.ID) != nil || sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a seat with no creature that connected, or the controller, was asked")
	}
	answerSacrifice(t, g, opp1.ID, hit1b)
	answerSacrifice(t, g, opp2.ID, hit2)

	if hasCard(g, hit1b) || hasCard(g, hit2) {
		t.Error("the chosen creatures were not sacrificed")
	}
	for name, id := range map[string]uuid.UUID{"hit1": hit1, "idle": idle, "idle2": idle2, "mine": mine} {
		if !hasCard(g, id) {
			t.Errorf("%s was sacrificed; one trigger is one edict per opponent", name)
		}
	}
	// The Ring tempts you, after the sacrifices.
	if ringPrompt(g, me.ID) == nil {
		t.Fatal("the Ring did not tempt the controller")
	}
	if got := ringCount(g, me.ID); got != 1 {
		t.Errorf("Ring temptations = %d, want 1", got)
	}
}

func TestWitchKingIgnoresACreatureThatWasFlickeredAfterConnecting(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Witch-king of Angmar", "Legendary Creature — Wraith Noble", witchKingOfAngmarOracle, false)
	hit := b12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 2)

	dealCombatDamage(g, opp.ID, hit, me.ID)
	// In response to the trigger, the opponent flickers it: CR 400.7,
	// it is a new object that did not deal the damage.
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(hit); err != nil {
			t.Fatal(err)
		}
		if _, err := g.ReturnFromExileToBattlefieldForEffect(hit, uuid.Nil, false); err != nil {
			t.Fatal(err)
		}
	})
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, opp.ID) != nil {
		t.Error("the flickered creature is a new object and cannot be chosen")
	}
	if got := onBattlefieldNamed(g, "Hit Bear"); got != 1 {
		t.Errorf("the flickered Bear should still be in play, found %d", got)
	}
	if ringPrompt(g, me.ID) == nil && ringCount(g, me.ID) != 1 {
		t.Error("the Ring tempts you even when no opponent had a creature to sacrifice")
	}
}

func TestWitchKingDoesNotTriggerOnNoncombatDamageOrDamageToAnotherPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Witch-king of Angmar", "Legendary Creature — Wraith Noble", witchKingOfAngmarOracle, false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	dealNoncombatDamage(g, opp.ID, bear, me.ID)
	dealCombatDamage(g, opp.ID, bear, third.ID)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, opp.ID) != nil || ringCount(g, me.ID) != 0 {
		t.Error("only combat damage to Witch-king's controller triggers it")
	}
}

func TestWitchKingShieldGainsIndestructibleAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	wk := pushCatalogPermanent(g, me.ID, "Witch-king of Angmar", "Legendary Creature — Wraith Noble", witchKingOfAngmarOracle, false)
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if effectiveHasKeyword(ringBFCard(t, g, wk), "indestructible") {
		t.Fatal("setup: already indestructible")
	}
	if err := g.ActivateCatalogAbility(me.ID, wk, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	card := ringBFCard(t, g, wk)
	if !effectiveHasKeyword(card, "indestructible") {
		t.Error("Witch-king did not gain indestructible")
	}
	if !card.Tapped {
		t.Error("Witch-king was not tapped")
	}
}

func effectiveHasKeyword(c *game.Card, kw string) bool {
	for _, a := range c.Effective().Abilities {
		if a == kw {
			return true
		}
	}
	return false
}

func TestReciprocateExilesOnlyACreatureThatDamagedYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hit := b12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	dealNoncombatDamage(g, opp.ID, hit, me.ID)

	if err := castCatalogSpellErr(t, g, "Reciprocate", "Instant", reciprocateOracle, b16TargetCard(idle)); err == nil {
		t.Fatal("a creature that dealt no damage to you was a legal target")
	}
	castCatalogSpell(t, g, "Reciprocate", "Instant", reciprocateOracle, b16TargetCard(hit))
	passPriorityAroundTable(t, g)
	if hasCard(g, hit) {
		t.Error("Reciprocate left the creature on the battlefield")
	}
	if !hasCard(g, idle) {
		t.Error("Reciprocate hit the wrong creature")
	}
}

func TestRetaliateDestroysEveryCreatureThatDamagedYouAndNothingElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	hit := b12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 2)
	hit2 := b12Creature(g, third.ID, "Other Hit Bear", "Creature — Bear", 2, 2)
	flickered := b12Creature(g, opp.ID, "Flickered Bear", "Creature — Bear", 2, 2)
	hitAnother := b12Creature(g, opp.ID, "Hit Someone Else", "Creature — Bear", 2, 2)
	idle := b12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	dealCombatDamage(g, opp.ID, hit, me.ID)
	dealNoncombatDamage(g, third.ID, hit2, me.ID)
	dealCombatDamage(g, opp.ID, flickered, me.ID)
	dealCombatDamage(g, opp.ID, hitAnother, third.ID)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(flickered); err != nil {
			t.Fatal(err)
		}
		if _, err := g.ReturnFromExileToBattlefieldForEffect(flickered, uuid.Nil, false); err != nil {
			t.Fatal(err)
		}
	})

	castCatalogSpell(t, g, "Retaliate", "Instant", retaliateOracle, nil)
	passPriorityAroundTable(t, g)
	if hasCard(g, hit) || hasCard(g, hit2) {
		t.Error("a creature that damaged you survived")
	}
	for name, id := range map[string]uuid.UUID{"hitAnother": hitAnother, "idle": idle} {
		if !hasCard(g, id) {
			t.Errorf("%s was destroyed but never damaged you", name)
		}
	}
	if got := onBattlefieldNamed(g, "Flickered Bear"); got != 1 {
		t.Errorf("the flickered Bear is a new object and should survive, found %d", got)
	}
}

func TestSpearOfHeliodIsAnAnthemWhoseAbilityDestroysACreatureThatDamagedYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spear := pushCatalogPermanent(g, me.ID, "Spear of Heliod", "Legendary Enchantment Artifact", spearOfHeliodOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hit := b12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("anthem: power = %d, want 3", got)
	}
	dealCombatDamage(g, opp.ID, hit, me.ID)

	if !hasCard(g, spear) {
		t.Fatal("setup")
	}
	toMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	err := g.ActivateCatalogAbility(me.ID, spear, 0, game.ActivateAbilityParams{Targets: b16TargetCard(idle)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a creature that dealt no damage to you: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, spear, 0, game.ActivateAbilityParams{Targets: b16TargetCard(hit)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if hasCard(g, hit) {
		t.Error("the Spear did not destroy the creature that damaged you")
	}
	if !hasCard(g, idle) {
		t.Error("the Spear destroyed the wrong creature")
	}
}
