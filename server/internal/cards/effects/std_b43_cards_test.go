package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b43_cards_test.go — Standard set leftovers, batches 43-46 (#4452).
// Each test plays the real catalog entry and asserts what its text says.

const (
	stdSentryOracle  = "6b869c33-a77c-4eb7-9df9-9815d70f7444"
	stdGossipOracle  = "19dc29a7-8ccc-4e54-829b-86cc102b082d"
	stdHellionOracle = "e121f9cb-ed25-4be1-9822-b68cce2f2623"
)

func TestStdB43HellionDisguiseAndReach(t *testing.T) {
	spec, ok := Lookup(stdHellionOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Riftburst Hellion: %+v", spec)
	}
	ac := game.AlternativeCostByKey(stdHellionOracle, "disguise")
	if ac == nil || ac.FaceDown == nil || ac.FaceDown.Kind != game.FaceDownDisguised || ac.FaceDown.FaceUpCost != "{4}{R/G}{R/G}" {
		t.Fatalf("disguise offer: %+v", ac)
	}
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "reach" {
		t.Errorf("keywords %v, want reach", spec.PrintedKeywords)
	}
}

func TestStdB43SumalaSentryCountsBothWhenAPermanentTurnsUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sentry := b12Push(g, me.ID, "Sumala Sentry", "Creature — Elf Archer", stdSentryOracle, 3, 3)
	perf := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	// Entering face down is not turning face up.
	if c := tmtB1Card(t, g, sentry); c.Counters["+1/+1"] != 0 {
		t.Fatalf("a permanent entering face down gave the Sentry a counter")
	}
	mkmMPTurnFaceUp(t, g, me, perf, "R")
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	if c := tmtB1Card(t, g, sentry); c.Counters["+1/+1"] != 1 {
		t.Errorf("Sentry counters %d, want 1", c.Counters["+1/+1"])
	}
	if c := tmtB1Card(t, g, perf); c.Counters["+1/+1"] != 1 {
		t.Errorf("turned-up permanent counters %d, want 1", c.Counters["+1/+1"])
	}
}

func stdGossipMana(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	gossip := pushCatalogPermanent(g, me.ID, "Tin Street Gossip", "Creature — Lizard Advisor", stdGossipOracle, false)
	b28TapForMana(t, g, me.ID, gossip, "")
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("Gossip made %d mana, want {R}{G}", got)
	}
	return g, me
}

func TestStdB43TinStreetGossipManaPaysAFaceDownCast(t *testing.T) {
	g, me := stdGossipMana(t)
	b06AddMana(me, "C")
	id := handCardFull(me, "Pyrotechnic Performer", mkmMPPerformerTypeLine, "{1}{R}", mkmMPPerformerOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: "disguise", Strict: true}); err != nil {
		t.Fatalf("a face-down cast paid with Gossip mana: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left, want 0", got)
	}
}

func TestStdB43TinStreetGossipManaCannotPayAFaceUpCast(t *testing.T) {
	g, me := stdGossipMana(t)
	id := handCardFull(me, "Pyrotechnic Performer", mkmMPPerformerTypeLine, "{1}{R}", mkmMPPerformerOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a face-up cast was paid with Gossip's restricted mana")
	}
}

func TestStdB43TinStreetGossipManaTurnsACreatureFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gossip := pushCatalogPermanent(g, me.ID, "Tin Street Gossip", "Creature — Lizard Advisor", stdGossipOracle, false)
	perf := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	b28TapForMana(t, g, me.ID, gossip, "")
	if err := g.PerformSpecialAction(me.ID, perf, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up with Gossip mana: %v", err)
	}
	if findCardAnywhere(t, g, perf).FaceDown {
		t.Fatal("still face down")
	}
}

const stdTolsimirOracle = "0e9c7380-7584-43e6-8c31-5d1bb1685fe0"

func TestStdB43TolsimirCreatesVojaWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	castHoldCreature(t, g, "Tolsimir, Midnight's Light", "Legendary Creature — Elf Scout", stdTolsimirOracle, 3, 3)
	passPriorityAroundTable(t, g)
	var voja game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Voja Fenstalker" {
				voja = c
			}
		}
	})
	if voja.InstanceID == [16]byte{} || voja.Power != 5 || voja.Toughness != 5 || !voja.HasSubtype("Wolf") ||
		!hasKeywordOnBattlefield(t, g, voja.InstanceID, "trample") || len(voja.Colors) != 2 {
		t.Fatalf("Voja Fenstalker is %+v, want a legendary 5/5 green and white Wolf with trample", voja)
	}
}

func stdTolsimirBoard(g *game.Game) (me, opp *game.Player, tolsimir, wolf, victim uuid.UUID) {
	me, opp = seats1684(g)
	tolsimir = typedCreature(g, me.ID, "Tolsimir, Midnight's Light", "Legendary Creature — Elf Scout", stdTolsimirOracle, 3, 3, "haste")
	wolf = typedCreature(g, me.ID, "Test Wolf", "Creature — Wolf", "", 2, 2, "haste")
	victim = reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	return
}

func TestStdB43TolsimirForcesABlockOnAWolfWhenHeAttacksToo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, tolsimir, wolf, victim := stdTolsimirBoard(g)
	attackAndTrigger(t, g, tolsimir, wolf)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	recs := blocksAttackerRecords(g)
	if len(recs) != 1 || recs[0].Objects[0].ID != wolf {
		t.Fatalf("records = %+v, want one naming the Wolf", recs)
	}
	onToBlockers(t, g)
	reqRefusal(t, reqBlock(t, g, tolsimir, victim))
	if err := reqBlock(t, g, wolf, victim); err != nil {
		t.Fatalf("the forced block: %v", err)
	}
	_ = opp
}

func TestStdB43TolsimirDoesNothingWhenHeStaysHome(t *testing.T) {
	g := newCatalogGame(t)
	_, _, _, wolf, _ := stdTolsimirBoard(g)
	attackAndTrigger(t, g, wolf)
	passPriorityAroundTable(t, g)
	if recs := blocksAttackerRecords(g); len(recs) != 0 {
		t.Fatalf("records = %+v, want none: Tolsimir did not attack", recs)
	}
}

const stdSoulSearchOracle = "58f12f0d-e2c1-48f9-9e27-d1de8db743cb"

// stdSoulSearch casts Soul Search at the opponent holding `cards` and
// chooses `pick` (a hand index), returning the opponent and the picked id.
func stdSoulSearch(t *testing.T, pick int, cards ...game.Card) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ids := revealHand(opp, cards...)
	castAt(t, g, "Soul Search", "Sorcery", stdSoulSearchOracle, pbPlayer(opp.ID))
	p := openVariant(t, g)
	if err := g.ResolvePendingChoice(p.ID, me.ID, []uuid.UUID{ids[pick]}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	return g, me, opp, ids[pick]
}

func TestStdB43SoulSearchExilesAndMakesASpiritForAManaValueOneCard(t *testing.T) {
	g, me, _, id := stdSoulSearch(t, 1,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1})
	if !g.Exile.Contains(id) {
		t.Fatal("the chosen card was not exiled")
	}
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 1 {
		t.Fatalf("%d Spirit tokens, want 1", n)
	}
}

func TestStdB43SoulSearchMakesNoSpiritForAManaValueTwoCard(t *testing.T) {
	g, me, _, id := stdSoulSearch(t, 0,
		game.Card{Name: "Pricey", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2})
	if !g.Exile.Contains(id) {
		t.Fatal("the chosen card was not exiled")
	}
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 0 {
		t.Fatalf("%d Spirit tokens, want 0", n)
	}
}

func TestStdB43SoulSearchIgnoresLands(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealHand(opp, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	castAt(t, g, "Soul Search", "Sorcery", stdSoulSearchOracle, pbPlayer(opp.ID))
	if len(opp.Hand.Cards) != 1 {
		t.Fatalf("hand %d, want the land kept", len(opp.Hand.Cards))
	}
}

const stdRepulsiveOracle = "611bfaa6-b098-48d9-9e49-1e7dc8687f15"

// stdRepulsive: seat 0 casts a sorcery, seat 1 answers with Repulsive
// Mutation (X=2) on their 2/2, targeting that sorcery. Returns the table,
// the casting seat, the sorcery and the creature.
func stdRepulsive(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	caster, me := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	bear := reqCreature(g, me.ID, "My Bear", "", 2, 2)
	victim := castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	b06AddMana(me, "G", "U", "C", "C")
	id := handCardFull(me, "Repulsive Mutation", "Instant", "{X}{G}{U}", stdRepulsiveOracle, []string{"G", "U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}, {Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("casting Repulsive Mutation: %v", err)
	}
	passPriorityAroundTable(t, g)
	return g, caster, victim, bear
}

func TestStdB43RepulsiveMutationCountersUnlessThePowerIsPaid(t *testing.T) {
	g, caster, victim, bear := stdRepulsive(t)
	if c := tmtB1Card(t, g, bear); c.Counters["+1/+1"] != 2 {
		t.Fatalf("counters %d, want 2", c.Counters["+1/+1"])
	}
	// The greatest power among their creatures is the 4/4 Bear: {4}.
	pay := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the caster was not offered the payment")
	}
	if err := g.ResolvePayUnless(pay.ID, caster.ID, false); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if g.Stack.Contains(victim) {
		t.Fatal("the spell was not countered when its controller declined")
	}
}

func TestStdB43RepulsiveMutationLetsAPayerKeepTheSpell(t *testing.T) {
	g, caster, victim, _ := stdRepulsive(t)
	b06AddMana(caster, "C", "C", "C", "C")
	pay := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the caster was not offered the payment")
	}
	if err := g.ResolvePayUnless(pay.ID, caster.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if got := caster.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left, want the {4} paid", got)
	}
	_ = victim
}

