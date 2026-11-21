package testsso

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b25_cards_test.go — Secrets of Strixhaven (sos), batch 25 (#4456).
// Each test plays the real catalog entry through the engine and asserts what
// the printed text says.

const (
	socB25FlowStateOracle  = "e7194bca-0cf4-4a86-8b84-d2f9219a08b3"
	socB25LumaretsOracle   = "5bb38d08-582e-43aa-9507-7c20759978ec"
	socB25FoolishOracle    = "9c51893f-9118-4d95-b88a-e264f4cbd7a8"
	socB25NecroscribeOrcl  = "50c97054-77c9-4446-9a41-e4f8877b887d"
	socB25ForumOracle      = "349ea6c7-6b3e-417f-b082-b712e2b1635b"
	socB25AnomalyOracle    = "644379e7-782e-48e1-9b47-c4dc62e62d85"
	socB25FractalizeOracle = "9e7bd032-8946-4d72-9910-b224823bc940"
	socB25MascotOracle     = "cca3f3e8-15a7-4ad8-893a-d3570178a800"
	socB25TenderOracle     = "cb45860a-1937-4803-b6d8-331a35be39fd"
)

func TestSocB25EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB25FlowStateOracle:  "Flow State",
		socB25LumaretsOracle:   "Follow the Lumarets",
		socB25FoolishOracle:    "Foolish Fate",
		socB25NecroscribeOrcl:  "Forum Necroscribe",
		socB25ForumOracle:      "Forum of Amity",
		socB25AnomalyOracle:    "Fractal Anomaly",
		socB25FractalizeOracle: "Fractalize",
		socB25MascotOracle:     "Fractal Mascot",
		socB25TenderOracle:     "Fractal Tender",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB25Active is the seat whose turn it is.
func socB25Active(g *game.Game) *game.Player { return g.Seats[g.Turn.ActiveSeat] }

// socB25Named lists the battlefield permanents with that name.
func socB25Named(g *game.Game, name string) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				out = append(out, c)
			}
		}
	})
	return out
}

// socB25LibraryTop is the id of the top card of the player's library (the
// last entry of Library.Cards: index 0 is the bottom).
func socB25LibraryTop(p *game.Player) uuid.UUID {
	return p.Library.Cards[len(p.Library.Cards)-1].InstanceID
}

// socB25InGraveyard puts a typed card into the player's graveyard.
func socB25InGraveyard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

// socB25DrawnThisTurn is how many cards the player has drawn this turn.
func socB25DrawnThisTurn(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = g.TurnTallyFor(p.ID).CardsDrawn })
	return n
}

// --- Flow State -------------------------------------------------------------

func socB25FlowStateTable(t *testing.T) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	ids := effectstest.SetLibraryTopFirst(me,
		game.Card{Name: "Top", TypeLine: "Creature — Bear"},
		game.Card{Name: "Middle", TypeLine: "Sorcery"},
		game.Card{Name: "Third", TypeLine: "Instant"},
		game.Card{Name: "Fourth", TypeLine: "Sorcery"},
		game.Card{Name: "Fifth", TypeLine: "Sorcery"})
	return g, me, ids
}

