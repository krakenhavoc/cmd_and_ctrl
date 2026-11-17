package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rith, Liberated Primeval and the excess-damage turn tally (#5663).

const rithLiberatedOracle = "f89de40f-52d3-44b1-8faa-6ddaa32880d0"

func rithPush(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Push(g, owner, "Rith, Liberated Primeval", "Legendary Creature — Dragon", rithLiberatedOracle, 5, 5)
}

// rithDeal deals noncombat damage from source to a creature.
func rithDeal(t *testing.T, g *game.Game, source, target uuid.UUID, amount int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.DealDamageToCreatureForEffect(source, target, amount) })
	if err != nil {
		t.Fatalf("deal %d damage: %v", amount, err)
	}
}

// rithEndStep takes the active player to their end step and lets the
// table pass, so an end-step trigger has resolved.
func rithEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	mkAdvanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
}

func TestRithLiberatedPrimevalIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(rithLiberatedOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 || spec.Name != "Rith, Liberated Primeval" {
		t.Fatalf("spec %+v %v", spec, ok)
	}
	g := newCatalogGame(t)
	rith := rithPush(g, g.Seats[0].ID)
	if !hasEffectiveKeyword(t, g, rith, "flying") {
		t.Error("printed flying is missing")
	}
}

// rithWardPrompts reports whether an opponent's spell aimed at the
// permanent `seed` makes is met with a ward payment, beside a Rith I
// control.
func rithWardPrompts(t *testing.T, seed func(g *game.Game, me, opp *game.Player, rith uuid.UUID) uuid.UUID) bool {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rith := rithPush(g, me.ID)
	target := seed(g, me, opp, rith)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, target)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	return hasPayUnlessFor(g, opp.ID)
}

func TestRithLiberatedPrimevalWardAndTheGrantToOtherDragons(t *testing.T) {
	if !rithWardPrompts(t, func(_ *game.Game, _, _ *game.Player, rith uuid.UUID) uuid.UUID { return rith }) {
		t.Error("no ward prompt for a spell aimed at Rith")
	}
	if !rithWardPrompts(t, func(g *game.Game, me, _ *game.Player, _ uuid.UUID) uuid.UUID {
		return b12Creature(g, me.ID, "Whelp", "Creature — Dragon", 2, 2)
	}) {
		t.Error("no ward prompt for a spell aimed at another Dragon I control")
	}
	if rithWardPrompts(t, func(g *game.Game, me, _ *game.Player, _ uuid.UUID) uuid.UUID {
		return b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	}) {
		t.Error("a non-Dragon I control has ward")
	}
}

// "Other Dragons YOU control": an opponent's Dragon is not warded, here
// against a spell cast by a third player.
func TestRithLiberatedPrimevalDoesNotWardAnOpponentsDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	rithPush(g, me.ID)
	theirs := b12Creature(g, third.ID, "Their Whelp", "Creature — Dragon", 2, 2)
	advanceToMain(t, g)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	castAtWardedCreature(t, g, opp, theirs)
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if hasPayUnlessFor(g, opp.ID) {
		t.Error("another player's Dragon has ward from my Rith")
	}
}

func TestRithLiberatedPrimevalMakesADragonAfterExcessDamageToAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rith := rithPush(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	rithDeal(t, g, rith, theirs, 3)
	rithEndStep(t, g)
	if got := countOnBattlefield(g, "Dragon", me.ID); got != 1 {
		t.Fatalf("%d Dragon tokens, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Dragon" {
			continue
		}
		if !c.IsToken() || c.Power != 4 || c.Toughness != 4 || len(c.Colors) != 1 || c.Colors[0] != "R" ||
			!c.HasSubtype("Dragon") || !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
			t.Errorf("token %+v, want a 4/4 red Dragon with flying", c)
		}
	}
}

