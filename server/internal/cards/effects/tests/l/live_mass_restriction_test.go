package testsl

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castFalter casts and resolves Falter for the active seat.
func castFalter(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Falter", "Instant", effectstest.FalterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
}

// The issue's own test: Falter resolves, then a creature arrives, and
// it can't block either — checked on the restriction bit AND by a
// real block declaration the engine refuses. A creature with flying
// is untouched and may block.
func TestFalterStopsACreatureThatArrivesAfterItResolved(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := effectstest.PushSizedCreature(g, me.ID, "My Attacker", 3, 3)
	early := effectstest.PushSizedCreature(g, opp.ID, "Early Bear", 2, 2)
	flyer := effectstest.PushSizedCreature(g, opp.ID, "Their Flyer", 1, 1, "flying")

	castFalter(t, g)
	late := effectstest.PushSizedCreature(g, opp.ID, "Late Bear", 2, 2)

	effectstest.AssertRestrictions(t, g, early, game.CantBlock)
	effectstest.AssertRestrictions(t, g, late, game.CantBlock)
	effectstest.AssertRestrictions(t, g, flyer, 0)
	effectstest.AssertRestrictions(t, g, attacker, game.CantBlock)

	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(late, attacker); !errors.Is(err, game.ErrIllegalBlock) {
		t.Errorf("a creature that arrived after Falter blocked: err = %v, want ErrIllegalBlock", err)
	}
	if err := g.DeclareBlocker(flyer, attacker); err != nil {
		t.Errorf("a creature with flying may still block: %v", err)
	}
}

// "Without flying" is read against the FINISHED characteristics. A
// flying grant with a later timestamp than Falter lifts the
// restriction, and losing flying later brings it on — neither of
// which a set read inside layer 6 by timestamp could see.
func TestFalterReadsFlyingAfterEveryLayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	_ = me
	grounded := effectstest.PushSizedCreature(g, opp.ID, "Grounded Bear", 2, 2)
	bird := effectstest.PushSizedCreature(g, opp.ID, "Bird", 1, 1, "flying")

	castFalter(t, g)
	effectstest.AssertRestrictions(t, g, grounded, game.CantBlock)

	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(grounded),
			[]game.Mod{game.AddKeywordsMod("flying")}, g.UntilEndOfTurnDuration(), "test — gains flying")
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(bird),
			[]game.Mod{game.RemoveKeywordsMod("flying")}, g.UntilEndOfTurnDuration(), "test — loses flying")
	})
	effectstest.AssertRestrictions(t, g, grounded, 0)
	effectstest.AssertRestrictions(t, g, bird, game.CantBlock)
}

// CR 514.2: the effect ends in the cleanup step, for the creatures
// that were there and the one that arrived alike.
func TestFalterEndsAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	early := effectstest.PushSizedCreature(g, opp.ID, "Early Bear", 2, 2)
	castFalter(t, g)
	late := effectstest.PushSizedCreature(g, opp.ID, "Late Bear", 2, 2)
	effectstest.AssertRestrictions(t, g, late, game.CantBlock)

	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AssertRestrictions(t, g, early, 0)
	effectstest.AssertRestrictions(t, g, late, 0)
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("ScopedEffects = %d after cleanup, want 0", n)
	}
	// And nothing arriving next turn is caught by it.
	next := effectstest.PushSizedCreature(g, opp.ID, "Next Turn Bear", 2, 2)
	effectstest.AssertRestrictions(t, g, next, 0)
}

// The single-Target form is "target creature can't be blocked this
// turn" and stays exactly that: its target, and no creature that
// arrives later.
func TestSingleTargetRestrictionCoversOnlyItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	target := effectstest.PushSizedCreature(g, me.ID, "Targeted Bear", 2, 2)
	other := effectstest.PushSizedCreature(g, me.ID, "Other Bear", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Artful Dodge", "Sorcery", effectstest.ArtfulDodgeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.PushSizedCreature(g, me.ID, "Late Bear", 2, 2)

	effectstest.AssertRestrictions(t, g, target, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, other, 0)
	effectstest.AssertRestrictions(t, g, late, 0)
	if len(g.ScopedEffects) != 1 || g.ScopedEffects[0].Scope != game.ScopeNone {
		t.Errorf("the targeted form must register one pinned record, have %+v", g.ScopedEffects)
	}
}