func TestSocB25FlowStateTakesOneOfThreeAndOrdersTheRestOnTheBottom(t *testing.T) {
	g, me, ids := socB25FlowStateTable(t)
	effectstest.CastCatalogSpell(t, g, "Flow State", "Sorcery", socB25FlowStateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no take prompt: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 3 {
		t.Fatalf("prompt min %d max %d over %d cards, want exactly one of the top three",
			pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	hand := me.Hand.Size()
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("take the middle card: %v", err)
	}
	if !me.Hand.Contains(ids[1]) || me.Hand.Size() != hand+1 {
		t.Fatalf("the middle card is not the one new card in hand")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom {
		t.Fatalf("the rest are not ordered on the bottom: %+v", order)
	}
	if err := g.ResolvePutInLibrary(order.ID, me.ID, []uuid.UUID{ids[2], ids[0]}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if got := effectstest.LibraryBottomIDs(me, 2); !effectstest.SameIDs(got, []uuid.UUID{ids[2], ids[0]}) {
		t.Errorf("bottom %v, want the player's order", got)
	}
	if socB25LibraryTop(me) != ids[3] {
		t.Error("the fourth card is not on top: only three were looked at")
	}
}

func TestSocB25FlowStateTakesTwoWithAnInstantAndASorceryInTheGraveyard(t *testing.T) {
	g, me, ids := socB25FlowStateTable(t)
	socB25InGraveyard(me, "Old Instant", "Instant")
	socB25InGraveyard(me, "Old Sorcery", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Flow State", "Sorcery", socB25FlowStateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no take prompt: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 {
		t.Fatalf("prompt min %d max %d, want exactly two", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[2]}); err != nil {
		t.Fatalf("take two: %v", err)
	}
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[2]) {
		t.Fatal("the two picked cards are not in hand")
	}
	// One card is left: it goes to the bottom with nothing to order.
	if got := effectstest.LibraryBottomIDs(me, 1); !effectstest.SameIDs(got, []uuid.UUID{ids[1]}) {
		t.Errorf("bottom %v, want the one card left over", got)
	}
}

func TestSocB25FlowStateNeedsBothKindsAndDoesNotCountItself(t *testing.T) {
	for name, kinds := range map[string][]string{
		"only an instant":  {"Instant"},
		"only a sorcery":   {"Sorcery"},
		"a creature, land": {"Creature — Bear", "Land"},
	} {
		t.Run(name, func(t *testing.T) {
			g, me, _ := socB25FlowStateTable(t)
			for _, k := range kinds {
				socB25InGraveyard(me, "Old", k)
			}
			effectstest.CastCatalogSpell(t, g, "Flow State", "Sorcery", socB25FlowStateOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
			if pick == nil || pick.ChooseMax != 1 {
				t.Fatalf("prompt %+v, want a take of exactly one", pick)
			}
		})
	}
}

func TestSocB25FlowStateOnAShortLibraryTakesWhatThereIsWithoutAsking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	socB25InGraveyard(me, "Old Instant", "Instant")
	socB25InGraveyard(me, "Old Sorcery", "Sorcery")
	ids := effectstest.SetLibraryTopFirst(me, game.Card{Name: "Only", TypeLine: "Sorcery"}, game.Card{Name: "Last", TypeLine: "Sorcery"})
	effectstest.CastCatalogSpell(t, g, "Flow State", "Sorcery", socB25FlowStateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked to choose two of two cards")
	}
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[1]) {
		t.Error("both library cards should have been taken")
	}
}

// --- Follow the Lumarets ----------------------------------------------------

func socB25LumaretsTable(t *testing.T) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	ids := effectstest.SetLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"},
		game.Card{Name: "Beyond", TypeLine: "Creature — Elk"})
	return g, me, ids
}

func TestSocB25LumaretsRevealsOneCreatureOrLandAndBottomsTheRest(t *testing.T) {
	g, me, ids := socB25LumaretsTable(t)
	effectstest.CastCatalogSpell(t, g, "Follow the Lumarets", "Sorcery", socB25LumaretsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no take prompt: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("prompt min %d max %d, want \"may\" up to one", pick.ChooseMin, pick.ChooseMax)
	}
	got := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		got[id] = true
	}
	if len(got) != 3 || !got[ids[0]] || !got[ids[2]] || !got[ids[3]] || got[ids[1]] {
		t.Fatalf("offered %v, want the creature and land cards among the top four only", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[2]}); err != nil {
		t.Fatalf("take the land: %v", err)
	}
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the land is not in hand")
	}
	// The other three looked-at cards are on the bottom; the fifth is on top.
	if socB25LibraryTop(me) != ids[4] {
		t.Error("the fifth card is not on top: only four were looked at")
	}
	bottom := map[uuid.UUID]bool{}
	for _, id := range effectstest.LibraryBottomIDs(me, 3) {
		bottom[id] = true
	}
	if !bottom[ids[0]] || !bottom[ids[1]] || !bottom[ids[3]] {
		t.Errorf("bottom three %v, want the looked-at cards that were not taken", bottom)
	}
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("a random-order bottom must not ask the player to order anything")
	}
}

func TestSocB25LumaretsWithoutLifeGainedOffersOnlyOneCard(t *testing.T) {
	g, me, ids := socB25LumaretsTable(t)
	effectstest.CastCatalogSpell(t, g, "Follow the Lumarets", "Sorcery", socB25LumaretsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[2]}); err == nil {
		t.Error("two cards were accepted without life gained this turn")
	}
}

