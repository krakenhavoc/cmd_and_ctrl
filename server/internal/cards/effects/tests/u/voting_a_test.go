package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
		effectstest.PushHandCard(g, p)
	}
}

func TestTyrantsChoiceDeathEdictsAndTortureDrainsOnATie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, opp := order[0], order[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Victim", effectstest.VotingATestCreatureTypeLine, "", false)
	effectstest.CastCatalogSpell(t, g, "Tyrant's Choice", "Sorcery", effectstest.TyrantsChoiceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "death", "death", "torture", "torture") // a tie
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

	effectstest.CastCatalogSpell(t, g, "Tyrant's Choice", "Sorcery", effectstest.TyrantsChoiceOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "death", "death", "death", "torture")
	if opp.Life != life {
		t.Errorf("death won yet life moved: %d -> %d", life, opp.Life)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, votingAFirstCreatureOf(g, opp))
	if votingACreaturesOf(g, opp) != 0 {
		t.Error("death should have made the opponent sacrifice their creature")
	}
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
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
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, opp := order[0], order[1]
	mine := effectstest.CtrlPushCreature(g, me.ID, "Mine")
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Bite of the Black Rose", "Sorcery", effectstest.BiteOfTheBlackRoseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "sickness", "sickness", "psychosis", "sickness")
	if got := effectstest.EffectivePower(t, g, theirs); got != 1 {
		t.Errorf("sickness: an opponent's creature has power %d, want 1", got)
	}
	if got := effectstest.EffectivePower(t, g, mine); got != 2 {
		t.Errorf("sickness shrank the caster's own creature: %d", got)
	}

	// A tie is psychosis: each opponent chooses two cards to discard.
	votingAFillHand(g, order[1], 5)
	votingAFillHand(g, order[2], 5)
	before1, before2, beforeMe := order[1].Hand.Size(), order[2].Hand.Size(), me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Bite of the Black Rose", "Sorcery", effectstest.BiteOfTheBlackRoseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "sickness", "sickness", "psychosis", "psychosis")
	for _, p := range order[1:] {
		if effectstest.DiscardOwed(g, p.ID) != 2 {
			t.Fatalf("%s owes %d discards, want 2", p.Name, effectstest.DiscardOwed(g, p.ID))
		}
		effectstest.DiscardFromHand(t, g, p.ID)
	}
	if order[1].Hand.Size() != before1-2 || order[2].Hand.Size() != before2-2 {
		t.Errorf("opponents' hands %d,%d -> %d,%d, want two fewer each", before1, before2, order[1].Hand.Size(), order[2].Hand.Size())
	}
	if me.Hand.Size() != beforeMe-1 && me.Hand.Size() != beforeMe {
		t.Errorf("the caster's hand changed: %d -> %d", beforeMe, me.Hand.Size())
	}
}

func TestCapitalPunishmentSacrificesPerDeathVoteAndDiscardsPerTaxesVote(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, opp := order[0], order[1]
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, opp.ID, "Victim", effectstest.VotingATestCreatureTypeLine, "", false)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Keeper", effectstest.VotingATestCreatureTypeLine, "", false)
	votingAFillHand(g, opp, 4)
	handBefore := opp.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Capital Punishment", "Sorcery", effectstest.CapitalPunishmentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// two death votes, two taxes votes
	effectstest.VoteRound(t, g, "death", "taxes", "death", "taxes")
	// #3664: two death votes are one pick of two creatures.
	pick := effectstest.SacrificeChoiceFor(g, opp.ID)
	if pick == nil || pick.ChooseMax != 2 || len(effectstest.SacrificeOptions(pick)) < 2 {
		t.Fatalf("the opponent should name two of their creatures at once: %+v", pick)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, effectstest.SacrificeOptions(pick)[:2]...)
	if got := votingACreaturesOf(g, opp); got != 1 {
		t.Errorf("the opponent has %d creatures after two death votes, want 1", got)
	}
	if effectstest.DiscardOwed(g, opp.ID) != 2 {
		t.Fatalf("the opponent owes %d discards, want 2 (one per taxes vote)", effectstest.DiscardOwed(g, opp.ID))
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if got := opp.Hand.Size(); got != handBefore-2 {
		t.Errorf("hand %d -> %d, want two fewer", handBefore, got)
	}
	if votingACreaturesOf(g, me) != 1 || effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the caster was affected")
	}
}

