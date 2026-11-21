package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestHexproof1651CardsAreRegistered(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.DetectionTowerOracle:           "Detection Tower",
		effectstest.ArcaneLighthouseOracle:         "Arcane Lighthouse",
		effectstest.ArchetypeOfEnduranceOracle:     "Archetype of Endurance",
		effectstest.ArchetypeOfImaginationOracle:   "Archetype of Imagination",
		effectstest.ArchetypeOfFinalityOracle:      "Archetype of Finality",
		effectstest.B27ArchetypeOfAggressionOracle: "Archetype of Aggression",
		effectstest.B43ArchetypeOfCourageOracle:    "Archetype of Courage",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name {
			t.Errorf("%s is not registered under %s", name, oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want full", name, spec.Completeness)
		}
	}
}

// seatsAfter returns the two seats after the active one.
func seatsAfter(g *game.Game) (*game.Player, *game.Player) {
	n := len(g.Seats)
	return g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
}

// --- Detection Tower (B1) -------------------------------------------

func TestDetectionTowerWaivesHexproofForItsControllerUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, third := seatsAfter(g)
	theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	shrouded := effectstest.PushKeywordCreature(g, opp.ID, "Their Ledgewalker", "shroud")
	opp.Statics = append(opp.Statics, game.PlayerStatic{Keyword: game.KeywordHexproof})
	me := g.Seats[g.Turn.ActiveSeat]
	if effectstest.CreatureTargetableBy(g, me.ID, theirs) || effectstest.PlayerTargetableBy(g, me.ID, opp.ID) {
		t.Fatal("baseline: hexproof refuses an opponent")
	}

	me, tower := effectstest.ActivateLand(t, g, "Detection Tower", effectstest.DetectionTowerOracle)
	if !effectstest.IsTapped(g, tower) {
		t.Error("{T} is part of the cost")
	}
	if !effectstest.CreatureTargetableBy(g, me.ID, theirs) {
		t.Error("your spells may target an opponent's hexproof creature")
	}
	if !effectstest.PlayerTargetableBy(g, me.ID, opp.ID) {
		t.Error("your spells may target a hexproof opponent")
	}
	if effectstest.CreatureTargetableBy(g, third.ID, theirs) || effectstest.PlayerTargetableBy(g, third.ID, opp.ID) {
		t.Error(`"spells and abilities you control": a third player gets nothing`)
	}
	if effectstest.CreatureTargetableBy(g, me.ID, shrouded) {
		t.Error("shroud is not hexproof and is not waived")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("the creature keeps hexproof: the waiver removes nothing")
	}
	// The set is live: a hexproof creature arriving afterwards is covered.
	late := effectstest.PushKeywordCreature(g, opp.ID, "Late Bogle", "hexproof")
	if !effectstest.CreatureTargetableBy(g, me.ID, late) {
		t.Error("a hexproof creature that arrives after the ability resolves is covered too")
	}

	// A real spell announces and resolves through the same check.
	effectstest.CastDoomBladeAt(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("Doom Blade at the hexproof creature should have resolved")
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	opp.Statics = append(opp.Statics, game.PlayerStatic{Keyword: game.KeywordHexproof})
	if effectstest.CreatureTargetableBy(g, me.ID, late) || effectstest.PlayerTargetableBy(g, me.ID, opp.ID) {
		t.Error("the waiver ends at cleanup")
	}
}

func TestDetectionTowerSurvivesUndoAndARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, _ := seatsAfter(g)
	theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	me := g.Seats[g.Turn.ActiveSeat]
	before := g.Clone()
	effectstest.ActivateLand(t, g, "Detection Tower", effectstest.DetectionTowerOracle)
	after := g.Clone()

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if effectstest.CreatureTargetableBy(g, me.ID, theirs) {
		t.Error("undo past the activation takes the waiver away")
	}
	g.WithWriteLock(func() { g.RestoreFrom(after) })
	if !effectstest.CreatureTargetableBy(g, me.ID, theirs) {
		t.Error("redo brings the waiver back")
	}

	restored := effectstest.RestoreRoundTrip(t, g, true)
	if !effectstest.CreatureTargetableBy(restored, me.ID, theirs) {
		t.Fatal("the restored game keeps the waiver")
	}
	effectstest.AdvanceToNextSeatsTurn(t, restored)
	if effectstest.CreatureTargetableBy(restored, me.ID, theirs) {
		t.Error("and it still ends at cleanup there")
	}
}

// --- Arcane Lighthouse (B2, scoped) -----------------------------------

