package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// will_of_the_planeswalkers_5648_test.go — #5648: a spell that
// planeswalks or makes chaos ensue after a vote (CR 311.7, 701.31a,
// 701.38), the five Paths that print it, Ichor Elixir's extra planar die
// (CR 614.1a, 706.6), and Missy's "chaos ensues".

const (
	pathOfTheAnimistOracle     = "85ccae95-92c4-41ff-96fd-dcb515dc568e"
	pathOfTheEnigmaOracle      = "602f3203-91ae-423c-8a19-c6c1a5034f27"
	pathOfTheGhosthunterOracle = "b79c3ae8-ac25-4a16-a981-a41721040fad"
	pathOfThePyromancerOracle  = "b908fad9-25a6-441b-bba8-0d7fc12ee2dc"
	pathOfTheSchemerOracle     = "9cdb1cf6-2bf1-401b-83ae-6f3f7512e5a6"
	ichorElixirOracle          = "81b06e8d-ed30-4f90-b729-0aa93b0d870f"
)

func TestSeam5648CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		pathOfTheAnimistOracle:     "Path of the Animist",
		pathOfTheEnigmaOracle:      "Path of the Enigma",
		pathOfTheGhosthunterOracle: "Path of the Ghosthunter",
		pathOfThePyromancerOracle:  "Path of the Pyromancer",
		pathOfTheSchemerOracle:     "Path of the Schemer",
		ichorElixirOracle:          "Ichor Elixir",
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

// voteAll casts every ballot of the open vote for `label`, in the order
// the ballots open, and returns the voters in that order.
func voteAll(t *testing.T, g *game.Game, label string) []uuid.UUID {
	t.Helper()
	var order []uuid.UUID
	for range 16 {
		var voter uuid.UUID
		for _, p := range g.Seats {
			if p != nil && ballotFor(g, p.ID) != nil {
				voter = p.ID
				break
			}
		}
		if voter == uuid.Nil {
			return order
		}
		voteFor(t, g, voter, label)
		order = append(order, voter)
	}
	t.Fatal("the vote never finished")
	return nil
}

func alienAngels(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Alien Angel" && c.Controller == controller {
			n++
		}
	}
	return n
}

// Chaos gets more votes: chaos ensues (CR 311.7), so the face-up plane's
// chaos ability triggers, controlled by the planar controller, with no
// die rolled. Each player votes once, starting with the caster.
func TestWillOfThePlaneswalkersChaosEnsues(t *testing.T) {
	g, plane := newPlaneGame(t, planeAplanMortarium)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rolls := countEvents(g, game.EventRollDie)
	castCatalogSpell(t, g, "Path of the Ghosthunter", "Sorcery", pathOfTheGhosthunterOracle, nil)
	passPriorityAroundTable(t, g)
	if ballotFor(g, me.ID) == nil {
		t.Fatalf("the caster is not asked to vote first: %+v", g.PendingChoices)
	}
	if n := itemsFromPlane(g, plane.InstanceID); n != 0 {
		t.Fatal("the chaos ability triggered before the vote")
	}
	if order := voteAll(t, g, "chaos"); len(order) != 4 || order[0] != me.ID {
		t.Fatalf("voters %v, want all four starting with the caster", order)
	}
	if got := countEvents(g, game.EventChaosEnsues); got != 1 {
		t.Fatalf("%d chaos-ensues events, want 1", got)
	}
	if got := planeTriggerControllers(g, plane.InstanceID); len(got) != 1 || got[0] != me.ID {
		t.Fatalf("the chaos ability's controllers are %v, want the planar controller once", got)
	}
	passPriorityAroundTable(t, g)
	if got := alienAngels(g, me.ID); got != 2 {
		t.Errorf("Aplan Mortarium's chaos ability made %d Alien Angels, want 2", got)
	}
	if got := countEvents(g, game.EventRollDie); got != rolls {
		t.Errorf("chaos from a spell rolled a die: %d roll events, was %d", got, rolls)
	}
	if got := countEvents(g, game.EventPlaneswalkTo); got != 0 {
		t.Errorf("chaos won, yet %d planeswalks happened", got)
	}
	if up, ok := faceUpPlane(g, planeAplanMortarium.Oracle); !ok || up.InstanceID != plane.InstanceID {
		t.Error("the plane changed")
	}
}