func TestCapitalPunishmentAllTaxesNeverAsksForASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	opp := order[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Victim", effectstest.VotingATestCreatureTypeLine, "", false)
	votingAFillHand(g, opp, 3)
	effectstest.CastCatalogSpell(t, g, "Capital Punishment", "Sorcery", effectstest.CapitalPunishmentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "taxes", "taxes", "taxes", "taxes")
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		t.Fatal("no death vote, yet a sacrifice was asked")
	}
	if effectstest.DiscardOwed(g, opp.ID) != 4 && opp.Hand.Size() >= 4 {
		t.Errorf("the opponent owes %d discards, want 4", effectstest.DiscardOwed(g, opp.ID))
	}
	if votingACreaturesOf(g, opp) != 1 {
		t.Error("the creature was sacrificed")
	}
}

func TestLieutenantsOfTheGuardCountersAndSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastCatalogSpell(t, g, "Lieutenants of the Guard", "Creature — Human Soldier", effectstest.LieutenantsOfTheGuardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "strength", "numbers", "numbers", "strength")
	id, ok := effectstest.B43BattlefieldByName(g, "Lieutenants of the Guard")
	if !ok {
		t.Fatal("Lieutenants did not enter")
	}
	if got := effectstest.AangCounterCount(g, id, effectstest.VotingATestPlusOneCounterKind); got != 2 {
		t.Errorf("counters = %d, want 2 (two strength votes)", got)
	}
	if got := votingACountNamed(g, me, "Soldier"); got != 2 {
		t.Errorf("Soldier tokens = %d, want 2 (two numbers votes)", got)
	}
}

func TestOrchardElementalCountersAndLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastCatalogSpell(t, g, "Orchard Elemental", "Creature — Elemental", effectstest.OrchardElementalOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "sprout", "sprout", "sprout", "harvest")
	id, ok := effectstest.B43BattlefieldByName(g, "Orchard Elemental")
	if !ok {
		t.Fatal("Orchard Elemental did not enter")
	}
	if got := effectstest.AangCounterCount(g, id, effectstest.VotingATestPlusOneCounterKind); got != 6 {
		t.Errorf("counters = %d, want 6 (two per sprout vote)", got)
	}
	if me.Life != game.StartingLife+3 {
		t.Errorf("life = %d, want %d (3 for one harvest vote)", me.Life, game.StartingLife+3)
	}
}

func TestMessengerJaysCountersThenDrawsAndDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Messenger Jays", "Creature — Bird", effectstest.MessengerJaysOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "feather", "quill", "quill", "feather")
	id, ok := effectstest.B43BattlefieldByName(g, "Messenger Jays")
	if !ok {
		t.Fatal("Messenger Jays did not enter")
	}
	if got := effectstest.AangCounterCount(g, id, effectstest.VotingATestPlusOneCounterKind); got != 2 {
		t.Errorf("counters = %d, want 2 (two feather votes)", got)
	}
	// The Jays left the hand when cast; two quill votes draw two, then
	// ask for two discards.
	if effectstest.DiscardOwed(g, me.ID) != 2 {
		t.Fatalf("owes %d discards, want 2", effectstest.DiscardOwed(g, me.ID))
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	_ = handBefore
}

// A draw a replacement redirects is not "a card drawn this way": with an
// opponent's Notion Thief out, the quill draws go to the Thief's
// controller and the Jays' controller discards nothing.
func TestMessengerJaysDiscardsOnlyForCardsActuallyDrawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, thief := order[0], order[1]
	effectstest.SeedReplacementPermanent(g, effectstest.B10NotionThiefOracle, "Notion Thief", thief.ID)
	effectstest.CastCatalogSpell(t, g, "Messenger Jays", "Creature — Bird", effectstest.MessengerJaysOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	myHand, thiefHand := me.Hand.Size(), thief.Hand.Size()
	effectstest.VoteRound(t, g, "quill", "quill", "feather", "feather")
	if effectstest.DiscardOwed(g, me.ID) != 0 {
		t.Errorf("owes %d discards although no card was drawn", effectstest.DiscardOwed(g, me.ID))
	}
	if me.Hand.Size() != myHand {
		t.Errorf("hand %d -> %d, want unchanged (draws were stolen)", myHand, me.Hand.Size())
	}
	if thief.Hand.Size() != thiefHand+2 {
		t.Errorf("Notion Thief's controller hand %d -> %d, want +2", thiefHand, thief.Hand.Size())
	}
}

