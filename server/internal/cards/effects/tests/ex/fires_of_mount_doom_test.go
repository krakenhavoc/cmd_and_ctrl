package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const firesOfMountDoomOracle = "f4944fde-2066-4aed-85c0-6e5e00a4bfe6"

func TestFiresOfMountDoomIsRegisteredFull(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, firesOfMountDoomOracle, "Fires of Mount Doom")
}

func firesLives(g *game.Game) []int {
	var out []int
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			out = append(out, p.Life)
		}
	})
	return out
}

// The enters trigger deals 2 damage to the target and destroys every
// Equipment attached to it; an Equipment on another creature stays.
func TestFiresOfMountDoomEntersDamagesAndDestroysTheTargetsEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 3, 5)
	other := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	sword := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sword", TypeLine: "Artifact — Equipment", Owner: opp.ID, Controller: opp.ID})
	axe := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Axe", TypeLine: "Artifact — Equipment", Owner: opp.ID, Controller: opp.ID})
	effectstest.AttachForTest(t, g, sword, ogre)
	effectstest.AttachForTest(t, g, axe, other)
	effectstest.CastCatalogSpell(t, g, "Fires of Mount Doom", "Legendary Enchantment", firesOfMountDoomOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, ogre); got != 2 {
		t.Errorf("%d damage on the target, want 2", got)
	}
	if z := effectstest.E2Zone(g, sword); z != game.ZoneGraveyard {
		t.Errorf("the target's Equipment is in %q, want the graveyard", z)
	}
	if z := effectstest.E2Zone(g, axe); z != game.ZoneBattlefield {
		t.Errorf("another creature's Equipment is in %q, want the battlefield", z)
	}
}

// {2}{R} exiles the top card; casting it this turn makes Fires deal 2
// damage to each player, once.
func TestFiresOfMountDoomCastingTheExiledCardDamagesEachPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	fires := effectstest.PushCatalogPermanent(g, me.ID, "Fires of Mount Doom", "Legendary Enchantment", firesOfMountDoomOracle, false)
	top := effectstest.TopOfLibrary(me, "Shock Wave", "Sorcery", "{R}")
	effectstest.FloatMana(t, g, me, "{C}{C}{R}")
	effectstest.P7Activate(t, g, me, fires, 0, game.ActivateAbilityParams{})
	if z := effectstest.E2Zone(g, top); z != game.ZoneExile {
		t.Fatalf("the top card is in %q, want exile", z)
	}
	before := firesLives(g)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g) // the delayed trigger
	after := firesLives(g)
	for i := range before {
		if before[i]-after[i] != 2 {
			t.Errorf("seat %d lost %d life, want 2", i, before[i]-after[i])
		}
	}
	effectstest.PassPriorityAroundTable(t, g) // the spell
	if got := firesLives(g); got[0] != after[0] {
		t.Errorf("the spell's resolution dealt damage again: %v -> %v", after, got)
	}
}

// A land exiled this way is played as the turn's land drop, and the
// play also triggers. With Fires gone, the trigger still deals the damage
// (its source by last-known information). A card not played does nothing,
// and the window ends with the turn.
func TestFiresOfMountDoomLandPlayTriggersEvenAfterFiresLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	fires := effectstest.PushCatalogPermanent(g, me.ID, "Fires of Mount Doom", "Legendary Enchantment", firesOfMountDoomOracle, false)
	land := effectstest.TopOfLibrary(me, "Mountain", "Basic Land — Mountain", "")
	effectstest.FloatMana(t, g, me, "{C}{C}{R}")
	effectstest.P7Activate(t, g, me, fires, 0, game.ActivateAbilityParams{})
	effectstest.P1Destroy(t, g, fires)
	before := firesLives(g)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("play the exiled land: %v", err)
	}
	if z := effectstest.E2Zone(g, land); z != game.ZoneBattlefield {
		t.Fatalf("the land is in %q, want the battlefield", z)
	}
	effectstest.PassPriorityAroundTable(t, g)
	after := firesLives(g)
	for i := range before {
		if before[i]-after[i] != 2 {
			t.Errorf("seat %d lost %d life, want 2", i, before[i]-after[i])
		}
	}
}

// A card played from exile some other way, or a card just left in exile,
// triggers nothing; the permission ends with the turn.
func TestFiresOfMountDoomUnplayedCardLapsesAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	fires := effectstest.PushCatalogPermanent(g, me.ID, "Fires of Mount Doom", "Legendary Enchantment", firesOfMountDoomOracle, false)
	top := effectstest.TopOfLibrary(me, "Shock Wave", "Sorcery", "{R}")
	effectstest.FloatMana(t, g, me, "{C}{C}{R}")
	effectstest.P7Activate(t, g, me, fires, 0, game.ActivateAbilityParams{})
	effectstest.FloatMana(t, g, me, "{R}")
	if !effectstest.EnumeratedFor(g, me.ID, top) {
		t.Fatal("the exiled card is not offered this turn")
	}
	mine := g.Turn.ActiveSeat
	effectstest.AdvanceToMainOf(t, g, (mine+1)%len(g.Seats))
	effectstest.AdvanceToMainOf(t, g, mine)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("the exiled card was cast on a later turn")
	}
}
