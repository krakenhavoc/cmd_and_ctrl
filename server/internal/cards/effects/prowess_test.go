package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prowess_test.go — #706, CR 702.108. Prowess is a canonical keyword
// token the engine turns into one trigger per instance
// (game/prowess.go); these tests hold the rule's five edges — what
// triggers it, what doesn't, how instances add up, when it ends, and
// where the token can come from — plus one per proof card.

const (
	oracleMonasteryMentor   = "3979067a-9c68-443d-a85f-d9f07be880b9"
	oracleSokka             = "6b68acc2-b9d5-495b-8054-c04bae1349f1"
	oracleElementalEruption = "8629fdef-dbbe-4a97-a7e1-6b1646f4ca0b"
	oracleCoriSteelCutter   = "46b1a169-bc32-47bf-b379-fa0c3d13878f"
)

// pushProwessCreature is a creature with no catalog entry whose
// printed keywords the deck importer stamped — the Monastery
// Swiftspear case, and the proof that prowess needs no card file.
func pushProwessCreature(g *game.Game, controller uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	if len(keywords) == 0 {
		keywords = []string{game.KeywordProwess}
	}
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Human Monk",
		Power:      power,
		Toughness:  toughness,
		Keywords:   keywords,
		Owner:      controller,
		Controller: controller,
	})
}

// settleProwess passes priority until nothing is on or headed for the
// stack, answering any CR 603.3b trigger-order prompt in the order
// offered. A prowess creature and another trigger from a different
// source on the same cast is two different triggers, and the player
// orders them.
func settleProwess(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if ch := triggerOrderPrompt(g); ch != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				if triggerOrderPrompt(g) != nil {
					continue
				}
				t.Fatalf("a prompt other than trigger ordering is open: %+v", g.PendingChoices)
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

// castNoncreature casts a non-catalog instant from the active seat —
// the plainest noncreature spell there is — and returns its ID with the
// cast's triggers already harvested.
func castNoncreature(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, "Filler", "Instant", "", nil)
}

func prowessItemsFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source && item.Label == "Prowess — +1/+1 until end of turn" {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source && item.Label == "Prowess — +1/+1 until end of turn" {
			n++
		}
	}
	return n
}

func prowessCountOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := -1
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				n = game.ProwessCount(&g.Battlefield.Cards[i])
			}
		}
	})
	if n < 0 {
		t.Fatalf("card %s not on the battlefield", id)
	}
	return n
}

// TestProwessNoncreatureSpellPumpsUntilEndOfTurn — the rule, on a
// creature with no catalog entry: a noncreature spell puts ONE trigger
// on the stack (the pump waits for it to resolve, so it can be
// responded to), the pump is +1/+1, and it is gone after cleanup.
func TestProwessNoncreatureSpellPumpsUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := pushProwessCreature(g, me.ID, "Monastery Swiftspear", 1, 2, game.KeywordProwess, "haste")

	castNoncreature(t, g)
	if got := prowessItemsFrom(g, monk); got != 1 {
		t.Fatalf("%d prowess triggers on the stack, want 1", got)
	}
	if p := effectivePower(t, g, monk); p != 1 {
		t.Errorf("power %d before the trigger resolved, want 1 (a trigger, not a static)", p)
	}
	settleProwess(t, g)
	if p, tg := effectivePower(t, g, monk), effectiveToughness(t, g, monk); p != 2 || tg != 3 {
		t.Errorf("after one noncreature spell: %d/%d, want 2/3", p, tg)
	}

	castNoncreature(t, g)
	settleProwess(t, g)
	if p, tg := effectivePower(t, g, monk), effectiveToughness(t, g, monk); p != 3 || tg != 4 {
		t.Errorf("after two noncreature spells: %d/%d, want 3/4", p, tg)
	}

	advanceToNextSeatsTurn(t, g)
	if p, tg := effectivePower(t, g, monk), effectiveToughness(t, g, monk); p != 1 || tg != 2 {
		t.Errorf("after cleanup: %d/%d, want the printed 1/2 (CR 514.2)", p, tg)
	}
}

// TestProwessIgnoresCreatureSpellsAndOpponentsSpells — "you cast a
// NONCREATURE spell": a creature spell is not one, and a spell somebody
// else cast is not yours.
func TestProwessIgnoresCreatureSpellsAndOpponentsSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := pushProwessCreature(g, me.ID, "Mine", 1, 1)
	theirs := pushProwessCreature(g, opp.ID, "Theirs", 1, 1)

	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	if got := prowessItemsFrom(g, mine); got != 0 {
		t.Errorf("a creature spell put %d prowess triggers on the stack, want 0", got)
	}
	settleProwess(t, g)
	if p := effectivePower(t, g, mine); p != 1 {
		t.Errorf("power %d after a creature spell, want 1", p)
	}

	castNoncreature(t, g)
	if got := prowessItemsFrom(g, theirs); got != 0 {
		t.Errorf("the opponent's prowess triggered %d times on my spell, want 0", got)
	}
	settleProwess(t, g)
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("my prowess: power %d, want 2", p)
	}
	if p := effectivePower(t, g, theirs); p != 1 {
		t.Errorf("their prowess: power %d, want 1", p)
	}
}