func TestMessengerJaysWithNoQuillVotesDrawsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastCatalogSpell(t, g, "Messenger Jays", "Creature — Bird", effectstest.MessengerJaysOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	handBefore := me.Hand.Size()
	effectstest.VoteRound(t, g, "feather", "feather", "feather", "feather")
	if me.Hand.Size() != handBefore || effectstest.DiscardOwed(g, me.ID) != 0 {
		t.Errorf("no quill vote, yet hand %d -> %d and %d discards owed", handBefore, me.Hand.Size(), effectstest.DiscardOwed(g, me.ID))
	}
	id, _ := effectstest.B43BattlefieldByName(g, "Messenger Jays")
	if got := effectstest.AangCounterCount(g, id, effectstest.VotingATestPlusOneCounterKind); got != 4 {
		t.Errorf("counters = %d, want 4", got)
	}
}

func TestEmissaryGreenTreasuresAndCountersOnEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, opp := order[0], order[1]
	green := effectstest.PushCatalogPermanent(g, me.ID, "Emissary Green", "Legendary Creature — Human Advisor", effectstest.EmissaryGreenOracle, false)
	bear := effectstest.CtrlPushCreature(g, me.ID, "Bear")
	theirs := effectstest.CtrlPushCreature(g, opp.ID, "Theirs")
	effectstest.Pr7bAttack(t, g, opp.ID, green)
	effectstest.VoteRound(t, g, "profit", "security", "security", "profit")
	if got := votingACountNamed(g, me, "Treasure"); got != 4 {
		t.Errorf("Treasures = %d, want 4 (twice two profit votes)", got)
	}
	if got := effectstest.AangCounterCount(g, bear, effectstest.VotingATestPlusOneCounterKind); got != 2 {
		t.Errorf("Bear counters = %d, want 2 (two security votes)", got)
	}
	if got := effectstest.AangCounterCount(g, green, effectstest.VotingATestPlusOneCounterKind); got != 2 {
		t.Errorf("Emissary Green counters = %d, want 2", got)
	}
	if got := effectstest.AangCounterCount(g, theirs, effectstest.VotingATestPlusOneCounterKind); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

