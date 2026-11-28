package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b792_cards_test.go — Visions, S82 OLD batch 792 (#5631). Every test
// plays the real catalog entry through the engine.

const (
	oldB792Warrior   = "05a7ca83-e820-433f-b9e9-151e817d3708"
	oldB792Enforcer  = "86cc83d8-4a70-4b97-aa4d-160afd4cd6cb"
	oldB792Wishes    = "6b3110a5-fd38-4d09-9073-c8a6f3f17840"
	oldB792TimeTide  = "b6fee94a-30b8-415e-8567-ad03ccb1d658"
	oldB792Paradise  = "76c33d54-ce55-400e-bec5-79d33a5a20fb"
	oldB792Canyon    = "ba642c8b-9ade-4501-8393-672fd53d4955"
	oldB792SandsTime = "c8f69b48-a3b2-4324-852d-11fad19d6432"
)

func TestOldB792EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB792Warrior: "Tar Pit Warrior", oldB792Enforcer: "Ogre Enforcer", oldB792Wishes: "Three Wishes",
		oldB792TimeTide: "Time and Tide", oldB792Paradise: "Undiscovered Paradise", oldB792Canyon: "Griffin Canyon",
		oldB792SandsTime: "Sands of Time",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB792TarPitWarriorIsSacrificedWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := effectstest.B12Push(g, me.ID, "Tar Pit Warrior", "Creature — Cyclops Warrior", oldB792Warrior, 3, 4)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Actor: opp.ID, CardID: warrior})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(warrior) {
		t.Error("the Warrior was not sacrificed")
	}
	if !me.Graveyard.Contains(warrior) {
		t.Error("the Warrior is not in its owner's graveyard")
	}
}

func TestOldB792OgreEnforcerSurvivesLethalDamageFromSeveralSources(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Push(g, me.ID, "Ogre Enforcer", "Creature — Ogre", oldB792Enforcer, 4, 4)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
	effectstest.DamageAndSettle(t, g, a, ogre, 2)
	effectstest.DamageAndSettle(t, g, b, ogre, 2)
	if !g.Battlefield.Contains(ogre) {
		t.Fatal("the Ogre died to lethal damage that no single source dealt")
	}
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Giant", 4, 4)
	effectstest.DamageAndSettle(t, g, c, ogre, 4)
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre survived lethal damage from a single source")
	}
}

func TestOldB792OgreEnforcerDiesToASingleLethalSourceAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := effectstest.B12Push(g, me.ID, "Ogre Enforcer", "Creature — Ogre", oldB792Enforcer, 4, 4)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Giant", 5, 5)
	effectstest.DamageAndSettle(t, g, src, ogre, 5)
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre survived 5 damage from one source")
	}
}

func TestOldB792ThreeWishesExilesThreeFaceDownAndBinsTheUnplayed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var top []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		top = append(top, id)
	}
	libBefore := me.Library.Size()
	effectstest.CastInPlace(t, g, me.ID, "Three Wishes", oldB792Wishes)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != libBefore-3 {
		t.Fatalf("library %d → %d, want -3", libBefore, got)
	}
	for _, id := range top {
		if !g.Exile.Contains(id) {
			t.Fatalf("card %s was not exiled", id)
		}
	}
	// Until the caster's next turn the cards stay in exile; at the beginning of
	// that turn's upkeep the ones not played go to the graveyard.
	effectstest.AdvanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range top {
		if g.Exile.Contains(id) {
			t.Errorf("card %s is still exiled after the caster's next upkeep", id)
		}
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %s did not reach the graveyard", id)
		}
	}
}

func TestOldB792TimeAndTideSwapsPhasedOutAndPhasingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	out := effectstest.B12Creature(g, opp.ID, "Out", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	phaser := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Phaser", TypeLine: "Creature — Drake", Power: 2, Toughness: 2,
		Keywords: []string{game.KeywordPhasing}, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		if err := g.PhaseOutForEffect(uuid.Nil, out); err != nil {
			t.Fatal(err)
		}
	})
	if !effectstest.PhasedOutInCatalogGame(g, out) {
		t.Fatal("setup: the creature did not phase out")
	}
	effectstest.CastInPlace(t, g, me.ID, "Time and Tide", oldB792TimeTide)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PhasedOutInCatalogGame(g, out) {
		t.Error("the phased-out creature did not phase in")
	}
	if !effectstest.PhasedOutInCatalogGame(g, phaser) {
		t.Error("the creature with phasing did not phase out")
	}
	if effectstest.PhasedOutInCatalogGame(g, plain) {
		t.Error("a creature without phasing phased out")
	}
}

func TestOldB792UndiscoveredParadiseReturnsAtItsControllersNextUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Undiscovered Paradise", "Land", oldB792Paradise, false)
	if err := g.ActivateManaAbility(me.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tapping for mana: %v", err)
	}
	effectstest.ResolveManaPickAnyColor(t, g, me.ID)
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land left the battlefield before the untap step")
	}
	// Pass the turn around to the same player's next untap step.
	for i := 0; i < 600 && g.Battlefield.Contains(land); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if g.Battlefield.Contains(land) {
		t.Fatal("the land never returned")
	}
	if !me.Hand.Contains(land) {
		t.Error("the land is not in its owner's hand")
	}
}

func TestOldB792GriffinCanyonUntapsAGriffinAndPumpsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	canyon := effectstest.PushCatalogPermanent(g, me.ID, "Griffin Canyon", "Land", oldB792Canyon, false)
	griffin := effectstest.B12Creature(g, me.ID, "Griffin", "Creature — Griffin", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, canyon, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("Griffin Canyon targeted a creature that is not a Griffin")
	}
	if err := g.TapCard(griffin, true); err != nil {
		t.Fatal(err)
	}
	effectstest.B16Activate(t, g, me.ID, canyon, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(griffin)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HmlTapped(t, g, griffin) {
		t.Error("the Griffin is still tapped")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, griffin); p != 3 || tough != 3 {
		t.Errorf("the Griffin is %d/%d, want 3/3", p, tough)
	}
}

func TestOldB792GriffinCanyonDoesNotPumpANoncreatureGriffin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	canyon := effectstest.PushCatalogPermanent(g, me.ID, "Griffin Canyon", "Land", oldB792Canyon, false)
	statue := effectstest.B12Permanent(g, me.ID, "Griffin Statue", "Artifact — Griffin")
	if err := g.TapCard(statue, true); err != nil {
		t.Fatal(err)
	}
	effectstest.B16Activate(t, g, me.ID, canyon, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(statue)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HmlTapped(t, g, statue) {
		t.Error("the noncreature Griffin is still tapped")
	}
}

func TestOldB792SandsOfTimeFlipsTheUpkeepPlayersPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.PushCatalogPermanent(g, g.Seats[0].ID, "Sands of Time", "Artifact", oldB792SandsTime, false)
	tapped := effectstest.B12Creature(g, me.ID, "Tapped Bear", "Creature — Bear", 2, 2)
	untapped := effectstest.B12Creature(g, me.ID, "Untapped Bear", "Creature — Bear", 2, 2)
	land := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	if err := g.TapCard(tapped, true); err != nil {
		t.Fatal(err)
	}
	if err := g.TapCard(land, true); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HmlTapped(t, g, tapped) {
		t.Error("the tapped creature stayed tapped")
	}
	if effectstest.HmlTapped(t, g, land) {
		t.Error("the tapped land stayed tapped")
	}
	if !effectstest.HmlTapped(t, g, untapped) {
		t.Error("the untapped creature was not tapped")
	}
}