// Undo is Clone + RestoreFrom. The live scope is data on the record,
// so a restored game still reaches a creature that arrives after the
// undo, and an undo from before the cast takes the effect away.
func TestFalterLiveScopeSurvivesUndo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := g.Clone()
	castFalter(t, g)
	after := g.Clone()

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if n := len(g.ScopedEffects); n != 0 {
		t.Fatalf("undo past the cast left %d scoped effects", n)
	}
	unaffected := effectstest.PushSizedCreature(g, opp.ID, "Bear Before", 2, 2)
	effectstest.AssertRestrictions(t, g, unaffected, 0)

	g.WithWriteLock(func() { g.RestoreFrom(after) })
	late := effectstest.PushSizedCreature(g, opp.ID, "Bear After Undo", 2, 2)
	effectstest.AssertRestrictions(t, g, late, game.CantBlock)
}

// A restore point carries the scope: capture → JSON → RestoreStrict,
// then a creature arrives in the restored game and can't block, and
// the effect still ends at cleanup there.
func TestFalterLiveScopeSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	castFalter(t, g)

	restored := effectstest.RestoreRoundTrip(t, g, true)
	if len(restored.ScopedEffects) != 1 || restored.ScopedEffects[0].Scope != game.ScopeCreaturesWithoutFlying {
		t.Fatalf("the restored game lost the live scope: %+v", restored.ScopedEffects)
	}
	late := effectstest.PushSizedCreature(restored, opp.ID, "Late Bear", 2, 2)
	effectstest.AssertRestrictions(t, restored, late, game.CantBlock)

	effectstest.AdvanceToNextSeatsTurn(t, restored)
	effectstest.AssertRestrictions(t, restored, late, 0)
}

// Magmatic Chasm and Seismic Stomp are Falter at sorcery speed: the
// same live set, and a flyer is left alone.
func TestSorceryFaltersReachLateArrivals(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Magmatic Chasm", effectstest.MagmaticChasmOracle},
		{"Seismic Stomp", effectstest.SeismicStompOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			flyer := effectstest.PushSizedCreature(g, opp.ID, "Their Flyer", 1, 1, "flying")
			effectstest.CastCatalogSpell(t, g, c.name, "Sorcery", c.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			late := effectstest.PushSizedCreature(g, opp.ID, "Late Bear", 2, 2)
			effectstest.AssertRestrictions(t, g, late, game.CantBlock)
			effectstest.AssertRestrictions(t, g, flyer, 0)
		})
	}
}

// Cosmotronic Wave and Hazardous Blast: the damage is one-shot and the
// "can't block" is live. An opponent's 1/1 dies, their 2/2 survives
// and can't block, a creature an opponent gets later can't block
// either, and the caster's own creatures are untouched by both halves.
func TestOpponentsCantBlockSorceries(t *testing.T) {
	for _, c := range []struct{ name, oracle string }{
		{"Cosmotronic Wave", effectstest.CosmotronicWaveOracle},
		{"Hazardous Blast", effectstest.HazardousBlastOracle},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			mine := effectstest.PushSizedCreature(g, me.ID, "My Squirrel", 1, 1)
			small := effectstest.PushSizedCreature(g, opp.ID, "Their Squirrel", 1, 1)
			big := effectstest.PushSizedCreature(g, opp.ID, "Their Bear", 2, 2)

			effectstest.CastCatalogSpell(t, g, c.name, "Sorcery", c.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			late := effectstest.PushSizedCreature(g, opp.ID, "Their Late Bear", 2, 2)

			if g.Battlefield.Contains(small) {
				t.Error("the opponent's 1/1 should have died to the 1 damage")
			}
			if !g.Battlefield.Contains(mine) {
				t.Error("the caster's own 1/1 is not an opponent's creature and takes no damage")
			}
			effectstest.AssertRestrictions(t, g, big, game.CantBlock)
			effectstest.AssertRestrictions(t, g, late, game.CantBlock)
			effectstest.AssertRestrictions(t, g, mine, 0)
		})
	}
}

