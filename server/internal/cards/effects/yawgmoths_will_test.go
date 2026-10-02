package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	yawgmothsWillOracle    = "322f0459-f394-44f0-977b-55fd0cbe0712"
	gaeasWillOracle        = "87efff06-b6cb-4a8f-937b-e50e367fd896"
	magusOfTheWillOracle   = "f32c5530-6692-4d47-8789-c73da23fd5b7"
	walkInClosetOracle     = "52e77cc3-f8e9-4a20-811b-fe1e46a96ad7"
	graveyardPlayCardName  = "Dead Forest"
	graveyardPlayTypeLine  = "Basic Land — Forest"
	graveyardPlayBoltLabel = "Lightning Bolt"
)

// yawgmothsWillBoard is the shared setup: a land and a Lightning Bolt
// in my graveyard.
func yawgmothsWillBoard(g *game.Game, me *game.Player) (land, bolt uuid.UUID) {
	land = pushGraveyardCardTyped(me, graveyardPlayCardName, graveyardPlayTypeLine)
	bolt = uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: bolt, Name: graveyardPlayBoltLabel, TypeLine: "Instant", OracleID: lightningBoltOracle,
		Owner: me.ID, Controller: me.ID,
	})
	return land, bolt
}

// assertYawgmothsWillTurn checks every printed clause on the board
// the effect left: a land is played from the graveyard, a spell is cast
// from it and then exiled, my dying creature and discarded card are
// exiled, an opponent's dying creature is not, and the record ends
// with the turn.
func assertYawgmothsWillTurn(t *testing.T, g *game.Game, me, opp *game.Player, land, bolt uuid.UUID) {
	t.Helper()
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("play the land from the graveyard: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land never reached the battlefield")
	}
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: string(game.ZoneGraveyard),
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast the Bolt from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bolt) || !g.Exile.Contains(bolt) {
		t.Error("the Bolt cast from the graveyard went back to the graveyard instead of exile")
	}

	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	fodder := uuid.New()
	me.Hand.Cards = nil
	me.Hand.PushTop(game.Card{InstanceID: fodder, Name: "Fodder", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(mine)
		_ = g.DestroyPermanentForEffect(theirs)
		_ = g.DiscardRandomForEffect(me.ID, 1)
	})
	if me.Graveyard.Contains(mine) || !g.Exile.Contains(mine) {
		t.Error("my creature that died this turn was not exiled instead")
	}
	if me.Graveyard.Contains(fodder) || !g.Exile.Contains(fodder) {
		t.Error("my discarded card was not exiled instead")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's creature was exiled; the clause names only my graveyard")
	}

	g.WithWriteLock(func() { g.ClearEndOfTurnScopedStaticsLocked() })
	next := pushVanillaCreature(g, me.ID, "Next Turn's Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(next) })
	if !me.Graveyard.Contains(next) {
		t.Error("the replacement outlived the turn")
	}
}

func TestYawgmothsWillPlaysLandsAndSpellsFromTheGraveyardAndExilesInstead(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, bolt := yawgmothsWillBoard(g, me)
	will := castCatalogSpell(t, g, "Yawgmoth's Will", "Sorcery", yawgmothsWillOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(will) || !g.Exile.Contains(will) {
		t.Error("Yawgmoth's Will did not exile itself: it goes to the graveyard after its effect starts")
	}
	assertYawgmothsWillTurn(t, g, me, opp, land, bolt)
}

func TestGaeasWillIsSuspendAndRunsTheSameClausePair(t *testing.T) {
	if sa := suspendOf(gaeasWillOracle); sa == nil || sa.Counters != 4 || sa.Cost != "{G}" {
		t.Fatalf("Gaea's Will's suspend = %+v, want suspend 4—{G}", sa)
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, bolt := yawgmothsWillBoard(g, me)
	will := castCatalogSpell(t, g, "Gaea's Will", "Sorcery", gaeasWillOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(will) || !g.Exile.Contains(will) {
		t.Error("Gaea's Will did not exile itself")
	}
	assertYawgmothsWillTurn(t, g, me, opp, land, bolt)
}

func TestMagusOfTheWillExilesItselfAsACostAndRunsTheClausePair(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, bolt := yawgmothsWillBoard(g, me)
	advanceToMain(t, g)
	magus := pushCatalogPermanent(g, me.ID, "Magus of the Will", "Creature — Human Wizard", magusOfTheWillOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(magus) || me.Graveyard.Contains(magus) {
		t.Error("the Magus was not exiled as the cost")
	}
	assertYawgmothsWillTurn(t, g, me, opp, land, bolt)
}

func TestWalkInClosetPlaysLandsAndForgottenCellarOnlyCastsSpellsWhenUnlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, bolt := yawgmothsWillBoard(g, me)
	advanceToMain(t, g)
	room := roomCardC(me.ID, walkInClosetOracle, "Walk-In Closet", "{2}{G}", "Forgotten Cellar", "{3}{G}{G}", "G")
	castRoomC(t, g, me.ID, room, 0)
	settleOrdering(t, g, me.ID)

	// Walk-In Closet alone: lands from the graveyard, and nothing is
	// exiled instead.
	if p := grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard); p.Granted() {
		t.Fatalf("Walk-In Closet alone opened a spell cast from the graveyard: %+v", p)
	}
	dying := pushVanillaCreature(g, me.ID, "Early Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dying) })
	if !me.Graveyard.Contains(dying) {
		t.Error("a locked Forgotten Cellar exiled a card")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("Walk-In Closet: play the land from the graveyard: %v", err)
	}

	// Unlock Forgotten Cellar: spells from the graveyard, and the
	// replacement, but no land play of its own.
	unlockDoorC(t, g, me.ID, room.InstanceID, game.DoorRight)
	settleOrdering(t, g, me.ID)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: string(game.ZoneGraveyard),
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("Forgotten Cellar: cast the Bolt from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bolt) || !g.Exile.Contains(bolt) {
		t.Error("the Bolt did not go to exile instead of the graveyard")
	}
	later := pushVanillaCreature(g, me.ID, "Later Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(later)
		_ = g.DestroyPermanentForEffect(theirs)
	})
	if me.Graveyard.Contains(later) || !g.Exile.Contains(later) {
		t.Error("a creature of mine that died after the unlock was not exiled instead")
	}
	if !opp.Graveyard.Contains(theirs) {
		t.Error("an opponent's creature was exiled; the clause names only my graveyard")
	}
}

func TestForgottenCellarAloneOpensNoLandPlay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land, _ := yawgmothsWillBoard(g, me)
	advanceToMain(t, g)
	room := roomCardC(me.ID, walkInClosetOracle, "Walk-In Closet", "{2}{G}", "Forgotten Cellar", "{3}{G}{G}", "G")
	castRoomC(t, g, me.ID, room, 1)
	settleOrdering(t, g, me.ID)
	if p := grantedPermissionOn(g, me.ID, land, game.ZoneGraveyard); p.Granted() {
		t.Errorf("Forgotten Cellar opened a land play from the graveyard, which its text does not: %+v", p)
	}
}
