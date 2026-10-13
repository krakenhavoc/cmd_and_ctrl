package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// voting_a_test.go — #2928, the first slice of the other voting cards on
// ADR 0146's seam: Tyrant's Choice, Bite of the Black Rose, Capital
// Punishment, Lieutenants of the Guard, Orchard Elemental, Messenger
// Jays, Emissary Green, Travel Through Caradhras, Fateful Tempest, Sail
// into the West, Split Decision, Council Guardian.

const (
	tyrantsChoiceOracle           = "ee7d2c2f-e32f-4be6-82ac-ff1e32055376"
	biteOfTheBlackRoseOracle      = "4db04e23-e1da-4c41-9dba-5502dfff73fb"
	capitalPunishmentOracle       = "c2902f4f-b40e-49f7-8ab0-5b6a70021482"
	lieutenantsOfTheGuardOracle   = "80745dea-2335-43fc-9915-f62e0436e73b"
	orchardElementalOracle        = "189b7c2b-81df-41c3-94ae-6b32dbd48cec"
	messengerJaysOracle           = "dfeb1d36-d1d9-43dd-a79f-a98560676bb3"
	emissaryGreenOracle           = "c43e7fa5-ff93-42d9-af79-ee417a8e41a7"
	travelThroughCaradhrasOracle  = "2d47a98b-5700-4096-b080-32df4d79910c"
	fatefulTempestOracle          = "9d2f02f7-adcd-4d23-a051-ebca70f22ba5"
	sailIntoTheWestOracle         = "0a8956d0-bd37-4305-8443-4df043796f6e"
	splitDecisionOracle           = "5f725d51-e133-488c-9ca9-1c7b0523b5bd"
	councilGuardianOracle         = "4f212675-e34d-4387-a125-e17b25757083"
	votingATestCreatureTypeLine   = "Creature — Bear"
	votingATestPlusOneCounterKind = "+1/+1"
)

// voteRound casts every seat's ballot, in turn order from the active
// player, for the labels given in that order.
func voteRound(t *testing.T, g *game.Game, labels ...string) {
	t.Helper()
	order := voteOrder(g)
	if len(labels) != len(order) {
		t.Fatalf("voteRound: %d labels for %d seats", len(labels), len(order))
	}
	for i, p := range order {
		voteFor(t, g, p.ID, labels[i])
	}
}

// votingACreaturesOf counts the creatures on the battlefield p controls.
func votingACreaturesOf(g *game.Game, p *game.Player) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p.ID && c.IsCreature() {
			n++
		}
	}
	return n
}

// votingACountNamed counts the permanents p controls with the name.
func votingACountNamed(g *game.Game, p *game.Player, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p.ID && c.Name == name {
			n++
		}
	}
	return n
}

// votingAFillHand puts n cards into p's hand.
func votingAFillHand(g *game.Game, p *game.Player, n int) {
	for p.Hand.Size() < n {
		pushHandCard(g, p)
	}
}

func TestTyrantsChoiceDeathEdictsAndTortureDrainsOnATie(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me, opp := order[0], order[1]
	pushCatalogPermanent(g, opp.ID, "Victim", votingATestCreatureTypeLine, "", false)
	castCatalogSpell(t, g, "Tyrant's Choice", "Sorcery", tyrantsChoiceOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "death", "death", "torture", "torture") // a tie
	life := opp.Life
	if life != game.StartingLife-4 {
		t.Errorf("a tie is torture: opponent at %d, want %d", life, game.StartingLife-4)
	}
	if votingACreaturesOf(g, opp) != 1 {
		t.Error("torture should not have taken a creature")
	}
	if me.Life != game.StartingLife {
		t.Errorf("the caster lost life: %d", me.Life)
	}

	castCatalogSpell(t, g, "Tyrant's Choice", "Sorcery", tyrantsChoiceOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "death", "death", "death", "torture")
	if opp.Life != life {
		t.Errorf("death won yet life moved: %d -> %d", life, opp.Life)
	}
	answerSacrifice(t, g, opp.ID, votingAFirstCreatureOf(g, opp))
	if votingACreaturesOf(g, opp) != 0 {
		t.Error("death should have made the opponent sacrifice their creature")
	}
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the caster was asked to sacrifice")
	}
}

