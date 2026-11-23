package testsp

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestProwessNoncreatureSpellPumpsUntilEndOfTurn — the rule, on a
// creature with no catalog entry: a noncreature spell puts ONE trigger
// on the stack (the pump waits for it to resolve, so it can be
// responded to), the pump is +1/+1, and it is gone after cleanup.
func TestProwessNoncreatureSpellPumpsUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := effectstest.PushProwessCreature(g, me.ID, "Monastery Swiftspear", 1, 2, game.KeywordProwess, "haste")

	effectstest.CastNoncreature(t, g)
	if got := effectstest.ProwessItemsFrom(g, monk); got != 1 {
		t.Fatalf("%d prowess triggers on the stack, want 1", got)
	}
	if p := effectstest.EffectivePower(t, g, monk); p != 1 {
		t.Errorf("power %d before the trigger resolved, want 1 (a trigger, not a static)", p)
	}
	effectstest.SettleProwess(t, g)
	if p, tg := effectstest.EffectivePower(t, g, monk), effectstest.EffectiveToughness(t, g, monk); p != 2 || tg != 3 {
		t.Errorf("after one noncreature spell: %d/%d, want 2/3", p, tg)
	}

	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	if p, tg := effectstest.EffectivePower(t, g, monk), effectstest.EffectiveToughness(t, g, monk); p != 3 || tg != 4 {
		t.Errorf("after two noncreature spells: %d/%d, want 3/4", p, tg)
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, tg := effectstest.EffectivePower(t, g, monk), effectstest.EffectiveToughness(t, g, monk); p != 1 || tg != 2 {
		t.Errorf("after cleanup: %d/%d, want the printed 1/2 (CR 514.2)", p, tg)
	}
}

// TestProwessIgnoresCreatureSpellsAndOpponentsSpells — "you cast a
// NONCREATURE spell": a creature spell is not one, and a spell somebody
// else cast is not yours.
func TestProwessIgnoresCreatureSpellsAndOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.PushProwessCreature(g, me.ID, "Mine", 1, 1)
	theirs := effectstest.PushProwessCreature(g, opp.ID, "Theirs", 1, 1)

	effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if got := effectstest.ProwessItemsFrom(g, mine); got != 0 {
		t.Errorf("a creature spell put %d prowess triggers on the stack, want 0", got)
	}
	effectstest.SettleProwess(t, g)
	if p := effectstest.EffectivePower(t, g, mine); p != 1 {
		t.Errorf("power %d after a creature spell, want 1", p)
	}

	effectstest.CastNoncreature(t, g)
	if got := effectstest.ProwessItemsFrom(g, theirs); got != 0 {
		t.Errorf("the opponent's prowess triggered %d times on my spell, want 0", got)
	}
	effectstest.SettleProwess(t, g)
	if p := effectstest.EffectivePower(t, g, mine); p != 2 {
		t.Errorf("my prowess: power %d, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, theirs); p != 1 {
		t.Errorf("their prowess: power %d, want 1", p)
	}
}

// TestTwoInstancesOfProwessTriggerSeparately — CR 702.108b, through the
// deck's own pair: Ty Lee prints prowess and is an Ally, and Sokka
// gives other Allies prowess. Two instances, two triggers, +2/+2. (The
// two share a source and a label, so on their own they would need no
// CR 603.3b ordering; Sokka's triggers on the same cast do.)
func TestTwoInstancesOfProwessTriggerSeparately(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tyLee := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ty Lee, Chi Blocker", OracleID: effectstest.OracleTyLee,
		TypeLine: "Legendary Creature — Human Performer Ally", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	sokka := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sokka, Tenacious Tactician", OracleID: effectstest.OracleSokka,
		TypeLine: "Legendary Creature — Human Warrior Ally", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	if n := effectstest.ProwessCountOf(t, g, tyLee); n != 2 {
		t.Fatalf("Ty Lee under Sokka has %d instances of prowess, want 2 (printed + granted)", n)
	}
	if n := effectstest.ProwessCountOf(t, g, sokka); n != 1 {
		t.Fatalf("Sokka has %d instances of prowess, want 1 (\"other\" Allies)", n)
	}

	effectstest.CastNoncreature(t, g)
	if got := effectstest.ProwessItemsFrom(g, tyLee); got != 2 {
		t.Fatalf("Ty Lee put %d prowess triggers on the stack, want 2", got)
	}
	effectstest.SettleProwess(t, g)
	if p, tg := effectstest.EffectivePower(t, g, tyLee), effectstest.EffectiveToughness(t, g, tyLee); p != 4 || tg != 3 {
		t.Errorf("Ty Lee: %d/%d, want 4/3 (+2/+2)", p, tg)
	}
	if p, tg := effectstest.EffectivePower(t, g, sokka), effectstest.EffectiveToughness(t, g, sokka); p != 4 || tg != 4 {
		t.Errorf("Sokka: %d/%d, want 4/4", p, tg)
	}

	// The Ally Sokka made is an Ally too, so it has menace and prowess
	// from the moment it enters — but it entered after the spell was
	// cast, so it did not pump for it.
	ally := effectstest.ProwessFirstNamed(t, g, "Ally", me.ID)
	if !effectstest.EffectiveAbilitiesContain(t, g, ally, "menace") || effectstest.ProwessCountOf(t, g, ally) != 1 {
		t.Errorf("Sokka's Ally token: abilities %v, want menace and one prowess", effectstest.EffectiveAbilities(t, g, ally))
	}
	if p := effectstest.EffectivePower(t, g, ally); p != 1 {
		t.Errorf("the Ally token pumped for a spell cast before it existed: power %d", p)
	}
}

