package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hone_counters_cards_test.go — hone counters (#4500, CR 122.1j) through
// the two cards that use them: Sting, Bilbo's Sword and Dwalin,
// Weaponmaster. The +1/+0 itself is the engine's
// (game/hone_counters_test.go); these tests cast and play the real
// catalog entries and read the creature's power the way combat does.

const (
	honeStingOracle  = "9779f32c-b1a2-42a3-8e78-14c28c3ad254"
	honeDwalinOracle = "cee583b7-7cc3-40ea-a227-b760839ec291"
)

func honeCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[game.CounterHone]
}

// honePowerOf is the creature's power after layers and counters.
func honePowerOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.DcCard(t, g, id).CurrentPower()
}

func honeCastSting(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return effectstest.CastCatalogSpell(t, g, "Sting, Bilbo's Sword", "Legendary Artifact — Equipment", honeStingOracle, nil)
}

// honeAnswerSting answers Sting's two target prompts, one per clause: the
// opponent first and then, when given, the creature.
func honeAnswerSting(t *testing.T, g *game.Game, chooser, opponent uuid.UUID, creature ...uuid.UUID) {
	t.Helper()
	if err := honeAnswerStingErr(g, chooser, opponent, creature...); err != nil {
		t.Fatalf("answering Sting's targets: %v", err)
	}
}

func honeAnswerStingErr(g *game.Game, chooser, opponent uuid.UUID, creature ...uuid.UUID) error {
	pick := effectstest.LatestPickTarget(g, chooser)
	if pick == nil {
		return game.ErrPendingChoiceNotFound
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, []game.TargetRef{{Kind: game.TargetPlayer, ID: opponent}}); err != nil {
		return err
	}
	second := effectstest.LatestPickTarget(g, chooser)
	if second == nil {
		return nil
	}
	refs := make([]game.TargetRef, 0, len(creature))
	for _, c := range creature {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: c})
	}
	return g.ResolvePickTargets(second.ID, chooser, refs)
}

func TestHoneCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		honeStingOracle:  "Sting, Bilbo's Sword",
		honeDwalinOracle: "Dwalin, Weaponmaster",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q, completeness %s", name, spec.Name, spec.Completeness)
		}
	}
}

// Sting: one counter per creature the TARGET opponent controls, then the
// attach, and the equipped creature is +N/+0.
func TestHoneStingCountsTheTargetOpponentsCreaturesAndAttaches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	third := g.Seats[(g.Turn.ActiveSeat+2)%4]
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.PushVanillaCreature(g, me.ID, "My Other Bear", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Their A", 1, 1)
	effectstest.PushVanillaCreature(g, opp.ID, "Their B", 1, 1)
	effectstest.PushVanillaCreature(g, opp.ID, "Their C", 1, 1)
	effectstest.PushVanillaCreature(g, third.ID, "Third's", 1, 1)

	sting := honeCastSting(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	honeAnswerSting(t, g, me.ID, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if got := honeCounters(t, g, sting); got != 3 {
		t.Errorf("hone counters = %d, want 3 (only the target opponent's creatures count)", got)
	}
	if host := effectstest.AttachmentHostOf(t, g, sting); host.Kind != game.TargetCard || host.ID != bear {
		t.Fatalf("Sting is attached to %+v, want the bear", host)
	}
	if got := honePowerOf(t, g, bear); got != 5 {
		t.Errorf("equipped bear power = %d, want 5 (2 + 3 hone counters)", got)
	}
}

// "Up to one": with no creature chosen Sting keeps its counters and
// attaches to nothing; nobody gets the bonus.
func TestHoneStingWithNoCreatureChosenStaysUnattached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Their A", 1, 1)
	effectstest.PushVanillaCreature(g, opp.ID, "Their B", 1, 1)

	sting := honeCastSting(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	honeAnswerSting(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)

	if got := honeCounters(t, g, sting); got != 2 {
		t.Errorf("hone counters = %d, want 2", got)
	}
	if host := effectstest.AttachmentHostOf(t, g, sting); host.ID != uuid.Nil {
		t.Errorf("Sting is attached to %+v, want nothing", host)
	}
	if got := honePowerOf(t, g, bear); got != 2 {
		t.Errorf("bear power = %d, want 2: an unattached Sting gives nothing", got)
	}
}

// An opponent with no creatures: no counters, and the attach still
// happens (the two halves do not depend on each other).
func TestHoneStingAgainstAnOpponentWithNoCreaturesStillAttaches(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)

	sting := honeCastSting(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	honeAnswerSting(t, g, me.ID, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)

	if got := honeCounters(t, g, sting); got != 0 {
		t.Errorf("hone counters = %d, want 0", got)
	}
	if host := effectstest.AttachmentHostOf(t, g, sting); host.ID != bear {
		t.Errorf("Sting is attached to %+v, want the bear", host)
	}
	if got := honePowerOf(t, g, bear); got != 2 {
		t.Errorf("bear power = %d, want 2", got)
	}
}

// The second clause only takes a creature YOU control.
func TestHoneStingRefusesAnOpponentsCreatureAsTheAttachTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	honeCastSting(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if err := honeAnswerStingErr(g, me.ID, opp.ID, theirs); err == nil {
		t.Error("an opponent's creature was accepted as the creature you control")
	}
}

// Sting has flash: it can be cast at instant speed.
func TestHoneStingHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := effectstest.PushCatalogHandCard(me, "Sting, Bilbo's Sword", "Legendary Artifact — Equipment", honeStingOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Errorf("Sting was refused at instant speed: %v", err)
	}
}

// Equip {3} moves Sting, and the bonus moves with it.
func TestHoneStingEquipMovesTheBonus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	first := effectstest.PushVanillaCreature(g, me.ID, "First Bear", 2, 2)
	second := effectstest.PushVanillaCreature(g, me.ID, "Second Bear", 1, 1)
	effectstest.PushVanillaCreature(g, opp.ID, "Their A", 1, 1)
	effectstest.PushVanillaCreature(g, opp.ID, "Their B", 1, 1)

	sting := honeCastSting(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	honeAnswerSting(t, g, me.ID, opp.ID, first)
	effectstest.PassPriorityAroundTable(t, g)
	if honePowerOf(t, g, first) != 4 || honePowerOf(t, g, second) != 1 {
		t.Fatalf("after the enter trigger: %d and %d, want 4 and 1", honePowerOf(t, g, first), honePowerOf(t, g, second))
	}

	effectstest.EquipTo(t, g, me.ID, sting, second)
	if honePowerOf(t, g, first) != 2 || honePowerOf(t, g, second) != 3 {
		t.Errorf("after equipping the second bear: %d and %d, want 2 and 3", honePowerOf(t, g, first), honePowerOf(t, g, second))
	}
}

// Dwalin has first strike.
func TestHoneDwalinHasFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dwalin := effectstest.PushCatalogPermanent(g, me.ID, "Dwalin, Weaponmaster", "Legendary Creature — Dwarf Warrior", honeDwalinOracle, false)
	if !effectstest.EffectiveAbilitiesContain(t, g, dwalin, "first strike") {
		t.Error("Dwalin lacks first strike")
	}
}

