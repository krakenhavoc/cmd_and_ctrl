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