func TestSocB25LumaretsAfterGainingLifeTakesTwoCreatureAndOrLandCards(t *testing.T) {
	g, me, ids := socB25LumaretsTable(t)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2); err != nil {
			t.Errorf("gain: %v", err)
		}
	})
	effectstest.CastCatalogSpell(t, g, "Follow the Lumarets", "Sorcery", socB25LumaretsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 2 {
		t.Fatalf("prompt min %d max %d, want \"may\" up to two", pick.ChooseMin, pick.ChooseMax)
	}
	// Two creatures is as good as a creature and a land.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[3]}); err != nil {
		t.Fatalf("take two creatures: %v", err)
	}
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[3]) {
		t.Fatal("the two creature cards are not in hand")
	}
}

func TestSocB25LumaretsMayTakeNothing(t *testing.T) {
	g, me, ids := socB25LumaretsTable(t)
	effectstest.CastCatalogSpell(t, g, "Follow the Lumarets", "Sorcery", socB25LumaretsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no take prompt")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	for _, id := range ids {
		if me.Hand.Contains(id) {
			t.Error("a card reached the hand after declining")
		}
	}
	if socB25LibraryTop(me) != ids[4] {
		t.Error("the four looked-at cards should all have gone to the bottom")
	}
}

// --- Foolish Fate -----------------------------------------------------------

func socB25CastFoolishFate(t *testing.T, g *game.Game, victim uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Foolish Fate", "Instant", socB25FoolishOracle, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB25FoolishFateDestroysAndCostsNothingWithoutLifeGained(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	before := opp.Life
	socB25CastFoolishFate(t, g, theirs)
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature survived")
	}
	if opp.Life != before {
		t.Errorf("their controller went from %d to %d life with no life gained", before, opp.Life)
	}
}

func TestSocB25FoolishFateInfusionMakesTheCreaturesControllerLoseThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Errorf("gain: %v", err)
		}
	})
	before, mine := opp.Life, me.Life
	socB25CastFoolishFate(t, g, theirs)
	if g.Battlefield.Contains(theirs) {
		t.Error("the creature survived")
	}
	if opp.Life != before-3 {
		t.Errorf("their controller is at %d, want %d", opp.Life, before-3)
	}
	if me.Life != mine {
		t.Errorf("the caster lost life: %d to %d", mine, me.Life)
	}
}

func TestSocB25FoolishFateOnYourOwnCreatureCostsYouTheLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Errorf("gain: %v", err)
		}
	})
	before := me.Life
	socB25CastFoolishFate(t, g, mine)
	if me.Life != before-3 {
		t.Errorf("caster at %d, want %d: the creature's controller is the caster", me.Life, before-3)
	}
}

func TestSocB25FoolishFateStillDrainsWhenTheCreatureIsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Theirs", TypeLine: "Creature — Golem", Power: 2, Toughness: 2,
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1); err != nil {
			t.Errorf("gain: %v", err)
		}
	})
	before := opp.Life
	socB25CastFoolishFate(t, g, theirs)
	if !g.Battlefield.Contains(theirs) {
		t.Error("an indestructible creature was destroyed")
	}
	if opp.Life != before-3 {
		t.Errorf("their controller is at %d, want %d", opp.Life, before-3)
	}
}

func TestSocB25FoolishFateRefusesANonCreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Foolish Fate", "Instant", socB25FoolishOracle, effectstest.CardRefs(rock)); err == nil {
		t.Error("an artifact was accepted as the target")
	}
}

// --- Forum of Amity ---------------------------------------------------------

func TestSocB25ForumOfAmityEntersTappedMakesWOrBAndSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	effectstest.SeedLibrary(me, "Top", "Next")
	id := effectstest.PlayLandFromHand(t, g, "Forum of Amity", socB25ForumOracle)
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.BattlefieldCard(g, id)
	if !ok || !card.Tapped {
		t.Fatalf("on battlefield %v, tapped %v; want it to enter tapped", ok, card.Tapped)
	}
	effectstest.UntapForTest(g, id)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Fatal("the land made {R}")
	}
	for _, color := range []string{"W", "B"} {
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{color}}); err != nil {
			t.Fatalf("{T}: Add {%s}: %v", color, err)
		}
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		effectstest.UntapForTest(g, id)
	}

	effectstest.FloatForTest(g, me, "WBC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("surveil activated with three mana")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "WBCC")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("{2}{W}{B}, {T}: surveil 1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID)
	if c == nil || len(c.ScryCards) != 1 {
		t.Fatalf("surveil prompt %v, want one looking at one card", c)
	}
	top := c.ScryCards[0]
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card is not in the graveyard")
	}
}