// A tie is chaos ("or the vote is tied"); more votes for planeswalk
// planeswalk the caster away from the plane and to the next one, and no
// chaos ability triggers.
func TestWillOfThePlaneswalkersTieIsChaosAndPlaneswalkPlaneswalks(t *testing.T) {
	g, plane := newPlaneGame(t, planeAplanMortarium)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Path of the Ghosthunter", "Sorcery", pathOfTheGhosthunterOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range voteOrder(g) {
		label := "planeswalk"
		if i%2 == 1 {
			label = "chaos"
		}
		voteFor(t, g, p.ID, label)
	}
	if countEvents(g, game.EventChaosEnsues) != 1 || countEvents(g, game.EventPlaneswalkTo) != 0 {
		t.Fatalf("2-2 is a tie and goes to chaos: chaos %d, planeswalks %d",
			countEvents(g, game.EventChaosEnsues), countEvents(g, game.EventPlaneswalkTo))
	}
	passPriorityAroundTable(t, g)

	castCatalogSpell(t, g, "Path of the Ghosthunter", "Sorcery", pathOfTheGhosthunterOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range voteOrder(g) {
		label := "planeswalk"
		if i == 3 {
			label = "chaos"
		}
		voteFor(t, g, p.ID, label)
	}
	if countEvents(g, game.EventChaosEnsues) != 1 {
		t.Error("planeswalk won 3-1, yet chaos ensued")
	}
	if _, ok := faceUpPlane(g, planeAplanMortarium.Oracle); ok {
		t.Error("the caster did not planeswalk away from the plane")
	}
	up := g.PlanarFaceUpCards()
	if len(up) != 1 || up[0].InstanceID == plane.InstanceID {
		t.Fatalf("face up after the planeswalk: %+v", up)
	}
	walked := false
	for _, ev := range g.Events {
		walked = walked || (ev.Kind == game.EventPlaneswalkTo && ev.Actor == me.ID && ev.CardID == up[0].InstanceID)
	}
	if !walked {
		t.Error("no planeswalk-to event names the caster and the new plane")
	}
}

// CR 701.31a: only the planar controller may planeswalk. The vote of a
// player who is not the planar controller planeswalks nobody; their chaos
// still ensues, and the planar controller controls the chaos ability.
func TestWillOfThePlaneswalkersOnlyThePlanarControllerPlaneswalks(t *testing.T) {
	g, plane := newPlaneGame(t, planeAplanMortarium)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	start := func() {
		g.WithWriteLock(func() {
			item := &game.StackItem{Kind: game.StackItemSpell, Controller: opp.ID, Owner: opp.ID}
			if err := (WillOfThePlaneswalkers{Card: "Test Path"}).Apply(NewContext(g, item)); err != nil {
				t.Fatalf("the vote: %v", err)
			}
		})
	}
	start()
	if order := voteAll(t, g, "planeswalk"); len(order) != 4 || order[0] != opp.ID {
		t.Fatalf("voters %v, want all four starting with the vote's controller", order)
	}
	if countEvents(g, game.EventPlaneswalkTo) != 0 {
		t.Fatal("a player who is not the planar controller planeswalked")
	}
	if _, ok := faceUpPlane(g, planeAplanMortarium.Oracle); !ok {
		t.Fatal("the plane changed")
	}
	start()
	voteAll(t, g, "chaos")
	g.RunStateChecksForTest()
	if got := planeTriggerControllers(g, plane.InstanceID); len(got) != 1 || got[0] != me.ID {
		t.Errorf("the chaos ability's controllers are %v, want the planar controller once", got)
	}
}

