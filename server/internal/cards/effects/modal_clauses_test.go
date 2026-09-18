package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modal_clauses_test.go — #764 / ADR 0065: the proof set for per-mode
// targets, repeated modes, "choose one or more", and modal triggers.
//
// Every assertion here fails with the machinery backed out: before
// #764 Kolaghan's Command could not be REGISTERED (effects.Register
// panicked at boot on a Max > 1 spec with two targeted options),
// Mystic Confluence could not be ANNOUNCED (validateModes rejected a
// repeated index), and a trigger had no mode slot at all.

const (
	kolaghansCommandOracle = "45f1e957-09f0-4d46-8e32-238f26060a87"
	mysticConfluenceOracle = "cd11c27b-9368-4622-bc5a-2e0b993cc42b"
	sublimeEpiphanyOracle  = "56148ae7-a9df-4771-8d53-d9ffb815c884"
	glissaSunslayerOracle  = "8859a692-d107-4de2-9110-4681fc2761c7"
)

// modeRef is a target ref that names its mode OCCURRENCE and clause
// slot — what a #764 client sends and what the engine validates
// against the clause of that occurrence.
func modeRef(kind game.TargetRefKind, id uuid.UUID, mode, slot int) game.TargetRef {
	return game.TargetRef{Kind: kind, ID: id, Mode: mode, Slot: slot}
}

// --- per-mode targets: Kolaghan's Command ---------------------------

func TestKolaghansCommandAnnouncesTwoTargetedModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Two targeting bullets, each with its own target group. Before
	// #764 this cast was ErrInvalidParam and the card could not even
	// be registered.
	castModal(t, g, "Kolaghan's Command", "Instant", kolaghansCommandOracle,
		[]int{2, 3},
		[]game.TargetRef{
			modeRef(game.TargetCard, rock, 0, 0), // occurrence 0 = "destroy target artifact"
			modeRef(game.TargetCard, bear, 1, 0), // occurrence 1 = "2 damage to any target"
		})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact bullet destroyed its own target")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the damage bullet killed the 2/2 it targeted")
	}
	_ = me
}

func TestKolaghansCommandRefusesATargetForTheWrongMode(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	rock := b12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)

	// "Destroy target artifact" and "target player discards": the
	// artifact ref names occurrence 0, whose clause is the DISCARD
	// bullet's "target player". Each clause is checked on its own, so
	// this is an illegal target rather than a spell that resolves and
	// does nothing.
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Kolaghan's Command", TypeLine: "Instant",
		OracleID: kolaghansCommandOracle, Owner: active.ID, Controller: active.ID,
	})
	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Modes: []int{1, 2},
		Targets: []game.TargetRef{
			modeRef(game.TargetCard, rock, 0, 0),
			modeRef(game.TargetCard, rock, 1, 0),
		},
	})
	if err != game.ErrIllegalTarget {
		t.Errorf("an artifact in the 'target player' slot: err %v, want ErrIllegalTarget", err)
	}
}

func TestKolaghansCommandResolvesItsModesInAnnounceOrder(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := b17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	_ = opp

	// Modes [0, 1]: return the creature card, then the opponent
	// discards. CR 700.2c resolves them in the order announced.
	castModal(t, g, "Kolaghan's Command", "Instant", kolaghansCommandOracle,
		[]int{0, 1},
		[]game.TargetRef{
			modeRef(game.TargetCard, dead, 0, 0),
			modeRef(game.TargetPlayer, opp.ID, 1, 0),
		})
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(dead) {
		t.Error("the graveyard bullet returned its card to hand")
	}
	if discardOwed(g, opp.ID) != 1 {
		t.Error("the discard bullet queued the opponent's discard")
	}
}

// --- repeated modes (CR 700.2d): Mystic Confluence ------------------

func TestMysticConfluenceTakesTheSameModeThreeTimes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	before := len(me.Hand.Cards)

	castModal(t, g, "Mystic Confluence", "Instant", mysticConfluenceOracle,
		[]int{2, 2, 2}, nil)
	passPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards) - before; got != 3 {
		t.Errorf("three draw bullets drew %d cards, want 3", got)
	}
}

func TestMysticConfluenceGivesEachOccurrenceItsOwnTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	for i := 0; i < 3; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}

	// Bounce, bounce, draw: two occurrences of the same bullet, two
	// separate targets. A single clause could not say this — before
	// #764 the second bounce had to share the first one's target.
	castModal(t, g, "Mystic Confluence", "Instant", mysticConfluenceOracle,
		[]int{1, 1, 2},
		[]game.TargetRef{
			modeRef(game.TargetCard, a, 0, 0),
			modeRef(game.TargetCard, b, 1, 0),
		})
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both bounce occurrences returned their own target")
	}
}

func TestMysticConfluenceRefusesTheWrongCount(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	for _, modes := range [][]int{{2, 2}, {2, 2, 2, 2}, nil} {
		id := uuid.New()
		active.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Mystic Confluence", TypeLine: "Instant",
			OracleID: mysticConfluenceOracle, Owner: active.ID, Controller: active.ID,
		})
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes}); err == nil {
			t.Errorf("choose THREE: %v was accepted", modes)
		}
	}
}

