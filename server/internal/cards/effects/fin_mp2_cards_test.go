package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_mp2_cards_test.go — Final Fantasy (FIN), cards M–P, second batch
// (#4462). Each test plays the real catalog entry and asserts what its
// text says.

const (
	finMP2ScytheOracle    = "0d5bbde7-6f33-4708-b3ab-34d4528af649"
	finMP2MemoriesOracle  = "6cdc6548-cb32-437f-850a-9002542ad3d5"
	finMP2NibelheimOracle = "eab5eef8-d180-4b08-a770-6b3a1df03745"
	finMP2OmegaOracle     = "3173ca98-d46a-4de3-a95e-3eaad28d4017"
	finMP2PrisheOracle    = "bc92988c-d628-45ea-8606-516a1211cef3"
)

func TestFinMP2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		finMP2ScytheOracle:    "Magitek Scythe",
		finMP2MemoriesOracle:  "Memories Returning",
		finMP2NibelheimOracle: "Nibelheim Aflame",
		finMP2OmegaOracle:     "Omega, Heartless Evolution",
		finMP2PrisheOracle:    "Prishe's Wanderings",
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

// --- Magitek Scythe -----------------------------------------------------------------

// finMP2Scythe casts the Scythe, targets `bear` with its enters trigger
// and answers the "you may" with `attach`.
func finMP2Scythe(t *testing.T, g *game.Game, me, bear uuid.UUID, attach bool) uuid.UUID {
	t.Helper()
	scythe := castCatalogSpell(t, g, "Magitek Scythe", "Artifact — Equipment", finMP2ScytheOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me, bear)
	passPriorityAroundTable(t, g)
	ask := latestTriggerPromptFor(g, me)
	if ask == nil {
		t.Fatalf("the enters trigger asked nothing: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(ask.ID, me, attach); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	passPriorityAroundTable(t, g)
	return scythe
}

func TestFinMP2MagitekScytheAttachesAndItsCreatureMustBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	wall := reqCreature(g, opp.ID, "Wall", "", 0, 4)
	scythe := finMP2Scythe(t, g, me.ID, bear, true)
	if c, _ := g.LookupCardForEffect(scythe); c.AttachedTo.ID != bear {
		t.Fatalf("the Scythe is attached to %v, want the Bear", c.AttachedTo)
	}
	if v := s58p6View(t, g, bear); v.CurrentPower() != 4 || v.CurrentToughness() != 3 {
		t.Errorf("the Bear is %d/%d, want 4/3", v.CurrentPower(), v.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the Bear did not gain first strike")
	}
	reqToBlockers(t, g, bear)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); !strings.Contains(got, "must be blocked if able") {
		t.Errorf("sentence = %q", got)
	}
	if err := reqBlock(t, g, bear, wall); err != nil {
		t.Fatal(err)
	}
}

func TestFinMP2MagitekScytheGivesNothingWhenNotAttached(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := seats1684(g)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	scythe := finMP2Scythe(t, g, me.ID, bear, false)
	if c, _ := g.LookupCardForEffect(scythe); c.AttachedTo.ID == bear {
		t.Fatal("the Scythe attached after a no")
	}
	if hasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the Bear gained first strike without the attach")
	}
	if v := s58p6View(t, g, bear); v.CurrentPower() != 2 {
		t.Errorf("the Bear's power is %d, want 2", v.CurrentPower())
	}
}

// --- Memories Returning --------------------------------------------------------------

