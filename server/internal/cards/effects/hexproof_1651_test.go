package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hexproof_1651_test.go — #1651's proof cards (ADR 0038's amendment of
// 2026-09-28): Detection Tower's turn-scoped hexproof waiver (B1), and
// "loses <keyword> and can't have <keyword>" in its scoped (Arcane
// Lighthouse) and static (the Archetype cycle) forms (B2). The engine
// half is in game/cant_have_1651_test.go.

const (
	detectionTowerOracle         = "93695c16-c441-492d-af12-b57df9739846"
	arcaneLighthouseOracle       = "30ac68e6-160a-41f9-9f0f-0e0eef383150"
	archetypeOfEnduranceOracle   = "79254223-3fa6-4b7c-8163-1e48bf5cb708"
	archetypeOfImaginationOracle = "a5458de0-0f61-49a3-a013-d90f92559809"
	archetypeOfFinalityOracle    = "eb4e9ce9-1917-46b1-b01c-645f5920ef9c"
)

func TestHexproof1651CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		detectionTowerOracle:           "Detection Tower",
		arcaneLighthouseOracle:         "Arcane Lighthouse",
		archetypeOfEnduranceOracle:     "Archetype of Endurance",
		archetypeOfImaginationOracle:   "Archetype of Imagination",
		archetypeOfFinalityOracle:      "Archetype of Finality",
		b27ArchetypeOfAggressionOracle: "Archetype of Aggression",
		b43ArchetypeOfCourageOracle:    "Archetype of Courage",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s is not registered under %s", name, oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want full", name, spec.Completeness)
		}
	}
}

// activateLand puts the land under the active player, floats {C} for
// the {1}, and activates its first non-mana ability, resolving it.
func activateLand(t *testing.T, g *game.Game, name, oracle string) (me *game.Player, land uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	land = pushStampedCatalogCard(g, me.ID, name, "Land", oracle)
	advanceToStepInTurn(t, g, game.StepPrecombatMain)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return me, land
}

// pushStampedCatalogCard seeds a catalog permanent through the entry
// event, so it carries a battlefield timestamp (CR 613.7d) and the
// layer version moves.
func pushStampedCatalogCard(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
}

// seatsAfter returns the two seats after the active one.
func seatsAfter(g *game.Game) (*game.Player, *game.Player) {
	n := len(g.Seats)
	return g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
}

// --- Detection Tower (B1) -------------------------------------------

func TestDetectionTowerWaivesHexproofForItsControllerUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	opp, third := seatsAfter(g)
	theirs := pushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	shrouded := pushKeywordCreature(g, opp.ID, "Their Ledgewalker", "shroud")
	opp.Statics = append(opp.Statics, game.PlayerStatic{Keyword: game.KeywordHexproof})
	me := g.Seats[g.Turn.ActiveSeat]
	if creatureTargetableBy(g, me.ID, theirs) || playerTargetableBy(g, me.ID, opp.ID) {
		t.Fatal("baseline: hexproof refuses an opponent")
	}

	me, tower := activateLand(t, g, "Detection Tower", detectionTowerOracle)
	if !isTapped(g, tower) {
		t.Error("{T} is part of the cost")
	}
	if !creatureTargetableBy(g, me.ID, theirs) {
		t.Error("your spells may target an opponent's hexproof creature")
	}
	if !playerTargetableBy(g, me.ID, opp.ID) {
		t.Error("your spells may target a hexproof opponent")
	}
	if creatureTargetableBy(g, third.ID, theirs) || playerTargetableBy(g, third.ID, opp.ID) {
		t.Error(`"spells and abilities you control": a third player gets nothing`)
	}
	if creatureTargetableBy(g, me.ID, shrouded) {
		t.Error("shroud is not hexproof and is not waived")
	}
	if !effectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("the creature keeps hexproof: the waiver removes nothing")
	}
	// The set is live: a hexproof creature arriving afterwards is covered.
	late := pushKeywordCreature(g, opp.ID, "Late Bogle", "hexproof")
	if !creatureTargetableBy(g, me.ID, late) {
		t.Error("a hexproof creature that arrives after the ability resolves is covered too")
	}

	// A real spell announces and resolves through the same check.
	castDoomBladeAt(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("Doom Blade at the hexproof creature should have resolved")
	}

	advanceToNextSeatsTurn(t, g)
	opp.Statics = append(opp.Statics, game.PlayerStatic{Keyword: game.KeywordHexproof})
	if creatureTargetableBy(g, me.ID, late) || playerTargetableBy(g, me.ID, opp.ID) {
		t.Error("the waiver ends at cleanup")
	}
}