// --- "choose one or more": Sublime Epiphany -------------------------

func TestSublimeEpiphanyTakesOneThroughN(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	pushLibraryCardForTest(opp, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})

	spec, _ := Lookup(sublimeEpiphanyOracle)
	if spec.Modes == nil || spec.Modes.Min != 1 || spec.Modes.Max != len(spec.Modes.Options) {
		t.Fatalf("choose one or more: %+v", spec.Modes)
	}

	// One bullet.
	castModal(t, g, "Sublime Epiphany", "Instant", sublimeEpiphanyOracle,
		[]int{1}, []game.TargetRef{modeRef(game.TargetCard, theirs, 0, 0)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("one bullet: the Elk was bounced")
	}

	// Three bullets, three target groups.
	before := len(opp.Hand.Cards)
	castModal(t, g, "Sublime Epiphany", "Instant", sublimeEpiphanyOracle,
		[]int{1, 2, 3},
		[]game.TargetRef{
			modeRef(game.TargetCard, mine, 0, 0),
			modeRef(game.TargetCard, mine, 1, 0),
			modeRef(game.TargetPlayer, opp.ID, 2, 0),
		})
	passPriorityAroundTable(t, g)
	if len(opp.Hand.Cards) <= before {
		t.Error("three bullets: the draw bullet resolved too")
	}
}

func TestSublimeEpiphanyRefusesZeroModes(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Sublime Epiphany", TypeLine: "Instant",
		OracleID: sublimeEpiphanyOracle, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err == nil {
		t.Error("choose ONE or more: zero bullets is illegal")
	}
}

// --- modal trigger with targets: Glissa Sunslayer -------------------

func TestGlissaSunslayerPicksModeThenTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glissa := b12Push(g, me.ID, "Glissa Sunslayer", "Legendary Creature — Phyrexian Zombie Elf",
		glissaSunslayerOracle, 3, 3)
	b64ReadyToAttack(g, glissa)
	shrine := b12Push(g, opp.ID, "Their Shrine", "Enchantment", "", 0, 0)

	declareAttack(t, g, opp.ID, glissa)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)

	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("combat damage to a player asks for the mode first (CR 603.3c)")
	}
	if triggerOnStack(g, glissa) != nil {
		t.Error("the ability is not on the stack until the mode is chosen")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("CR 603.3d: targets come AFTER the mode, not before")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}

	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enchantment bullet asks for its target")
	}
	if len(p.PickTargetCards) != 1 || p.PickTargetCards[0] != shrine {
		t.Fatalf("only the enchantment is a legal target: %+v", p.PickTargetCards)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: shrine}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shrine) {
		t.Error("the chosen bullet destroyed the enchantment")
	}
}

func TestGlissaSunslayerHidesAModeWithNoLegalTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glissa := b12Push(g, me.ID, "Glissa Sunslayer", "Legendary Creature — Phyrexian Zombie Elf",
		glissaSunslayerOracle, 3, 3)
	b64ReadyToAttack(g, glissa)

	// No enchantment and no counter anywhere: the only bullet that
	// can be taken is the draw, so it is the only one offered
	// (CR 603.3d).
	declareAttack(t, g, opp.ID, glissa)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)

	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the draw bullet is still takeable, so the trigger still asks")
	}
	if len(c.ModeOptionIndex) != 1 || c.ModeOptionIndex[0] != 0 {
		t.Fatalf("only the untargeted bullet is offered: %+v", c.ModeOptionIndex)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err == nil {
		t.Error("a bullet that was not offered cannot be chosen")
	}
	before := len(me.Hand.Cards)
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != before+1 {
		t.Error("the draw bullet resolved")
	}
}

// --- per-slot clauses on an activated ability: Resourceful Defense --

func TestResourcefulDefenseRefusesOnePermanentInBothSlots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	def := b12Push(g, me.ID, "Resourceful Defense", "Enchantment",
		"83e78565-e61f-4bbc-b834-f47941f7e3ec", 0, 0)
	a := pushCounterCreature(g, me.ID, "My Hydra", "+1/+1", 3)
	b := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	b06AddMana(me, "C", "C", "C", "C", "W")

	// "a SECOND target permanent you control" is a Distinct clause:
	// naming one permanent for both slots is refused at announce.
	if err := g.ActivateCatalogAbility(me.ID, def, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			modeRef(game.TargetCard, a, 0, 0),
			modeRef(game.TargetCard, a, 0, 1),
		},
	}); err == nil {
		t.Error("one permanent cannot fill both slots")
	}
	if err := g.ActivateCatalogAbility(me.ID, def, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			modeRef(game.TargetCard, a, 0, 0),
			modeRef(game.TargetCard, b, 0, 1),
		},
	}); err != nil {
		t.Fatalf("two distinct permanents: %v", err)
	}
	passPriorityAroundTable(t, g)
	if counterCount(g, b, "+1/+1") != 3 || counterCount(g, a, "+1/+1") != 0 {
		t.Error("the counters moved from slot 0 to slot 1")
	}
}

// b64ReadyToAttack clears the summoning-sickness flag so a freshly
// seeded permanent can attack this turn.
func b64ReadyToAttack(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].SummonedThisTurn = false
			return
		}
	}
}
