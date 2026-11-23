package testsm

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardslu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/lu"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const maestrosAscendancyOracle = "d1438681-241b-4d53-9470-4d04a7797ea8"

func TestMaestrosAscendancyIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(maestrosAscendancyOracle)
	if !ok || spec.Name != "Maestros Ascendancy" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registered %v, name %q, completeness %s, caveats %v", ok, spec.Name, spec.Completeness, spec.Caveats)
	}
}

// maestrosParams is a strict graveyard cast through the Ascendancy naming
// the given permanents as the sacrifice.
func maestrosParams(sacrifice ...uuid.UUID) game.CastSpellParams {
	return game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: cardslu.MaestrosAscendancyAltCost, AltCostIDs: sacrifice}
}

func TestMaestrosAscendancyCastsASorceryFromTheGraveyardForItsCostAndACreatureThenExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Maestros Ascendancy", "Enchantment", maestrosAscendancyOracle, false)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	sorcery := effectstest.PushToGraveyard(g, me, game.Card{Name: "Plain Sorcery", TypeLine: "Sorcery", ManaCost: "{1}"})
	creature := effectstest.PushToGraveyard(g, me, game.Card{Name: "Ox", TypeLine: "Creature — Ox", ManaCost: "{1}", Power: 2, Toughness: 2})
	effectstest.FloatForTest(g, me, "C")

	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("cast from the graveyard without the sacrifice offer")
	}
	if err := g.CastSpell(me.ID, sorcery, maestrosParams()); err == nil {
		t.Fatal("cast without naming a creature to sacrifice")
	}
	if err := g.CastSpell(me.ID, sorcery, maestrosParams(relic)); err == nil {
		t.Fatal("sacrificed a noncreature permanent")
	}
	if err := g.CastSpell(me.ID, creature, maestrosParams(bear)); err == nil {
		t.Fatal("cast a creature spell from the graveyard")
	}
	if !g.Battlefield.Contains(bear) || !g.Battlefield.Contains(relic) || me.ManaPool.Total() != 1 {
		t.Fatal("a refused cast paid something")
	}
	if err := g.CastSpell(me.ID, sorcery, maestrosParams(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Error("the creature was not sacrificed as the spell was cast")
	}
	if me.ManaPool.Total() != 0 {
		t.Error("the mana cost was not paid")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(sorcery) || !g.Exile.Contains(sorcery) {
		t.Error("the resolved spell was not exiled instead of going to the graveyard")
	}
}

func TestMaestrosAscendancyIsOnceEachTurnAndOnlyInYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Maestros Ascendancy", "Enchantment", maestrosAscendancyOracle, false)
	first := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	second := effectstest.B12Push(g, me.ID, "Ox", "Creature — Ox", "", 2, 2)
	third := effectstest.B12Push(g, me.ID, "Elk", "Creature — Elk", "", 2, 2)
	instant := effectstest.PushToGraveyard(g, me, game.Card{Name: "Plain Instant", TypeLine: "Instant", ManaCost: "{1}"})
	other := effectstest.PushToGraveyard(g, me, game.Card{Name: "Other Instant", TypeLine: "Instant", ManaCost: "{1}"})
	effectstest.FloatForTest(g, me, "CC")

	if err := g.CastSpell(me.ID, instant, maestrosParams(first)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, other, maestrosParams(second)); err == nil {
		t.Fatal("a second graveyard cast in one turn")
	}
	if !g.Battlefield.Contains(second) {
		t.Error("the refused cast sacrificed a creature")
	}

	// An opponent's turn: an instant could be cast, but not this way.
	effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
	effectstest.FloatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, other, maestrosParams(second)); err == nil {
		t.Fatal("cast from the graveyard in an opponent's turn")
	}

	// My next turn opens a fresh use.
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.FloatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, other, maestrosParams(third)); err != nil {
		t.Fatalf("cast in my next turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(other) || !g.Exile.Contains(instant) {
		t.Error("the spells cast this way were not exiled")
	}
}

func TestMaestrosAscendancySacrificesByTheTypesOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Maestros Ascendancy", "Enchantment", maestrosAscendancyOracle, false)
	arbor := effectstest.B12Push(g, me.ID, "Dryad Arbor", "Land Creature — Forest Dryad", "", 1, 1)
	instant := effectstest.PushToGraveyard(g, me, game.Card{Name: "Plain Instant", TypeLine: "Instant", ManaCost: "{1}"})
	effectstest.FloatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, instant, maestrosParams(arbor)); err != nil {
		t.Fatalf("a land creature is a creature: %v", err)
	}
	if g.Battlefield.Contains(arbor) {
		t.Error("the land creature was not sacrificed")
	}
}

func TestMaestrosAscendancyWithNoCreatureCannotBeClaimed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Top100ActiveSeat(g), g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Maestros Ascendancy", "Enchantment", maestrosAscendancyOracle, false)
	theirs := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	instant := effectstest.PushToGraveyard(g, me, game.Card{Name: "Plain Instant", TypeLine: "Instant", ManaCost: "{1}"})
	effectstest.FloatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, instant, maestrosParams(theirs)); err == nil {
		t.Fatal("sacrificed an opponent's creature")
	}
	if !me.Graveyard.Contains(instant) || !g.Battlefield.Contains(theirs) {
		t.Error("the refused cast moved something")
	}
}

// The exile-instead clause and the spent use are carried state: both must
// survive a restore taken with the spell on the stack.
func TestMaestrosAscendancySurvivesASnapshotRoundTripOnTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Maestros Ascendancy", "Enchantment", maestrosAscendancyOracle, false)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	ox := effectstest.B12Push(g, me.ID, "Ox", "Creature — Ox", "", 2, 2)
	instant := effectstest.PushToGraveyard(g, me, game.Card{Name: "Plain Instant", TypeLine: "Instant", ManaCost: "{1}"})
	other := effectstest.PushToGraveyard(g, me, game.Card{Name: "Other Instant", TypeLine: "Instant", ManaCost: "{1}"})
	effectstest.FloatForTest(g, me, "CC")
	if err := g.CastSpell(me.ID, instant, maestrosParams(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}

	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	restoredMe := restored.Seats[seat]
	if err := restored.CastSpell(restoredMe.ID, other, maestrosParams(ox)); err == nil {
		t.Fatal("the restored table forgot this turn's use")
	}
	effectstest.PassPriorityAroundTable(t, restored)
	if restoredMe.Graveyard.Contains(instant) || !restored.Exile.Contains(instant) {
		t.Error("the restored spell was not exiled instead of going to the graveyard")
	}
}

// "Would be put into your graveyard" is however it gets there (CR 701.6a):
// a countered spell cast this way is exiled too.
func TestMaestrosAscendancyExilesACounteredSpellCastThisWay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Maestros Ascendancy", "Enchantment", maestrosAscendancyOracle, false)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	instant := effectstest.PushToGraveyard(g, me, game.Card{Name: "Plain Instant", TypeLine: "Instant", ManaCost: "{1}"})
	effectstest.FloatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, instant, maestrosParams(bear)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if err := g.CounterSpell(instant, nil); err != nil {
		t.Fatalf("counter: %v", err)
	}
	if me.Graveyard.Contains(instant) || !g.Exile.Contains(instant) {
		t.Error("the countered spell was not exiled instead of going to the graveyard")
	}
}