func TestArcaneLighthouseStripsHexproofAndShroudAndBeatsALaterGrant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, _ := seatsAfter(g)
	bogle := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	ledge := effectstest.PushKeywordCreature(g, opp.ID, "Their Ledgewalker", "shroud")

	me, _ := effectstest.ActivateLand(t, g, "Arcane Lighthouse", effectstest.ArcaneLighthouseOracle)
	mine := effectstest.PushKeywordCreature(g, me.ID, "My Bogle", "hexproof")
	for _, id := range []uuid.UUID{bogle, ledge} {
		if got := effectstest.EffectiveAbilities(t, g, id); effectstest.ContainsString(got, "hexproof") || effectstest.ContainsString(got, "shroud") {
			t.Errorf("abilities = %v: an opponent's creature loses hexproof and shroud", got)
		}
		if !effectstest.CreatureTargetableBy(g, me.ID, id) {
			t.Error("with neither keyword it is a legal target")
		}
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "hexproof") {
		t.Error("your own creatures are untouched")
	}

	// The opponent answers with Heroic Intervention: its hexproof grant
	// has a LATER timestamp, and "can't have" still wins (CR 101.2).
	// Its indestructible is not refused.
	effectstest.B29CastAs(t, g, opp, "Heroic Intervention", "Instant", effectstest.HeroicInterventionOracle, "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bogle, ledge} {
		got := effectstest.EffectiveAbilities(t, g, id)
		if effectstest.ContainsString(got, "hexproof") || effectstest.ContainsString(got, "shroud") {
			t.Errorf("abilities = %v: a later hexproof grant must not stick this turn", got)
		}
		if !effectstest.ContainsString(got, "indestructible") {
			t.Errorf("abilities = %v: the other half of the grant still applies", got)
		}
	}
	if !effectstest.CreatureTargetableBy(g, me.ID, bogle) {
		t.Error("the creature is still a legal target after Heroic Intervention")
	}

	// CR 611.2c: the set was locked as the ability resolved.
	late := effectstest.PushKeywordCreature(g, opp.ID, "Late Bogle", "hexproof")
	if !effectstest.EffectiveAbilitiesContain(t, g, late, "hexproof") {
		t.Error("a creature that arrived afterwards keeps its hexproof")
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, ledge, "shroud") {
		t.Error("the effect ends at cleanup, and the printed shroud comes back")
	}
}

func TestArcaneLighthouseSurvivesUndoAndARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp, _ := seatsAfter(g)
	bogle := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	before := g.Clone()
	effectstest.ActivateLand(t, g, "Arcane Lighthouse", effectstest.ArcaneLighthouseOracle)
	after := g.Clone()

	g.WithWriteLock(func() { g.RestoreFrom(before) })
	if !effectstest.EffectiveAbilitiesContain(t, g, bogle, "hexproof") {
		t.Error("undo past the activation gives the hexproof back")
	}
	g.WithWriteLock(func() { g.RestoreFrom(after) })
	if effectstest.EffectiveAbilitiesContain(t, g, bogle, "hexproof") {
		t.Error("redo takes it away again")
	}

	restored := effectstest.RestoreRoundTrip(t, g, true)
	if effectstest.EffectiveAbilitiesContain(t, restored, bogle, "hexproof") {
		t.Fatal("the restored game keeps the can't-have")
	}
	effectstest.B29CastAs(t, restored, restored.Seats[(restored.Turn.ActiveSeat+1)%len(restored.Seats)],
		"Heroic Intervention", "Instant", effectstest.HeroicInterventionOracle, "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, restored)
	if effectstest.EffectiveAbilitiesContain(t, restored, bogle, "hexproof") {
		t.Error("and it still beats a later grant there")
	}
}

// --- The Archetypes (B2, static) -------------------------------------

func TestArchetypeOfEnduranceHexproofForYouCantHaveForThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp, _ := seatsAfter(g)
	mine := effectstest.PushSizedCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	archetype := effectstest.PushStampedCatalogCard(g, me.ID, "Archetype of Endurance", "Enchantment Creature — Boar", effectstest.ArchetypeOfEnduranceOracle)

	effectstest.AssertKeywords(t, g, archetype, "hexproof")
	effectstest.AssertKeywords(t, g, mine, "hexproof")
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("an opponent's creature loses hexproof")
	}
	// Can't gain it: an opponent's Heroic Intervention, cast after the
	// Archetype entered, gives their creatures no hexproof.
	effectstest.B29CastAs(t, g, opp, "Heroic Intervention", "Instant", effectstest.HeroicInterventionOracle, "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("a later hexproof grant does not stick to an opponent's creature")
	}
	if !effectstest.CreatureTargetableBy(g, me.ID, theirs) {
		t.Error("so your spells can target it")
	}
	if effectstest.CreatureTargetableBy(g, opp.ID, mine) {
		t.Error("your creatures have hexproof against the opponent")
	}
}