// A vote waiting on a ballot is a restore point at a Planechase table,
// and the restored game makes chaos ensue and resolves the plane's chaos
// ability.
func TestWillOfThePlaneswalkersSurvivesARestore(t *testing.T) {
	g, _ := newPlaneGame(t, planeAplanMortarium)
	advanceToMain(t, g)
	order := voteOrder(g)
	me := order[0]
	castCatalogSpell(t, g, "Path of the Ghosthunter", "Sorcery", pathOfTheGhosthunterOracle, nil)
	passPriorityAroundTable(t, g)
	voteFor(t, g, order[0].ID, "chaos")
	voteFor(t, g, order[1].ID, "planeswalk")
	restored := restoreRoundTrip(t, g, true)
	voteFor(t, restored, order[2].ID, "chaos")
	voteFor(t, restored, order[3].ID, "chaos")
	if got := countEvents(restored, game.EventChaosEnsues); got != 1 {
		t.Fatalf("%d chaos-ensues events after the restore, want 1", got)
	}
	// The chaos ability waiting to resolve is a restore point too.
	restored.RunStateChecksForTest()
	again := restoreRoundTrip(t, restored, true)
	passPriorityAroundTable(t, again)
	if got := alienAngels(again, me.ID); got != 2 {
		t.Errorf("the chaos ability made %d Alien Angels across the restores, want 2", got)
	}
}

// At a table that is not playing Planechase the vote still happens, and
// neither result does anything.
func TestWillOfThePlaneswalkersWithoutAPlanarDeck(t *testing.T) {
	for _, label := range []string{"planeswalk", "chaos"} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		castCatalogSpell(t, g, "Path of the Ghosthunter", "Sorcery", pathOfTheGhosthunterOracle, nil)
		passPriorityAroundTable(t, g)
		if order := voteAll(t, g, label); len(order) != 4 || order[0] != me.ID {
			t.Fatalf("%s: voters %v, want all four starting with the caster", label, order)
		}
		if got := countEvents(g, game.EventVoteCast); got != 4 {
			t.Errorf("%s: %d votes logged, want 4", label, got)
		}
		if countEvents(g, game.EventChaosEnsues) != 0 || countEvents(g, game.EventPlaneswalkTo) != 0 {
			t.Errorf("%s: something happened without a planar deck", label)
		}
		if len(g.StackMeta) != 0 || len(g.PendingTriggers) != 0 || len(g.PendingChoices) != 0 {
			t.Errorf("%s: the spell left something behind", label)
		}
	}
}

// --- the cards -----------------------------------------------------------------

func TestPathOfTheAnimistFetchesTwoBasicsTappedThenVotes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	f1 := pushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	f2 := pushLibraryCard(me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	dual := pushLibraryCard(me, game.Card{Name: "Dual", TypeLine: "Land — Forest Island"})
	castCatalogSpell(t, g, "Path of the Animist", "Sorcery", pathOfTheAnimistOracle, nil)
	passPriorityAroundTable(t, g)
	if ballotFor(g, me.ID) != nil {
		t.Fatal("the vote opened before the search was answered")
	}
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{dual}); err == nil {
		t.Fatal("a nonbasic land was accepted")
	}
	answerSearchByID(t, g, me.ID, f1, f2)
	for _, id := range []uuid.UUID{f1, f2} {
		if !g.Battlefield.Contains(id) || !b12Card(t, g, id).Tapped {
			t.Errorf("basic land %v is not on the battlefield tapped", id)
		}
	}
	if order := voteAll(t, g, "planeswalk"); len(order) != 4 || order[0] != me.ID {
		t.Fatalf("voters after the search: %v", order)
	}
	spec, _ := Lookup(pathOfTheAnimistOracle)
	if spec.Purpose.Lands != 2 {
		t.Errorf("declared lands %d, want 2", spec.Purpose.Lands)
	}
}

func TestPathOfTheEnigmaTargetPlayerDrawsFourThenVotes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	if err := castCatalogSpellErr(t, g, "Path of the Enigma", "Sorcery", pathOfTheEnigmaOracle, nil); err == nil {
		t.Fatal("cast without a target")
	}
	mine, theirs := handSize(me), handSize(opp)
	castCatalogSpell(t, g, "Path of the Enigma", "Sorcery", pathOfTheEnigmaOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := handSize(opp); got != theirs+4 {
		t.Errorf("the target's hand %d → %d, want +4", theirs, got)
	}
	// The refused cast left its card in the caster's hand.
	if got := handSize(me); got != mine {
		t.Errorf("the caster's hand %d → %d, want unchanged", mine, got)
	}
	if order := voteAll(t, g, "chaos"); len(order) != 4 || order[0] != me.ID {
		t.Fatalf("voters after the draw: %v", order)
	}
}