func TestRithLiberatedPrimevalNeedsExcessDamageToAnOpponentsPermanent(t *testing.T) {
	for name, deal := range map[string]func(t *testing.T, g *game.Game, rith, mine, theirs uuid.UUID){
		"no damage":                 func(*testing.T, *game.Game, uuid.UUID, uuid.UUID, uuid.UUID) {},
		"exactly lethal":            func(t *testing.T, g *game.Game, rith, _, theirs uuid.UUID) { rithDeal(t, g, rith, theirs, 2) },
		"excess to my own creature": func(t *testing.T, g *game.Game, rith, mine, _ uuid.UUID) { rithDeal(t, g, rith, mine, 9) },
		"less than lethal":          func(t *testing.T, g *game.Game, rith, _, theirs uuid.UUID) { rithDeal(t, g, rith, theirs, 1) },
	} {
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			rith := rithPush(g, me.ID)
			mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
			theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			advanceToMain(t, g)
			deal(t, g, rith, mine, theirs)
			rithEndStep(t, g)
			if got := countOnBattlefield(g, "Dragon", me.ID); got != 0 {
				t.Errorf("%d Dragon tokens, want 0", got)
			}
		})
	}
}

// The damage need not be Rith's, nor dealt while Rith was on the
// battlefield, and the permanent need not survive: an opponent's own
// source deals the excess, the creature dies, and Rith arrives after.
func TestRithLiberatedPrimevalCountsDamageDealtBeforeItArrivedByAnySource(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pinger := b12Creature(g, opp.ID, "Their Pinger", "Creature — Goblin", 1, 1)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	rithDeal(t, g, pinger, theirs, 5)
	g.RunStateChecksForTest()
	if countOnBattlefield(g, "Their Bear", opp.ID) != 0 {
		t.Fatal("the 2/2 survived 5 damage")
	}
	rithPush(g, me.ID)
	rithEndStep(t, g)
	if got := countOnBattlefield(g, "Dragon", me.ID); got != 1 {
		t.Errorf("%d Dragon tokens, want 1", got)
	}
}

// "At the beginning of YOUR end step", and "this turn": the excess
// damage of my turn makes one Dragon at my end step, none at the next
// player's, and the tally is gone by then.
func TestRithLiberatedPrimevalTriggersOnlyAtYourEndStepForThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rith := rithPush(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	rithDeal(t, g, rith, theirs, 3)
	rithEndStep(t, g)
	if got := countOnBattlefield(g, "Dragon", me.ID); got != 1 {
		t.Fatalf("%d Dragon tokens after my end step, want 1", got)
	}
	mine := g.Turn.ActiveSeat
	for i := 0; i < 64 && g.Turn.ActiveSeat == mine; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	other := b12Creature(g, opp.ID, "Another Bear", "Creature — Bear", 2, 2)
	rithDeal(t, g, rith, other, 3)
	rithEndStep(t, g)
	if got := countOnBattlefield(g, "Dragon", me.ID); got != 1 {
		t.Errorf("%d Dragon tokens after another player's end step, want 1", got)
	}
}

// The tally is carried through a snapshot: excess damage dealt before a
// restore still makes the Dragon at the end step after it.
func TestRithLiberatedPrimevalExcessDamageSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rith := rithPush(g, me.ID)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	rithDeal(t, g, rith, theirs, 3)
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatal("the snapshot is not restorable")
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	rithEndStep(t, restored)
	if got := countOnBattlefield(restored, "Dragon", me.ID); got != 1 {
		t.Errorf("%d Dragon tokens in the restored game, want 1", got)
	}
}

func TestRithLiberatedPrimevalEndStepTriggerDeclaresItsFootprint(t *testing.T) {
	rows := game.CatalogTriggers(rithLiberatedOracle)
	if len(rows) != 3 {
		t.Fatalf("%d rows, want 3", len(rows))
	}
	fp := rows[2].Footprint
	if len(fp) != 2 || fp[0].Kind != game.FootprintCondition ||
		fp[0].Cond != game.FootprintCondOpponentPermanentDealtExcessDamage || fp[1].Kind != game.FootprintCreateToken {
		t.Errorf("footprint %+v, want the excess-damage condition then a token", fp)
	}
}