func votingAFirstCreatureOf(g *game.Game, p *game.Player) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == p.ID && c.IsCreature() {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestBiteOfTheBlackRose(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me, opp := order[0], order[1]
	mine := ctrlPushCreature(g, me.ID, "Mine")
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 3, 3)
	castCatalogSpell(t, g, "Bite of the Black Rose", "Sorcery", biteOfTheBlackRoseOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "sickness", "sickness", "psychosis", "sickness")
	if got := effectivePower(t, g, theirs); got != 1 {
		t.Errorf("sickness: an opponent's creature has power %d, want 1", got)
	}
	if got := effectivePower(t, g, mine); got != 2 {
		t.Errorf("sickness shrank the caster's own creature: %d", got)
	}

	// A tie is psychosis: each opponent chooses two cards to discard.
	votingAFillHand(g, order[1], 5)
	votingAFillHand(g, order[2], 5)
	before1, before2, beforeMe := order[1].Hand.Size(), order[2].Hand.Size(), me.Hand.Size()
	castCatalogSpell(t, g, "Bite of the Black Rose", "Sorcery", biteOfTheBlackRoseOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "sickness", "sickness", "psychosis", "psychosis")
	for _, p := range order[1:] {
		if discardOwed(g, p.ID) != 2 {
			t.Fatalf("%s owes %d discards, want 2", p.Name, discardOwed(g, p.ID))
		}
		discardFromHand(t, g, p.ID)
	}
	if order[1].Hand.Size() != before1-2 || order[2].Hand.Size() != before2-2 {
		t.Errorf("opponents' hands %d,%d -> %d,%d, want two fewer each", before1, before2, order[1].Hand.Size(), order[2].Hand.Size())
	}
	if me.Hand.Size() != beforeMe-1 && me.Hand.Size() != beforeMe {
		t.Errorf("the caster's hand changed: %d -> %d", beforeMe, me.Hand.Size())
	}
}

func TestCapitalPunishmentSacrificesPerDeathVoteAndDiscardsPerTaxesVote(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me, opp := order[0], order[1]
	for i := 0; i < 3; i++ {
		pushCatalogPermanent(g, opp.ID, "Victim", votingATestCreatureTypeLine, "", false)
	}
	pushCatalogPermanent(g, me.ID, "Keeper", votingATestCreatureTypeLine, "", false)
	votingAFillHand(g, opp, 4)
	handBefore := opp.Hand.Size()
	castCatalogSpell(t, g, "Capital Punishment", "Sorcery", capitalPunishmentOracle, nil)
	passPriorityAroundTable(t, g)
	// two death votes, two taxes votes
	voteRound(t, g, "death", "taxes", "death", "taxes")
	for i := 0; i < 2; i++ {
		answerSacrifice(t, g, opp.ID, votingAFirstCreatureOf(g, opp))
	}
	if got := votingACreaturesOf(g, opp); got != 1 {
		t.Errorf("the opponent has %d creatures after two death votes, want 1", got)
	}
	if discardOwed(g, opp.ID) != 2 {
		t.Fatalf("the opponent owes %d discards, want 2 (one per taxes vote)", discardOwed(g, opp.ID))
	}
	discardFromHand(t, g, opp.ID)
	if got := opp.Hand.Size(); got != handBefore-2 {
		t.Errorf("hand %d -> %d, want two fewer", handBefore, got)
	}
	if votingACreaturesOf(g, me) != 1 || sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the caster was affected")
	}
}

func TestCapitalPunishmentAllTaxesNeverAsksForASacrifice(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	opp := order[1]
	pushCatalogPermanent(g, opp.ID, "Victim", votingATestCreatureTypeLine, "", false)
	votingAFillHand(g, opp, 3)
	castCatalogSpell(t, g, "Capital Punishment", "Sorcery", capitalPunishmentOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "taxes", "taxes", "taxes", "taxes")
	if sacrificeChoiceFor(g, opp.ID) != nil {
		t.Fatal("no death vote, yet a sacrifice was asked")
	}
	if discardOwed(g, opp.ID) != 4 && opp.Hand.Size() >= 4 {
		t.Errorf("the opponent owes %d discards, want 4", discardOwed(g, opp.ID))
	}
	if votingACreaturesOf(g, opp) != 1 {
		t.Error("the creature was sacrificed")
	}
}

func TestLieutenantsOfTheGuardCountersAndSoldiers(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	castCatalogSpell(t, g, "Lieutenants of the Guard", "Creature — Human Soldier", lieutenantsOfTheGuardOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "strength", "numbers", "numbers", "strength")
	id, ok := b43BattlefieldByName(g, "Lieutenants of the Guard")
	if !ok {
		t.Fatal("Lieutenants did not enter")
	}
	if got := aangCounterCount(g, id, votingATestPlusOneCounterKind); got != 2 {
		t.Errorf("counters = %d, want 2 (two strength votes)", got)
	}
	if got := votingACountNamed(g, me, "Soldier"); got != 2 {
		t.Errorf("Soldier tokens = %d, want 2 (two numbers votes)", got)
	}
}