// "Your creatures keep hexproof against a removal" holds for a removal
// OLDER than the Archetype: its grant sorts after it (CR 613.7). A
// NEWER "loses hexproof" (Shadowspear's activation) does take it: the
// Archetype prints "can't have" for its controller's opponents only,
// not for its controller.
func TestArchetypeOfEnduranceAgainstAnOpposingRemoval(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.PushSizedCreature(g, me.ID, "My Bear", 2, 2)
	removeHexproof := func() {
		g.WithWriteLock(func() {
			g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(mine),
				[]game.Mod{game.RemoveKeywordsMod("hexproof")}, g.UntilEndOfTurnDuration(), "test — Shadowspear")
		})
	}
	removeHexproof()
	effectstest.PushStampedCatalogCard(g, me.ID, "Archetype of Endurance", "Enchantment Creature — Boar", effectstest.ArchetypeOfEnduranceOracle)
	if !effectstest.EffectiveAbilitiesContain(t, g, mine, "hexproof") {
		t.Error("a removal older than the Archetype does not take the hexproof it grants")
	}
	removeHexproof()
	if effectstest.EffectiveAbilitiesContain(t, g, mine, "hexproof") {
		t.Error("a newer removal does: the Archetype does not say your creatures can't lose it")
	}
}

// Two opposing Archetypes of Endurance: each side's creatures have
// hexproof from one and can't have it from the other, and "can't" wins.
func TestOpposingArchetypesOfEnduranceCancelOut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp, _ := seatsAfter(g)
	mine := effectstest.PushStampedCatalogCard(g, me.ID, "Archetype of Endurance", "Enchantment Creature — Boar", effectstest.ArchetypeOfEnduranceOracle)
	theirs := effectstest.PushStampedCatalogCard(g, opp.ID, "Archetype of Endurance", "Enchantment Creature — Boar", effectstest.ArchetypeOfEnduranceOracle)
	if effectstest.EffectiveAbilitiesContain(t, g, mine, "hexproof") || effectstest.EffectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("with an opposing Archetype out, neither side's creatures have hexproof")
	}
}

// The rest of the cycle is the same two lines for another keyword.
func TestArchetypeCycleLocksOutItsKeyword(t *testing.T) {
	for _, c := range []struct {
		name, oracle, typeLine, keyword string
	}{
		{"Archetype of Imagination", effectstest.ArchetypeOfImaginationOracle, "Enchantment Creature — Human Wizard", "flying"},
		{"Archetype of Finality", effectstest.ArchetypeOfFinalityOracle, "Enchantment Creature — Gorgon", "deathtouch"},
		{"Archetype of Courage", effectstest.B43ArchetypeOfCourageOracle, "Enchantment Creature — Human Soldier", "first strike"},
		{"Archetype of Aggression", effectstest.B27ArchetypeOfAggressionOracle, "Enchantment Creature — Human Warrior", "trample"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			opp, _ := seatsAfter(g)
			mine := effectstest.PushSizedCreature(g, me.ID, "My Bear", 2, 2)
			theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Creature", c.keyword)
			effectstest.PushStampedCatalogCard(g, me.ID, c.name, c.typeLine, c.oracle)
			if !effectstest.EffectiveAbilitiesContain(t, g, mine, c.keyword) {
				t.Errorf("creatures you control have %s", c.keyword)
			}
			if effectstest.EffectiveAbilitiesContain(t, g, theirs, c.keyword) {
				t.Errorf("an opponent's creature loses %s", c.keyword)
			}
			// A later grant — here the opponent's own until-end-of-turn
			// one, registered after the Archetype entered — does not
			// stick.
			g.WithWriteLock(func() {
				g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(theirs),
					[]game.Mod{game.AddKeywordsMod(c.keyword)}, g.UntilEndOfTurnDuration(), "test — later grant")
			})
			if effectstest.EffectiveAbilitiesContain(t, g, theirs, c.keyword) {
				t.Errorf("can't have or gain %s: a later grant does not stick", c.keyword)
			}
		})
	}
}

// CR 613.6: an Archetype that has lost its abilities records no
// can't-have, so the opponent's creature has its keyword again.
func TestASilencedArchetypeLocksOutNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp, _ := seatsAfter(g)
	theirs := effectstest.PushKeywordCreature(g, opp.ID, "Their Bogle", "hexproof")
	archetype := effectstest.PushStampedCatalogCard(g, me.ID, "Archetype of Endurance", "Enchantment Creature — Boar", effectstest.ArchetypeOfEnduranceOracle)
	if effectstest.EffectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Fatal("setup: the Archetype strips the opponent's hexproof")
	}
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(archetype),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "test — loses all abilities")
	})
	if !effectstest.EffectiveAbilitiesContain(t, g, theirs, "hexproof") {
		t.Error("a silenced Archetype's can't-have no longer applies")
	}
}