// TestTwoInstancesOfProwessTriggerSeparately — CR 702.108b, through the
// deck's own pair: Ty Lee prints prowess and is an Ally, and Sokka
// gives other Allies prowess. Two instances, two triggers, +2/+2. (The
// two share a source and a label, so on their own they would need no
// CR 603.3b ordering; Sokka's triggers on the same cast do.)
func TestTwoInstancesOfProwessTriggerSeparately(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tyLee := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ty Lee, Chi Blocker", OracleID: oracleTyLee,
		TypeLine: "Legendary Creature — Human Performer Ally", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	sokka := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sokka, Tenacious Tactician", OracleID: oracleSokka,
		TypeLine: "Legendary Creature — Human Warrior Ally", Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID,
	})
	if n := prowessCountOf(t, g, tyLee); n != 2 {
		t.Fatalf("Ty Lee under Sokka has %d instances of prowess, want 2 (printed + granted)", n)
	}
	if n := prowessCountOf(t, g, sokka); n != 1 {
		t.Fatalf("Sokka has %d instances of prowess, want 1 (\"other\" Allies)", n)
	}

	castNoncreature(t, g)
	if got := prowessItemsFrom(g, tyLee); got != 2 {
		t.Fatalf("Ty Lee put %d prowess triggers on the stack, want 2", got)
	}
	settleProwess(t, g)
	if p, tg := effectivePower(t, g, tyLee), effectiveToughness(t, g, tyLee); p != 4 || tg != 3 {
		t.Errorf("Ty Lee: %d/%d, want 4/3 (+2/+2)", p, tg)
	}
	if p, tg := effectivePower(t, g, sokka), effectiveToughness(t, g, sokka); p != 4 || tg != 4 {
		t.Errorf("Sokka: %d/%d, want 4/4", p, tg)
	}

	// The Ally Sokka made is an Ally too, so it has menace and prowess
	// from the moment it enters — but it entered after the spell was
	// cast, so it did not pump for it.
	ally := prowessFirstNamed(t, g, "Ally", me.ID)
	if !effectiveAbilitiesContain(t, g, ally, "menace") || prowessCountOf(t, g, ally) != 1 {
		t.Errorf("Sokka's Ally token: abilities %v, want menace and one prowess", effectiveAbilities(t, g, ally))
	}
	if p := effectivePower(t, g, ally); p != 1 {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	thor := pushProwessCreature(g, me.ID, "Thor Odinson", 6, 6,
		"flying", "vigilance", game.KeywordProwess, game.KeywordProwess)

	if n := prowessCountOf(t, g, thor); n != 2 {
		t.Fatalf("a card printing prowess twice has %d instances, want 2", n)
	}
	if !effectiveAbilitiesContain(t, g, thor, "flying") || !effectiveAbilitiesContain(t, g, thor, "vigilance") {
		t.Errorf("Thor Odinson's ordinary printed keywords: %v, want flying and vigilance", effectiveAbilities(t, g, thor))
	}

	castNoncreature(t, g)
	if got := prowessItemsFrom(g, thor); got != 2 {
		t.Fatalf("Thor Odinson put %d prowess triggers on the stack, want 2", got)
	}
	settleProwess(t, g)
	if p, tg := effectivePower(t, g, thor), effectiveToughness(t, g, thor); p != 8 || tg != 8 {
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
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := pushProwessCreature(g, me.ID, "Watcher", 1, 1)

	castNoncreature(t, g)
	settleProwess(t, g)
	castCatalogSpell(t, g, "Elemental Eruption", "Sorcery", oracleElementalEruption, nil)
	settleProwess(t, g)

	if got := countOnBattlefield(g, "Dragon Elemental", me.ID); got != 2 {
		t.Fatalf("%d Dragon Elementals, want 2 (the spell and one storm copy)", got)
	}
	if p := effectivePower(t, g, monk); p != 3 {
		t.Errorf("power %d, want 3: +1 for the filler, +1 for the Eruption, nothing for its copy", p)
	}
}

// TestProwessWorksFromAToken — Monastery Mentor's Monks carry prowess
// on the token template, with no catalog entry and no trigger written.
// The first spell makes a Monk; the second pumps it.
func TestProwessWorksFromAToken(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mentor := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Monastery Mentor", OracleID: oracleMonasteryMentor,
		TypeLine: "Creature — Human Monk", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID,
	})

	castNoncreature(t, g)
	settleProwess(t, g)
	if got := countOnBattlefield(g, "Monk", me.ID); got != 1 {
		t.Fatalf("%d Monk tokens after one spell, want 1", got)
	}
	first := prowessFirstNamed(t, g, "Monk", me.ID)
	if p := effectivePower(t, g, first); p != 1 {
		t.Errorf("the first Monk pumped for the spell that made it: power %d", p)
	}
	if p := effectivePower(t, g, mentor); p != 3 {
		t.Errorf("Mentor: power %d after one spell, want 3", p)
	}

	castNoncreature(t, g)
	settleProwess(t, g)
	if got := countOnBattlefield(g, "Monk", me.ID); got != 2 {
		t.Fatalf("%d Monk tokens after two spells, want 2", got)
	}
	if p, tg := effectivePower(t, g, first), effectiveToughness(t, g, first); p != 2 || tg != 2 {
		t.Errorf("the first Monk after the second spell: %d/%d, want 2/2 (its own prowess)", p, tg)
	}
	if p := effectivePower(t, g, mentor); p != 4 {
		t.Errorf("Mentor: power %d after two spells, want 4", p)
	}

	advanceToNextSeatsTurn(t, g)
	if p := effectivePower(t, g, first); p != 1 {
		t.Errorf("the Monk kept its pump past cleanup: power %d", p)
	}
}