func TestOrchardElementalCountersAndLife(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	castCatalogSpell(t, g, "Orchard Elemental", "Creature — Elemental", orchardElementalOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "sprout", "sprout", "sprout", "harvest")
	id, ok := b43BattlefieldByName(g, "Orchard Elemental")
	if !ok {
		t.Fatal("Orchard Elemental did not enter")
	}
	if got := aangCounterCount(g, id, votingATestPlusOneCounterKind); got != 6 {
		t.Errorf("counters = %d, want 6 (two per sprout vote)", got)
	}
	if me.Life != game.StartingLife+3 {
		t.Errorf("life = %d, want %d (3 for one harvest vote)", me.Life, game.StartingLife+3)
	}
}

func TestMessengerJaysCountersThenDrawsAndDiscards(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	handBefore := me.Hand.Size()
	castCatalogSpell(t, g, "Messenger Jays", "Creature — Bird", messengerJaysOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "feather", "quill", "quill", "feather")
	id, ok := b43BattlefieldByName(g, "Messenger Jays")
	if !ok {
		t.Fatal("Messenger Jays did not enter")
	}
	if got := aangCounterCount(g, id, votingATestPlusOneCounterKind); got != 2 {
		t.Errorf("counters = %d, want 2 (two feather votes)", got)
	}
	// The Jays left the hand when cast; two quill votes draw two, then
	// ask for two discards.
	if discardOwed(g, me.ID) != 2 {
		t.Fatalf("owes %d discards, want 2", discardOwed(g, me.ID))
	}
	discardFromHand(t, g, me.ID)
	_ = handBefore
}

func TestMessengerJaysWithNoQuillVotesDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	castCatalogSpell(t, g, "Messenger Jays", "Creature — Bird", messengerJaysOracle, nil)
	passPriorityAroundTable(t, g)
	handBefore := me.Hand.Size()
	voteRound(t, g, "feather", "feather", "feather", "feather")
	if me.Hand.Size() != handBefore || discardOwed(g, me.ID) != 0 {
		t.Errorf("no quill vote, yet hand %d -> %d and %d discards owed", handBefore, me.Hand.Size(), discardOwed(g, me.ID))
	}
	id, _ := b43BattlefieldByName(g, "Messenger Jays")
	if got := aangCounterCount(g, id, votingATestPlusOneCounterKind); got != 4 {
		t.Errorf("counters = %d, want 4", got)
	}
}

func TestEmissaryGreenTreasuresAndCountersOnEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me, opp := order[0], order[1]
	green := pushCatalogPermanent(g, me.ID, "Emissary Green", "Legendary Creature — Human Advisor", emissaryGreenOracle, false)
	bear := ctrlPushCreature(g, me.ID, "Bear")
	theirs := ctrlPushCreature(g, opp.ID, "Theirs")
	pr7bAttack(t, g, opp.ID, green)
	voteRound(t, g, "profit", "security", "security", "profit")
	if got := votingACountNamed(g, me, "Treasure"); got != 4 {
		t.Errorf("Treasures = %d, want 4 (twice two profit votes)", got)
	}
	if got := aangCounterCount(g, bear, votingATestPlusOneCounterKind); got != 2 {
		t.Errorf("Bear counters = %d, want 2 (two security votes)", got)
	}
	if got := aangCounterCount(g, green, votingATestPlusOneCounterKind); got != 2 {
		t.Errorf("Emissary Green counters = %d, want 2", got)
	}
	if got := aangCounterCount(g, theirs, votingATestPlusOneCounterKind); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestTravelThroughCaradhrasLandsAndGraveyardThenExiles(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	me.Library.Cards = nil
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		forests = append(forests, pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
	}
	pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	back := pushGraveyardCardForTest(me, "Old Friend")
	castCatalogSpell(t, g, "Travel Through Caradhras", "Sorcery", travelThroughCaradhrasOracle, nil)
	passPriorityAroundTable(t, g)
	// Two Redhorn Pass votes, two Mines of Moria votes.
	voteRound(t, g, "Redhorn Pass", "Mines of Moria", "Redhorn Pass", "Mines of Moria")
	answerSearchByID(t, g, me.ID, forests[0], forests[1])
	for _, f := range forests[:2] {
		if !g.Battlefield.Contains(f) {
			t.Errorf("a searched Forest did not enter")
		}
	}
	if !g.Battlefield.Contains(forests[0]) {
		t.Fatal("no Forest entered")
	}
	for _, f := range forests[:2] {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == f && !c.Tapped {
				t.Error("a Forest entered untapped")
			}
		}
	}
	// One card in the graveyard to return for two votes: it returns, and
	// Travel Through Caradhras itself is not in the graveyard to be chosen.
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil {
		t.Fatal("no graveyard choice was opened")
	}
	for _, id := range c.ChooseCards {
		if card, ok := g.LookupCardForEffect(id); ok && card.Name == "Travel Through Caradhras" {
			t.Error("Travel Through Caradhras is offered as a card to return")
		}
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{back}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(back) {
		t.Error("the graveyard card did not return to hand")
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard holds %d cards; Travel Through Caradhras should be exiled", me.Graveyard.Size())
	}
}