// --- Forum Necroscribe ------------------------------------------------------

func socB25NecroscribeOnBoard(g *game.Game) (necroscribe, theirs, dead uuid.UUID) {
	me, opp := g.Seats[0], g.Seats[1]
	necroscribe = effectstest.B12Push(g, me.ID, "Forum Necroscribe", "Creature — Troll Warlock", socB25NecroscribeOrcl, 5, 4)
	theirs = effectstest.B12Creature(g, opp.ID, "Theirs", socB22BearType, 2, 2)
	dead = socB25InGraveyard(me, "Fallen Bear", "Creature — Bear")
	return necroscribe, theirs, dead
}

func TestSocB25NecroscribeReparteeReturnsACreatureCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	necroscribe, theirs, dead := socB25NecroscribeOnBoard(g)
	effectstest.AdvanceToMain(t, g)
	if _, err := effectstest.CastBoltAtFrom(t, g, me, effectstest.CardRef(theirs)); err != nil {
		t.Fatalf("bolt a creature: %v", err)
	}
	if n := effectstest.TriggersOnStackFrom(g, necroscribe); n != 1 {
		// The target is chosen as the trigger goes on the stack.
		effectstest.B04WaitForPick(t, g, me.ID)
		effectstest.PickCard(t, g, me.ID, dead)
	}
	if n := effectstest.TriggersOnStackFrom(g, necroscribe); n != 1 {
		t.Fatalf("Repartee triggers %d, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card did not return to the battlefield")
	}
	c, _ := g.LookupCardForEffect(dead)
	if c.Controller != me.ID {
		t.Errorf("it came back under %v's control, want yours", c.Controller)
	}
}

func TestSocB25NecroscribeIgnoresASpellThatTargetsOnlyAPlayerAndAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	necroscribe, theirs, _ := socB25NecroscribeOnBoard(g)
	effectstest.AdvanceToMain(t, g)
	if _, err := effectstest.CastBoltAtFrom(t, g, me, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("bolt a player: %v", err)
	}
	if n := effectstest.TriggersOnStackFrom(g, necroscribe); n != 0 {
		t.Errorf("Repartee triggered %d times for a spell that targets no creature", n)
	}
	effectstest.PassPriorityAroundTable(t, g)

	// With no creature card in YOUR graveyard there is no target, so no
	// trigger, even though an opponent's graveyard holds one.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	nec2 := effectstest.B12Push(g2, me2.ID, "Forum Necroscribe", "Creature — Troll Warlock", socB25NecroscribeOrcl, 5, 4)
	their2 := effectstest.B12Creature(g2, opp2.ID, "Theirs", socB22BearType, 2, 2)
	theirDead := socB25InGraveyard(opp2, "Their Fallen", "Creature — Bear")
	effectstest.AdvanceToMain(t, g2)
	if _, err := effectstest.CastBoltAtFrom(t, g2, me2, effectstest.CardRef(their2)); err != nil {
		t.Fatalf("bolt a creature: %v", err)
	}
	if n := effectstest.TriggersOnStackFrom(g2, nec2); n != 0 {
		t.Errorf("Repartee put a trigger on the stack with no creature card in your graveyard: %d", n)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(theirDead) {
		t.Error("a creature card from an opponent's graveyard was returned")
	}
	_ = theirs
}

func TestSocB25NecroscribeHasWardDiscardACard(t *testing.T) {
	g, _, opp, warded := effectstest.WardTable(t, "Forum Necroscribe", "Creature — Troll Warlock", socB25NecroscribeOrcl)
	hand := effectstest.FillHand(opp, "Filler A", "Filler B")
	bolt := effectstest.BoltWardedCreature(t, g, opp, warded)
	prompt := effectstest.PassUntilConfirmFor(t, g, opp.ID)
	if prompt == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, hand[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(hand[0]) {
		t.Error("the ward's discard did not happen")
	}
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the Lightning Bolt should have resolved and gone to the graveyard")
	}
}

// --- Fractal Anomaly --------------------------------------------------------