func TestTravelThroughCaradhrasLandsAndGraveyardThenExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	me.Library.Cards = nil
	var forests []uuid.UUID
	for i := 0; i < 3; i++ {
		forests = append(forests, effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
	}
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	back := effectstest.PushGraveyardCardForTest(me, "Old Friend")
	effectstest.CastCatalogSpell(t, g, "Travel Through Caradhras", "Sorcery", effectstest.TravelThroughCaradhrasOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Two Redhorn Pass votes, two Mines of Moria votes.
	effectstest.VoteRound(t, g, "Redhorn Pass", "Mines of Moria", "Redhorn Pass", "Mines of Moria")
	effectstest.AnswerSearchByID(t, g, me.ID, forests[0], forests[1])
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
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
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
	effectstest.CastCatalogSpell(t, g, "Fateful Tempest", "Sorcery", effectstest.FatefulTempestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// two past votes, two present votes
	effectstest.VoteRound(t, g, "past", "past", "present", "present")
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
		if perm := effectstest.ExiledPermission(g, id); perm.Player != me.ID || perm.CastOnly || !effectstest.PermissionLive(g, perm, me.ID) {
			t.Errorf("%+v: the caster may play it", perm)
		}
	}
	if !me.Library.Contains(ids[4]) {
		t.Error("a third card left the library")
	}
}

func TestFatefulTempestAllPresentDealsNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me := order[0]
	effectstest.CastCatalogSpell(t, g, "Fateful Tempest", "Sorcery", effectstest.FatefulTempestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	graveBefore := me.Graveyard.Size()
	effectstest.VoteRound(t, g, "present", "present", "present", "present")
	for _, p := range order[1:] {
		if p.Life != game.StartingLife {
			t.Errorf("%s took damage with no past vote", p.Name)
		}
	}
	// The spell itself is the one new card: it waits on the stack for
	// the vote and goes to the graveyard after it (#3613, CR 608.2n).
	if me.Graveyard.Size() != graveBefore+1 {
		t.Error("a card was milled with no past vote")
	}
	if g.Exile.Size() != 4 {
		t.Errorf("exile holds %d cards, want 4", g.Exile.Size())
	}
}

func TestSailIntoTheWestReturnExilesItselfAndReturnsUpToTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, opp := order[0], order[1]
	a := effectstest.PushGraveyardCardForTest(me, "Mine A")
	b := effectstest.PushGraveyardCardForTest(me, "Mine B")
	c := effectstest.PushGraveyardCardForTest(me, "Mine C")
	theirs := effectstest.PushGraveyardCardForTest(opp, "Theirs")
	sail := effectstest.CastCatalogSpell(t, g, "Sail into the West", "Instant", effectstest.SailIntoTheWestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "return", "return", "embark", "return")
	// Each player picks in their own prompt; the caster asks for three
	// and is refused.
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
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
	pick = effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, opp.ID)
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
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, yes, no := order[0], order[1], order[2]
	votingAFillHand(g, me, 3)
	votingAFillHand(g, yes, 3)
	votingAFillHand(g, no, 3)
	oldMine, oldYes, oldNo := votingAHandIDs(me), votingAHandIDs(yes), votingAHandIDs(no)
	effectstest.CastCatalogSpell(t, g, "Sail into the West", "Instant", effectstest.SailIntoTheWestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "return", "embark", "return", "embark") // a tie
	for _, p := range order {
		c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, p.ID)
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
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	me, opp := order[0], order[1]
	_ = me
	// A Tyrant's Choice from the first seat is the target; Split Decision
	// is cast by the same seat in response (it is an instant).
	target := effectstest.CastCatalogSpell(t, g, "Tyrant's Choice", "Sorcery", effectstest.TyrantsChoiceOracle, nil)
	effectstest.CastCatalogSpell(t, g, "Split Decision", "Instant", effectstest.SplitDecisionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "denial", "denial", "duplication", "denial")
	if g.Stack.Contains(target) {
		t.Error("the targeted spell is still on the stack after denial")
	}
	if opp.Life != game.StartingLife {
		t.Errorf("the countered spell resolved: %d", opp.Life)
	}
}

func TestSplitDecisionDuplicationCopiesTheSpellForItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	order := effectstest.VoteOrder(g)
	opp := order[1]
	target := effectstest.CastCatalogSpell(t, g, "Tyrant's Choice", "Sorcery", effectstest.TyrantsChoiceOracle, nil)
	effectstest.CastCatalogSpell(t, g, "Split Decision", "Instant", effectstest.SplitDecisionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "denial", "duplication", "duplication", "denial") // a tie
	effectstest.PassPriorityAroundTable(t, g)
	// The copy and the original each ask for their own vote.
	effectstest.VoteRound(t, g, "torture", "torture", "torture", "torture")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "torture", "torture", "torture", "torture")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != game.StartingLife-8 {
		t.Errorf("life %d, want %d: the original and the copy each drained 4", opp.Life, game.StartingLife-8)
	}
}

func TestCouncilGuardianGainsProtectionFromEachTiedColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Council Guardian", "Creature — Giant Soldier", effectstest.CouncilGuardianOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "blue", "red", "red", "blue") // blue and red tie for most
	id, ok := effectstest.B43BattlefieldByName(g, "Council Guardian")
	if !ok {
		t.Fatal("Council Guardian did not enter")
	}
	abilities := effectstest.EffectiveAbilities(t, g, id)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Council Guardian", "Creature — Giant Soldier", effectstest.CouncilGuardianOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.VoteRound(t, g, "green", "green", "green", "green")
	id, _ := effectstest.B43BattlefieldByName(g, "Council Guardian")
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, id, game.ProtectionFromColor("G")) {
		t.Error("protection from green ended with the turn")
	}
}