const stdFlotsamOracle = "b0c28b2b-a2dd-4b76-bb1c-cec55a0a6784"

// stdCastFlotsamJetsam puts the split card in the active player's hand
// with a few cards on top of every opponent's library, casts the face and
// returns the table.
func stdCastFlotsamJetsam(t *testing.T, g *game.Game, face int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		for i := 0; i < 5; i++ {
			p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", ManaCost: "{9}", Owner: p.ID, Controller: p.ID})
		}
	}
	card := game.Card{
		InstanceID: uuid.New(), OracleID: stdFlotsamOracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Flotsam", TypeLine: "Instant", ManaCost: "{1}{G/U}", Colors: []string{"G", "U"}},
			{Name: "Jetsam", TypeLine: "Sorcery", ManaCost: "{4}{U/B}{U/B}", Colors: []string{"U", "B"}},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: face}); err != nil {
		t.Fatalf("cast the split card: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestStdB43FlotsamMillsThreeAndInvestigates(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gy := me.Graveyard.Size()
	stdCastFlotsamJetsam(t, g, 0)
	// The split card itself lands in the graveyard after resolving.
	if got := me.Graveyard.Size() - gy; got != 4 {
		t.Errorf("graveyard grew by %d, want 3 milled plus the spell", got)
	}
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Errorf("%d Clues, want 1", n)
	}
}

func TestStdB43JetsamMillsEachOpponentThenCastsFromTheirGraveyardFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	before := opp.Graveyard.Size()
	stdCastFlotsamJetsam(t, g, 1)
	if got := opp.Graveyard.Size() - before; got < 3 {
		t.Errorf("an opponent's graveyard grew by %d, want 3 milled", got)
	}
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast {
			for _, id := range c.DiscardOptions {
				if id == bolt {
					offer = c
				}
			}
		}
	}
	if offer == nil {
		t.Fatal("no cast offer naming the Bolt after the mill")
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, bolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("casting Lightning Bolt free from their graveyard: %v", err)
	}
	// Each opponent's graveyard has its own offer; leave the others.
	for _, c := range append([]*game.PendingChoice(nil), g.PendingChoices...) {
		if c != nil && c.Kind == game.PendingChoiceResolutionCast {
			if err := g.ResolvePendingChoice(c.ID, c.Chooser, nil); err != nil {
				t.Fatalf("declining another offer: %v", err)
			}
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bolt) {
		t.Error("the Bolt cast this way was not exiled instead of going to the graveyard")
	}
}

const stdBranchOracle = "7a30316b-dcd5-4a4b-b959-eecde7ca92e7"

func TestStdB43BranchOfVituGhaziAddsTwoManaOfOneColorWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := mkmMPCastFaceDown(t, g, "Branch of Vitu-Ghazi", "Land", stdBranchOracle, 0, 0)
	if c := findCardAnywhere(t, g, id); !c.FaceDownIsPermanent() {
		t.Fatal("did not enter face down")
	}
	mkmMPTurnFaceUp(t, g, me, id, "C", "C", "C")
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	if n := b10ResolveAllManaPicks(t, g, me.ID, "G"); n == 0 {
		t.Fatal("no colour was asked for")
	}
	g.ReadSnapshot(func() {
		if len(me.ManaPool) != 2 {
			t.Fatalf("pool %+v, want two mana", me.ManaPool)
		}
		for _, tok := range me.ManaPool {
			if tok.Color != "G" {
				t.Errorf("token %+v is not green: one colour for both", tok)
			}
			if len(tok.Riders) == 0 {
				t.Errorf("token %+v carries no keep-until-end-of-turn mark", tok)
			}
		}
	})
}

func TestStdB43BranchOfVituGhaziTapsForColorless(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	land := pushCatalogPermanent(g, me.ID, "Branch of Vitu-Ghazi", "Land", stdBranchOracle, false)
	b28TapForMana(t, g, me.ID, land, "")
	if got := me.ManaPool.Total(); got != 1 {
		t.Fatalf("%d mana, want {C}", got)
	}
}

const stdAssembleOracle = "180ba940-2b11-4953-b7ce-74847e1aebfd"

func stdAssembleTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Assemble the Players", "Enchantment", stdAssembleOracle, false)
	return g, me
}

func stdLibraryTop(me *game.Player, name string, power int, oracle string, extra func(*game.Card)) uuid.UUID {
	id := uuid.New()
	c := game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", ManaCost: "{0}", OracleID: oracle,
		Power: power, Toughness: 2, Owner: me.ID, Controller: me.ID}
	if extra != nil {
		extra(&c)
	}
	me.Library.PushTop(c)
	return id
}

func TestStdB43AssembleThePlayersCastsATwoPowerCreatureFromTheTopOncePerTurn(t *testing.T) {
	g, me := stdAssembleTable(t)
	first := stdLibraryTop(me, "Small Bear", 2, "", nil)
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("a power 2 creature from the top of the library: %v", err)
	}
	passPriorityAroundTable(t, g)
	second := stdLibraryTop(me, "Other Small Bear", 1, "", nil)
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a second creature spell from the top of the library was allowed in the same turn")
	}
}

func TestStdB43AssembleThePlayersRefusesAThreePowerCreature(t *testing.T) {
	g, me := stdAssembleTable(t)
	big := stdLibraryTop(me, "Big Bear", 3, "", nil)
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a power 3 creature was cast from the top of the library")
	}
}

func TestStdB43AssembleThePlayersRefusesANoncreatureSpell(t *testing.T) {
	g, me := stdAssembleTable(t)
	id := stdLibraryTop(me, "Ring", 0, "", func(c *game.Card) { c.TypeLine = "Artifact"; c.Toughness = 0 })
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("an artifact was cast from the top of the library")
	}
}

func TestStdB43AssembleThePlayersLetsYouLookAtTheTop(t *testing.T) {
	spec, ok := Lookup(stdAssembleOracle)
	if !ok || spec.Completeness != CompletenessFull || spec.LibraryTopVisible != game.LibraryTopOwner {
		t.Fatalf("Assemble the Players: %+v", spec)
	}
}

// A disguise card is judged as the 2/2 it is cast as (ruling): a power 5
// creature with disguise can be cast face down from the top.
func TestStdB43AssembleThePlayersCastsABigDisguiseCreatureFaceDown(t *testing.T) {
	g, me := stdAssembleTable(t)
	id := stdLibraryTop(me, "Hulking Spy", 5, mkmMPPerformerOracle, nil)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "library", AlternativeCost: "disguise", Strict: true}); err == nil {
		// {3} is not paid: a pool-less strict cast is expected to refuse on
		// mana, never on the permission. Add mana and try again below.
		t.Fatal("cast without mana")
	}
	b06AddMana(me, "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "library", AlternativeCost: "disguise", Strict: true}); err != nil {
		t.Fatalf("a power 5 disguise creature cast face down from the top: %v", err)
	}
}

const stdIntrudeOracle = "eedb1527-548a-435a-a3e1-8ad7d9d6d8a9"

func stdIntrudeThopter(g *game.Game, owner uuid.UUID) (game.Card, bool) {
	var out game.Card
	var ok bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == owner && c.Name == "Thopter" {
				out, ok = c, true
			}
		}
	})
	return out, ok
}

// You separate the five cards; the opponent you name chooses the pile that
// goes to your hand; the Thopter gets one counter per card binned.
func TestStdB43IntrudeOnTheMindPutsThePileInHandAndCountsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := seedFactOrFictionLibrary(me, "IM One", "IM Two", "IM Three", "IM Four", "IM Five")
	hand, graves := me.Hand.Size(), me.Graveyard.Size()

	castCatalogSpell(t, g, "Intrude on the Mind", "Instant", stdIntrudeOracle, nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	split := latestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatalf("the caster separates the piles: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, revealed[:2]); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	pick := latestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the named opponent chooses a pile: %+v", g.PendingChoices)
	}
	// Take the pile of three (the second option): two cards are binned.
	answerOptionPick(t, g, opp.ID, 1)
	passPriorityAroundTable(t, g)

	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d -> %d, want +3", hand, me.Hand.Size())
	}
	if me.Graveyard.Size() != graves+3 {
		t.Errorf("graveyard %d -> %d, want +3 (two binned cards and the spell)", graves, me.Graveyard.Size())
	}
	th, ok := stdIntrudeThopter(g, me.ID)
	if !ok {
		t.Fatal("no Thopter")
	}
	if th.Counters["+1/+1"] != 2 || th.CurrentPower() != 2 || !hasKeywordOnBattlefield(t, g, th.InstanceID, "flying") {
		t.Errorf("Thopter %d/%d with %d counters, want a 2/2 flyer", th.CurrentPower(), th.CurrentToughness(), th.Counters["+1/+1"])
	}
}

