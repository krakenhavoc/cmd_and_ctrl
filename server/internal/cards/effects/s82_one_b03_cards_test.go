package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 ONE batch 3 — simple enters and dies triggers (#5479).

const (
	s82OneMalcatorsWatcher = "b0493924-dc6d-4655-865f-7e9b40f547ea"
	s82OneChimneyRabble    = "f6b499d5-0e39-46ce-9de9-f64a2cc2cb32"
	s82OneQuicksilver      = "b3e3d037-992b-4718-913d-7ab509d2047e"
	s82OneChromeProwler    = "6d16c2fa-4118-4b81-92ca-6c8e9efe03b6"
	s82OneTestament        = "ce0f6cd6-35d5-4f52-b81e-6d15e3bf916e"
	s82OneAmbulatory       = "b46351cf-73c7-493c-b8a7-5d7bedc8b205"
	s82OneShrapnel         = "b400536c-ea8b-4237-90da-1d3ac139881d"
	s82OneMyrCustodian     = "43923f24-7e95-4934-9a5c-8d9a5e06ddea"
	s82OneGrimnarch        = "4c27b7cd-dcbf-45a7-9e15-a6d8dfd3a087"
	s82OneDuneMover        = "38d2176c-1a1e-4030-9c63-e28ba4a69b6d"
	s82OneMyrKinsmith      = "97bb382b-16b2-44db-a2bd-de09ca90bfcc"
	s82OneNimraiser        = "c9daff24-1af8-4133-bff1-e0e4d1042db4"
	s82OneMeldweb          = "26d6a089-b867-4b89-be8e-2ad11bff0d64"
)

func TestS82OneB03CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82OneMalcatorsWatcher: "Malcator's Watcher", s82OneChimneyRabble: "Chimney Rabble",
		s82OneQuicksilver: "Quicksilver Fisher", s82OneChromeProwler: "Chrome Prowler",
		s82OneTestament: "Testament Bearer", s82OneAmbulatory: "Ambulatory Edifice",
		s82OneShrapnel: "Shrapnel Slinger", s82OneMyrCustodian: "Myr Custodian",
		s82OneGrimnarch: "Cruel Grimnarch", s82OneDuneMover: "Dune Mover",
		s82OneMyrKinsmith: "Myr Kinsmith", s82OneNimraiser: "Nimraiser Paladin",
		s82OneMeldweb: "Meldweb Curator",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OneMalcatorsWatcherDrawsWhenItDiesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	w := b27Push(g, me.ID, "Malcator's Watcher", "Artifact Creature — Phyrexian Drone", s82OneMalcatorsWatcher, "{1}{U}", 1, 1, "U")
	assertKeywords(t, g, w, "flying", "vigilance")
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatal("drew before dying")
	}
	b27Kill(g, w)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestS82OneChimneyRabbleMakesAPhyrexianGoblin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rabble := fdnCCast(t, g, "Chimney Rabble", "Creature — Phyrexian Goblin Warrior", s82OneChimneyRabble, 3, 3)
	assertKeywords(t, g, rabble, "haste")
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Phyrexian Goblin" {
			found = true
			if c.Power != 1 || c.Toughness != 1 || !c.IsCreature() || !c.HasSubtype("Goblin") {
				t.Errorf("token is %s %d/%d", c.TypeLine, c.Power, c.Toughness)
			}
		}
	}
	if !found {
		t.Error("no Phyrexian Goblin token")
	}
}

func TestS82OneQuicksilverFisherLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Top", "Next")
	fdnCCast(t, g, "Quicksilver Fisher", "Creature — Phyrexian Drake", s82OneQuicksilver, 4, 3)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt after the draw")
	}
	drawn := me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID
	hand := me.Hand.Size()
	answerDiscard(t, g, me.ID, drawn)
	if me.Hand.Size() != hand-1 || !me.Graveyard.Contains(drawn) {
		t.Errorf("hand %d -> %d; the drawn card must be a legal discard", hand, me.Hand.Size())
	}
}

func TestS82OneChromeProwlerTapsAnOpposingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	prowler := castHoldCreature(t, g, "Chrome Prowler", "Artifact Creature — Phyrexian Cat", s82OneChromeProwler, 3, 2)
	assertKeywords(t, g, prowler, "flash")
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target asked")
	}
	if hasID(pick.PickTargetCards, mine) || hasID(pick.PickTargetCards, prowler) {
		t.Error("a creature you control was offered")
	}
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, theirs) {
		t.Error("the target was not tapped")
	}
	if got := b18Counters(t, g, theirs, game.CounterStun); got != 0 {
		t.Errorf("%d stun counters, want none", got)
	}
}