func TestSocB25FractalAnomalyPutsOneCounterPerCardDrawnThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.DrawN(g, me, 3)
	want := socB25DrawnThisTurn(g, me)
	if want < 3 {
		t.Fatalf("setup: only %d cards drawn this turn", want)
	}
	effectstest.CastCatalogSpell(t, g, "Fractal Anomaly", "Instant", socB25AnomalyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	fractals := socB25Named(g, "Fractal")
	if len(fractals) != 1 {
		t.Fatalf("%d Fractals, want 1", len(fractals))
	}
	if p, tg := effectstest.SocB14PT(t, g, fractals[0].InstanceID); p != want || tg != want {
		t.Errorf("Fractal is %d/%d, want %d/%d", p, tg, want, want)
	}
	ch := effectstest.EffectiveCardForTest(t, g, fractals[0].InstanceID).Effective()
	if !reflect.DeepEqual(ch.Colors, []string{"G", "U"}) && !reflect.DeepEqual(ch.Colors, []string{"U", "G"}) {
		t.Errorf("Fractal colours %v, want green and blue", ch.Colors)
	}
}

func TestSocB25FractalAnomalyCountsOnlyYourOwnDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(opp, "A", "B", "C")
	effectstest.SeedLibrary(me, "X")
	base := socB25DrawnThisTurn(g, me)
	effectstest.DrawN(g, opp, 3)
	effectstest.CastCatalogSpell(t, g, "Fractal Anomaly", "Instant", socB25AnomalyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	fractals := socB25Named(g, "Fractal")
	if base == 0 {
		if len(fractals) != 0 {
			t.Errorf("%d Fractals survived with no card drawn by you", len(fractals))
		}
		return
	}
	if len(fractals) != 1 {
		t.Fatalf("%d Fractals, want 1", len(fractals))
	}
	if p, _ := effectstest.SocB14PT(t, g, fractals[0].InstanceID); p != base {
		t.Errorf("Fractal power %d, want %d: an opponent's draws do not count", p, base)
	}
}

func TestSocB25FractalAnomalyWithNoCardsDrawnLeavesAZeroZeroThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { g.TurnTally.Players[me.ID] = game.PlayerTurnTally{} })
	effectstest.CastCatalogSpell(t, g, "Fractal Anomaly", "Instant", socB25AnomalyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB25Named(g, "Fractal")); n != 0 {
		t.Errorf("%d Fractals on the battlefield, want the 0/0 gone to state-based actions", n)
	}
}

// --- Fractalize -------------------------------------------------------------

func socB25CastFractalize(t *testing.T, g *game.Game, victim uuid.UUID, x int) {
	t.Helper()
	me := socB25Active(g)
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Fractalize", TypeLine: "Instant", ManaCost: "{X}{U}", OracleID: socB25FractalizeOracle}
	if err := effectstest.SocB19CastWithPool(g, me, card, "U", x+1, game.CastSpellParams{XValue: x, Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("cast Fractalize for X=%d: %v", x, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB25FractalizeMakesTheCreatureAGreenAndBlueFractalOfSizeXPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Dragon", TypeLine: "Artifact Creature — Dragon Golem", Colors: []string{"R"},
		Power: 5, Toughness: 5, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	socB25CastFractalize(t, g, victim, 2)
	if p, tg := effectstest.SocB14PT(t, g, victim); p != 3 || tg != 3 {
		t.Errorf("creature is %d/%d, want 3/3", p, tg)
	}
	d := effectstest.EffectiveCardForTest(t, g, victim)
	ch := d.Effective()
	if !reflect.DeepEqual(ch.Subtypes, []string{"Fractal"}) {
		t.Errorf("subtypes %v, want only Fractal", ch.Subtypes)
	}
	if !reflect.DeepEqual(ch.Types, []string{"Artifact", "Creature"}) {
		t.Errorf("types %v, want it still an artifact creature", ch.Types)
	}
	if len(ch.Colors) != 2 || !((ch.Colors[0] == "G" && ch.Colors[1] == "U") || (ch.Colors[0] == "U" && ch.Colors[1] == "G")) {
		t.Errorf("colours %v, want green and blue only", ch.Colors)
	}
	if !effectstest.EffectiveHasKeyword(&d, "flying") {
		t.Error("it lost flying: Fractalize does not remove abilities")
	}
}

func TestSocB25FractalizeAtZeroIsAOneOneAndCountersStackOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Beast", "Creature — Beast", 6, 6)
	effectstest.AddCountersForTest(g, victim, game.CounterPlusOne, 2)
	socB25CastFractalize(t, g, victim, 0)
	if p, tg := effectstest.SocB14PT(t, g, victim); p != 3 || tg != 3 {
		t.Errorf("creature is %d/%d, want base 1/1 plus two counters", p, tg)
	}
}

func TestSocB25FractalizeWearsOffAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Beast", "Creature — Beast", 6, 6)
	socB25CastFractalize(t, g, victim, 1)
	if p, _ := effectstest.SocB14PT(t, g, victim); p != 2 {
		t.Fatalf("power %d during the turn, want 2", p)
	}
	effectstest.EndTurn(t, g)
	if p, tg := effectstest.SocB14PT(t, g, victim); p != 6 || tg != 6 {
		t.Errorf("creature is %d/%d next turn, want it back at 6/6", p, tg)
	}
	if ch := effectstest.EffectiveCardForTest(t, g, victim).Effective(); !reflect.DeepEqual(ch.Subtypes, []string{"Beast"}) {
		t.Errorf("subtypes %v next turn, want Beast", ch.Subtypes)
	}
}

func TestSocB25FractalizeCanTargetYourOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	socB25CastFractalize(t, g, mine, 4)
	if p, tg := effectstest.SocB14PT(t, g, mine); p != 5 || tg != 5 {
		t.Errorf("creature is %d/%d, want 5/5", p, tg)
	}
}