func TestFinMP2MemoriesReturningAlternatesHandAndBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	revealed := seedFactOrFictionLibrary(me, "MR One", "MR Two", "MR Three", "MR Four", "MR Five")
	hand := me.Hand.Size()

	castCatalogSpell(t, g, "Memories Returning", "Sorcery", finMP2MemoriesOracle, nil)
	passPriorityAroundTable(t, g)

	finMP2PickYours(t, g, me.ID, 5, revealed[0])
	answerChoosePlayer(t, g, me.ID, opp)
	finMP2PickTheirs(t, g, opp.ID, me.ID, 4, revealed[1])
	finMP2PickYours(t, g, me.ID, 3, revealed[2])
	finMP2PickTheirs(t, g, opp.ID, me.ID, 2, revealed[3])

	for _, id := range []uuid.UUID{revealed[0], revealed[2], revealed[4]} {
		if _, ok := cardInZone(me.Hand, id); !ok {
			t.Errorf("%s is not in hand", id)
		}
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want three cards more", hand, me.Hand.Size())
	}
	lib := me.Library.Cards
	if len(lib) < 2 || lib[0].InstanceID != revealed[3] || lib[1].InstanceID != revealed[1] {
		t.Errorf("the bottom of the library is not the opponent's two picks, last one lowest")
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("prompts left open: %+v", g.PendingChoices)
	}
}

// finMP2PickYours answers the controller's "put one into your hand".
func finMP2PickYours(t *testing.T, g *game.Game, me uuid.UUID, offered int, pick uuid.UUID) {
	t.Helper()
	p := latestChooseCardsFor(g, me)
	if p == nil {
		t.Fatalf("no hand pick for the controller: %+v", g.PendingChoices)
	}
	if len(p.ChooseCards) != offered || p.ChooseMin != 1 || p.ChooseMax != 1 {
		t.Fatalf("hand pick offers %d cards (%d..%d), want one of %d", len(p.ChooseCards), p.ChooseMin, p.ChooseMax, offered)
	}
	if err := g.ResolveChooseCards(p.ID, me, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// finMP2PickTheirs answers the opponent's "put one on the bottom".
func finMP2PickTheirs(t *testing.T, g *game.Game, opp, owner uuid.UUID, offered int, pick uuid.UUID) {
	t.Helper()
	p := latestRevealPickFor(g, opp)
	if p == nil {
		t.Fatalf("no bottom pick for the opponent: %+v", g.PendingChoices)
	}
	if p.FromPlayer != owner || len(p.ChooseCards) != offered || p.ChooseMin != 1 || p.ChooseMax != 1 {
		t.Fatalf("bottom pick from %s offers %d cards (%d..%d), want one of %d", p.FromPlayer, len(p.ChooseCards), p.ChooseMin, p.ChooseMax, offered)
	}
	for _, id := range p.ChooseCards {
		if c, ok := g.LookupCardForEffect(id); !ok || !c.KnownBy[opp] {
			t.Fatalf("the opponent can't see %s", id)
		}
	}
	if err := g.ResolveRevealPick(p.ID, opp, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
}

func TestFinMP2MemoriesReturningHasFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedFactOrFictionLibrary(me, "MR One", "MR Two", "MR Three", "MR Four", "MR Five")
	id := seedGraveyardCard(t, g, "Memories Returning", "Sorcery", finMP2MemoriesOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) == nil {
		t.Fatal("the flashed-back spell did not reveal and ask")
	}
}

// --- Nibelheim Aflame ----------------------------------------------------------------

func TestFinMP2NibelheimAflameBurnsEveryOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hero := b12Creature(g, me.ID, "Hero", "Creature — Human", 3, 3)
	mine := b12Creature(g, me.ID, "Big Friend", "Creature — Giant", 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := castCatalogSpellErr(t, g, "Nibelheim Aflame", "Sorcery", finMP2NibelheimOracle, finABCard(theirs)); err == nil {
		t.Fatal("Nibelheim Aflame targeted a creature you don't control")
	}
	pushCatalogHandCard(me, "Kept", "Instant", "")
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Nibelheim Aflame", "Sorcery", finMP2NibelheimOracle, finABCard(hero))
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(theirs) {
		t.Error("the 2/2 survived 3 damage")
	}
	if c, _ := g.LookupCardForEffect(mine); c.DamageMarked != 3 {
		t.Errorf("your other creature has %d damage, want 3", c.DamageMarked)
	}
	if c, _ := g.LookupCardForEffect(hero); c.DamageMarked != 0 {
		t.Errorf("the chosen creature dealt %d damage to itself", c.DamageMarked)
	}
	if me.Hand.Size() != hand {
		t.Errorf("cast from hand, the hand changed %d → %d", hand, me.Hand.Size())
	}
}

func TestFinMP2NibelheimAflameFromAGraveyardWheelsForFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hero := b12Creature(g, me.ID, "Hero", "Creature — Human", 1, 1)
	id := seedGraveyardCard(t, g, "Nibelheim Aflame", "Sorcery", finMP2NibelheimOracle)
	old := []uuid.UUID{pushCatalogHandCard(me, "Old One", "Instant", ""), pushCatalogHandCard(me, "Old Two", "Instant", "")}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Targets: finABCard(hero),
	}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range old {
		if _, ok := cardInZone(me.Graveyard, c); !ok {
			t.Errorf("%s was not discarded", c)
		}
	}
	if me.Hand.Size() != 4 {
		t.Errorf("hand is %d cards, want the four drawn", me.Hand.Size())
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back sorcery was not exiled")
	}
}