func TestStdB43IntrudeOnTheMindMakesAZeroZeroThopterWhenNothingIsBinned(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	revealed := seedFactOrFictionLibrary(me, "IM One", "IM Two", "IM Three", "IM Four", "IM Five")
	castCatalogSpell(t, g, "Intrude on the Mind", "Instant", stdIntrudeOracle, nil)
	passPriorityAroundTable(t, g)
	answerChoosePlayer(t, g, me.ID, opp)
	split := latestRevealPickFor(g, me.ID)
	if err := g.ResolveRevealPick(split.ID, me.ID, revealed); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	// All five in one pile, and the opponent hands it over.
	pick := latestOptionPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("no pile choice: %+v", g.PendingChoices)
	}
	for i, o := range pick.PickOptions {
		if len(o.Cards) == 5 {
			answerOptionPick(t, g, opp.ID, i)
		}
	}
	passPriorityAroundTable(t, g)
	// A 0/0 with no counters dies as state-based actions are checked.
	if _, ok := stdIntrudeThopter(g, me.ID); ok {
		t.Error("a 0/0 Thopter with no counters stayed on the battlefield")
	}
}

// --- batch 44 ------------------------------------------------------------

const (
	stdExitOracle      = "3d145811-adb2-46a4-b173-bf8152e5de48"
	stdWispdrinkerOrac = "6317d45a-fe02-4879-b515-bc5327ecb5ae"
)

func TestStdB44UnauthorizedExitBouncesThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Unauthorized Exit", "Instant", stdExitOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Hand.Contains(bear) {
		t.Fatal("the target was not returned to its owner's hand")
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoiceSurveil, me.ID); c == nil {
		t.Fatal("no surveil 1 after the bounce")
	}
}

func TestStdB44UnauthorizedExitCannotTargetALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	if err := castCatalogSpellErr(t, g, "Unauthorized Exit", "Instant", stdExitOracle, []game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Fatal("a land was targeted")
	}
}

func TestStdB44WispdrinkerVampireDrainsForASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opps := mkmMPOpponents(g)
	b12Push(g, me.ID, "Wispdrinker Vampire", "Creature — Vampire Rogue", stdWispdrinkerOrac, 3, 3)
	before, mine := mkmMPLives(opps), me.Life
	castHoldCreature(t, g, "Small Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	for i, p := range opps {
		if p.Life != before[i]-1 {
			t.Errorf("opponent %d life %d -> %d, want -1", i, before[i], p.Life)
		}
	}
	if me.Life != mine+1 {
		t.Errorf("life %d -> %d, want +1", mine, me.Life)
	}
}

func TestStdB44WispdrinkerVampireIgnoresABigCreatureAndItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opps := mkmMPOpponents(g)
	before := mkmMPLives(opps)
	castHoldCreature(t, g, "Wispdrinker Vampire", "Creature — Vampire Rogue", stdWispdrinkerOrac, 2, 2)
	castHoldCreature(t, g, "Big Bear", "Creature — Bear", "", 3, 3)
	passPriorityAroundTable(t, g)
	for i, p := range opps {
		if p.Life != before[i] {
			t.Errorf("opponent %d life changed %d -> %d", i, before[i], p.Life)
		}
	}
	_ = me
}

func TestStdB44WispdrinkerVampireGivesSmallCreaturesDeathtouchAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vamp := b12Push(g, me.ID, "Wispdrinker Vampire", "Creature — Vampire Rogue", stdWispdrinkerOrac, 3, 3)
	small := reqCreature(g, me.ID, "Small Bear", "", 2, 2)
	big := reqCreature(g, me.ID, "Big Bear", "", 4, 4)
	advanceToMain(t, g)
	floatForTest(g, me, "CCCCCWB")
	if err := g.ActivateCatalogAbility(me.ID, vamp, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !hasKeywordOnBattlefield(t, g, small, kw) {
			t.Errorf("the 2/2 lacks %s", kw)
		}
		if hasKeywordOnBattlefield(t, g, big, kw) {
			t.Errorf("the 4/4 gained %s", kw)
		}
	}
}

const stdNivOracle = "46570366-5e58-40b7-b725-833321a3b7b1"

func stdColored(g *game.Game, owner uuid.UUID, name string, colors ...string) {
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", Power: 1, Toughness: 1,
		Colors: colors, Owner: owner, Controller: owner,
	})
}

func TestStdB44NivMizzetCountsDifferentTwoColorPairs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	niv := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Niv-Mizzet, Guildpact", TypeLine: "Legendary Creature — Dragon Avatar", OracleID: stdNivOracle,
		Colors: []string{"W", "U", "B", "R", "G"}, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	stdColored(g, me.ID, "WU one", "W", "U")
	stdColored(g, me.ID, "UW two", "U", "W") // the same pair
	stdColored(g, me.ID, "BR", "B", "R")
	stdColored(g, me.ID, "GWU", "G", "W", "U") // three colours: no pair
	stdColored(g, me.ID, "Mono", "G")
	stdColored(g, opp.ID, "Their GR", "G", "R") // not yours
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	oppLife, myLife, hand := opp.Life, me.Life, me.Hand.Size()

	dealCombatDamageToPlayer(g, niv, opp.ID, 6)
	// Targets: any target (the opponent), then target player (me).
	pickPlayer(t, g, me.ID, opp.ID)
	pickPlayer(t, g, me.ID, me.ID)
	passPriorityAroundTable(t, g)

	if opp.Life != oppLife-2 {
		t.Errorf("opponent life %d, want %d (2 damage: two different pairs)", opp.Life, oppLife-2)
	}
	if me.Life != myLife+2 {
		t.Errorf("my life %d, want %d", me.Life, myLife+2)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

func TestStdB44NivMizzetHasHexproofFromMulticolored(t *testing.T) {
	spec, ok := Lookup(stdNivOracle)
	if !ok || spec.Completeness != CompletenessFull {
		t.Fatalf("Niv-Mizzet: %+v", spec)
	}
	q, ok := game.ParseHexproofFromQuality(game.HexproofFromMulticolored)
	if !ok || q.Kind != game.ProtectionQualityMulticolored {
		t.Fatalf("hexproof from multicolored does not parse: %+v %v", q, ok)
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	niv := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Niv-Mizzet, Guildpact", TypeLine: "Legendary Creature — Dragon Avatar", OracleID: stdNivOracle,
		Colors: []string{"W", "U", "B", "R", "G"}, Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	b06AddMana(opp, "R", "R")
	multi := handCardFull(opp, "Multi Bolt", "Instant", "{R}", lightningBoltOracle, []string{"R", "G"})
	mono := handCardFull(opp, "Mono Bolt", "Instant", "{R}", lightningBoltOracle, []string{"R"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass to the opponent: %v", err)
	}
	target := []game.TargetRef{{Kind: game.TargetCard, ID: niv}}
	if err := g.CastSpell(opp.ID, multi, game.CastSpellParams{Strict: true, Targets: target}); err == nil {
		t.Error("a multicolored spell targeted Niv-Mizzet")
	}
	if err := g.CastSpell(opp.ID, mono, game.CastSpellParams{Strict: true, Targets: target}); err != nil {
		t.Errorf("a monocolored spell could not target Niv-Mizzet: %v", err)
	}
}

const stdCodebreakerOracle = "af2b36ee-2945-41d3-a186-8fd28f3cde16"

func stdCodebreakerFaceDown(t *testing.T, g *game.Game, graveInstants int) (*game.Player, uuid.UUID) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < graveInstants; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Spent", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	// A creature card in the graveyard does not count.
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	id := mkmMPCastFaceDown(t, g, "Fugitive Codebreaker", "Creature — Goblin Rogue", stdCodebreakerOracle, 2, 2)
	return me, id
}

func TestStdB44FugitiveCodebreakerDisguiseCostIsReducedPerInstantAndSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, id := stdCodebreakerFaceDown(t, g, 3)
	// {5}{R} less three generic is {2}{R}: two mana is short.
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("R", "C")...) })
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("turned face up for {R}{C}: the cost is {2}{R} after three reductions")
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C")...) })
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up for {2}{R}: %v", err)
	}
}

func TestStdB44FugitiveCodebreakerReductionNeverTouchesTheColoredSymbol(t *testing.T) {
	g := newCatalogGame(t)
	me, id := stdCodebreakerFaceDown(t, g, 9)
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("C")...) })
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err == nil {
		t.Fatal("turned face up with no {R}")
	}
	g.WithWriteLock(func() { me.ManaPool.AddMana(manaTokens("R")...) })
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionTurnFaceUp, game.SpecialActionParams{Strict: true}); err != nil {
		t.Fatalf("turn face up for {R}: %v", err)
	}
}

func TestStdB44FugitiveCodebreakerDiscardsTheHandThenDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, id := stdCodebreakerFaceDown(t, g, 4)
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Draw", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	me.Hand.Cards = nil
	for i := 0; i < 2; i++ {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Held", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	graves := me.Graveyard.Size()
	mkmMPTurnFaceUp(t, g, me, id, "R", "C")
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3 drawn cards (the two held cards discarded)", got)
	}
	if got := me.Graveyard.Size() - graves; got != 2 {
		t.Errorf("graveyard grew by %d, want the 2 discarded cards", got)
	}
}

func TestStdB44FugitiveCodebreakerStillDrawsThreeWithAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me, id := stdCodebreakerFaceDown(t, g, 4)
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Draw", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	me.Hand.Cards = nil
	mkmMPTurnFaceUp(t, g, me, id, "R", "C")
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3", got)
	}
}

const stdWorldsoulOracle = "3b6ef144-bb98-4686-9718-204f1c3cf020"

