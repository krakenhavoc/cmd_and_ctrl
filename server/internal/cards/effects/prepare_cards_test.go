package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prepare_cards_test.go — the four ADR 0090 proof cards (#1328): one
// test per card, through the catalog and the real cast path. The
// engine's own lifecycle tests (the sweep, phasing, snapshots) are
// game/prepare_test.go.

// preparationCard builds a preparation card the way deck.toGameCard
// does: per-face data, front-face keywords at card level, SetFace(0).
func preparationCard(owner uuid.UUID, oracle string, keywords []string, front, spell game.Face) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   oracle,
		Layout:     game.LayoutPrepare,
		Owner:      owner,
		Controller: owner,
		Keywords:   keywords,
		Faces:      []game.Face{front, spell},
	}
	c.SetFace(0)
	return c
}

func skycoachConductorCard(owner uuid.UUID) game.Card {
	return preparationCard(owner, skycoachConductorOracleID, []string{"flash", "flying", "vigilance"},
		game.Face{Name: "Skycoach Conductor", TypeLine: "Creature — Bird Pilot", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 2, Toughness: 3},
		game.Face{Name: "All Aboard", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}})
}

func landscapePainterCard(owner uuid.UUID) game.Card {
	return preparationCard(owner, landscapePainterOracleID, nil,
		game.Face{Name: "Landscape Painter", TypeLine: "Creature — Merfolk Wizard", ManaCost: "{1}{U}", Colors: []string{"U"}, Power: 2, Toughness: 1},
		game.Face{Name: "Vibrant Idea", TypeLine: "Sorcery", ManaCost: "{4}{U}", Colors: []string{"U"}})
}

func encouragingAviatorCard(owner uuid.UUID) game.Card {
	return preparationCard(owner, encouragingAviatorOracleID, []string{"flying"},
		game.Face{Name: "Encouraging Aviator", TypeLine: "Creature — Bird Wizard", ManaCost: "{2}{U}", Colors: []string{"U"}, Power: 2, Toughness: 3},
		game.Face{Name: "Jump", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}})
}

// prepareCopyOf finds the CR 722.3c copy in exile by name.
func prepareCopyOf(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.PrepareCopy && c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

func countPrepareCopies(g *game.Game) int {
	n := 0
	for _, c := range g.Exile.Cards {
		if c.PrepareCopy {
			n++
		}
	}
	return n
}

// castFromHandAndResolve casts a card from the active seat's hand and
// lets the table resolve it.
func castFromHandAndResolve(t *testing.T, g *game.Game, p *game.Player, c game.Card) {
	t.Helper()
	p.Hand.PushTop(c)
	if err := g.CastSpell(p.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c.InstanceID) {
		t.Fatalf("%s did not resolve onto the battlefield", c.Name)
	}
}

// Skycoach Conductor: flashed in, it ENTERS prepared (the CR 614.1d
// self-replacement), with All Aboard's copy in exile. Casting the copy
// unprepares the Conductor (CR 601.2i), blinks the target as a new
// object, and the copy ceases to exist. All Aboard cannot target the
// Conductor, which is a Pilot.
func TestSkycoachConductorEntersPreparedAndAllAboardBlinks(t *testing.T) {
	g := newCatalogGame(t) // seat 0's draw step: flash is what makes the cast legal
	me := g.Seats[0]
	conductor := skycoachConductorCard(me.ID)
	castFromHandAndResolve(t, g, me, conductor)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)

	if !g.IsPreparedForEffect(conductor.InstanceID) {
		t.Fatal("Skycoach Conductor did not enter prepared")
	}
	cp, ok := prepareCopyOf(g, "All Aboard")
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
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(bear) {
		t.Error("the bear was not blinked — the old object is still on the battlefield")
	}
	if findBattlefieldByName(g, "Bear") == uuid.Nil {
		t.Error("the bear did not come back")
	}
	if countPrepareCopies(g) != 0 {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	painter := landscapePainterCard(me.ID)
	castFromHandAndResolve(t, g, me, painter)
	cp, ok := prepareCopyOf(g, "Vibrant Idea")
	if !ok || !g.IsPreparedForEffect(painter.InstanceID) {
		t.Fatal("Landscape Painter did not enter prepared with Vibrant Idea in exile")
	}

	advanceTo(t, g, game.StepBeginCombat)
	err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)})
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Fatalf("Vibrant Idea at beginning of combat: err = %v, want ErrSorcerySpeedRequired", err)
	}

	advanceTo(t, g, game.StepPostcombatMain)
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("cast Vibrant Idea in main 2: %v", err)
	}
	passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aviator := encouragingAviatorCard(me.ID)
	g.Battlefield.PushTop(aviator)

	declareAttack(t, g, opp.ID, aviator.InstanceID)
	if g.IsPreparedForEffect(aviator.InstanceID) {
		t.Fatal("the Aviator was prepared before its trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(aviator.InstanceID) {
		t.Fatal("the Aviator did not become prepared")
	}
	if _, ok := prepareCopyOf(g, "Jump"); !ok || countPrepareCopies(g) != 1 {
		t.Fatalf("exile holds %d prepare copies, want exactly one Jump", countPrepareCopies(g))
	}
}

func TestEncouragingAviatorAlreadyPreparedMakesNoSecondCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aviator := encouragingAviatorCard(me.ID)
	g.Battlefield.PushTop(aviator)
	g.WithWriteLock(func() {
		if _, err := g.BecomePreparedForEffect(aviator.InstanceID); err != nil {
			t.Fatalf("BecomePreparedForEffect: %v", err)
		}
	})
	first, _ := prepareCopyOf(g, "Jump")

	declareAttack(t, g, opp.ID, aviator.InstanceID)
	passPriorityAroundTable(t, g)
	again, _ := prepareCopyOf(g, "Jump")
	if countPrepareCopies(g) != 1 || again.InstanceID != first.InstanceID {
		t.Error("CR 722.3a: an already-prepared Aviator made a second copy on attacking")
	}
}

// Skycoach Waypoint: "target creature becomes prepared" targets any
// creature, and a creature with no prepare spell resolves into
// nothing; a preparation creature that is not prepared becomes so.
func TestSkycoachWaypointPreparesOnlyAPreparationCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	const waypointOracle = "2ac2b815-2d72-48e6-b43a-18884a74bf95"
	w1 := pushPermanentForTest(g, me.ID, "Skycoach Waypoint", waypointOracle, "Land")
	w2 := pushPermanentForTest(g, me.ID, "Skycoach Waypoint", waypointOracle, "Land")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	painter := landscapePainterCard(me.ID)
	g.Battlefield.PushTop(painter)

	activate := func(src, target uuid.UUID) {
		t.Helper()
		floatForTest(g, me, "CCC")
		if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		}); err != nil {
			t.Fatalf("ActivateCatalogAbility: %v", err)
		}
		passPriorityAroundTable(t, g)
	}

	activate(w1, bear)
	if g.IsPreparedForEffect(bear) || countPrepareCopies(g) != 0 {
		t.Error("CR 722.3a: a creature with no prepare spell became prepared")
	}
	activate(w2, painter.InstanceID)
	if !g.IsPreparedForEffect(painter.InstanceID) {
		t.Fatal("the Painter did not become prepared")
	}
	if _, ok := prepareCopyOf(g, "Vibrant Idea"); !ok {
		t.Error("no copy of Vibrant Idea in exile")
	}
}
