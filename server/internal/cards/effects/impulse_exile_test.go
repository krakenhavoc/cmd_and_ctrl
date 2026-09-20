package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// impulse_exile_test.go — S21 sub-PR 6. Both cards hang off combat
// damage, so the tests declare an attack and let the damage step
// drain the trigger, the same shape combat_triggers_test.go uses.

const (
	ragavanOracle  = "37108cd4-bbab-4ce3-9ed6-f60e8422e703"
	breechesOracle = "eb77f7dc-e9e4-44ef-8616-9f4e737e8ca5"
)

// stackLibrary replaces a player's library with one named card on
// top, so the test knows exactly what an impulse trigger will steal.
func stackLibrary(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	p.Library.Cards = nil
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// exiledPermission returns the impulse grant on a card in exile.
func exiledPermission(g *game.Game, id uuid.UUID) *game.CastPermission {
	if perm := g.CastPermissionOnCardByIDForEffect(id); perm != nil {
		return perm
	}
	return &game.CastPermission{}
}

// permissionLive asks the engine's ONE liveness predicate (#945): the
// permission names this player, its CR 702.185a floor has been
// reached, and its CR 611.2 duration has not run out. Card tests ask
// it rather than reading a window field, so they cannot disagree with
// the cast path about what "live" means.
func permissionLive(g *game.Game, perm *game.CastPermission, player uuid.UUID) bool {
	return g.CastPermissionActiveForEffect(perm, player)
}

func TestRagavanStealsTheTopCardAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := stackLibrary(victim, "Stolen Bolt", "Instant", "{R}")
	ragavan := pushCatalogPermanent(g, me.ID, "Ragavan, Nimble Pilferer",
		"Legendary Creature — Monkey Pirate", ragavanOracle, false)

	attackWith(t, g, victim.ID, ragavan)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(loot) {
		t.Fatalf("Ragavan should have exiled the victim's top card")
	}
	perm := exiledPermission(g, loot)
	if perm.Player != me.ID {
		t.Errorf("permission holder = %v, want Ragavan's controller", perm.Player)
	}
	if !perm.CastOnly {
		t.Errorf(`Ragavan says "you may CAST that card" — CastOnly should be set`)
	}
	if perm.AnyColor {
		t.Errorf("Ragavan grants no colour relaxation")
	}
	if findBattlefieldByName(g, "Treasure") == uuid.Nil {
		t.Errorf("the trigger should also make a Treasure")
	}
	// And the whole point: it's castable right now, by the thief.
	if err := g.CastSpell(me.ID, loot, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Errorf("casting the stolen card: %v", err)
	}
}

// The victim doesn't get to cast their own stolen card back.
func TestStolenCardIsOnlyCastableByTheThief(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := stackLibrary(victim, "Stolen Bolt", "Instant", "{R}")
	ragavan := pushCatalogPermanent(g, me.ID, "Ragavan, Nimble Pilferer",
		"Legendary Creature — Monkey Pirate", ragavanOracle, false)
	attackWith(t, g, victim.ID, ragavan)
	passPriorityAroundTable(t, g)

	if err := g.CastSpell(victim.ID, loot, game.CastSpellParams{FromZone: "exile"}); err != game.ErrNoPlayPermission {
		t.Errorf("owner casting their own stolen card: %v, want ErrNoPlayPermission", err)
	}
}

// Ragavan's grant is cast-only, so a land off the top is a blank.
func TestRagavanStrandsALand(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := stackLibrary(victim, "Stolen Island", "Basic Land — Island", "")
	ragavan := pushCatalogPermanent(g, me.ID, "Ragavan, Nimble Pilferer",
		"Legendary Creature — Monkey Pirate", ragavanOracle, false)
	attackWith(t, g, victim.ID, ragavan)
	passPriorityAroundTable(t, g)

	if err := g.CastSpell(me.ID, loot, game.CastSpellParams{FromZone: "exile"}); err != game.ErrNoPlayPermission {
		t.Errorf("land under a cast-only grant: %v, want ErrNoPlayPermission", err)
	}
}

func TestBreechesGrantsPlayAndAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := stackLibrary(victim, "Stolen Counterspell", "Instant", "{U}{U}")
	breeches := pushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", breechesOracle, false)

	attackWith(t, g, victim.ID, breeches)
	passPriorityAroundTable(t, g)

	perm := exiledPermission(g, loot)
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
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := stackLibrary(victim, "Safe", "Instant", "{U}")
	pushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", breechesOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	attackWith(t, g, victim.ID, bear)
	passPriorityAroundTable(t, g)

	if g.Exile.Contains(loot) {
		t.Errorf("a non-Pirate attacker should not trigger Breeches")
	}
}

// Breeches prints "deal damage", not "deal combat damage" — unlike
// Ragavan, a Pirate's noncombat damage (an ability, not an attack)
// should trigger it too.
func TestBreechesTriggersOnNoncombatPirateDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, victim := g.Seats[0], g.Seats[1]
	loot := stackLibrary(victim, "Safe", "Instant", "{U}")
	pushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", breechesOracle, false)
	pirate := pushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pirate, victim.ID, 1) })
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(loot) {
		t.Errorf("noncombat damage from a Pirate you control should still trigger Breeches")
	}
}

// "exile the top card of EACH of those opponents' libraries": two
// Pirates hitting two DIFFERENT opponents in the same damage batch
// should exile from both, not just one.
func TestBreechesExilesFromEachOpponentDamagedInOneBatch(t *testing.T) {
	g := newCatalogGame(t)
	me, v1, v2 := g.Seats[0], g.Seats[1], g.Seats[2]
	loot1 := stackLibrary(v1, "Safe One", "Instant", "{U}")
	loot2 := stackLibrary(v2, "Safe Two", "Instant", "{U}")
	pushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", breechesOracle, false)
	p1 := pushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)
	p2 := pushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(p1, v1.ID, 1)
		_ = g.DealDamageToPlayerForEffect(p2, v2.ID, 1)
	})
	passPriorityAroundTable(t, g)

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
	g := newCatalogGame(t)
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
	pushCatalogPermanent(g, me.ID, "Breeches, Brazen Plunderer",
		"Legendary Creature — Goblin Pirate", breechesOracle, false)
	p1 := pushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)
	p2 := pushCatalogPermanent(g, me.ID, "Corsair", "Creature — Human Pirate", "", false)

	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(p1, victim.ID, 1)
		_ = g.DealDamageToPlayerForEffect(p2, victim.ID, 1)
	})
	passPriorityAroundTable(t, g)

	if exiled := before - victim.Library.Size(); exiled != 1 {
		t.Errorf("two Pirates hitting the same opponent exiled %d cards, want exactly 1", exiled)
	}
	if !g.Exile.Contains(top) {
		t.Errorf("the one exiled card should be the top of the library")
	}
}