// Dwalin entering puts a hone counter on each Equipment its controller
// controls, and on nothing else.
func TestHoneDwalinEntersPutsAHoneCounterOnEachEquipmentYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	axe := effectstest.PushPermanentForTest(g, me.ID, "Axe", "", "Artifact — Equipment")
	sword := effectstest.PushPermanentForTest(g, me.ID, "Sword", "", "Artifact — Equipment")
	rock := effectstest.PushPermanentForTest(g, me.ID, "Rock", "", "Artifact")
	theirs := effectstest.PushPermanentForTest(g, opp.ID, "Their Axe", "", "Artifact — Equipment")

	effectstest.CastCatalogSpell(t, g, "Dwalin, Weaponmaster", "Legendary Creature — Dwarf Warrior", honeDwalinOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"Axe": axe, "Sword": sword} {
		if got := honeCounters(t, g, id); got != 1 {
			t.Errorf("%s hone counters = %d, want 1", name, got)
		}
	}
	if got := honeCounters(t, g, rock); got != 0 {
		t.Errorf("a non-Equipment artifact got %d hone counters", got)
	}
	if got := honeCounters(t, g, theirs); got != 0 {
		t.Errorf("an opponent's Equipment got %d hone counters", got)
	}
}

// Dwalin with no Equipment: the trigger does nothing and nothing breaks.
func TestHoneDwalinEntersWithNoEquipmentDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	dwalin := effectstest.CastCatalogSpell(t, g, "Dwalin, Weaponmaster", "Legendary Creature — Dwarf Warrior", honeDwalinOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := honeCounters(t, g, dwalin) + honeCounters(t, g, bear); got != 0 {
		t.Errorf("%d hone counters appeared with no Equipment", got)
	}
}

// Dwalin attacking puts another counter on each Equipment, and the
// Equipment's host hits harder for it.
func TestHoneDwalinAttackingAddsAnotherHoneCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	dwalin := effectstest.PushCatalogPermanent(g, me.ID, "Dwalin, Weaponmaster", "Legendary Creature — Dwarf Warrior", honeDwalinOracle, false)
	sword := effectstest.PushCatalogPermanent(g, me.ID, "Bonesplitter", "Artifact — Equipment", "", false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: dwalin}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
	})
	before := honePowerOf(t, g, dwalin)

	effectstest.DeclareAttack(t, g, opp.ID, dwalin)
	effectstest.PassPriorityAroundTable(t, g)

	if got := honeCounters(t, g, sword); got != 1 {
		t.Errorf("hone counters after the attack trigger = %d, want 1", got)
	}
	if got := honePowerOf(t, g, dwalin); got != before+1 {
		t.Errorf("Dwalin's power = %d, want %d (the attack trigger's counter, +1/+0)", got, before+1)
	}
}
