package testscr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

func hasCard(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() { found = g.Battlefield.Contains(id) })
	return found
}

func TestWitchKingMakesEachOpponentSacrificeACreatureThatConnected(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.PushCatalogPermanent(g, me.ID, "Witch-king of Angmar", "Legendary Creature — Wraith Noble", witchKingOfAngmarOracle, false)
	hit1 := effectstest.B12Creature(g, opp1.ID, "Hit Bear", "Creature — Bear", 2, 2)
	hit1b := effectstest.B12Creature(g, opp1.ID, "Hit Ox", "Creature — Ox", 2, 2)
	idle := effectstest.B12Creature(g, opp1.ID, "Idle Bear", "Creature — Bear", 2, 2)
	hit2 := effectstest.B12Creature(g, opp2.ID, "Other Hit Bear", "Creature — Bear", 2, 2)
	idle2 := effectstest.B12Creature(g, opp2.ID, "Other Idle Bear", "Creature — Bear", 2, 2)
	_ = effectstest.B12Creature(g, opp3.ID, "Bystander", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	// Two of opponent one's creatures and one of opponent two's connect
	// in the same damage step: ONE trigger (CR 603.2c), so each
	// opponent sacrifices exactly one creature.
	effectstest.DealCombatDamage(g, opp1.ID, hit1, me.ID)
	effectstest.DealCombatDamage(g, opp1.ID, hit1b, me.ID)
	effectstest.DealCombatDamage(g, opp2.ID, hit2, me.ID)
	effectstest.PassPriorityAroundTable(t, g)

	c1 := effectstest.SacrificeChoiceFor(g, opp1.ID)
	if c1 == nil {
		t.Fatal("opponent one was not asked to sacrifice")
	}
	if len(effectstest.SacrificeOptions(c1)) != 2 {
		t.Errorf("opponent one is offered %d creatures, want only the two that connected", len(effectstest.SacrificeOptions(c1)))
	}
	if effectstest.SacrificeChoiceFor(g, opp3.ID) != nil || effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a seat with no creature that connected, or the controller, was asked")
	}
	effectstest.AnswerSacrifice(t, g, opp1.ID, hit1b)
	effectstest.AnswerSacrifice(t, g, opp2.ID, hit2)

	if hasCard(g, hit1b) || hasCard(g, hit2) {
		t.Error("the chosen creatures were not sacrificed")
	}
	for name, id := range map[string]uuid.UUID{"hit1": hit1, "idle": idle, "idle2": idle2, "mine": mine} {
		if !hasCard(g, id) {
			t.Errorf("%s was sacrificed; one trigger is one edict per opponent", name)
		}
	}
	// The Ring tempts you, after the sacrifices.
	if effectstest.RingPrompt(g, me.ID) == nil {
		t.Fatal("the Ring did not tempt the controller")
	}
	if got := effectstest.RingCount(g, me.ID); got != 1 {
		t.Errorf("Ring temptations = %d, want 1", got)
	}
}

func TestWitchKingIgnoresACreatureThatWasFlickeredAfterConnecting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Witch-king of Angmar", "Legendary Creature — Wraith Noble", witchKingOfAngmarOracle, false)
	hit := effectstest.B12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 2)

	effectstest.DealCombatDamage(g, opp.ID, hit, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Error("the flickered creature is a new object and cannot be chosen")
	}
	if got := effectstest.OnBattlefieldNamed(g, "Hit Bear"); got != 1 {
		t.Errorf("the flickered Bear should still be in play, found %d", got)
	}
	if effectstest.RingPrompt(g, me.ID) == nil && effectstest.RingCount(g, me.ID) != 1 {
		t.Error("the Ring tempts you even when no opponent had a creature to sacrifice")
	}
}