func stdWorldsoulRage(t *testing.T, x int) (g *game.Game, me, opp *game.Player, handLand, graveLand, extra uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := func(z *game.Zone, name string) uuid.UUID {
		id := uuid.New()
		z.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
		return id
	}
	handLand, graveLand, extra = land(me.Hand, "Hand Forest"), land(me.Graveyard, "Grave Forest"), land(me.Hand, "Second Hand Forest")
	id := handCardFull(me, "Worldsoul's Rage", "Sorcery", "{X}{R}{G}", stdWorldsoulOracle, []string{"R", "G"})
	advanceToMain(t, g)
	b06AddMana(me, "R", "G")
	for i := 0; i < x; i++ {
		b06AddMana(me, "C")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: x,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	return
}

func TestStdB44WorldsoulsRageDamagesThenPutsUpToXLandsFromHandAndGraveyardTapped(t *testing.T) {
	life := 40
	g, me, opp, handLand, graveLand, extra := stdWorldsoulRage(t, 2)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatalf("no land pick: %+v", g.PendingChoices)
	}
	// Three candidate lands, a ceiling of X = 2 in all: one from each zone.
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{handLand, graveLand, extra}); err == nil {
		t.Fatal("three lands were put for X = 2")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{handLand, graveLand}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{handLand, graveLand} {
		c, ok := battlefieldCard(g, id)
		if !ok || !c.Tapped {
			t.Errorf("land %s: on battlefield %v tapped %v, want tapped on the battlefield", id, ok, c.Tapped)
		}
	}
	if _, ok := battlefieldCard(g, extra); ok {
		t.Error("the land that was not picked entered")
	}
}

func TestStdB44WorldsoulsRageMayPutNone(t *testing.T) {
	g, me, _, handLand, _, _ := stdWorldsoulRage(t, 1)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no land pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("choosing none: %v", err)
	}
	if _, ok := battlefieldCard(g, handLand); ok {
		t.Error("a land entered although none were chosen")
	}
}

const stdPrideOracle = "bbd7b756-bbee-451e-9446-06019791cb36"

func stdPrideCast(t *testing.T, toughness int, pool ...string) error {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if toughness > 0 {
		reqCreature(g, me.ID, "Wall", "", 0, toughness)
	}
	reqCreature(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Their Wall", "", 0, 50) // not yours
	id := handCardFull(me, "The Pride of Hull Clade", "Legendary Creature — Crocodile Elk Turtle", "{10}{G}", stdPrideOracle, []string{"G"})
	advanceToMain(t, g)
	b06AddMana(me, pool...)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
}

func TestStdB44PrideOfHullCladeCostsLessByTotalToughnessDownToItsColoredSymbol(t *testing.T) {
	if err := stdPrideCast(t, 12, "G"); err != nil {
		t.Fatalf("12 toughness should leave {G}: %v", err)
	}
	if err := stdPrideCast(t, 4, "G", "C", "C", "C", "C", "C", "C"); err != nil {
		t.Fatalf("4 toughness should leave {6}{G}: %v", err)
	}
	if err := stdPrideCast(t, 4, "G", "C", "C", "C", "C", "C"); err == nil {
		t.Fatal("cast for {5}{G} with only 4 toughness among your creatures")
	}
	if err := stdPrideCast(t, 12, "C"); err == nil {
		t.Fatal("cast without the {G}")
	}
}

func TestStdB44PrideOfHullCladeGrantsTheDrawTheDefenderWaiverAndPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pride := b12Push(g, me.ID, "The Pride of Hull Clade", "Legendary Creature — Crocodile Elk Turtle", stdPrideOracle, 5, 5)
	wall := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"},
		Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 6; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	advanceToMain(t, g)
	floatForTest(g, me, "CCUU")
	if err := g.ActivateCatalogAbility(me.ID, pride, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wall}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := mkmMPPT(t, g, wall); p != 1 {
		t.Errorf("Wall power %d, want 1", p)
	}
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, wall, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d cards, want 4 (the Wall's toughness)", got)
	}
}

func TestStdB44PrideOfHullCladeLetsADefenderAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pride := b12Push(g, me.ID, "The Pride of Hull Clade", "Legendary Creature — Crocodile Elk Turtle", stdPrideOracle, 5, 5)
	wall := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4, Keywords: []string{"defender"},
		Owner: me.ID, Controller: me.ID,
	})
	advanceToMain(t, g)
	floatForTest(g, me, "CCUU")
	if err := g.ActivateCatalogAbility(me.ID, pride, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: wall}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(wall, opp.ID); err != nil {
		t.Fatalf("the Wall could not attack: %v", err)
	}
}

// --- batch 45 ------------------------------------------------------------

const stdGreedOracle = "5483f787-7aca-4efe-a39a-ea29f8c07f0d"

func stdGreedCast(t *testing.T, dealt bool) (*game.Game, *game.Player, error, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	fighter := reqCreature(g, me.ID, "Fighter", "", 2, 2, "haste")
	idle := reqCreature(g, me.ID, "Idle", "", 2, 2)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	if dealt {
		dealCombatDamageToPlayer(g, fighter, opp.ID, 2)
		passPriorityAroundTable(t, g)
	}
	b06AddMana(me, "C", "W", "B")
	id := handCardFull(me, "Treacherous Greed", "Instant", "{1}{W}{B}", stdGreedOracle, []string{"W", "B"})
	victim := idle
	if dealt {
		victim = fighter
	}
	return g, me, g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{victim}}), victim
}

func TestStdB45TreacherousGreedSacrificesADamageDealerAndDrains(t *testing.T) {
	g, me, err, victim := stdGreedCast(t, true)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	opp := g.Seats[1]
	hand, oppLife, myLife := me.Hand.Size(), opp.Life, me.Life
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not sacrificed")
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if opp.Life != oppLife-3 || me.Life != myLife+3 {
		t.Errorf("life: opponent %d -> %d, mine %d -> %d; want -3 and +3", oppLife, opp.Life, myLife, me.Life)
	}
}

func TestStdB45TreacherousGreedRefusesACreatureThatDealtNoDamage(t *testing.T) {
	_, _, err, _ := stdGreedCast(t, false)
	if err == nil {
		t.Fatal("sacrificed a creature that dealt no damage this turn")
	}
}

const stdProftOracle = "b749e9e4-0f2a-4157-8cca-facd6b48157f"

func stdProftTable(t *testing.T, draws int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	for i := 0; i < 8; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	proft := b12Permanent(g, me.ID, "Proft's Eidetic Memory", "Legendary Enchantment")
	g.WithWriteLock(func() { g.Battlefield.Cards[len(g.Battlefield.Cards)-1].OracleID = stdProftOracle })
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2)
	if draws > 0 {
		g.WithWriteLock(func() {
			if err := g.DrawNForEffect(me.ID, draws); err != nil {
				t.Fatalf("draw: %v", err)
			}
		})
	}
	return g, me, proft, bear
}

func TestStdB45ProftsEideticMemoryPutsDrawsMinusOneCounters(t *testing.T) {
	// The draw step drew one card; two more make three this turn.
	g, me, _, bear := stdProftTable(t, 2)
	advanceTo(t, g, game.StepBeginCombat)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, bear); c.Counters["+1/+1"] != 2 {
		t.Errorf("counters %d, want 2 (three draws minus one)", c.Counters["+1/+1"])
	}
}

func TestStdB45ProftsEideticMemoryNeedsMoreThanOneDraw(t *testing.T) {
	// Only the draw step's card has been drawn.
	g, me, _, bear := stdProftTable(t, 0)
	advanceTo(t, g, game.StepBeginCombat)
	if latestPickTarget(g, me.ID) != nil || triggerOnStackAny(g) {
		t.Fatalf("triggered with only one card drawn this turn (tally %d, pick %v, stack %v)", g.TurnTallyFor(me.ID).CardsDrawn, latestPickTarget(g, me.ID) != nil, triggerOnStackAny(g))
	}
	if c := tmtB1Card(t, g, bear); c.Counters["+1/+1"] != 0 {
		t.Errorf("counters %d, want none", c.Counters["+1/+1"])
	}
}

func triggerOnStackAny(g *game.Game) bool {
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered {
			return true
		}
	}
	return len(g.PendingTriggers) > 0
}

func TestStdB45ProftsEideticMemoryDrawsWhenItEntersAndRemovesTheHandLimit(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Proft's Eidetic Memory", "Legendary Enchantment", stdProftOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1", got)
	}
	spec, _ := Lookup(stdProftOracle)
	if len(spec.HandSize) != 1 || spec.HandSize[0].Kind != game.HandSizeNoMaximum || spec.HandSize[0].Players != game.HandSizeYou {
		t.Errorf("hand size statics %+v, want no maximum for you", spec.HandSize)
	}
}

const stdHideOracle = "94b15f1d-827b-45e9-a684-5d0c52ea0a09"