// --- Fractal Mascot ---------------------------------------------------------

func TestSocB25MascotTapsAnOpponentsCreatureAndStunsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", socB22BearType, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", socB22BearType, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", socB22BearType, 2, 2)
	id := effectstest.CastCatalogSpell(t, g, "Fractal Mascot", "Creature — Fractal Elk", socB25MascotOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The mascot is on the battlefield; its enters trigger asks for a target.
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Mascot is not on the battlefield")
	}
	tc, _ := g.LookupCardForEffect(theirs)
	if !tc.Tapped || tc.Counters[game.CounterStun] != 1 {
		t.Errorf("their creature tapped=%v stun=%d, want tapped with one stun counter", tc.Tapped, tc.Counters[game.CounterStun])
	}
	oc, _ := g.LookupCardForEffect(other)
	if oc.Tapped || oc.Counters[game.CounterStun] != 0 {
		t.Error("the other creature was affected")
	}
	mc, _ := g.LookupCardForEffect(mine)
	if mc.Tapped || mc.Counters[game.CounterStun] != 0 {
		t.Error("your own creature was affected")
	}
	c := effectstest.EffectiveCardForTest(t, g, id)
	if !effectstest.EffectiveHasKeyword(&c, "trample") {
		t.Error("Mascot has no trample")
	}
}

// --- Fractal Tender ---------------------------------------------------------

func socB25TenderOnBoard(g *game.Game) uuid.UUID {
	me := socB25Active(g)
	return effectstest.B12Push(g, me.ID, "Fractal Tender", "Creature — Elf Wizard", socB25TenderOracle, 3, 3)
}

