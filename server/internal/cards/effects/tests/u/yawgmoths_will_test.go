package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// yawgmothsWillBoard is the shared setup: a land and a Lightning Bolt
// in my graveyard.
func yawgmothsWillBoard(g *game.Game, me *game.Player) (land, bolt uuid.UUID) {
	land = effectstest.PushGraveyardCardTyped(me, effectstest.GraveyardPlayCardName, effectstest.GraveyardPlayTypeLine)
	bolt = uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: bolt, Name: effectstest.GraveyardPlayBoltLabel, TypeLine: "Instant", OracleID: effectstest.LightningBoltOracle,
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
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bolt) || !g.Exile.Contains(bolt) {
		t.Error("the Bolt cast from the graveyard went back to the graveyard instead of exile")
	}

	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
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
	next := effectstest.PushVanillaCreature(g, me.ID, "Next Turn's Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(next) })
	if !me.Graveyard.Contains(next) {
		t.Error("the replacement outlived the turn")
	}
}

func TestYawgmothsWillPlaysLandsAndSpellsFromTheGraveyardAndExilesInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, bolt := yawgmothsWillBoard(g, me)
	will := effectstest.CastCatalogSpell(t, g, "Yawgmoth's Will", "Sorcery", effectstest.YawgmothsWillOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(will) || !g.Exile.Contains(will) {
		t.Error("Yawgmoth's Will did not exile itself: it goes to the graveyard after its effect starts")
	}
	assertYawgmothsWillTurn(t, g, me, opp, land, bolt)
}

func TestGaeasWillIsSuspendAndRunsTheSameClausePair(t *testing.T) {
	if sa := effectstest.SuspendOf(effectstest.GaeasWillOracle); sa == nil || sa.Counters != 4 || sa.Cost != "{G}" {
		t.Fatalf("Gaea's Will's suspend = %+v, want suspend 4—{G}", sa)
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, bolt := yawgmothsWillBoard(g, me)
	will := effectstest.CastCatalogSpell(t, g, "Gaea's Will", "Sorcery", effectstest.GaeasWillOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(will) || !g.Exile.Contains(will) {
		t.Error("Gaea's Will did not exile itself")
	}
	assertYawgmothsWillTurn(t, g, me, opp, land, bolt)
}

func TestMagusOfTheWillExilesItselfAsACostAndRunsTheClausePair(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, bolt := yawgmothsWillBoard(g, me)
	effectstest.AdvanceToMain(t, g)
	magus := effectstest.PushCatalogPermanent(g, me.ID, "Magus of the Will", "Creature — Human Wizard", effectstest.MagusOfTheWillOracle, false)
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateCatalogAbility(me.ID, magus, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(magus) || me.Graveyard.Contains(magus) {
		t.Error("the Magus was not exiled as the cost")
	}
	assertYawgmothsWillTurn(t, g, me, opp, land, bolt)
}

func TestWalkInClosetPlaysLandsAndForgottenCellarOnlyCastsSpellsWhenUnlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land, bolt := yawgmothsWillBoard(g, me)
	effectstest.AdvanceToMain(t, g)
	room := effectstest.RoomCardC(me.ID, effectstest.WalkInClosetOracle, "Walk-In Closet", "{2}{G}", "Forgotten Cellar", "{3}{G}{G}", "G")
	effectstest.CastRoomC(t, g, me.ID, room, 0)
	effectstest.SettleOrdering(t, g, me.ID)

	// Walk-In Closet alone: lands from the graveyard, and nothing is
	// exiled instead.
	if p := effectstest.GrantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard); p.Granted() {
		t.Fatalf("Walk-In Closet alone opened a spell cast from the graveyard: %+v", p)
	}
	dying := effectstest.PushVanillaCreature(g, me.ID, "Early Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dying) })
	if !me.Graveyard.Contains(dying) {
		t.Error("a locked Forgotten Cellar exiled a card")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: string(game.ZoneGraveyard)}); err != nil {
		t.Fatalf("Walk-In Closet: play the land from the graveyard: %v", err)
	}

	// Unlock Forgotten Cellar: spells from the graveyard, and the
	// replacement, but no land play of its own.
	effectstest.UnlockDoorC(t, g, me.ID, room.InstanceID, game.DoorRight)
	effectstest.SettleOrdering(t, g, me.ID)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: string(game.ZoneGraveyard),
		Targets:  []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("Forgotten Cellar: cast the Bolt from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bolt) || !g.Exile.Contains(bolt) {
		t.Error("the Bolt did not go to exile instead of the graveyard")
	}
	later := effectstest.PushVanillaCreature(g, me.ID, "Later Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land, _ := yawgmothsWillBoard(g, me)
	effectstest.AdvanceToMain(t, g)
	room := effectstest.RoomCardC(me.ID, effectstest.WalkInClosetOracle, "Walk-In Closet", "{2}{G}", "Forgotten Cellar", "{3}{G}{G}", "G")
	effectstest.CastRoomC(t, g, me.ID, room, 1)
	effectstest.SettleOrdering(t, g, me.ID)
	if p := effectstest.GrantedPermissionOn(g, me.ID, land, game.ZoneGraveyard); p.Granted() {
		t.Errorf("Forgotten Cellar opened a land play from the graveyard, which its text does not: %+v", p)
	}
}