func TestStdB45HideInPlainSightCloaksTwoAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	var ids []uuid.UUID
	for i := 0; i < 8; i++ {
		id := uuid.New()
		tl := "Creature — Bear"
		if i == 6 {
			tl = "Sorcery" // a nonpermanent card may be cloaked too
		}
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Lib", TypeLine: tl, ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		ids = append(ids, id)
	}
	top5 := []uuid.UUID{ids[7], ids[6], ids[5], ids[4], ids[3]}
	castCatalogSpell(t, g, "Hide in Plain Sight", "Sorcery", stdHideOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatalf("no pick: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 5 {
		t.Fatalf("%d candidates, want the top five", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[6]}); err == nil {
		t.Fatal("one card accepted for \"cloak two\"")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[6], ids[4]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{ids[6], ids[4]} {
		c, ok := battlefieldCard(g, id)
		if !ok || !c.FaceDownIsPermanent() || c.FaceDownKind != game.FaceDownCloaked || c.Controller != me.ID {
			t.Fatalf("card %s: on battlefield %v face down %v kind %q, want a cloaked permanent", id, ok, c.FaceDown, c.FaceDownKind)
		}
		if p, tough := mkmMPPT(t, g, id); p != 2 || tough != 2 || !c.FaceDownKind.HasWard() {
			t.Errorf("cloaked card is %d/%d (ward %v), want a 2/2 with ward", p, tough, c.FaceDownKind.HasWard())
		}
	}
	// Three cards were not chosen: they are at the bottom, below the three
	// cards that were never looked at.
	if me.Library.Size() != 6 {
		t.Fatalf("library %d, want 6", me.Library.Size())
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:3] {
		bottom[c.InstanceID] = true
	}
	for _, id := range []uuid.UUID{ids[7], ids[5], ids[3]} {
		if !bottom[id] {
			t.Errorf("unchosen card %s is not among the bottom three", id)
		}
	}
	_ = top5
}

const stdVannifarOracle = "e8aa0d7d-e5f6-4d66-bdc1-315483e1b256"

func TestStdB45VannifarCloaksACardFromYourHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Vannifar, Evolved Enigma", "Legendary Creature — Elf Ooze Wizard", stdVannifarOracle, 3, 3)
	me.Hand.Cards = nil
	keep := handCardFull(me, "Kept", "Sorcery", "{1}", "", nil)
	cloak := handCardFull(me, "Cloaked Card", "Instant", "{U}", "", []string{"U"})
	advanceTo(t, g, game.StepBeginCombat)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	choose := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if choose == nil {
		t.Fatalf("no card pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(choose.ID, me.ID, []uuid.UUID{cloak}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, cloak)
	if !ok || !c.FaceDownIsPermanent() || c.FaceDownKind != game.FaceDownCloaked {
		t.Fatalf("the chosen card: on battlefield %v, kind %q; want cloaked", ok, c.FaceDownKind)
	}
	if !me.Hand.Contains(keep) {
		t.Error("the other card left the hand")
	}
}

func TestStdB45VannifarPutsCountersOnColorlessCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Vannifar, Evolved Enigma", "Legendary Creature — Elf Ooze Wizard", stdVannifarOracle, 3, 3)
	stdColored(g, me.ID, "Green", "G")
	colorless := reqCreature(g, me.ID, "Golem", "", 2, 2)
	theirs := reqCreature(g, g.Seats[1].ID, "Their Golem", "", 2, 2)
	faceDown := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	advanceTo(t, g, game.StepBeginCombat)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{colorless: 1, faceDown: 1, theirs: 0} {
		if c := tmtB1Card(t, g, id); c.Counters["+1/+1"] != want {
			t.Errorf("%s has %d counters, want %d", c.Name, c.Counters["+1/+1"], want)
		}
	}
}

const stdEtrataOracle = "87957d09-7e22-40fe-8770-ff66d1934a38"

func stdEtrataBoard(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Etrata, Deadly Fugitive", "Legendary Creature — Vampire Assassin", stdEtrataOracle, 2, 3)
	return g, me
}

func TestStdB45EtrataGivesFaceDownCreaturesATurnFaceUpAbility(t *testing.T) {
	g, me := stdEtrataBoard(t)
	perf := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	floatForTest(g, me, "CCUB")
	if err := g.ActivateCatalogAbility(me.ID, perf, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findCardAnywhere(t, g, perf).FaceDown {
		t.Fatal("the creature is still face down")
	}
}

func TestStdB45EtrataExilesAndOffersAFreeCastWhenItCannotTurnFaceUp(t *testing.T) {
	g, me := stdEtrataBoard(t)
	opp := g.Seats[1]
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: lightningBoltOracle,
		Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		if err := g.PutCardsFromHandOntoBattlefieldKeyedForEffect([]uuid.UUID{bolt},
			game.HandEntryOptions{Controller: me.ID, FaceDown: game.FaceDownCloaked}, game.KeyedTail{}); err != nil {
			t.Fatalf("cloak: %v", err)
		}
	})
	if c, ok := battlefieldCard(g, bolt); !ok || !c.FaceDownIsPermanent() {
		t.Fatal("the cloaked Bolt is not a face-down permanent")
	}
	floatForTest(g, me, "CCUB")
	if err := g.ActivateCatalogAbility(me.ID, bolt, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bolt) {
		t.Fatal("an instant that cannot be turned face up was not exiled")
	}
	offer := latestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatalf("no free cast offer: %+v", g.PendingChoices)
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, bolt, game.CastSpellParams{FromZone: "exile",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast the exiled Bolt for free: %v", err)
	}
}

func TestStdB45EtrataCloaksTheTopCardWhenAnAssassinHitsAnOpponent(t *testing.T) {
	g, me := stdEtrataBoard(t)
	opp := g.Seats[1]
	topID := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: topID, Name: "Their Top", TypeLine: "Creature — Bear", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	assassin := reqCreature(g, me.ID, "Hired Knife", "", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == assassin {
				g.Battlefield.Cards[i].TypeLine = "Creature — Human Assassin"
			}
		}
	})
	other := reqCreature(g, me.ID, "Plain Bear", "", 1, 1)
	dealCombatDamageToPlayer(g, other, opp.ID, 1)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, topID); ok {
		t.Fatal("a non-Assassin's damage cloaked a card")
	}
	dealCombatDamageToPlayer(g, assassin, opp.ID, 1)
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, topID)
	if !ok || c.FaceDownKind != game.FaceDownCloaked || c.Controller != me.ID || c.Owner != opp.ID {
		t.Fatalf("their top card: on battlefield %v kind %q controller %v, want cloaked under my control", ok, c.FaceDownKind, c.Controller == me.ID)
	}
}

const stdGatekeeperOracle = "3c8fa799-7f30-4532-8317-f37a7ab787de"

func stdGatekeeperTurnUp(t *testing.T, target func(g *game.Game, me, opp *game.Player) uuid.UUID) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := mkmMPCastFaceDown(t, g, "Unyielding Gatekeeper", "Creature — Elephant Cleric", stdGatekeeperOracle, 2, 2)
	victim := target(g, me, opp)
	mkmMPTurnFaceUp(t, g, me, id, "W", "C")
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	return g, me, opp, victim
}

func TestStdB45UnyieldingGatekeeperReturnsYourPermanentTapped(t *testing.T) {
	g, _, _, _ := stdGatekeeperTurnUp(t, func(g *game.Game, me, _ *game.Player) uuid.UUID {
		return reqCreature(g, me.ID, "My Bear", "", 2, 2)
	})
	// The returned card is a new object (CR 400.7), found by name.
	var found, tapped int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "My Bear" {
				found++
				if c.Tapped {
					tapped++
				}
			}
		}
	})
	if found != 1 || tapped != 1 {
		t.Fatalf("%d My Bear on the battlefield, %d tapped; want it back, tapped", found, tapped)
	}
}

func TestStdB45UnyieldingGatekeeperGivesTheirControllerADetective(t *testing.T) {
	g, me, opp, theirs := stdGatekeeperTurnUp(t, func(g *game.Game, _, opp *game.Player) uuid.UUID {
		return reqCreature(g, opp.ID, "Their Bear", "", 2, 2)
	})
	if _, ok := battlefieldCard(g, theirs); ok || !g.Exile.Contains(theirs) {
		t.Fatal("their permanent was not exiled for good")
	}
	if n, _ := countTokensNamed(g, opp.ID, "Detective"); n != 1 {
		t.Errorf("their controller has %d Detectives, want 1", n)
	}
	if n, _ := countTokensNamed(g, me.ID, "Detective"); n != 0 {
		t.Errorf("I got %d Detectives, want 0", n)
	}
}

const stdFalconOracle = "4c495e94-3194-47c7-8129-dd0a0f915daa"

func TestStdB45CovetedFalconGivesPermanentsAwayAndDrawsForEach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := mkmMPCastFaceDown(t, g, "Coveted Falcon", "Artifact Creature — Bird", stdFalconOracle, 2, 2)
	a := reqCreature(g, me.ID, "Gift A", "", 1, 1)
	b := reqCreature(g, me.ID, "Gift B", "", 1, 1)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	hand := me.Hand.Size()
	mkmMPTurnFaceUp(t, g, me, id, "U", "C")
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	t.Logf("first prompt: players %v cards %v", pick.PickTargetPlayers, pick.PickTargetCards)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("player: %v", err)
	}
	pick = latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no second prompt: %+v", g.PendingChoices)
	}
	t.Logf("second prompt: players %v cards %v", pick.PickTargetPlayers, pick.PickTargetCards)
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("cards: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, p := range []uuid.UUID{a, b} {
		if c := tmtB1Card(t, g, p); c.Controller != opp.ID {
			t.Errorf("%s is controlled by %v, want the opponent", c.Name, c.Controller)
		}
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (one per permanent given)", got)
	}
}

func TestStdB45CovetedFalconTakesBackAPermanentYouOwnWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	falcon := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Coveted Falcon", TypeLine: "Artifact Creature — Bird", OracleID: stdFalconOracle,
		Power: 2, Toughness: 2, Keywords: []string{"flying", "haste"}, Owner: me.ID, Controller: me.ID,
	})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Lent Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID})
	theirs := reqCreature(g, opp.ID, "Their Own Bear", "", 2, 2)
	attackAndTrigger(t, g, falcon)
	// Only a permanent I own and don't control is a legal target.
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != mine {
		t.Fatalf("offered %v, want only the lent Bear (not %v)", pick.PickTargetCards, theirs)
	}
	pickCard(t, g, me.ID, mine)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, mine); c.Controller != me.ID {
		t.Error("the lent Bear was not taken back")
	}
}