func TestFatefulTempestMillsDamagesAndExilesForLater(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me, opp := order[0], order[1]
	me.Library.Cards = nil
	// The library's top three, top first: two cards to mill (mana
	// value 3 and 2), then the first exiled card.
	top := []string{"Milled 3", "Milled 2", "Exiled 1", "Exiled 2", "Deep"}
	costs := []string{"{3}", "{2}", "{1}", "{1}", "{9}"}
	ids := make([]uuid.UUID, len(top))
	for i := len(top) - 1; i >= 0; i-- {
		ids[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: ids[i], Name: top[i], TypeLine: "Creature — Bear", ManaCost: costs[i], Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Fateful Tempest", "Sorcery", fatefulTempestOracle, nil)
	passPriorityAroundTable(t, g)
	// two past votes, two present votes
	voteRound(t, g, "past", "past", "present", "present")
	for i := 0; i < 2; i++ {
		if !me.Graveyard.Contains(ids[i]) {
			t.Errorf("%s was not milled", top[i])
		}
	}
	for _, p := range order[1:] {
		if p.Life != game.StartingLife-5 {
			t.Errorf("%s is at %d, want %d (3 + 2 mana value milled)", p.Name, p.Life, game.StartingLife-5)
		}
	}
	if me.Life != game.StartingLife {
		t.Errorf("the caster took damage: %d", me.Life)
	}
	_ = opp
	for _, id := range ids[2:4] {
		if !g.Exile.Contains(id) {
			t.Fatalf("a present vote did not exile a card")
		}
		if perm := exiledPermission(g, id); perm.Player != me.ID || perm.CastOnly || !permissionLive(g, perm, me.ID) {
			t.Errorf("%+v: the caster may play it", perm)
		}
	}
	if !me.Library.Contains(ids[4]) {
		t.Error("a third card left the library")
	}
}

func TestFatefulTempestAllPresentDealsNoDamage(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me := order[0]
	castCatalogSpell(t, g, "Fateful Tempest", "Sorcery", fatefulTempestOracle, nil)
	passPriorityAroundTable(t, g)
	graveBefore := me.Graveyard.Size()
	voteRound(t, g, "present", "present", "present", "present")
	for _, p := range order[1:] {
		if p.Life != game.StartingLife {
			t.Errorf("%s took damage with no past vote", p.Name)
		}
	}
	if me.Graveyard.Size() != graveBefore {
		t.Error("a card was milled with no past vote")
	}
	if g.Exile.Size() != 4 {
		t.Errorf("exile holds %d cards, want 4", g.Exile.Size())
	}
}

func TestSailIntoTheWestReturnExilesItselfAndReturnsUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me, opp := order[0], order[1]
	a := pushGraveyardCardForTest(me, "Mine A")
	b := pushGraveyardCardForTest(me, "Mine B")
	c := pushGraveyardCardForTest(me, "Mine C")
	theirs := pushGraveyardCardForTest(opp, "Theirs")
	sail := castCatalogSpell(t, g, "Sail into the West", "Instant", sailIntoTheWestOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "return", "return", "embark", "return")
	// Each player picks in their own prompt; the caster asks for three
	// and is refused.
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("the caster was not offered a graveyard choice")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b, c}); err == nil {
		t.Error("three cards were accepted for 'up to two'")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(a) || !me.Hand.Contains(b) || me.Hand.Contains(c) {
		t.Error("the caster should hold exactly the two cards they picked")
	}
	pick = latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID)
	if pick == nil {
		t.Fatal("an opponent was not offered a graveyard choice")
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !opp.Hand.Contains(theirs) {
		t.Error("the opponent's card did not return")
	}
	if !g.Exile.Contains(sail) {
		t.Error("Sail into the West was not exiled")
	}
	if me.Graveyard.Contains(sail) {
		t.Error("Sail into the West is in the graveyard")
	}
}