func TestS82OneTestamentBearerTakesOneOfTheTopThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	top := samiLibraryTopFirst(me,
		game.Card{Name: "A", TypeLine: "Sorcery"}, game.Card{Name: "B", TypeLine: "Sorcery"},
		game.Card{Name: "C", TypeLine: "Sorcery"}, game.Card{Name: "D", TypeLine: "Sorcery"})
	bearer := b27Push(g, me.ID, "Testament Bearer", "Creature — Phyrexian Warrior", s82OneTestament, "{3}{B}", 4, 1, "B")
	passPriorityAroundTable(t, g)
	if me.Library.Size() != 4 {
		t.Fatal("looked at the library before dying")
	}
	b27Kill(g, bearer)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top three")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of three", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{top[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(top[1]) || !me.Graveyard.Contains(top[0]) || !me.Graveyard.Contains(top[2]) {
		t.Error("one should be in hand and the other two in the graveyard")
	}
	if !me.Library.Contains(top[3]) || me.Library.Size() != 1 {
		t.Error("the fourth card must stay in the library")
	}
}

func TestS82OneAmbulatoryEdificePaysLifeForMinusOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	fdnCCast(t, g, "Ambulatory Edifice", "Artifact Creature — Phyrexian Construct", s82OneAmbulatory, 3, 2)
	answerMayChoice(t, g, me.ID, true)
	if me.Life != life-2 {
		t.Fatalf("life %d -> %d, want 2 paid", life, me.Life)
	}
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, victim); p != 1 || th != 1 {
		t.Errorf("target is %d/%d, want 1/1", p, th)
	}
}

func TestS82OneAmbulatoryEdificeDeclinedPaysNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	life := me.Life
	fdnCCast(t, g, "Ambulatory Edifice", "Artifact Creature — Phyrexian Construct", s82OneAmbulatory, 3, 2)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if me.Life != life || latestPickTarget(g, me.ID) != nil {
		t.Error("declining must pay nothing and raise no trigger")
	}
	if p, _ := effectivePTOf(t, g, victim); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
}

func TestS82OneAmbulatoryEdificeAtOneLifeIsNotAsked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Life = 1
	fdnCCast(t, g, "Ambulatory Edifice", "Artifact Creature — Phyrexian Construct", s82OneAmbulatory, 3, 2)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("asked to pay 2 life with only 1 (CR 119.4)")
	}
}

func TestS82OneShrapnelSlingerSacrificesToDestroyAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	rock := b12Creature(g, opp.ID, "Their Rock", "Artifact", 0, 0)
	mine := b12Creature(g, me.ID, "My Rock", "Artifact", 0, 0)
	castHoldCreature(t, g, "Shrapnel Slinger", "Artifact Creature — Phyrexian Beast", s82OneShrapnel, 2, 2)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{fodder}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no artifact target asked")
	}
	if hasID(pick.PickTargetCards, mine) {
		t.Error("your own artifact was offered")
	}
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("your artifact was destroyed")
	}
}

func TestS82OneShrapnelSlingerDeclinedDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	rock := b12Creature(g, opp.ID, "Their Rock", "Artifact", 0, 0)
	castHoldCreature(t, g, "Shrapnel Slinger", "Artifact Creature — Phyrexian Beast", s82OneShrapnel, 2, 2)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || latestPickTarget(g, me.ID) != nil {
		t.Error("declining the sacrifice must leave the artifact alone")
	}
}

func TestS82OneMyrCustodianScriesThenEachOpponentMay(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		seedLibrary(p, "A", "B", "C")
	}
	fdnCCast(t, g, "Myr Custodian", "Artifact Creature — Myr", s82OneMyrCustodian, 2, 3)
	if c := scryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("no scry 2 for the controller: %+v", g.PendingChoices)
	}
	if latestConfirmFor(g, g.Seats[1].ID) != nil {
		t.Error("an opponent was asked before the controller finished scrying")
	}
	answerScryKeepAll(t, g, me.ID)
	// Opponents in turn order: the first declines, the second scries 1, the third declines.
	answerMayChoice(t, g, g.Seats[1].ID, false)
	answerMayChoice(t, g, g.Seats[2].ID, true)
	if c := scryChoiceFor(g, g.Seats[2].ID); c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("no scry 1 for the second opponent: %+v", g.PendingChoices)
	}
	answerScryKeepAll(t, g, g.Seats[2].ID)
	answerMayChoice(t, g, g.Seats[3].ID, false)
	if scryChoiceFor(g, g.Seats[1].ID) != nil || scryChoiceFor(g, g.Seats[3].ID) != nil {
		t.Error("a declining opponent scried")
	}
}