const stdYarusOracle = "30fd6171-3e57-4c05-a86b-e1c0479ba236"

func stdYarusTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	yarus := b12Push(g, me.ID, "Yarus, Roar of the Old Gods", "Legendary Creature — Centaur Druid", stdYarusOracle, 3, 3)
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	return g, me, yarus
}

func TestStdB45YarusGivesOtherCreaturesHaste(t *testing.T) {
	g, me, yarus := stdYarusTable(t)
	other := reqCreature(g, me.ID, "Bear", "", 2, 2)
	theirs := reqCreature(g, g.Seats[1].ID, "Their Bear", "", 2, 2)
	if !hasKeywordOnBattlefield(t, g, other, "haste") {
		t.Error("another creature you control lacks haste")
	}
	if hasKeywordOnBattlefield(t, g, yarus, "haste") {
		t.Error("Yarus gave himself haste")
	}
	if hasKeywordOnBattlefield(t, g, theirs, "haste") {
		t.Error("an opponent's creature has haste")
	}
}

func TestStdB45YarusDrawsOncePerPlayerHitByFaceDownCreatures(t *testing.T) {
	g, me, _ := stdYarusTable(t)
	opp := g.Seats[1]
	a := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	b := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	faceUp := reqCreature(g, me.ID, "Bear", "", 2, 2)
	hand := me.Hand.Size()
	dealCombatDamageToPlayer(g, faceUp, opp.ID, 2)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Fatalf("a face-up creature's damage drew %d", me.Hand.Size()-hand)
	}
	dealCombatDamageToPlayer(g, a, opp.ID, 2)
	dealCombatDamageToPlayer(g, b, opp.ID, 2)
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1 (one player hit by two face-down creatures)", got)
	}
}

func TestStdB45YarusReturnsADeadFaceDownCreatureFaceUp(t *testing.T) {
	g, me, _ := stdYarusTable(t)
	id := mkmMPCastFaceDown(t, g, "Pyrotechnic Performer", mkmMPPerformerTypeLine, mkmMPPerformerOracle, 3, 2)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}) })
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	passPriorityAroundTable(t, g)
	var found, faceDown int
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Pyrotechnic Performer" && c.Controller == me.ID {
				found++
				if c.FaceDown {
					faceDown++
				}
			}
		}
	})
	if found != 1 || faceDown != 0 {
		t.Fatalf("%d Performers on the battlefield, %d face down; want one, turned face up", found, faceDown)
	}
}

func TestStdB45YarusIgnoresAFaceUpCreatureDying(t *testing.T) {
	g, me, _ := stdYarusTable(t)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{bear}) })
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, bear); ok {
		t.Fatal("the bear is on the battlefield")
	}
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Bear" {
				n++
			}
		}
	})
	if n != 0 {
		t.Errorf("a face-up Bear returned %d times", n)
	}
}

const stdVoltstriderOracle = "e7b746c8-1b32-42ed-8328-4e16274209d8"

func TestStdB45KyloxsVoltstriderLinksTheEvidenceAndCastsAnInstantFromIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	vehicle := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Kylox's Voltstrider", TypeLine: "Artifact — Vehicle", OracleID: stdVoltstriderOracle,
		Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	big := gyCard(me, "Big Creature", "Creature — Bear", "{5}", "")
	creatureCard := gyCard(me, "Dead Bear", "Creature — Bear", "{9}", "")
	_ = creatureCard

	if err := g.ActivateCatalogAbility(me.ID, vehicle, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bolt}}); err == nil {
		t.Fatal("collect evidence 6 paid with a total of 1")
	}
	if err := g.ActivateCatalogAbility(me.ID, vehicle, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{bolt, big}}); err != nil {
		t.Fatalf("collect evidence 6: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.InstanceID != bolt && c.InstanceID != big {
				continue
			}
			if c.ExiledWith.ID != vehicle {
				t.Errorf("%s is not linked to the Vehicle: %+v", c.Name, c.ExiledWith)
			}
		}
	})
	if c := tmtB1Card(t, g, vehicle); !c.IsCreature() {
		t.Fatal("the Vehicle did not become a creature")
	}

	attackAndTrigger(t, g, vehicle)
	passPriorityAroundTable(t, g)
	offer := latestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatalf("no cast offer: %+v", g.PendingChoices)
	}
	for _, id := range offer.DiscardOptions {
		if id == big {
			t.Error("a creature card was offered")
		}
	}
	b06AddMana(me, "R")
	if err := g.CastForResolutionChoice(me.ID, offer.ID, bolt, game.CastSpellParams{FromZone: "exile",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast the exiled Bolt: %v", err)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left: the Bolt must be paid for (it is not free)", got)
	}
	passPriorityAroundTable(t, g)
	if g.Exile.Contains(bolt) || me.Graveyard.Contains(bolt) {
		t.Errorf("the resolved Bolt is in exile or the graveyard (zone %v), not at the bottom of its owner's library; stack %d, pending %d", g.FindCardZoneForEffect(bolt).Kind, g.Stack.Size(), len(g.PendingChoices))
	}
	if me.Library.Size() == 0 || me.Library.Cards[0].InstanceID != bolt {
		t.Error("the Bolt is not on the bottom of its owner's library")
	}
}

const stdCoverUpOracle = "04e2afbe-0c08-42bc-93e6-662066a958bf"

func stdCoverUpCast(t *testing.T, evidence bool) (extractionTable, uuid.UUID, uuid.UUID) {
	t.Helper()
	x := extractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	advanceToMain(t, x.g)
	mine := reqCreature(x.g, x.me.ID, "My Bear", "", 2, 2)
	theirs := x.permanent("bear on board", "Grizzly Bears", "Creature — Bear", "G")
	var ev []uuid.UUID
	for _, mv := range []string{"{3}", "{3}"} {
		ev = append(ev, gyCard(x.me, "Evidence "+mv, "Sorcery", mv, ""))
	}
	b06AddMana(x.me, "C", "C", "C", "B", "B")
	id := handCardFull(x.me, "Deadly Cover-Up", "Sorcery", "{3}{B}{B}", stdCoverUpOracle, []string{"B"})
	params := game.CastSpellParams{Strict: true}
	if evidence {
		params.OptionalCosts = []int{0}
		params.ExileIDs = ev
	}
	if err := x.g.CastSpell(x.me.ID, id, params); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, x.g)
	return x, mine, theirs
}

func TestStdB45DeadlyCoverUpWithoutEvidenceOnlyDestroysCreatures(t *testing.T) {
	x, mine, theirs := stdCoverUpCast(t, false)
	if x.g.Battlefield.Contains(mine) || x.g.Battlefield.Contains(theirs) {
		t.Fatal("a creature survived")
	}
	if c := latestChoiceOfKindFor(x.g, game.PendingChoiceChooseCards, x.me.ID); c != nil {
		t.Fatalf("asked to exile a card without evidence: %+v", c)
	}
	x.notExiled(t, "gy", "hand", "lib")
}

func TestStdB45DeadlyCoverUpWithEvidenceExilesTheNameFromEveryZoneAndDrawsForTheHand(t *testing.T) {
	x, _, _ := stdCoverUpCast(t, true)
	pick := latestChoiceOfKindFor(x.g, game.PendingChoiceChooseCards, x.me.ID)
	if pick == nil {
		t.Fatalf("no pick of a card from an opponent's graveyard: %+v", x.g.PendingChoices)
	}
	for _, id := range pick.ChooseCards {
		c, _ := x.g.LookupCardForEffect(id)
		if c.Owner == x.me.ID {
			t.Fatal("one of my own cards was offered")
		}
	}
	if err := x.g.ResolveChooseCards(pick.ID, x.me.ID, []uuid.UUID{x.cards["gy"]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	x.exiled(t, "gy")
	hand, lib := x.opp.Hand.Size(), x.opp.Library.Size()
	answerSearchByID(t, x.g, x.me.ID, x.cards["gy 2"], x.cards["hand"], x.cards["lib"])
	passPriorityAroundTable(t, x.g)
	x.exiled(t, "gy", "gy 2", "hand", "lib")
	x.notExiled(t, "gy other", "hand other", "lib other", "gy forest")
	// One card left their hand, so they draw one; the library lost one and
	// was shuffled.
	if got := x.opp.Hand.Size() - (hand - 1); got != 1 {
		t.Errorf("they drew %d, want 1 (one card was exiled from their hand)", got)
	}
	if x.opp.Library.Size() != lib-1-1 {
		t.Errorf("library %d, want %d (one exiled, one drawn)", x.opp.Library.Size(), lib-2)
	}
	if x.searches() != 1 {
		t.Errorf("%d shuffles, want 1", x.searches())
	}
}

func TestStdB45DeadlyCoverUpMayExileABasicLand(t *testing.T) {
	x, _, _ := stdCoverUpCast(t, true)
	pick := latestChoiceOfKindFor(x.g, game.PendingChoiceChooseCards, x.me.ID)
	if pick == nil {
		t.Fatal("no pick")
	}
	if err := x.g.ResolveChooseCards(pick.ID, x.me.ID, []uuid.UUID{x.cards["gy forest"]}); err != nil {
		t.Fatalf("a basic land was refused: %v", err)
	}
	x.exiled(t, "gy forest")
}

// --- batch 46 ------------------------------------------------------------

const stdVojaOracle = "d94ff0b8-5499-4eeb-80f0-20c7fbf07499"

func TestStdB46VojaCountsElvesForCountersAndWolvesForCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	voja := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Voja, Jaws of the Conclave", TypeLine: "Legendary Creature — Wolf", OracleID: stdVojaOracle,
		Power: 5, Toughness: 5, Keywords: []string{"haste"}, Owner: me.ID, Controller: me.ID,
	})
	elfA := typedCreature(g, me.ID, "Elf A", "Creature — Elf Warrior", "", 1, 1)
	elfB := typedCreature(g, me.ID, "Elf B", "Creature — Elf Druid", "", 1, 1)
	wolf := typedCreature(g, me.ID, "Wolf", "Creature — Wolf", "", 2, 2)
	theirs := reqCreature(g, g.Seats[1].ID, "Their Elf", "", 1, 1)
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	hand := me.Hand.Size()
	attackAndTrigger(t, g, voja)
	passPriorityAroundTable(t, g)
	// Two Elves: two counters on each creature I control.
	for _, id := range []uuid.UUID{voja, elfA, elfB, wolf} {
		if c := tmtB1Card(t, g, id); c.Counters["+1/+1"] != 2 {
			t.Errorf("%s has %d counters, want 2", c.Name, c.Counters["+1/+1"])
		}
	}
	if c := tmtB1Card(t, g, theirs); c.Counters["+1/+1"] != 0 {
		t.Error("an opponent's creature got counters")
	}
	// Voja and the Wolf: two cards.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 (Voja and one other Wolf)", got)
	}
}

