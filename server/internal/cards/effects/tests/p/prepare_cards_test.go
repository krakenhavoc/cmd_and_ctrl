package testsp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	cardssh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/sh"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func skycoachConductorCard(owner uuid.UUID) game.Card {
	return effectstest.PreparationCard(owner, cardssh.SkycoachConductorOracleID, []string{"flash", "flying", "vigilance"},
		game.Face{Name: "Skycoach Conductor", TypeLine: "Creature — Bird Pilot", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 2, Toughness: 3},
		game.Face{Name: "All Aboard", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}})
}

func landscapePainterCard(owner uuid.UUID) game.Card {
	return effectstest.PreparationCard(owner, cardskr.LandscapePainterOracleID, nil,
		game.Face{Name: "Landscape Painter", TypeLine: "Creature — Merfolk Wizard", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
		game.Face{Name: "Vibrant Idea", TypeLine: "Sorcery", ManaCost: "{4}{U}", Colors: []string{"U"}})
}

func encouragingAviatorCard(owner uuid.UUID) game.Card {
	return effectstest.PreparationCard(owner, cardse.EncouragingAviatorOracleID, []string{"flying"},
		game.Face{Name: "Encouraging Aviator", TypeLine: "Creature — Bird Wizard", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 2, Toughness: 3},
		game.Face{Name: "Jump", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}})
}

// Skycoach Conductor: flashed in, it ENTERS prepared (the CR 614.1d
// self-replacement), with All Aboard's copy in exile. Casting the copy
// unprepares the Conductor (CR 601.2i), blinks the target as a new
// object, and the copy ceases to exist. All Aboard cannot target the
// Conductor, which is a Pilot.
func TestSkycoachConductorEntersPreparedAndAllAboardBlinks(t *testing.T) {
	g := effectstest.NewCatalogGame(t) // seat 0's draw step: flash is what makes the cast legal
	me := g.Seats[0]
	conductor := skycoachConductorCard(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, conductor)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	if !g.IsPreparedForEffect(conductor.InstanceID) {
		t.Fatal("Skycoach Conductor did not enter prepared")
	}
	cp, ok := effectstest.PrepareCopyOf(g, "All Aboard")
	if !ok {
		t.Fatal("no copy of All Aboard in exile")
	}

	err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: string(game.ZoneExile),
		Targets:  []game.TargetRef{{Kind: game.TargetCard, ID: conductor.InstanceID}},
	})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("All Aboard on the Conductor itself: err = %v, want ErrIllegalTarget (it is a Pilot)", err)
	}

	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{
		FromZone: string(game.ZoneExile),
		Targets:  []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("cast the copy of All Aboard: %v", err)
	}
	if g.IsPreparedForEffect(conductor.InstanceID) {
		t.Error("CR 601.2i: the Conductor is still prepared once its copy was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("the bear was not blinked — the old object is still on the battlefield")
	}
	if effectstest.FindBattlefieldByName(g, "Bear") == uuid.Nil {
		t.Error("the bear did not come back")
	}
	if effectstest.CountPrepareCopies(g) != 0 {
		t.Error("the resolved copy is still in exile")
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name == "All Aboard" || c.PrepareCopy {
			t.Error("the resolved copy reached the graveyard — a copy is not a card (CR 707.10)")
		}
	}
}

// Landscape Painter: the copy of Vibrant Idea is a SORCERY and keeps
// sorcery timing — refused at beginning of combat, cast in the second
// main phase for its printed {4}{U}, and draws two.
func TestLandscapePainterVibrantIdeaIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	painter := landscapePainterCard(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, painter)
	cp, ok := effectstest.PrepareCopyOf(g, "Vibrant Idea")
	if !ok || !g.IsPreparedForEffect(painter.InstanceID) {
		t.Fatal("Landscape Painter did not enter prepared with Vibrant Idea in exile")
	}

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)})
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("Vibrant Idea at beginning of combat: err = %v, want ErrSorcerySpeedRequired", err)
	}

	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("cast Vibrant Idea in main 2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand = %d, want %d — Vibrant Idea draws two", got, hand+2)
	}
	if g.IsPreparedForEffect(painter.InstanceID) {
		t.Error("the Painter is still prepared")
	}
}

// Encouraging Aviator: attacking PREPARES it through a trigger that
// uses the stack, and an Aviator that is already prepared gains
// nothing from attacking (CR 722.3a) — one copy, the same one.
func TestEncouragingAviatorBecomesPreparedWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aviator := encouragingAviatorCard(me.ID)
	g.Battlefield.PushTop(aviator)

	effectstest.DeclareAttack(t, g, opp.ID, aviator.InstanceID)
	if g.IsPreparedForEffect(aviator.InstanceID) {
		t.Fatal("the Aviator was prepared before its trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(aviator.InstanceID) {
		t.Fatal("the Aviator did not become prepared")
	}
	if _, ok := effectstest.PrepareCopyOf(g, "Jump"); !ok || effectstest.CountPrepareCopies(g) != 1 {
		t.Fatalf("exile holds %d prepare copies, want exactly one Jump", effectstest.CountPrepareCopies(g))
	}
}

func TestEncouragingAviatorAlreadyPreparedMakesNoSecondCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aviator := encouragingAviatorCard(me.ID)
	g.Battlefield.PushTop(aviator)
	g.WithWriteLock(func() {
		if _, err := g.BecomePreparedForEffect(aviator.InstanceID); err != nil {
			t.Fatalf("BecomePreparedForEffect: %v", err)
		}
	})
	first, _ := effectstest.PrepareCopyOf(g, "Jump")

	effectstest.DeclareAttack(t, g, opp.ID, aviator.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	again, _ := effectstest.PrepareCopyOf(g, "Jump")
	if effectstest.CountPrepareCopies(g) != 1 || again.InstanceID != first.InstanceID {
		t.Error("CR 722.3a: an already-prepared Aviator made a second copy on attacking")
	}
}

// Skycoach Waypoint: "target creature becomes prepared" targets any
// creature, and a creature with no prepare spell resolves into
// nothing; a preparation creature that is not prepared becomes so.
func TestSkycoachWaypointPreparesOnlyAPreparationCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	const waypointOracle = "2ac2b815-2d72-48e6-b43a-18884a74bf95"
	w1 := effectstest.PushPermanentForTest(g, me.ID, "Skycoach Waypoint", waypointOracle, "Land")
	w2 := effectstest.PushPermanentForTest(g, me.ID, "Skycoach Waypoint", waypointOracle, "Land")
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	painter := landscapePainterCard(me.ID)
	g.Battlefield.PushTop(painter)

	activate := func(src, target uuid.UUID) {
		t.Helper()
		effectstest.FloatForTest(g, me, "CCC")
		if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		}); err != nil {
			t.Fatalf("ActivateCatalogAbility: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}

	activate(w1, bear)
	if g.IsPreparedForEffect(bear) || effectstest.CountPrepareCopies(g) != 0 {
		t.Error("CR 722.3a: a creature with no prepare spell became prepared")
	}
	activate(w2, painter.InstanceID)
	if !g.IsPreparedForEffect(painter.InstanceID) {
		t.Fatal("the Painter did not become prepared")
	}
	if _, ok := effectstest.PrepareCopyOf(g, "Vibrant Idea"); !ok {
		t.Error("no copy of Vibrant Idea in exile")
	}
}