// TestPrintedProwessTwiceTriggersTwice — #1510, CR 702.108b. A card
// that PRINTS prowess twice ("Prowess, prowess" — Thor Odinson, Ruric
// Thar, Biomagus, Cursed Firebreathing Yogurt) has no catalog entry
// (the deck importer's own keyword line does the whole job, the
// Monastery Swiftspear case), and the two instances are the deck
// importer's own Card.Keywords rather than a printed one plus a
// layer-6 grant — the shape TestTwoInstancesOfProwessTriggerSeparately
// exercises above. One noncreature spell still puts two triggers on
// the stack and pumps +2/+2.
func TestPrintedProwessTwiceTriggersTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thor := effectstest.PushProwessCreature(g, me.ID, "Thor Odinson", 6, 6,
		"flying", "vigilance", game.KeywordProwess, game.KeywordProwess)

	if n := effectstest.ProwessCountOf(t, g, thor); n != 2 {
		t.Fatalf("a card printing prowess twice has %d instances, want 2", n)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, thor, "flying") || !effectstest.EffectiveAbilitiesContain(t, g, thor, "vigilance") {
		t.Errorf("Thor Odinson's ordinary printed keywords: %v, want flying and vigilance", effectstest.EffectiveAbilities(t, g, thor))
	}

	effectstest.CastNoncreature(t, g)
	if got := effectstest.ProwessItemsFrom(g, thor); got != 2 {
		t.Fatalf("Thor Odinson put %d prowess triggers on the stack, want 2", got)
	}
	effectstest.SettleProwess(t, g)
	if p, tg := effectstest.EffectivePower(t, g, thor), effectstest.EffectiveToughness(t, g, thor); p != 8 || tg != 8 {
		t.Errorf("Thor Odinson: %d/%d, want 8/8 (+2/+2 from one noncreature spell)", p, tg)
	}
}

// TestProwessIsNotTriggeredByCopies — a copy of a spell is put on the
// stack, not cast (CR 707.10), and prowess says "cast". Elemental
// Eruption is the card that puts both on one stack: with one spell
// before it, it makes two Dragon Elementals (the spell and a storm
// copy), and a prowess creature on the battlefield pumps for the
// filler and the Eruption — twice, not three times.
func TestProwessIsNotTriggeredByCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := effectstest.PushProwessCreature(g, me.ID, "Watcher", 1, 1)

	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	effectstest.CastCatalogSpell(t, g, "Elemental Eruption", "Sorcery", effectstest.OracleElementalEruption, nil)
	effectstest.SettleProwess(t, g)

	if got := effectstest.CountOnBattlefield(g, "Dragon Elemental", me.ID); got != 2 {
		t.Fatalf("%d Dragon Elementals, want 2 (the spell and one storm copy)", got)
	}
	if p := effectstest.EffectivePower(t, g, monk); p != 3 {
		t.Errorf("power %d, want 3: +1 for the filler, +1 for the Eruption, nothing for its copy", p)
	}
}

// TestProwessWorksFromAToken — Monastery Mentor's Monks carry prowess
// on the token template, with no catalog entry and no trigger written.
// The first spell makes a Monk; the second pumps it.
func TestProwessWorksFromAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mentor := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Monastery Mentor", OracleID: effectstest.OracleMonasteryMentor,
		TypeLine: "Creature — Human Monk", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	if got := effectstest.CountOnBattlefield(g, "Monk", me.ID); got != 1 {
		t.Fatalf("%d Monk tokens after one spell, want 1", got)
	}
	first := effectstest.ProwessFirstNamed(t, g, "Monk", me.ID)
	if p := effectstest.EffectivePower(t, g, first); p != 1 {
		t.Errorf("the first Monk pumped for the spell that made it: power %d", p)
	}
	if p := effectstest.EffectivePower(t, g, mentor); p != 3 {
		t.Errorf("Mentor: power %d after one spell, want 3", p)
	}

	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	if got := effectstest.CountOnBattlefield(g, "Monk", me.ID); got != 2 {
		t.Fatalf("%d Monk tokens after two spells, want 2", got)
	}
	if p, tg := effectstest.EffectivePower(t, g, first), effectstest.EffectiveToughness(t, g, first); p != 2 || tg != 2 {
		t.Errorf("the first Monk after the second spell: %d/%d, want 2/2 (its own prowess)", p, tg)
	}
	if p := effectstest.EffectivePower(t, g, mentor); p != 4 {
		t.Errorf("Mentor: power %d after two spells, want 4", p)
	}

	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p := effectstest.EffectivePower(t, g, first); p != 1 {
		t.Errorf("the Monk kept its pump past cleanup: power %d", p)
	}
}