func TestS82OneCruelGrimnarchDiscardsOrGainsFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.Seats[1].Hand.Cards = nil // can't discard
	g.Seats[2].Hand.Cards = nil // can't discard
	life := me.Life
	hand3 := g.Seats[3].Hand.Size()
	if hand3 == 0 {
		t.Fatal("test setup: seat 3 has no hand")
	}
	fdnCCast(t, g, "Cruel Grimnarch", "Creature — Phyrexian Cleric", s82OneGrimnarch, 5, 5)
	if discardChoiceFor(g, g.Seats[3].ID) == nil {
		t.Fatal("the opponent with cards was not asked to discard")
	}
	discardFromHand(t, g, g.Seats[3].ID)
	if me.Life != life+8 {
		t.Errorf("life %d -> %d, want +8 (two opponents who can't discard)", life, me.Life)
	}
	if g.Seats[3].Hand.Size() != hand3-1 {
		t.Errorf("hand %d -> %d", hand3, g.Seats[3].Hand.Size())
	}
}

func TestS82OneCruelGrimnarchGainsNothingWhenEveryoneDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	fdnCCast(t, g, "Cruel Grimnarch", "Creature — Phyrexian Cleric", s82OneGrimnarch, 5, 5)
	if me.Life != life {
		t.Errorf("life %d -> %d, want unchanged", life, me.Life)
	}
	for _, p := range g.Seats[1:] {
		if discardChoiceFor(g, p.ID) == nil {
			t.Errorf("%s was not asked to discard", p.Name)
		}
	}
}

func TestS82OneDuneMoverPutsABasicLandOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		searchTestLand("Mystic Sanctuary", "Land — Island"))
	mover := fdnCCast(t, g, "Dune Mover", "Artifact Creature — Phyrexian Golem", s82OneDuneMover, 2, 1)
	assertKeywords(t, g, mover, "toxic 1")
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	if searchOptionNamed(g, c, "Bear") != uuid.Nil || searchOptionNamed(g, c, "Mystic Sanctuary") != uuid.Nil {
		t.Error("only a basic land card is a legal find")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	passPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.Name != "Forest" {
		t.Fatalf("top card %q, want Forest", top.Name)
	}
	if me.Library.Size() != 3 {
		t.Errorf("library %d, the card never left it", me.Library.Size())
	}
}

func TestS82OneMyrKinsmithFindsAMyr(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Myr Token Maker", TypeLine: "Artifact Creature — Myr"})
	fdnCCast(t, g, "Myr Kinsmith", "Artifact Creature — Myr", s82OneMyrKinsmith, 3, 1)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search offered")
	}
	if searchOptionNamed(g, c, "Bear") != uuid.Nil {
		t.Error("a non-Myr was offered")
	}
	answerSearchNamed(t, g, me.ID, "Myr Token Maker")
	passPriorityAroundTable(t, g)
	found := false
	for _, h := range me.Hand.Cards {
		found = found || h.Name == "Myr Token Maker"
	}
	if !found {
		t.Error("the Myr is not in hand")
	}
}

func TestS82OneNimraiserPaladinReturnsACheapCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cheap := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	dear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear", TypeLine: "Creature — Bear", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	paladin := castHoldCreature(t, g, "Nimraiser Paladin", "Creature — Phyrexian Knight", s82OneNimraiser, 4, 4)
	assertKeywords(t, g, paladin, "toxic 2")
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target asked")
	}
	if hasID(pick.PickTargetCards, dear) {
		t.Error("a mana value 4 creature card was offered")
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) || !me.Graveyard.Contains(dear) {
		t.Error("the cheap creature should be in hand and the dear one still in the graveyard")
	}
}

func TestS82OneMeldwebCuratorPutsASpellOnTopAndMayChooseNone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	bear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	castHoldCreature(t, g, "Meldweb Curator", "Creature — Phyrexian Wizard", s82OneMeldweb, 3, 4)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target asked")
	}
	if hasID(pick.PickTargetCards, bear) {
		t.Error("a creature card was offered")
	}
	pickCard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != spell {
		t.Errorf("top card is %q, want Bolt", top.Name)
	}
}

func TestS82OneMeldwebCuratorWithNoSpellInGraveyardDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lib := me.Library.Size()
	castHoldCreature(t, g, "Meldweb Curator", "Creature — Phyrexian Wizard", s82OneMeldweb, 3, 4)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != lib {
		t.Error("library changed with no card to put")
	}
}