// socB25CastForMana casts a vanilla sorcery with that many mana from `me`.
func socB25CastForMana(t *testing.T, g *game.Game, me *game.Player, n int) {
	t.Helper()
	cost := "{" + string(rune('0'+n-1)) + "}{G}"
	if err := effectstest.SocB19CastWithPool(g, me, game.Card{Name: "Filler Spell", TypeLine: "Sorcery", ManaCost: cost}, "G", n, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast a %d-mana spell: %v", n, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// socB25ToEndStep advances to the end step and settles its triggers.
func socB25ToEndStep(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestSocB25TenderMakesAFractalWithThreeCountersAtTheEndStepIfItGrewThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	tender := socB25TenderOnBoard(g)
	effectstest.AdvanceToMain(t, g)
	socB25CastForMana(t, g, me, 4)
	if p, tg := effectstest.SocB14PT(t, g, tender); p != 4 || tg != 4 {
		t.Fatalf("Tender is %d/%d after a four-mana spell, want Increment to make it 4/4", p, tg)
	}
	if n := len(socB25Named(g, "Fractal")); n != 0 {
		t.Fatalf("%d Fractals before the end step", n)
	}
	socB25ToEndStep(t, g)
	fractals := socB25Named(g, "Fractal")
	if len(fractals) != 1 {
		t.Fatalf("%d Fractals at the end step, want 1", len(fractals))
	}
	if p, tg := effectstest.SocB14PT(t, g, fractals[0].InstanceID); p != 3 || tg != 3 {
		t.Errorf("Fractal is %d/%d, want 3/3", p, tg)
	}
	if fractals[0].Controller != me.ID {
		t.Error("the Fractal is not yours")
	}
}

func TestSocB25TenderMakesNothingWhenNoCounterWasPutOnIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	tender := socB25TenderOnBoard(g)
	effectstest.AdvanceToMain(t, g)
	// Three mana is not greater than its power or toughness.
	socB25CastForMana(t, g, me, 3)
	if p, _ := effectstest.SocB14PT(t, g, tender); p != 3 {
		t.Fatalf("Tender grew to %d on a three-mana spell", p)
	}
	socB25ToEndStep(t, g)
	if n := len(socB25Named(g, "Fractal")); n != 0 {
		t.Errorf("%d Fractals with no counter put on Tender", n)
	}
}

func TestSocB25TenderMakesOnlyOneFractalForTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := socB25Active(g)
	tender := socB25TenderOnBoard(g)
	effectstest.AdvanceToMain(t, g)
	socB25CastForMana(t, g, me, 4)
	socB25CastForMana(t, g, me, 5)
	if p, _ := effectstest.SocB14PT(t, g, tender); p != 5 {
		t.Fatalf("Tender is %d power, want 5 after two increments", p)
	}
	socB25ToEndStep(t, g)
	if n := len(socB25Named(g, "Fractal")); n != 1 {
		t.Errorf("%d Fractals, want exactly 1 however many counters", n)
	}
}

func TestSocB25TenderIgnoresACounterAnotherPlayerPutOnIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	tender := socB25TenderOnBoard(g)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(opp.ID, tender, game.CounterPlusOne, 1); err != nil {
			t.Errorf("AddCounterBy: %v", err)
		}
	})
	socB25ToEndStep(t, g)
	if n := len(socB25Named(g, "Fractal")); n != 0 {
		t.Errorf("%d Fractals for a counter an opponent put on Tender", n)
	}
}

func TestSocB25TenderTriggersAtAnOpponentsEndStepToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tender := effectstest.B12Push(g, me.ID, "Fractal Tender", "Creature — Elf Wizard", socB25TenderOracle, 3, 3)
	effectstest.EndTurn(t, g)
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("setup: it is still my turn")
	}
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, tender, game.CounterPlusOne, 1); err != nil {
			t.Errorf("AddCounterBy: %v", err)
		}
	})
	socB25ToEndStep(t, g)
	fractals := socB25Named(g, "Fractal")
	if len(fractals) != 1 || fractals[0].Controller != me.ID {
		t.Fatalf("Fractals %d, want one under my control at the opponent's end step", len(fractals))
	}
	if p, _ := effectstest.SocB14PT(t, g, fractals[0].InstanceID); p != 3 {
		t.Errorf("Fractal power %d, want 3", p)
	}
}

func TestSocB25TenderCounterFromAnEarlierTurnDoesNotCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tender := effectstest.B12Push(g, me.ID, "Fractal Tender", "Creature — Elf Wizard", socB25TenderOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(me.ID, tender, game.CounterPlusOne, 1); err != nil {
			t.Errorf("AddCounterBy: %v", err)
		}
	})
	effectstest.EndTurn(t, g)
	effectstest.AdvanceToMain(t, g)
	socB25ToEndStep(t, g)
	if n := len(socB25Named(g, "Fractal")); n != 0 {
		t.Errorf("%d Fractals for a counter put on a previous turn", n)
	}
}

func TestSocB25TenderHasWardTwo(t *testing.T) {
	g, _, opp, warded := effectstest.WardTable(t, "Fractal Tender", "Creature — Elf Wizard", socB25TenderOracle)
	bolt := effectstest.BoltWardedCreature(t, g, opp, warded)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePayUnless, opp.ID)
	if prompt == nil {
		t.Fatalf("no ward {2} prompt for the spell's controller: %+v", g.PendingChoices)
	}
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(bolt) {
		t.Error("the countered spell should be in its owner's graveyard")
	}
	if effectstest.BoltResolved(g, warded) {
		t.Error("a countered Lightning Bolt must not damage the creature")
	}
}