// TestProwessEndsWithItsObject — "this creature" is the object that
// triggered (CR 400.7). A creature that leaves the battlefield in
// response gets nothing, and the trigger still resolves cleanly.
func TestProwessEndsWithItsObject(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := effectstest.PushProwessCreature(g, me.ID, "Doomed", 1, 1)

	effectstest.CastNoncreature(t, g)
	if effectstest.ProwessItemsFrom(g, monk) != 1 {
		t.Fatal("setup: no prowess trigger on the stack")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(monk); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.SettleProwess(t, g)
	g.ReadSnapshot(func() {
		for _, st := range g.ScopedEffects {
			if st.Label == "Prowess — +1/+1 until end of turn" {
				t.Errorf("a pump was registered for a creature that had left: %+v", st.SourceName)
			}
		}
	})
}

// TestProwessIsLostWithAllAbilities — a CR 613.1f "loses all
// abilities" empties the list the triggers are counted from.
func TestProwessIsLostWithAllAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := effectstest.PushProwessCreature(g, me.ID, "Silenced", 1, 1)
	// A layer-6 removal (Kenrith's Transformation's shape), pinned to
	// the monk.
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(monk),
			[]game.Mod{game.LoseAllAbilitiesMod()}, g.UntilEndOfTurnDuration(), "test: loses all abilities")
	})
	if n := effectstest.ProwessCountOf(t, g, monk); n != 0 {
		t.Fatalf("a creature that lost all abilities still has %d prowess", n)
	}
	effectstest.CastNoncreature(t, g)
	if got := effectstest.ProwessItemsFrom(g, monk); got != 0 {
		t.Errorf("%d prowess triggers from a creature with no abilities", got)
	}
}

// TestCoriSteelCutterFlurryMakesAProwessMonkAndMayAttach — the batch-40
// skip, now whole: the second spell of the turn makes a Monk with
// prowess and offers to move the Cutter onto it.
func TestCoriSteelCutterFlurryMakesAProwessMonkAndMayAttach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cutter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cori-Steel Cutter", OracleID: effectstest.OracleCoriSteelCutter,
		TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID,
	})

	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	if got := effectstest.CountOnBattlefield(g, "Monk", me.ID); got != 0 {
		t.Fatalf("the FIRST spell made %d Monks; flurry is the second", got)
	}

	effectstest.CastNoncreature(t, g)
	for i := 0; i < 16 && effectstest.LatestConfirmFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	ask := effectstest.LatestConfirmFor(g, me.ID)
	if ask == nil {
		t.Fatal("no \"attach it to the Monk?\" prompt")
	}
	monk := effectstest.ProwessFirstNamed(t, g, "Monk", me.ID)
	if effectstest.ProwessCountOf(t, g, monk) != 1 {
		t.Errorf("the Monk has no prowess: %v", effectstest.EffectiveAbilities(t, g, monk))
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.SettleProwess(t, g)
	var attached uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == cutter {
				attached = c.AttachedTo.ID
			}
		}
	})
	if attached != monk {
		t.Fatalf("the Cutter is attached to %v, want the Monk %v", attached, monk)
	}
	// +1/+1 from the Cutter; trample and haste from it too.
	if p, tg := effectstest.EffectivePower(t, g, monk), effectstest.EffectiveToughness(t, g, monk); p != 2 || tg != 2 {
		t.Errorf("the equipped Monk: %d/%d, want 2/2", p, tg)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, monk, "trample") || !effectstest.EffectiveAbilitiesContain(t, g, monk, "haste") {
		t.Errorf("the equipped Monk: abilities %v, want trample and haste", effectstest.EffectiveAbilities(t, g, monk))
	}
}

// TestTyLeeProwessPumpsTyLee — the deck card (#1306) with its last
// caveat gone.
func TestTyLeeProwessPumpsTyLee(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tyLee := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ty Lee, Chi Blocker", OracleID: effectstest.OracleTyLee,
		TypeLine: "Legendary Creature — Human Performer Ally", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastNoncreature(t, g)
	effectstest.SettleProwess(t, g)
	if p, tg := effectstest.EffectivePower(t, g, tyLee), effectstest.EffectiveToughness(t, g, tyLee); p != 3 || tg != 2 {
		t.Errorf("Ty Lee: %d/%d, want 3/2", p, tg)
	}
}