// --- Omega, Heartless Evolution -------------------------------------------------------

func TestFinMP2OmegaTapsAndStunsOnePermanentPerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	b12Permanent(g, me.ID, "Glacial Fortress", "Land")
	b12Permanent(g, me.ID, "Hallowed Fountain", "Land — Plains Island")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	aBear := b12Creature(g, a.ID, "A's Bear", "Creature — Bear", 2, 2)
	bRock := b12Permanent(g, b.ID, "B's Rock", "Artifact")
	cLand := b12Permanent(g, c.ID, "C's Land", "Land")
	life := me.Life

	castHoldCreature(t, g, "Omega, Heartless Evolution", "Legendary Artifact Creature — Robot", finMP2OmegaOracle, 8, 8)
	for i := 0; i < 12; i++ {
		hobAnswerTargets(t, g, me.ID, aBear, bRock, cLand)
		if len(g.StackMeta) == 0 && len(g.Stack.Cards) == 0 {
			break
		}
		passPriorityAroundTable(t, g)
	}
	for _, id := range []uuid.UUID{aBear, bRock} {
		card, _ := g.LookupCardForEffect(id)
		if !card.Tapped || card.Counters[game.CounterStun] != 2 {
			t.Errorf("%s: tapped %v, %d stun counters, want tapped with 2", card.Name, card.Tapped, card.Counters[game.CounterStun])
		}
	}
	if land, _ := g.LookupCardForEffect(cLand); land.Tapped || land.Counters[game.CounterStun] != 0 {
		t.Error("a land was tapped or stunned")
	}
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want two more (two nonbasic lands)", life, me.Life)
	}
}

// --- Prishe's Wanderings --------------------------------------------------------------

func TestFinMP2PrishesWanderingsFetchesATownAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	town := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: town, Name: "Gilded Town", TypeLine: "Land — Town", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Grove", TypeLine: "Land — Forest", Owner: me.ID, Controller: me.ID})

	castCatalogSpell(t, g, "Prishe's Wanderings", "Instant", finMP2PrisheOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Grove") != uuid.Nil {
		t.Error("a nonbasic land without the Town type was offered")
	}
	if searchOptionNamed(g, c, "Forest") == uuid.Nil {
		t.Error("a basic land was not offered")
	}
	answerSearchNamed(t, g, me.ID, "Gilded Town")
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(town) || !isTapped(g, town) {
		t.Error("the Town did not enter tapped")
	}
	if card, _ := g.LookupCardForEffect(bear); card.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", card.Counters[game.CounterPlusOne])
	}
}

func TestFinMP2PrishesWanderingsGrowsEvenWhenItFindsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Prishe's Wanderings", "Instant", finMP2PrisheOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchFailToFind(t, g, me.ID)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if card, _ := g.LookupCardForEffect(bear); card.Counters[game.CounterPlusOne] != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", card.Counters[game.CounterPlusOne])
	}
}