// CR 608.2b: its only target gone, the spell does nothing, the vote
// included.
func TestPathOfTheEnigmaWithNoLegalTargetDoesNotVote(t *testing.T) {
	g := newCatalogGame(t)
	opp := opponentOf(g)
	castCatalogSpell(t, g, "Path of the Enigma", "Sorcery", pathOfTheEnigmaOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if err := g.Concede(opp.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countEvents(g, game.EventVoteCast) != 0 || len(g.PendingChoices) != 0 {
		t.Errorf("a spell with no legal target called a vote: %+v", g.PendingChoices)
	}
}

func TestPathOfTheGhosthunterMakesXSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	paCast(t, g, "Path of the Ghosthunter", "Sorcery", "{X}{1}{W}", pathOfTheGhosthunterOracle, game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	spirits := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID || !c.IsToken() {
			continue
		}
		spirits++
		if c.Name != "Spirit" || c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !game.HasKeyword(&c, "flying") ||
			len(c.Colors) != 1 || c.Colors[0] != "W" || !c.HasSubtypeWord("Spirit") {
			t.Errorf("token %+v is not a 1/1 white Spirit with flying", c)
		}
	}
	if spirits != 3 {
		t.Errorf("%d Spirits, want X = 3", spirits)
	}
	if order := voteAll(t, g, "chaos"); len(order) != 4 {
		t.Fatalf("voters after the tokens: %v", order)
	}

	// X = 0: no token, and the vote still happens.
	paCast(t, g, "Path of the Ghosthunter", "Sorcery", "{X}{1}{W}", pathOfTheGhosthunterOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	after := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() {
			after++
		}
	}
	if after != 3 {
		t.Errorf("X = 0 made %d tokens", after-3)
	}
	if order := voteAll(t, g, "chaos"); len(order) != 4 {
		t.Fatalf("X = 0: voters %v", order)
	}
}

func TestPathOfThePyromancerDiscardsAddsRedAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Path of the Pyromancer", "Sorcery", pathOfThePyromancerOracle, nil)
	hand := handSize(me)
	if hand == 0 {
		t.Fatal("setup: the caster has no hand to discard")
	}
	before := append([]game.Card(nil), me.Hand.Cards...)
	yard := me.Graveyard.Size()
	passPriorityAroundTable(t, g)
	for _, c := range before {
		if !me.Graveyard.Contains(c.InstanceID) {
			t.Errorf("hand card %s was not discarded", c.Name)
		}
	}
	if got := handSize(me); got != hand+1 {
		t.Errorf("hand after the draw is %d, want the %d discarded plus one", got, hand)
	}
	pool := poolColors(me)
	if len(pool) != hand {
		t.Fatalf("pool %v, want %d red mana", pool, hand)
	}
	for _, c := range pool {
		if c != "R" {
			t.Errorf("pool %v holds mana that is not red", pool)
		}
	}
	if order := voteAll(t, g, "chaos"); len(order) != 4 || order[0] != me.ID {
		t.Fatalf("voters after the draw: %v", order)
	}
	// The spell itself joins the discarded cards once it has resolved.
	if got := me.Graveyard.Size(); got != yard+hand+1 {
		t.Errorf("graveyard %d → %d, want the hand and the spell", yard, got)
	}
}

// An empty hand discards nothing, adds nothing and draws one card.
func TestPathOfThePyromancerWithAnEmptyHandDrawsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	g.WithWriteLock(func() {
		for me.Hand.Size() > 0 {
			c, err := me.Hand.PopTop()
			if err != nil {
				t.Fatalf("PopTop: %v", err)
			}
			me.Library.PushBottom(c)
		}
	})
	castCatalogSpell(t, g, "Path of the Pyromancer", "Sorcery", pathOfThePyromancerOracle, nil)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != 1 {
		t.Errorf("hand %d, want 1", got)
	}
	if pool := poolColors(me); len(pool) != 0 {
		t.Errorf("pool %v, want empty", pool)
	}
	voteAll(t, g, "chaos")
}

