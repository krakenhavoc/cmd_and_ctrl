package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- per-mode targets: Kolaghan's Command ---------------------------

func TestKolaghansCommandAnnouncesTwoTargetedModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)

	// Two targeting bullets, each with its own target group. Before
	// #764 this cast was ErrInvalidParam and the card could not even
	// be registered.
	effectstest.CastModal(t, g, "Kolaghan's Command", "Instant", effectstest.KolaghansCommandOracle,
		[]int{2, 3},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetCard, rock, 0, 0), // occurrence 0 = "destroy target artifact"
			effectstest.ModeRef(game.TargetCard, bear, 1, 0), // occurrence 1 = "2 damage to any target"
		})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(rock) {
		t.Error("the artifact bullet destroyed its own target")
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the damage bullet killed the 2/2 it targeted")
	}
	_ = me
}

func TestKolaghansCommandRefusesATargetForTheWrongMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)

	// "Destroy target artifact" and "target player discards": the
	// artifact ref names occurrence 0, whose clause is the DISCARD
	// bullet's "target player". Each clause is checked on its own, so
	// this is an illegal target rather than a spell that resolves and
	// does nothing.
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Kolaghan's Command", TypeLine: "Instant",
		OracleID: effectstest.KolaghansCommandOracle, Owner: active.ID, Controller: active.ID,
	})
	err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Modes: []int{1, 2},
		Targets: []game.TargetRef{
			effectstest.ModeRef(game.TargetCard, rock, 0, 0),
			effectstest.ModeRef(game.TargetCard, rock, 1, 0),
		},
	})
	if err != game.ErrIllegalTarget {
		t.Errorf("an artifact in the 'target player' slot: err %v, want ErrIllegalTarget", err)
	}
}

func TestKolaghansCommandResolvesItsModesInAnnounceOrder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.B17GraveyardCard(me, "Dead Bear", "Creature — Bear", "{1}{G}")
	_ = opp

	// Modes [0, 1]: return the creature card, then the opponent
	// discards. CR 608.2c resolves them in the order announced.
	effectstest.CastModal(t, g, "Kolaghan's Command", "Instant", effectstest.KolaghansCommandOracle,
		[]int{0, 1},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetCard, dead, 0, 0),
			effectstest.ModeRef(game.TargetPlayer, opp.ID, 1, 0),
		})
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Hand.Contains(dead) {
		t.Error("the graveyard bullet returned its card to hand")
	}
	if effectstest.DiscardOwed(g, opp.ID) != 1 {
		t.Error("the discard bullet queued the opponent's discard")
	}
}

// --- repeated modes (CR 700.2d): Mystic Confluence ------------------

func TestMysticConfluenceTakesTheSameModeThreeTimes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}
	before := len(me.Hand.Cards)

	effectstest.CastModal(t, g, "Mystic Confluence", "Instant", effectstest.MysticConfluenceOracle,
		[]int{2, 2, 2}, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := len(me.Hand.Cards) - before; got != 3 {
		t.Errorf("three draw bullets drew %d cards, want 3", got)
	}
}

func TestMysticConfluenceGivesEachOccurrenceItsOwnTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	for i := 0; i < 3; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	}

	// Bounce, bounce, draw: two occurrences of the same bullet, two
	// separate targets. A single clause could not say this — before
	// #764 the second bounce had to share the first one's target.
	effectstest.CastModal(t, g, "Mystic Confluence", "Instant", effectstest.MysticConfluenceOracle,
		[]int{1, 1, 2},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetCard, a, 0, 0),
			effectstest.ModeRef(game.TargetCard, b, 1, 0),
		})
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("both bounce occurrences returned their own target")
	}
}

func TestMysticConfluenceRefusesTheWrongCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for _, modes := range [][]int{{2, 2}, {2, 2, 2, 2}, nil} {
		id := uuid.New()
		active.Hand.PushTop(game.Card{
			InstanceID: id, Name: "Mystic Confluence", TypeLine: "Instant",
			OracleID: effectstest.MysticConfluenceOracle, Owner: active.ID, Controller: active.ID,
		})
		if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes}); err == nil {
			t.Errorf("choose THREE: %v was accepted", modes)
		}
	}
}

// --- "choose one or more": Sublime Epiphany -------------------------