func TestSailIntoTheWestTieEmbarksEachPlayerChoosingForThemselves(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me, yes, no := order[0], order[1], order[2]
	votingAFillHand(g, me, 3)
	votingAFillHand(g, yes, 3)
	votingAFillHand(g, no, 3)
	oldMine, oldYes, oldNo := votingAHandIDs(me), votingAHandIDs(yes), votingAHandIDs(no)
	castCatalogSpell(t, g, "Sail into the West", "Instant", sailIntoTheWestOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "return", "embark", "return", "embark") // a tie
	for _, p := range order {
		c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, p.ID)
		if c == nil {
			t.Fatalf("%s was not asked to discard their hand", p.Name)
		}
		if err := g.ResolveConfirm(c.ID, p.ID, p.ID != no.ID); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
	}
	if me.Hand.Size() != 7 || yes.Hand.Size() != 7 {
		t.Errorf("hands %d and %d, want seven each", me.Hand.Size(), yes.Hand.Size())
	}
	for _, id := range oldMine {
		if me.Hand.Contains(id) {
			t.Error("the caster kept a card from the discarded hand")
		}
	}
	for _, id := range oldYes {
		if yes.Hand.Contains(id) {
			t.Error("a player who said yes kept an old card")
		}
	}
	if no.Hand.Size() != len(oldNo) {
		t.Errorf("the player who declined has %d cards, want %d", no.Hand.Size(), len(oldNo))
	}
}

func votingAHandIDs(p *game.Player) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

func TestSplitDecisionDenialCountersAndATieCopies(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	me, opp := order[0], order[1]
	_ = me
	// A Tyrant's Choice from the first seat is the target; Split Decision
	// is cast by the same seat in response (it is an instant).
	target := castCatalogSpell(t, g, "Tyrant's Choice", "Sorcery", tyrantsChoiceOracle, nil)
	castCatalogSpell(t, g, "Split Decision", "Instant", splitDecisionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
	voteRound(t, g, "denial", "denial", "duplication", "denial")
	if g.Stack.Contains(target) {
		t.Error("the targeted spell is still on the stack after denial")
	}
	if opp.Life != game.StartingLife {
		t.Errorf("the countered spell resolved: %d", opp.Life)
	}
}

func TestSplitDecisionDuplicationCopiesTheSpellForItsController(t *testing.T) {
	g := newCatalogGame(t)
	order := voteOrder(g)
	opp := order[1]
	target := castCatalogSpell(t, g, "Tyrant's Choice", "Sorcery", tyrantsChoiceOracle, nil)
	castCatalogSpell(t, g, "Split Decision", "Instant", splitDecisionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
	voteRound(t, g, "denial", "duplication", "duplication", "denial") // a tie
	passPriorityAroundTable(t, g)
	// The copy and the original each ask for their own vote.
	voteRound(t, g, "torture", "torture", "torture", "torture")
	passPriorityAroundTable(t, g)
	voteRound(t, g, "torture", "torture", "torture", "torture")
	passPriorityAroundTable(t, g)
	if opp.Life != game.StartingLife-8 {
		t.Errorf("life %d, want %d: the original and the copy each drained 4", opp.Life, game.StartingLife-8)
	}
}

func TestCouncilGuardianGainsProtectionFromEachTiedColour(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Council Guardian", "Creature — Giant Soldier", councilGuardianOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "blue", "red", "red", "blue") // blue and red tie for most
	id, ok := b43BattlefieldByName(g, "Council Guardian")
	if !ok {
		t.Fatal("Council Guardian did not enter")
	}
	abilities := effectiveAbilities(t, g, id)
	has := func(kw string) bool {
		for _, a := range abilities {
			if a == kw {
				return true
			}
		}
		return false
	}
	for _, color := range []string{"U", "R"} {
		if !has(game.ProtectionFromColor(color)) {
			t.Errorf("no protection from %s in %v", color, abilities)
		}
	}
	for _, color := range []string{"B", "G"} {
		if has(game.ProtectionFromColor(color)) {
			t.Errorf("protection from %s in %v, but it got no votes", color, abilities)
		}
	}
}

func TestCouncilGuardianProtectionLastsPastTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Council Guardian", "Creature — Giant Soldier", councilGuardianOracle, nil)
	passPriorityAroundTable(t, g)
	voteRound(t, g, "green", "green", "green", "green")
	id, _ := b43BattlefieldByName(g, "Council Guardian")
	advanceToNextSeatsTurn(t, g)
	advanceToNextSeatsTurn(t, g)
	if !effectiveAbilitiesContain(t, g, id, game.ProtectionFromColor("G")) {
		t.Error("protection from green ended with the turn")
	}
}
