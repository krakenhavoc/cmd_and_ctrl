package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestRagavanStealsTheTopCardAndMakesATreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := effectstest.StackLibrary(victim, "Stolen Bolt", "Instant", "{R}")
	ragavan := effectstest.PushCatalogPermanent(g, me.ID, "Ragavan, Nimble Pilferer",
		"Legendary Creature — Monkey Pirate", effectstest.RagavanOracle, false)

	effectstest.AttackWith(t, g, victim.ID, ragavan)
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(loot) {
		t.Fatalf("Ragavan should have exiled the victim's top card")
	}
	perm := effectstest.ExiledPermission(g, loot)
	if perm.Player != me.ID {
		t.Errorf("permission holder = %v, want Ragavan's controller", perm.Player)
	}
	if !perm.CastOnly {
		t.Errorf(`Ragavan says "you may CAST that card" — CastOnly should be set`)
	}
	if perm.AnyColor {
		t.Errorf("Ragavan grants no colour relaxation")
	}
	if effectstest.FindBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Errorf("the trigger should also make a Treasure")
	}
	// And the whole point: it's castable right now, by the thief.
	if err := g.CastSpell(me.ID, loot, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Errorf("casting the stolen card: %v", err)
	}
}

// The victim doesn't get to cast their own stolen card back.
func TestStolenCardIsOnlyCastableByTheThief(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := effectstest.StackLibrary(victim, "Stolen Bolt", "Instant", "{R}")
	ragavan := effectstest.PushCatalogPermanent(g, me.ID, "Ragavan, Nimble Pilferer",
		"Legendary Creature — Monkey Pirate", effectstest.RagavanOracle, false)
	effectstest.AttackWith(t, g, victim.ID, ragavan)
	effectstest.PassPriorityAroundTable(t, g)

	if err := g.CastSpell(victim.ID, loot, game.CastSpellParams{FromZone: "exile"}); err != game.ErrNoPlayPermission {
		t.Errorf("owner casting their own stolen card: %v, want ErrNoPlayPermission", err)
	}
}

// Ragavan's grant is cast-only, so a land off the top is a blank.
func TestRagavanStrandsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := effectstest.StackLibrary(victim, "Stolen Island", "Basic Land — Island", "")
	ragavan := effectstest.PushCatalogPermanent(g, me.ID, "Ragavan, Nimble Pilferer",
		"Legendary Creature — Monkey Pirate", effectstest.RagavanOracle, false)
	effectstest.AttackWith(t, g, victim.ID, ragavan)
	effectstest.PassPriorityAroundTable(t, g)

	if err := g.CastSpell(me.ID, loot, game.CastSpellParams{FromZone: "exile"}); err != game.ErrNoPlayPermission {
		t.Errorf("land under a cast-only grant: %v, want ErrNoPlayPermission", err)
	}
}

func TestBreechesGrantsPlayAndAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := effectstest.StackLibrary(victim, "Stolen Counterspell", "Instant", "{U}{U}")
	breeches := effectstest.PushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", effectstest.BreechesOracle, false)

	effectstest.AttackWith(t, g, victim.ID, breeches)
	effectstest.PassPriorityAroundTable(t, g)

	perm := effectstest.ExiledPermission(g, loot)
	if perm.Player != me.ID {
		t.Fatalf("Breeches should have exiled and granted: %+v", perm)
	}
	if perm.CastOnly {
		t.Errorf(`Breeches says "you may PLAY those cards" — CastOnly should be clear`)
	}
	if !perm.AnyColor {
		t.Errorf("Breeches lets you spend mana as though it were any colour")
	}
	// The relaxation is the point: a mono-red pool casts the blue
	// card under the strict gate.
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "R"})
	if err := g.CastSpell(me.ID, loot, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Errorf("any-colour cast under strict mana: %v", err)
	}
}

// Breeches only cares about damage from Pirates.
func TestBreechesIgnoresNonPirates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := effectstest.StackLibrary(victim, "Safe", "Instant", "{U}")
	effectstest.PushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", effectstest.BreechesOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.AttackWith(t, g, victim.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Exile.Contains(loot) {
		t.Errorf("a non-Pirate attacker should not trigger Breeches")
	}
}

// Breeches prints "deal damage", not "deal combat damage" — unlike
// Ragavan, a Pirate's noncombat damage (an ability, not an attack)
// should trigger it too.
func TestBreechesTriggersOnNoncombatPirateDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := effectstest.StackLibrary(victim, "Safe", "Instant", "{U}")
	effectstest.PushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", effectstest.BreechesOracle, false)
	pirate := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pirate, victim.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(loot) {
		t.Errorf("noncombat damage from a Pirate you control should still trigger Breeches")
	}
}

// "exile the top card of EACH of those opponents' libraries": two
// Pirates hitting two DIFFERENT opponents in the same damage batch
// should exile from both, not just one.
func TestBreechesExilesFromEachOpponentDamagedInOneBatch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, v1, v2 := g.Seats[0], g.Seats[1], g.Seats[2]
	loot1 := effectstest.StackLibrary(v1, "Safe One", "Instant", "{U}")
	loot2 := effectstest.StackLibrary(v2, "Safe Two", "Instant", "{U}")
	effectstest.PushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", effectstest.BreechesOracle, false)
	p1 := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)
	p2 := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(p1, v1.ID, 1)
		_ = g.DealDamageToPlayerForEffect(p2, v2.ID, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)

	if !g.Exile.Contains(loot1) {
		t.Errorf("opponent 1's top card should have been exiled")
	}
	if !g.Exile.Contains(loot2) {
		t.Errorf("opponent 2's top card should have been exiled")
	}
}

// Two Pirates hitting the SAME opponent in one batch collapse to one
// trigger (CR 603.2c, #784) and so exile exactly one card, not two.
func TestBreechesCollapsesTwoPiratesHittingTheSameOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	victim.Library.Cards = nil
	// Pushed bottom-to-top, so "Top Card" ends up on top.
	victim.Library.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Second Card", TypeLine: "Instant",
		Owner: victim.ID, Controller: victim.ID,
	})
	top := uuid.New()
	victim.Library.PushTop(game.Card{
		InstanceID: top, Name: "Top Card", TypeLine: "Instant",
		Owner: victim.ID, Controller: victim.ID,
	})
	before := victim.Library.Size()
	effectstest.PushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", effectstest.BreechesOracle, false)
	p1 := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)
	p2 := effectstest.PushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(p1, victim.ID, 1)
		_ = g.DealDamageToPlayerForEffect(p2, victim.ID, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)

	if exiled := before - victim.Library.Size(); exiled != 1 {
		t.Errorf("two Pirates hitting the same opponent exiled %d cards, want exactly 1", exiled)
	}
	if !g.Exile.Contains(top) {
		t.Errorf("the one exiled card should be the top of the library")
	}
}