func TestPathOfTheSchemerMillsThenReanimatesAsAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	dead := b17GraveyardCard(opp, "Their Dead Bear", "Creature — Bear", "{1}{G}")
	bolt := b17GraveyardCard(opp, "Spent Bolt", "Instant", "{R}")
	milled := b10LibraryTop(me, "Milled Golem", "Artifact Creature — Golem", "{4}", 3, 3)
	libs := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		libs[i] = p.Library.Size()
	}
	castCatalogSpell(t, g, "Path of the Schemer", "Sorcery", pathOfTheSchemerOracle, nil)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if p.Library.Size() != libs[i]-2 {
			t.Errorf("seat %d mills two: %d → %d", i, libs[i], p.Library.Size())
		}
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil || !hasID(pick.ChooseCards, dead) || !hasID(pick.ChooseCards, milled) || hasID(pick.ChooseCards, bolt) {
		t.Fatalf("the pick offers every creature card in every graveyard, the milled one included, and no instant: %+v", pick)
	}
	if ballotFor(g, me.ID) != nil {
		t.Fatal("the vote opened before the creature was chosen")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err == nil {
		t.Fatal("choosing no creature card was accepted")
	}
	answerChooseCards(t, g, me.ID, dead)
	c := b12Card(t, g, dead)
	if !g.Battlefield.Contains(dead) || c.Controller != me.ID {
		t.Fatal("the creature is not on the battlefield under the caster's control")
	}
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtypeWord("Bear") {
		t.Errorf("it is %v %v, want an artifact creature that is still a Bear", c.Effective().Types, c.Effective().Subtypes)
	}
	if order := voteAll(t, g, "chaos"); len(order) != 4 || order[0] != me.ID {
		t.Fatalf("voters after the reanimation: %v", order)
	}
	endTurn(t, g)
	if !b12Card(t, g, dead).IsArtifact() {
		t.Error("it stopped being an artifact at end of turn; the effect has no duration")
	}
}

// With no creature card in any graveyard nothing enters, and the vote
// still happens.
func TestPathOfTheSchemerWithNoCreatureCardStillVotes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Path of the Schemer", "Sorcery", pathOfTheSchemerOracle, nil)
	passPriorityAroundTable(t, g)
	if chooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("a pick opened with no creature card in any graveyard")
	}
	if order := voteAll(t, g, "planeswalk"); len(order) != 4 {
		t.Fatalf("voters: %v", order)
	}
}

// --- Ichor Elixir ----------------------------------------------------------------

// planarRollsSince lists the faces of the planar rolls announced after
// the first `from` events.
func planarRollsSince(g *game.Game, from int) []string {
	var out []string
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventRollDie && ev.PlanarFace != "" {
			out = append(out, ev.PlanarFace)
		}
	}
	return out
}

// ignorePrompt is the open "choose a roll to ignore" question.
func ignorePrompt(t *testing.T, g *game.Game, chooser uuid.UUID, dice int) *game.PendingChoice {
	t.Helper()
	c := latestOptionPickFor(g, chooser)
	if c == nil || len(c.PickOptions) != dice {
		t.Fatalf("no ignore prompt with %d dice for the chooser: %+v", dice, g.PendingChoices)
	}
	return c
}