func TestWitchKingDoesNotTriggerOnNoncombatDamageOrDamageToAnotherPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushCatalogPermanent(g, me.ID, "Witch-king of Angmar", "Legendary Creature — Wraith Noble", witchKingOfAngmarOracle, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.DealNoncombatDamage(g, opp.ID, bear, me.ID)
	effectstest.DealCombatDamage(g, opp.ID, bear, third.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil || effectstest.RingCount(g, me.ID) != 0 {
		t.Error("only combat damage to Witch-king's controller triggers it")
	}
}

func TestWitchKingShieldGainsIndestructibleAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wk := effectstest.PushCatalogPermanent(g, me.ID, "Witch-king of Angmar", "Legendary Creature — Wraith Noble", witchKingOfAngmarOracle, false)
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if effectstest.EffectiveHasKeyword(effectstest.RingBFCard(t, g, wk), "indestructible") {
		t.Fatal("setup: already indestructible")
	}
	if err := g.ActivateCatalogAbility(me.ID, wk, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card := effectstest.RingBFCard(t, g, wk)
	if !effectstest.EffectiveHasKeyword(card, "indestructible") {
		t.Error("Witch-king did not gain indestructible")
	}
	if !card.Tapped {
		t.Error("Witch-king was not tapped")
	}
}

func TestReciprocateExilesOnlyACreatureThatDamagedYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hit := effectstest.B12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.DealNoncombatDamage(g, opp.ID, hit, me.ID)

	if err := effectstest.CastCatalogSpellErr(t, g, "Reciprocate", "Instant", reciprocateOracle, effectstest.B16TargetCard(idle)); err == nil {
		t.Fatal("a creature that dealt no damage to you was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Reciprocate", "Instant", reciprocateOracle, effectstest.B16TargetCard(hit))
	effectstest.PassPriorityAroundTable(t, g)
	if hasCard(g, hit) {
		t.Error("Reciprocate left the creature on the battlefield")
	}
	if !hasCard(g, idle) {
		t.Error("Reciprocate hit the wrong creature")
	}
}

func TestRetaliateDestroysEveryCreatureThatDamagedYouAndNothingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	hit := effectstest.B12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 2)
	hit2 := effectstest.B12Creature(g, third.ID, "Other Hit Bear", "Creature — Bear", 2, 2)
	flickered := effectstest.B12Creature(g, opp.ID, "Flickered Bear", "Creature — Bear", 2, 2)
	hitAnother := effectstest.B12Creature(g, opp.ID, "Hit Someone Else", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamage(g, opp.ID, hit, me.ID)
	effectstest.DealNoncombatDamage(g, third.ID, hit2, me.ID)
	effectstest.DealCombatDamage(g, opp.ID, flickered, me.ID)
	effectstest.DealCombatDamage(g, opp.ID, hitAnother, third.ID)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(flickered); err != nil {
			t.Fatal(err)
		}
		if _, err := g.ReturnFromExileToBattlefieldForEffect(flickered, uuid.Nil, false); err != nil {
			t.Fatal(err)
		}
	})

	effectstest.CastCatalogSpell(t, g, "Retaliate", "Instant", retaliateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if hasCard(g, hit) || hasCard(g, hit2) {
		t.Error("a creature that damaged you survived")
	}
	for name, id := range map[string]uuid.UUID{"hitAnother": hitAnother, "idle": idle} {
		if !hasCard(g, id) {
			t.Errorf("%s was destroyed but never damaged you", name)
		}
	}
	if got := effectstest.OnBattlefieldNamed(g, "Flickered Bear"); got != 1 {
		t.Errorf("the flickered Bear is a new object and should survive, found %d", got)
	}
}

func TestSpearOfHeliodIsAnAnthemWhoseAbilityDestroysACreatureThatDamagedYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spear := effectstest.PushCatalogPermanent(g, me.ID, "Spear of Heliod", "Legendary Enchantment Artifact", spearOfHeliodOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	hit := effectstest.B12Creature(g, opp.ID, "Hit Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("anthem: power = %d, want 3", got)
	}
	effectstest.DealCombatDamage(g, opp.ID, hit, me.ID)

	if !hasCard(g, spear) {
		t.Fatal("setup")
	}
	effectstest.ToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{W}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	err := g.ActivateCatalogAbility(me.ID, spear, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(idle)})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a creature that dealt no damage to you: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, spear, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(hit)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if hasCard(g, hit) {
		t.Error("the Spear did not destroy the creature that damaged you")
	}
	if !hasCard(g, idle) {
		t.Error("the Spear destroyed the wrong creature")
	}
}