const (
	stdCrocodelfOracle = "3ba223b7-743f-4f99-836e-869778fb2e99"
	stdTeysaOracle     = "d4f5748f-bb0e-4c48-8802-45a7ad5a80a4"
)

func TestStdB46UndercoverCrocodelfInvestigatesOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	c := b12Push(g, me.ID, "Undercover Crocodelf", "Creature — Elf Crocodile Detective", stdCrocodelfOracle, 5, 5)
	dealCombatDamageToPlayer(g, c, opp.ID, 5)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 1 {
		t.Fatalf("%d Clues, want 1", n)
	}
	ac := game.AlternativeCostByKey(stdCrocodelfOracle, "disguise")
	if ac == nil || ac.FaceDown == nil || ac.FaceDown.FaceUpCost != "{3}{G/U}{G/U}" {
		t.Errorf("disguise offer %+v", ac)
	}
}

func stdTeysaBoard(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Teysa, Opulent Oligarch", "Legendary Creature — Human Advisor", stdTeysaOracle, 2, 4)
	return g, me
}

func TestStdB46TeysaInvestigatesForEachOpponentWhoLostLife(t *testing.T) {
	g, me := stdTeysaBoard(t)
	g.WithWriteLock(func() {
		for _, i := range []int{1, 2} {
			if err := g.ChangePlayerLifeForEffect(uuid.Nil, g.Seats[i].ID, -1); err != nil {
				t.Fatalf("lose life: %v", err)
			}
		}
	})
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Clue"); n != 2 {
		t.Fatalf("%d Clues, want 2 (two opponents lost life)", n)
	}
}

func TestStdB46TeysaMakesASpiritOnlyOnceEachTurnForAClue(t *testing.T) {
	g, me := stdTeysaBoard(t)
	clueA := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clue", TypeLine: "Token Artifact — Clue", Owner: me.ID, Controller: me.ID})
	clueB := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clue", TypeLine: "Token Artifact — Clue", Owner: me.ID, Controller: me.ID})
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clue", TypeLine: "Token Artifact — Clue", Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{theirs}) })
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 0 {
		t.Fatalf("an opponent's Clue made %d Spirits", n)
	}
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{clueA}) })
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{clueB}) })
	g.RunStateChecksForTest()
	b30SettleOrder(t, g)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Spirit"); n != 1 {
		t.Fatalf("%d Spirits, want 1 (once each turn)", n)
	}
}

const stdRakdosOracle = "d08ce18a-0fd9-47ac-ad2f-6934b60070f1"

// stdRakdosEndStep starts Rakdos' end-step trigger against the first opponent,
// who controls `perms` nonland nontoken permanents (and a land and a token
// that never count). Returns the table and the opponent's permanents.
func stdRakdosEndStep(t *testing.T, perms int) (*game.Game, *game.Player, *game.Player, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Rakdos, Patron of Chaos", "Legendary Creature — Demon", stdRakdosOracle, 6, 6)
	var mine []uuid.UUID
	for i := 0; i < perms; i++ {
		mine = append(mine, reqCreature(g, opp.ID, "Their Perm", "", 2, 2))
	}
	b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Token", TypeLine: "Token Creature — Bear", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	for i := 0; i < 5; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	advanceTo(t, g, game.StepEnd)
	pickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	return g, me, opp, mine
}

func TestStdB46RakdosLetsYouDrawTwoWhenTheyDecline(t *testing.T) {
	g, me, opp, mine := stdRakdosEndStep(t, 3)
	hand := me.Hand.Size()
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if c == nil {
		t.Fatalf("the opponent was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, opp.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	for _, id := range mine {
		if _, ok := battlefieldCard(g, id); !ok {
			t.Error("a permanent was sacrificed although they declined")
		}
	}
}

func TestStdB46RakdosTakesTwoSacrificesAndNoCards(t *testing.T) {
	g, me, opp, mine := stdRakdosEndStep(t, 3)
	hand := me.Hand.Size()
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if err := g.ResolveConfirm(c.ID, opp.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID)
	if pick == nil {
		t.Fatalf("no pick: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 3 {
		t.Errorf("%d candidates, want 3 (not the land or the token)", len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, mine[:1]); err == nil {
		t.Fatal("one permanent accepted for \"two\"")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, mine[:2]); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range mine[:2] {
		if _, ok := battlefieldCard(g, id); ok {
			t.Error("a chosen permanent was not sacrificed")
		}
	}
	if _, ok := battlefieldCard(g, mine[2]); !ok {
		t.Error("the third permanent was sacrificed")
	}
	if me.Hand.Size() != hand {
		t.Errorf("drew %d although they sacrificed", me.Hand.Size()-hand)
	}
}

func TestStdB46RakdosDrawsWithoutAskingWhenTheyHaveFewerThanTwoPermanents(t *testing.T) {
	g, me, opp, _ := stdRakdosEndStep(t, 1)
	if c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID); c != nil {
		t.Fatal("asked an opponent who cannot sacrifice two permanents")
	}
	// The draw step's card, then Rakdos' two.
	if got := g.TurnTallyFor(me.ID).CardsDrawn; got != 3 {
		t.Errorf("drew %d cards this turn, want 3", got)
	}
}

const stdBreakOutOracle = "99eaac39-28d4-4af7-984b-f75fd37902ed"

// stdBreakOutTable stocks the library with eight cards (the top six are
// looked at): a creature of mana value `mv` at the top, then five others.
func stdBreakOutTable(t *testing.T, mv string, creatures bool) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	var rest []uuid.UUID
	for i := 0; i < 7; i++ {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: "Other", TypeLine: "Sorcery", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		rest = append(rest, id)
	}
	creature := uuid.New()
	tl := "Creature — Bear"
	if !creatures {
		tl = "Sorcery"
	}
	me.Library.PushTop(game.Card{InstanceID: creature, Name: "Prize", TypeLine: tl, ManaCost: mv, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Break Out", "Sorcery", stdBreakOutOracle, nil)
	passPriorityAroundTable(t, g)
	return g, me, creature, rest
}

func TestStdB46BreakOutPutsACheapCreatureOntoTheBattlefieldWithHaste(t *testing.T) {
	g, me, creature, _ := stdBreakOutTable(t, "{1}{G}", true)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatalf("no reveal pick: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if c == nil {
		t.Fatalf("a mana value 2 creature was not offered the battlefield: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	var found game.Card
	g.ReadSnapshot(func() {
		for _, bc := range g.Battlefield.Cards {
			if bc.Name == "Prize" {
				found = bc
			}
		}
	})
	if found.InstanceID == uuid.Nil || !hasKeywordOnBattlefield(t, g, found.InstanceID, "haste") {
		t.Fatalf("the Prize is not on the battlefield with haste: %+v", found)
	}
	// Seven cards were left in the library; the five other cards looked at
	// are at the bottom.
	if me.Library.Size() != 7 {
		t.Errorf("library %d, want 7", me.Library.Size())
	}
}

func TestStdB46BreakOutPutsAnExpensiveCreatureIntoYourHand(t *testing.T) {
	g, me, creature, _ := stdBreakOutTable(t, "{3}{G}", true)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature}); err != nil {
		t.Fatalf("reveal: %v", err)
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID); c != nil {
		t.Fatal("offered the battlefield for a mana value 4 creature")
	}
	if !me.Hand.Contains(creature) {
		t.Error("the revealed creature is not in the hand")
	}
}

func TestStdB46BreakOutDeclinedGoesToHandAndNoCreatureMeansAllToTheBottom(t *testing.T) {
	g, me, creature, _ := stdBreakOutTable(t, "{G}", true)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	_ = g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{creature})
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if err := g.ResolveConfirm(c.ID, me.ID, false); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if !me.Hand.Contains(creature) {
		t.Error("a declined card did not go to the hand")
	}

	g2, me2, _, _ := stdBreakOutTable(t, "{1}", false)
	if p := latestChoiceOfKindFor(g2, game.PendingChoiceChooseCards, me2.ID); p != nil {
		t.Fatalf("asked to reveal with no creature card among them: %+v", p)
	}
	if me2.Library.Size() != 8 {
		t.Errorf("library %d, want all 8 still there", me2.Library.Size())
	}
}

const stdGorgonCaseOracle = "a89f9072-cf06-429f-a9b8-c11b5731ddd3"

func TestStdB46CaseOfTheGorgonsKissDestroysADamagedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hurt := reqCreature(g, opp.ID, "Hurt Bear", "", 4, 4)
	healthy := reqCreature(g, opp.ID, "Healthy Bear", "", 4, 4)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, hurt, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	castCatalogSpell(t, g, "Case of the Gorgon's Kiss", "Enchantment — Case", stdGorgonCaseOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatalf("no target prompt: %+v", g.PendingChoices)
	}
	if len(pick.PickTargetCards) != 1 || pick.PickTargetCards[0] != hurt {
		t.Fatalf("offered %v, want only the damaged creature (not %v)", pick.PickTargetCards, healthy)
	}
	pickCard(t, g, me.ID, hurt)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, hurt); ok {
		t.Error("the damaged creature survived")
	}
	if _, ok := battlefieldCard(g, healthy); !ok {
		t.Error("the undamaged creature was destroyed")
	}
}