func TestIchorElixirRollsOneMorePlanarDieAndIgnoresOne(t *testing.T) {
	g, _ := newPlaneGame(t, planeAplanMortarium)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], opponentOf(g)
	elixir := b12Push(g, me.ID, "Ichor Elixir", "Artifact", ichorElixirOracle, 0, 0)

	// {T}: Add {C}{C}.
	b28TapForMana(t, g, me.ID, elixir, "")
	if got := poolColors(me); len(got) != 2 || got[0] != "C" || got[1] != "C" {
		t.Errorf("{T}: Add {C}{C}: pool %v", got)
	}

	// The planar die special action (CR 901.9): two dice, its controller
	// ignores one, one roll is announced.
	from := len(g.Events)
	if err := g.PerformSpecialAction(me.ID, uuid.Nil, game.SpecialActionPlanarDie, game.SpecialActionParams{}); err != nil {
		t.Fatalf("the planar die: %v", err)
	}
	if got := planarRollsSince(g, from); len(got) != 0 {
		t.Fatalf("a roll was announced before one was ignored: %v", got)
	}
	c := ignorePrompt(t, g, me.ID, 2)
	kept := c.PickOptions[1].Label
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ignore: %v", err)
	}
	got := planarRollsSince(g, from)
	if len(got) != 1 || kept != "Ignore the "+got[0]+" (die 2)" {
		t.Fatalf("kept rolls %v after ignoring die 1, with die 2 offered as %q", got, kept)
	}
	if g.Planechase.DieRolls != 1 {
		t.Errorf("the special action was counted %d times, want 1", g.Planechase.DieRolls)
	}
	passPriorityAroundTable(t, g)

	// A roll an effect asks for is replaced too, and the open question is
	// a restore point.
	from = len(g.Events)
	g.WithWriteLock(func() {
		if err := g.RollPlanarDieForEffect(me.ID, uuid.Nil); err != nil {
			t.Fatalf("RollPlanarDieForEffect: %v", err)
		}
	})
	ignorePrompt(t, g, me.ID, 2)
	restored := restoreRoundTrip(t, g, true)
	rc := ignorePrompt(t, restored, me.ID, 2)
	if err := restored.ResolveOptionPick(rc.ID, me.ID, 1); err != nil {
		t.Fatalf("ignore after the restore: %v", err)
	}
	if got := planarRollsSince(restored, from); len(got) != 1 {
		t.Fatalf("kept rolls after the restore: %v, want one", got)
	}

	// An opponent's planar roll is not "you"; a numbered die is not a
	// planar die.
	c = ignorePrompt(t, g, me.ID, 2)
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ignore: %v", err)
	}
	passPriorityAroundTable(t, g)
	from = len(g.Events)
	g.WithWriteLock(func() {
		if err := g.RollPlanarDieForEffect(opp.ID, uuid.Nil); err != nil {
			t.Fatalf("RollPlanarDieForEffect: %v", err)
		}
	})
	if got := planarRollsSince(g, from); len(got) != 1 || len(g.PendingChoices) != 0 {
		t.Fatalf("an opponent's roll was replaced: rolls %v, prompts %d", got, len(g.PendingChoices))
	}
	passPriorityAroundTable(t, g)
	rolls := countEvents(g, game.EventRollDie)
	g.WithWriteLock(func() {
		if err := g.RollDiceThenForEffect(game.DiceRoll{Player: me.ID, Sides: 20, N: 1}); err != nil {
			t.Fatalf("RollDiceThenForEffect: %v", err)
		}
	})
	if got := countEvents(g, game.EventRollDie); got != rolls+1 || len(g.PendingChoices) != 0 {
		t.Errorf("a d20 was replaced: %d roll events, prompts %d", got-rolls, len(g.PendingChoices))
	}
}

// Two Elixirs: that many plus one, twice, and one ignored for each.
func TestTwoIchorElixirsRollThreeAndIgnoreTwo(t *testing.T) {
	g, _ := newPlaneGame(t, planeAplanMortarium)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Ichor Elixir", "Artifact", ichorElixirOracle, 0, 0)
	b12Push(g, me.ID, "Ichor Elixir", "Artifact", ichorElixirOracle, 0, 0)
	from := len(g.Events)
	if err := g.PerformSpecialAction(me.ID, uuid.Nil, game.SpecialActionPlanarDie, game.SpecialActionParams{}); err != nil {
		t.Fatalf("the planar die: %v", err)
	}
	c := ignorePrompt(t, g, me.ID, 3)
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ignore: %v", err)
	}
	c = ignorePrompt(t, g, me.ID, 2)
	if err := g.ResolveOptionPick(c.ID, me.ID, 0); err != nil {
		t.Fatalf("ignore: %v", err)
	}
	if got := planarRollsSince(g, from); len(got) != 1 {
		t.Errorf("kept rolls %v, want one", got)
	}
}

// Missy's villainous choice says "chaos ensues" (CR 311.7).
func TestChaosEnsuesForEffectIsWhatMissySays(t *testing.T) {
	g, plane := newPlaneGame(t, planeAplanMortarium)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { g.ChaosEnsuesForEffect(opponentOf(g).ID, uuid.Nil) })
	g.RunStateChecksForTest()
	if got := planeTriggerControllers(g, plane.InstanceID); len(got) != 1 || got[0] != me.ID {
		t.Fatalf("the chaos ability's controllers are %v, want the planar controller once", got)
	}
}