// TestProwessEndsWithItsObject — "this creature" is the object that
// triggered (CR 400.7). A creature that leaves the battlefield in
// response gets nothing, and the trigger still resolves cleanly.
func TestProwessEndsWithItsObject(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := pushProwessCreature(g, me.ID, "Doomed", 1, 1)

	castNoncreature(t, g)
	if prowessItemsFrom(g, monk) != 1 {
		t.Fatal("setup: no prowess trigger on the stack")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(monk); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	settleProwess(t, g)
	g.ReadSnapshot(func() {
		for _, st := range g.ScopedStatics {
			if st.Label == "Prowess — +1/+1 until end of turn" {
				t.Errorf("a pump was registered for a creature that had left: %+v", st.Source.Name)
			}
		}
	})
}

// TestProwessIsLostWithAllAbilities — a CR 613.1f "loses all
// abilities" empties the list the triggers are counted from.
func TestProwessIsLostWithAllAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monk := pushProwessCreature(g, me.ID, "Silenced", 1, 1)
	// A layer-6 removal (Kenrith's Transformation's shape), pinned to
	// the monk.
	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(game.StaticAbility{
			Layer:            game.Layer6Ability,
			RemovesAbilities: true,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.InstanceID == monk
			},
			Apply: func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
		}, uuid.Nil, "test: loses all abilities", g.UntilEndOfTurnDuration())
	})
	if n := prowessCountOf(t, g, monk); n != 0 {
		t.Fatalf("a creature that lost all abilities still has %d prowess", n)
	}
	castNoncreature(t, g)
	if got := prowessItemsFrom(g, monk); got != 0 {
		t.Errorf("%d prowess triggers from a creature with no abilities", got)
	}
}

// TestCoriSteelCutterFlurryMakesAProwessMonkAndMayAttach — the batch-40
// skip, now whole: the second spell of the turn makes a Monk with
// prowess and offers to move the Cutter onto it.
func TestCoriSteelCutterFlurryMakesAProwessMonkAndMayAttach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cutter := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Cori-Steel Cutter", OracleID: oracleCoriSteelCutter,
		TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID,
	})

	castNoncreature(t, g)
	settleProwess(t, g)
	if got := countOnBattlefield(g, "Monk", me.ID); got != 0 {
		t.Fatalf("the FIRST spell made %d Monks; flurry is the second", got)
	}

	castNoncreature(t, g)
	for i := 0; i < 16 && latestConfirmFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	ask := latestConfirmFor(g, me.ID)
	if ask == nil {
		t.Fatal("no \"attach it to the Monk?\" prompt")
	}
	monk := prowessFirstNamed(t, g, "Monk", me.ID)
	if prowessCountOf(t, g, monk) != 1 {
		t.Errorf("the Monk has no prowess: %v", effectiveAbilities(t, g, monk))
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	settleProwess(t, g)
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
	if p, tg := effectivePower(t, g, monk), effectiveToughness(t, g, monk); p != 2 || tg != 2 {
		t.Errorf("the equipped Monk: %d/%d, want 2/2", p, tg)
	}
	if !effectiveAbilitiesContain(t, g, monk, "trample") || !effectiveAbilitiesContain(t, g, monk, "haste") {
		t.Errorf("the equipped Monk: abilities %v, want trample and haste", effectiveAbilities(t, g, monk))
	}
}

// TestTyLeeProwessPumpsTyLee — the deck card (#1306) with its last
// caveat gone.
func TestTyLeeProwessPumpsTyLee(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tyLee := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Ty Lee, Chi Blocker", OracleID: oracleTyLee,
		TypeLine: "Legendary Creature — Human Performer Ally", Power: 2, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	castNoncreature(t, g)
	settleProwess(t, g)
	if p, tg := effectivePower(t, g, tyLee), effectiveToughness(t, g, tyLee); p != 3 || tg != 2 {
		t.Errorf("Ty Lee: %d/%d, want 3/2", p, tg)
	}
}

func prowessFirstNamed(t *testing.T, g *game.Game, name string, controller uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller {
				id = c.InstanceID
				return
			}
		}
	})
	if id == uuid.Nil {
		t.Fatalf("no %q on the battlefield under %s", name, controller)
	}
	return id
}

func effectiveAbilitiesContain(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	for _, a := range effectiveAbilities(t, g, id) {
		if a == kw {
			return true
		}
	}
	return false
}