func TestDetectionTowerSurvivesUndoAndARestore(t *testing.T) {
	g := newCatalogGame(t)
	opp, _ := seatsAfter(g)
	theirs := pushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	me := g.Seats[g.Turn.ActiveSeat]
	before := g.Clone()
	activateLand(t, g, "Detection Tower", detectionTowerOracle)
	after := g.Clone()

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if creatureTargetableBy(g, me.ID, theirs) {
		t.Error("undo past the activation takes the waiver away")
	}
	g.WithWriteLock(func() { g.RestoreFrom(after) })
	if !creatureTargetableBy(g, me.ID, theirs) {
		t.Error("redo brings the waiver back")
	}

	restored := restoreRoundTrip(t, g, true)
	if !creatureTargetableBy(restored, me.ID, theirs) {
		t.Fatal("the restored game keeps the waiver")
	}
	advanceToNextSeatsTurn(t, restored)
	if creatureTargetableBy(restored, me.ID, theirs) {
		t.Error("and it still ends at cleanup there")
	}
}

// --- Arcane Lighthouse (B2, scoped) -----------------------------------

func TestArcaneLighthouseStripsHexproofAndShroudAndBeatsALaterGrant(t *testing.T) {
	g := newCatalogGame(t)
	opp, _ := seatsAfter(g)
	bogle := pushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	ledge := pushKeywordCreature(g, opp.ID, "Their Ledgewalker", "shroud")

	me, _ := activateLand(t, g, "Arcane Lighthouse", arcaneLighthouseOracle)
	mine := pushKeywordCreature(g, me.ID, "My Bogle", "hexproof")
	for _, id := range []uuid.UUID{bogle, ledge} {
		if got := effectiveAbilities(t, g, id); containsString(got, "hexproof") || containsString(got, "shroud") {
			t.Errorf("abilities = %v: an opponent's creature loses hexproof and shroud", got)
		}
		if !creatureTargetableBy(g, me.ID, id) {
			t.Error("with neither keyword it is a legal target")
		}
	}
	if !effectiveAbilitiesContain(t, g, mine, "hexproof") {
		t.Error("your own creatures are untouched")
	}

	// The opponent answers with Heroic Intervention: its hexproof grant
	// has a LATER timestamp, and "can't have" still wins (CR 101.2).
	// Its indestructible is not refused.
	b29CastAs(t, g, opp, "Heroic Intervention", "Instant", heroicInterventionOracle, "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bogle, ledge} {
		got := effectiveAbilities(t, g, id)
		if containsString(got, "hexproof") || containsString(got, "shroud") {
			t.Errorf("abilities = %v: a later hexproof grant must not stick this turn", got)
		}
		if !containsString(got, "indestructible") {
			t.Errorf("abilities = %v: the other half of the grant still applies", got)
		}
	}
	if !creatureTargetableBy(g, me.ID, bogle) {
		t.Error("the creature is still a legal target after Heroic Intervention")
	}

	// CR 611.2c: the set was locked as the ability resolved.
	late := pushKeywordCreature(g, opp.ID, "Late Bogle", "hexproof")
	if !effectiveAbilitiesContain(t, g, late, "hexproof") {
		t.Error("a creature that arrived afterwards keeps its hexproof")
	}

	advanceToNextSeatsTurn(t, g)
	if !effectiveAbilitiesContain(t, g, ledge, "shroud") {
		t.Error("the effect ends at cleanup, and the printed shroud comes back")
	}
}

func TestArcaneLighthouseSurvivesUndoAndARestore(t *testing.T) {
	g := newCatalogGame(t)
	opp, _ := seatsAfter(g)
	bogle := pushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	before := g.Clone()
	activateLand(t, g, "Arcane Lighthouse", arcaneLighthouseOracle)
	after := g.Clone()

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if !effectiveAbilitiesContain(t, g, bogle, "hexproof") {
		t.Error("undo past the activation gives the hexproof back")
	}
	g.WithWriteLock(func() { g.RestoreFrom(after) })
	if effectiveAbilitiesContain(t, g, bogle, "hexproof") {
		t.Error("redo takes it away again")
	}

	restored := restoreRoundTrip(t, g, true)
	if effectiveAbilitiesContain(t, restored, bogle, "hexproof") {
		t.Fatal("the restored game keeps the can't-have")
	}
	b29CastAs(t, restored, restored.Seats[(restored.Turn.ActiveSeat+1)%len(restored.Seats)],
		"Heroic Intervention", "Instant", heroicInterventionOracle, "", game.CastSpellParams{})
	passPriorityAroundTable(t, restored)
	if effectiveAbilitiesContain(t, restored, bogle, "hexproof") {
		t.Error("and it still beats a later grant there")
	}
}

// --- The Archetypes (B2, static) -------------------------------------