func TestSublimeEpiphanyTakesOneThroughN(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 3, 3)
	effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})

	spec, _ := effects.Lookup(effectstest.SublimeEpiphanyOracle)
	if spec.Modes == nil || spec.Modes.Min != 1 || spec.Modes.Max != len(spec.Modes.Options) {
		t.Fatalf("choose one or more: %+v", spec.Modes)
	}

	// One bullet. #1211 gave the card its printed SECOND bullet
	// ("counter target activated or triggered ability"), so the bounce
	// is option 2 and the token copy 3 — the indexes are the printed
	// order and move when a missing bullet is restored.
	effectstest.CastModal(t, g, "Sublime Epiphany", "Instant", effectstest.SublimeEpiphanyOracle,
		[]int{2}, []game.TargetRef{effectstest.ModeRef(game.TargetCard, theirs, 0, 0)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("one bullet: the Elk was bounced")
	}

	// Three bullets, three target groups.
	before := len(opp.Hand.Cards)
	effectstest.CastModal(t, g, "Sublime Epiphany", "Instant", effectstest.SublimeEpiphanyOracle,
		[]int{2, 3, 4},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetCard, mine, 0, 0),
			effectstest.ModeRef(game.TargetCard, mine, 1, 0),
			effectstest.ModeRef(game.TargetPlayer, opp.ID, 2, 0),
		})
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Hand.Cards) <= before {
		t.Error("three bullets: the draw bullet resolved too")
	}
}

func TestSublimeEpiphanyRefusesZeroModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Sublime Epiphany", TypeLine: "Instant",
		OracleID: effectstest.SublimeEpiphanyOracle, Owner: active.ID, Controller: active.ID,
	})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err == nil {
		t.Error("choose ONE or more: zero bullets is illegal")
	}
}

// --- modal trigger with targets: Glissa Sunslayer -------------------

func TestGlissaSunslayerPicksModeThenTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glissa := effectstest.B12Push(g, me.ID, "Glissa Sunslayer", "Legendary Creature — Phyrexian Zombie Elf",
		effectstest.GlissaSunslayerOracle, 3, 3)
	effectstest.B64ReadyToAttack(g, glissa)
	shrine := effectstest.B12Push(g, opp.ID, "Their Shrine", "Enchantment", "", 0, 0)

	effectstest.DeclareAttack(t, g, opp.ID, glissa)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)

	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("combat damage to a player asks for the mode first (CR 603.3c)")
	}
	if effectstest.TriggerOnStack(g, glissa) != nil {
		t.Error("the ability is not on the stack until the mode is chosen")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("CR 603.3d: targets come AFTER the mode, not before")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}

	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enchantment bullet asks for its target")
	}
	if len(p.PickTargetCards) != 1 || p.PickTargetCards[0] != shrine {
		t.Fatalf("only the enchantment is a legal target: %+v", p.PickTargetCards)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: shrine}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(shrine) {
		t.Error("the chosen bullet destroyed the enchantment")
	}
}

func TestGlissaSunslayerHidesAModeWithNoLegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	glissa := effectstest.B12Push(g, me.ID, "Glissa Sunslayer", "Legendary Creature — Phyrexian Zombie Elf",
		effectstest.GlissaSunslayerOracle, 3, 3)
	effectstest.B64ReadyToAttack(g, glissa)

	// No enchantment and no counter anywhere: the only bullet that
	// can be taken is the draw, so it is the only one offered
	// (CR 603.3d).
	effectstest.DeclareAttack(t, g, opp.ID, glissa)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)

	c := effectstest.ModePickChoiceFor(g, me.ID)
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
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != before+1 {
		t.Error("the draw bullet resolved")
	}
}

// --- per-slot clauses on an activated ability: Resourceful Defense --

func TestResourcefulDefenseRefusesOnePermanentInBothSlots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	def := effectstest.B12Push(g, me.ID, "Resourceful Defense", "Enchantment",
		"83e78565-e61f-4bbc-b834-f47941f7e3ec", 0, 0)
	a := effectstest.PushCounterCreature(g, me.ID, "My Hydra", "+1/+1", 3)
	b := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C", "W")

	// "a SECOND target permanent you control" is a Distinct clause:
	// naming one permanent for both slots is refused at announce.
	if err := g.ActivateCatalogAbility(me.ID, def, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			effectstest.ModeRef(game.TargetCard, a, 0, 0),
			effectstest.ModeRef(game.TargetCard, a, 0, 1),
		},
	}); err == nil {
		t.Error("one permanent cannot fill both slots")
	}
	if err := g.ActivateCatalogAbility(me.ID, def, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{
			effectstest.ModeRef(game.TargetCard, a, 0, 0),
			effectstest.ModeRef(game.TargetCard, b, 0, 1),
		},
	}); err != nil {
		t.Fatalf("two distinct permanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ResolvePayAmount(effectstest.PayAmountPrompt(g, me.ID).ID, me.ID, 3); err != nil {
		t.Fatalf("move all three: %v", err)
	}
	if effectstest.CounterCount(g, b, "+1/+1") != 3 || effectstest.CounterCount(g, a, "+1/+1") != 0 {
		t.Error("the counters moved from slot 0 to slot 1")
	}
}