func TestStdB46CaseOfTheGorgonsKissSolvesAndBecomesAFourFourGorgon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	caseID := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Case of the Gorgon's Kiss", TypeLine: "Enchantment — Case", OracleID: stdGorgonCaseOracle,
		Owner: me.ID, Controller: me.ID,
	})
	for i := 0; i < 2; i++ {
		id := reqCreature(g, me.ID, "Dying Bear", "", 1, 1)
		g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{id}) })
	}
	g.RunStateChecksForTest()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, caseID); c.Solved || c.IsCreature() {
		t.Fatal("solved with only two creature cards put into graveyards")
	}
	// A third one, then the next end step... this turn is over, so solve in a fresh turn.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	advanceToMain(t, g2)
	case2 := pushBattlefieldCardWithTimestamp(g2, game.Card{
		InstanceID: uuid.New(), Name: "Case of the Gorgon's Kiss", TypeLine: "Enchantment — Case", OracleID: stdGorgonCaseOracle,
		Owner: me2.ID, Controller: me2.ID,
	})
	var tokenGone uuid.UUID
	for i := 0; i < 3; i++ {
		id := reqCreature(g2, g2.Seats[1].ID, "Dying Bear", "", 1, 1) // any player's creature cards count
		g2.WithWriteLock(func() { g2.DestroyPermanentsForEffect([]uuid.UUID{id}) })
	}
	tokenGone = pushBattlefieldCardWithTimestamp(g2, game.Card{InstanceID: uuid.New(), Name: "Token", TypeLine: "Token Creature — Bear", Power: 1, Toughness: 1, Owner: me2.ID, Controller: me2.ID})
	g2.WithWriteLock(func() { g2.DestroyPermanentsForEffect([]uuid.UUID{tokenGone}) })
	g2.RunStateChecksForTest()
	advanceTo(t, g2, game.StepEnd)
	passPriorityAroundTable(t, g2)
	c := tmtB1Card(t, g2, case2)
	if !c.Solved {
		t.Fatal("not solved after three creature cards were put into graveyards")
	}
	if p, tough := mkmMPPT(t, g2, case2); !c.IsCreature() || p != 4 || tough != 4 || !c.HasSubtype("Gorgon") || !c.HasSubtype("Case") {
		t.Errorf("solved Case is %d/%d creature %v, want a 4/4 Gorgon Case creature", p, tough, c.IsCreature())
	}
	for _, kw := range []string{"deathtouch", "lifelink"} {
		if !hasKeywordOnBattlefield(t, g2, case2, kw) {
			t.Errorf("the solved Case lacks %s", kw)
		}
	}
}

const stdBatColonyOracle = "c2022f95-2c26-4a2a-b83b-60efa937cdde"

func TestStdB46BatColonyMakesABatForEachManaFromACave(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	t1 := pushCatalogPermanent(g, me.ID, "Secret Tunnel", "Land — Cave", avB20SecretTunnelOracle, false)
	t2 := pushCatalogPermanent(g, me.ID, "Secret Tunnel", "Land — Cave", avB20SecretTunnelOracle, false)
	b28TapForMana(t, g, me.ID, t1, "")
	b28TapForMana(t, g, me.ID, t2, "")
	b06AddMana(me, "W")
	id := handCardFull(me, "Bat Colony", "Enchantment", "{2}{W}", stdBatColonyOracle, []string{"W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Bat"); n != 2 {
		t.Fatalf("%d Bats, want 2 (two mana from a Cave)", n)
	}
}

func TestStdB46BatColonyMakesNoBatWithoutCaveMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	b06AddMana(me, "C", "C", "W")
	id := handCardFull(me, "Bat Colony", "Enchantment", "{2}{W}", stdBatColonyOracle, []string{"W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Bat"); n != 0 {
		t.Fatalf("%d Bats, want 0", n)
	}
}

func stdPlayLand(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: me.ID, Controller: me.ID})
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
}

func TestStdB46BatColonyPutsACounterWhenACaveEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Bat Colony", "Enchantment", stdBatColonyOracle, 0, 0)
	bear := reqCreature(g, me.ID, "Bear", "", 2, 2)
	stdPlayLand(t, g, me, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, bear); c.Counters["+1/+1"] != 0 {
		t.Fatalf("a non-Cave land entering gave %d counters", c.Counters["+1/+1"])
	}
	stdPlayLand(t, g, me, "Secret Tunnel", "Land — Cave", avB20SecretTunnelOracle)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if c := tmtB1Card(t, g, bear); c.Counters["+1/+1"] != 1 {
		t.Errorf("counters %d, want 1", c.Counters["+1/+1"])
	}
}

const stdReenactOracle = "ebb88893-427f-496c-82a8-94d2bafb5e4c"

func TestStdB44ReenactTheCrimeCopiesAndCastsACardPutIntoAGraveyardThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	bolt := handCardFull(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle, []string{"R"})
	old := gyCard(opp, "Old Bolt", "Instant", "{R}", lightningBoltOracle) // in the graveyard since before, never "put" this turn
	g.WithWriteLock(func() {
		for i := range opp.Graveyard.Cards {
			opp.Graveyard.Cards[i].PutInGraveyardThisTurn = false
		}
		if err := g.PutIntoGraveyardForEffect(bolt); err != nil {
			t.Fatalf("put into graveyard: %v", err)
		}
	})
	b06AddMana(me, "U", "U", "U", "C")
	id := handCardFull(me, "Reenact the Crime", "Instant", "{1}{U}{U}{U}", stdReenactOracle, []string{"U"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: old}}}); err == nil {
		t.Fatal("a card that was not put into a graveyard this turn was a legal target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bolt) {
		t.Fatal("the target was not exiled")
	}
	offer := latestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatalf("no free cast offer for the copy: %+v", g.PendingChoices)
	}
	copyID := offer.DiscardOptions[0]
	life := opp.Life
	if err := g.CastForResolutionChoice(me.ID, offer.ID, copyID, game.CastSpellParams{FromZone: "exile",
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("casting the copy for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d -> %d, want 3 damage from the copied Bolt", life, opp.Life)
	}
}

const stdAnzragOracle = "ce690a21-8117-420e-be00-cd6d9923a812"

func TestStdB44AnzragsRampageDestroysTheirArtifactsThenExilesAndReturnsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	myArtifact := b12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := []uuid.UUID{b12Permanent(g, opp.ID, "Their Rock", "Artifact"), b12Permanent(g, opp.ID, "Their Gem", "Artifact")}
	me.Library.Cards = nil
	filler := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: filler, Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	runner := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: runner, Name: "Runner", TypeLine: "Creature — Goblin", ManaCost: "{R}", Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID})
	third := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: third, Name: "Third", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	// One of MY artifacts went to a graveyard earlier this turn: it counts for X.
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{myArtifact}) })
	g.RunStateChecksForTest()

	b06AddMana(me, "R", "R", "C", "C", "C")
	id := handCardFull(me, "Anzrag's Rampage", "Sorcery", "{3}{R}{R}", stdAnzragOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, a := range theirs {
		if g.Battlefield.Contains(a) {
			t.Error("an artifact I don't control survived")
		}
	}
	// X = 3: my artifact and their two. The top three cards of my library are exiled.
	for _, c := range []uuid.UUID{third, runner, filler} {
		if !g.Exile.Contains(c) {
			t.Errorf("a card among the top three was not exiled")
		}
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatalf("no creature pick: %+v", g.PendingChoices)
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != runner {
		t.Fatalf("offered %v, want only the creature card", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{runner}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	var entered game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Runner" {
				entered = c
			}
		}
	})
	if entered.InstanceID == uuid.Nil || !hasKeywordOnBattlefield(t, g, entered.InstanceID, "haste") {
		t.Fatalf("the creature is not on the battlefield with haste: %+v", entered)
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if found := func() bool {
		in := false
		g.ReadSnapshot(func() {
			for _, c := range me.Hand.Cards {
				if c.Name == "Runner" {
					in = true
				}
			}
		})
		return in
	}(); !found {
		t.Error("the creature was not returned to my hand at the next end step")
	}
}

func TestStdB44AnzragsRampageDoesNothingFurtherWithNoArtifactsInGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	lib := me.Library.Size()
	b06AddMana(me, "R", "R", "C", "C", "C")
	id := handCardFull(me, "Anzrag's Rampage", "Sorcery", "{3}{R}{R}", stdAnzragOracle, []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Library.Size() != lib {
		t.Errorf("library %d -> %d: X was 0 so nothing is exiled", lib, me.Library.Size())
	}
}