// pushGlaringSpotlight puts the Spotlight and three Forests onto the
// battlefield under `owner`.
func pushGlaringSpotlight(g *game.Game, owner uuid.UUID) uuid.UUID {
	for i := 0; i < 3; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest",
			Owner: owner, Controller: owner,
		})
	}
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Glaring Spotlight", TypeLine: "Artifact",
		ManaCost: "{1}", OracleID: effectstest.GlaringSpotlightOracle, Owner: owner, Controller: owner,
	})
}

// The static: YOUR spells and abilities may target an opponent's
// hexproof creature. Another opponent's may not — the printed "you
// control" is the difference from Nowhere to Run — and your own
// hexproof creature is not "your opponents'".
func TestGlaringSpotlightWaivesHexproofForItsControllerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	mine := effectstest.PushKeywordCreature(g, me.ID, "My Bogle", "hexproof")
	pushGlaringSpotlight(g, me.ID)

	if !effectstest.CreatureTargetableBy(g, me.ID, theirs) {
		t.Error("the Spotlight's controller should be able to target an opponent's hexproof creature")
	}
	if effectstest.CreatureTargetableBy(g, third.ID, theirs) {
		t.Error(`"spells and abilities you control": another opponent must not benefit`)
	}
	if effectstest.CreatureTargetableBy(g, opp.ID, mine) {
		t.Error("the controller's own hexproof creature keeps its hexproof")
	}
}

// The activated ability: the hexproof grant is a characteristic and
// locks to the creatures there as it resolves; "can't be blocked" is a
// rule and covers the creature you get afterwards too. The opponent's
// creatures get neither, and the Spotlight is gone.
func TestGlaringSpotlightActivationLocksHexproofButNotEvasion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	early := effectstest.PushSizedCreature(g, me.ID, "My Early Bear", 2, 2)
	theirs := effectstest.PushSizedCreature(g, opp.ID, "Their Bear", 2, 2)
	spot := pushGlaringSpotlight(g, me.ID)
	effectstest.AdvanceToStepInTurn(t, g, game.StepPrecombatMain)

	if err := g.ActivateCatalogAbility(me.ID, spot, 0, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spot) {
		t.Error("the Spotlight is sacrificed as the cost")
	}
	late := effectstest.PushSizedCreature(g, me.ID, "My Late Bear", 2, 2)

	if !effectstest.EffectiveAbilitiesContain(t, g, early, "hexproof") {
		t.Error("a creature you controlled as it resolved gains hexproof")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, late, "hexproof") {
		t.Error("hexproof is a characteristic: CR 611.2c locks it to the creatures there at resolution")
	}
	effectstest.AssertRestrictions(t, g, early, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, late, game.CantBeBlocked)
	effectstest.AssertRestrictions(t, g, theirs, 0)
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("an opponent's creature gains nothing")
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AssertRestrictions(t, g, late, 0)
	if effectstest.EffectiveAbilitiesContain(t, g, early, "hexproof") {
		t.Error("the hexproof lasts until end of turn only")
	}
}

// The ratchet: a mass restriction through ScopedEffectFor's Match
// would take the CR 611.2c snapshot a restriction must not take, so it
// is refused and names the right primitive.
func TestScopedEffectForRefusesAMassRestriction(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushSizedCreature(g, me.ID, "Bear", 2, 2)
	var err error
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
		err = effects.ScopedEffectFor{
			Match:    effects.Creature(),
			Mods:     []game.Mod{game.AddRestrictionsMod(game.CantBlock)},
			Duration: effects.DurationUntilEndOfTurn(ctx),
			Label:    "test",
		}.Apply(ctx)
	})
	if err == nil {
		t.Fatal("a Match-scoped restriction must be refused")
	}
	if n := len(g.ScopedEffects); n != 0 {
		t.Errorf("the refusal registered %d records", n)
	}
}