func TestArchetypeOfEnduranceHexproofForYouCantHaveForThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp, _ := seatsAfter(g)
	mine := pushSizedCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	archetype := pushStampedCatalogCard(g, me.ID, "Archetype of Endurance", "Enchantment Creature — Boar", archetypeOfEnduranceOracle)

	assertKeywords(t, g, archetype, "hexproof")
	assertKeywords(t, g, mine, "hexproof")
	if effectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("an opponent's creature loses hexproof")
	}
	// Can't gain it: an opponent's Heroic Intervention, cast after the
	// Archetype entered, gives their creatures no hexproof.
	b29CastAs(t, g, opp, "Heroic Intervention", "Instant", heroicInterventionOracle, "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if effectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("a later hexproof grant does not stick to an opponent's creature")
	}
	if !creatureTargetableBy(g, me.ID, theirs) {
		t.Error("so your spells can target it")
	}
	if creatureTargetableBy(g, opp.ID, mine) {
		t.Error("your creatures have hexproof against the opponent")
	}
}

// "Your creatures keep hexproof against a removal" holds for a removal
// OLDER than the Archetype: its grant sorts after it (CR 613.7). A
// NEWER "loses hexproof" (Shadowspear's activation) does take it: the
// Archetype prints "can't have" for its controller's opponents only,
// not for its controller.
func TestArchetypeOfEnduranceAgainstAnOpposingRemoval(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := pushSizedCreature(g, me.ID, "My Bear", 2, 2)
	removeHexproof := func() {
		g.WithWriteLock(func() {
			g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(mine),
				[]game.Mod{game.RemoveKeywordsMod("hexproof")}, g.UntilEndOfTurnDuration(), "test — Shadowspear")
		})
	}
	removeHexproof()
	pushStampedCatalogCard(g, me.ID, "Archetype of Endurance", "Enchantment Creature — Boar", archetypeOfEnduranceOracle)
	if !effectiveAbilitiesContain(t, g, mine, "hexproof") {
		t.Error("a removal older than the Archetype does not take the hexproof it grants")
	}
	removeHexproof()
	if effectiveAbilitiesContain(t, g, mine, "hexproof") {
		t.Error("a newer removal does: the Archetype does not say your creatures can't lose it")
	}
}

// Two opposing Archetypes of Endurance: each side's creatures have
// hexproof from one and can't have it from the other, and "can't" wins.
func TestOpposingArchetypesOfEnduranceCancelOut(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp, _ := seatsAfter(g)
	mine := pushStampedCatalogCard(g, me.ID, "Archetype of Endurance", "Enchantment Creature — Boar", archetypeOfEnduranceOracle)
	theirs := pushStampedCatalogCard(g, opp.ID, "Archetype of Endurance", "Enchantment Creature — Boar", archetypeOfEnduranceOracle)
	if effectiveAbilitiesContain(t, g, mine, "hexproof") || effectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("with an opposing Archetype out, neither side's creatures have hexproof")
	}
}

// The rest of the cycle is the same two lines for another keyword.
func TestArchetypeCycleLocksOutItsKeyword(t *testing.T) {
	for _, c := range []struct {
		name, oracle, typeLine, keyword string
	}{
		{"Archetype of Imagination", archetypeOfImaginationOracle, "Enchantment Creature — Human Wizard", "flying"},
		{"Archetype of Finality", archetypeOfFinalityOracle, "Enchantment Creature — Gorgon", "deathtouch"},
		{"Archetype of Courage", b43ArchetypeOfCourageOracle, "Enchantment Creature — Human Soldier", "first strike"},
		{"Archetype of Aggression", b27ArchetypeOfAggressionOracle, "Enchantment Creature — Human Warrior", "trample"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp, _ := seatsAfter(g)
			mine := pushSizedCreature(g, me.ID, "My Bear", 2, 2)
			theirs := pushKeywordCreature(g, opp.ID, "Their Creature", c.keyword)
			pushStampedCatalogCard(g, me.ID, c.name, c.typeLine, c.oracle)
			if !effectiveAbilitiesContain(t, g, mine, c.keyword) {
				t.Errorf("creatures you control have %s", c.keyword)
			}
			if effectiveAbilitiesContain(t, g, theirs, c.keyword) {
				t.Errorf("an opponent's creature loses %s", c.keyword)
			}
			// A later grant — here the opponent's own until-end-of-turn
			// one, registered after the Archetype entered — does not
			// stick.
			g.WithWriteLock(func() {
				g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(theirs),
					[]game.Mod{game.AddKeywordsMod(c.keyword)}, g.UntilEndOfTurnDuration(), "test — later grant")
			})
			if effectiveAbilitiesContain(t, g, theirs, c.keyword) {
				t.Errorf("can't have or gain %s: a later grant does not stick", c.keyword)
			}
		})
	}
}

// CR 613.6: an Archetype that has lost its abilities records no
// can't-have, so the opponent's creature has its keyword again.
func TestASilencedArchetypeLocksOutNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp, _ := seatsAfter(g)
	theirs := pushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	archetype := pushStampedCatalogCard(g, me.ID, "Archetype of Endurance", "Enchantment Creature — Boar", archetypeOfEnduranceOracle)
	if effectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Fatal("setup: the Archetype strips the opponent's hexproof")
	}
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(archetype),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "test — loses all abilities")
	})
	if !effectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("a silenced Archetype's can't-have no longer applies")
	}
}
