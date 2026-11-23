package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsch "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ch"
	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	cardsd "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/d"
	cardsdi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/di"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_01_restore_test.go — #4253's closure sweep, batch 01: the card prompts
// that used to hold a Go closure and now name a registered key. Each test
// takes a STRICT restore (restoreRoundTrip with strict, the boot path) while
// the prompt is open, answers it on the restored table and checks what a
// player sees afterwards; a prompt that still held a closure would make the
// table unrestorable and fail the strict restore.

// sweep01Pick is the open choose_cards pick for chooser, asserted to carry
// the registered continuation `want`.
func sweep01Pick(t *testing.T, g *game.Game, chooser uuid.UUID, want game.ChooseCardsThen) *game.PendingChoice {
	t.Helper()
	c := effectstest.LatestChooseCardsFor(g, chooser)
	if c == nil || c.PickThen != want.Key() {
		t.Fatalf("the open pick is %+v, want one keyed %q", c, want.Key())
	}
	return c
}

// --- shared keys ----------------------------------------------------------

// Command Beacon with two commanders: the pick of which one goes to hand is
// the shared commander-into-hand key (Road of Return asks the same), and the
// restored answer still returns the chosen one.
func TestSweep01CommandBeaconPickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	beacon := effectstest.SeedPermanentWithOracle(g, me.ID, "Command Beacon", "Land", effectstest.CommandBeaconOracle)
	first, second := uuid.New(), uuid.New()
	for _, c := range []struct {
		id   uuid.UUID
		name string
	}{{first, "Tymna the Weaver"}, {second, "Thrasios, Triton Hero"}} {
		me.Command.PushTop(game.Card{
			InstanceID: c.id, Name: c.name, TypeLine: "Legendary Creature — Human",
			Owner: me.ID, Controller: me.ID, IsCommander: true,
		})
	}
	if err := g.ActivateCatalogAbility(me.ID, beacon, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	sweep01Pick(t, g, me.ID, effects.CommandBeaconPicked)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	sweep01Pick(t, g, me.ID, effects.CommandBeaconPicked)
	effectstest.AnswerChooseCards(t, g, me.ID, second)
	if offer := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionalReplacement); offer != nil {
		if err := g.ResolveOptionalReplacement(offer.ID, me.ID, false); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
	}
	if !me.Hand.Contains(second) || !me.Command.Contains(first) || me.Command.Contains(second) {
		t.Error("the chosen commander is in hand and the other stays in the command zone")
	}
}

// Chain of Vapor: the bounced permanent's controller picks the land to
// sacrifice (the shared chain-sacrifice key, Chain of Silence asks the same),
// and the copy question that follows the sacrifice names the spell, so the
// copy the restored answer makes is the Chain's.
func TestSweep01ChainOfVaporLandPickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushTribalCreature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.PushTribalCreature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLand(g, opp.ID, "Their Island")

	effectstest.CastCatalogSpell(t, g, "Chain of Vapor", "Instant", effectstest.ChainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	sweep01Pick(t, g, opp.ID, effects.ChainLandPicked)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	sac := sweep01Pick(t, g, opp.ID, effects.ChainLandPicked)
	if err := g.ResolveChooseCards(sac.ID, opp.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was sacrificed")
	}
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	prompt := effectstest.LatestPickTarget(g, opp.ID)
	if prompt == nil {
		t.Fatalf("the copy's controller re-targets it: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(prompt.ID, opp.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	copies := effectstest.CopiesOnStack(g)
	if len(copies) != 1 || copies[0].Controller != opp.ID {
		t.Fatalf("one copy controlled by the sacrificing player, got %+v", copies)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(mine) {
		t.Error("the copy bounced its own new target")
	}
}

// Compulsive Research: the discard's set rule ("two cards, or one land
// card") is registered with the keyed discard, so it holds on the restored
// table: one nonland card is refused, one land is a whole answer.
func TestSweep01CompulsiveResearchDiscardKeepsItsSetRuleAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.HandCardForTest(me, "Forest", "Basic Land — Forest", "")
	bear := effectstest.HandCardForTest(me, "Bear", "Creature — Bear", "")
	effectstest.CastCatalogSpell(t, g, "Compulsive Research", "Sorcery", effectstest.B42CompulsiveResearchOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Sweep01Discard(t, g, me.ID, cardsco.CompulsiveResearchDiscard.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.Sweep01Discard(t, g, me.ID, cardsco.CompulsiveResearchDiscard.Key())
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Fatal("one nonland card was accepted on the restored table")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("one land card is a complete discard: %v", err)
	}
	if me.Hand.Contains(land) || !me.Hand.Contains(bear) {
		t.Error("only the land was discarded")
	}
}

// Thirst for Knowledge shares the same registered rule shape (the helper
// behind Thirst for Discovery and Winternight Stories too): one artifact
// card is a whole answer after a restore, one creature card is not.
func TestSweep01ThirstForKnowledgeDiscardKeepsItsSetRuleAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.HandCardFull(me, "Sol Ring", "Artifact", "{1}", "", nil)
	bear := effectstest.HandCardFull(me, "Bear", "Creature — Bear", "{1}{G}", "", nil)
	effectstest.CastCatalogSpell(t, g, "Thirst for Knowledge", "Instant", effectstest.B18ThirstForKnowledgeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	const key = "choose-cards/discard-two-unless-artifact"
	effectstest.Sweep01Discard(t, g, me.ID, key)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.Sweep01Discard(t, g, me.ID, key)
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{bear}); err == nil {
		t.Fatal("one non-artifact card was accepted on the restored table")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("one artifact card is a complete discard: %v", err)
	}
	if me.Hand.Contains(rock) || !me.Hand.Contains(bear) {
		t.Error("only the artifact card was discarded")
	}
}

// Coin of Fate: naming the opponent (the keyed choose-a-player form) and the
// opponent's pick are both restore points; the exiled cards ride the
// prompts as data, so the restored answers still split them.
func TestSweep01CoinOfFateNamingAndPickAreRestorePoints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	libBefore := len(g.Seats[0].Library.Cards)
	_, one, _ := effectstest.CoinOfFateActivation(t, g)
	if c := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick); c == nil || c.PickThen != cardsco.CoinOfFateOpponentChosen.Key() {
		t.Fatalf("the naming prompt is %+v, want one keyed %q", c, cardsco.CoinOfFateOpponentChosen.Key())
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me := g.Seats[0]
	last := g.Seats[3]
	effectstest.AnswerChosenPlayer(t, g, me.ID, last.ID)
	sweep01Pick(t, g, last.ID, cardsco.CoinOfFatePicked)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	last = g.Seats[3]
	sweep01Pick(t, g, last.ID, cardsco.CoinOfFatePicked)
	effectstest.AnswerChooseCards(t, g, last.ID, one)
	effectstest.FinishCoinOfFate(t, g, me, libBefore)
}

// --- the rest of the batch ------------------------------------------------

// Chain of Smog: the discard is a keyed discard carrying the spell as data,
// so after a restore the copy offer that follows still names the Chain and
// the copy is made under the discarding player's control.
func TestSweep01ChainOfSmogDiscardIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Chain of Smog", "Sorcery", effectstest.ChainOfSmogOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Sweep01Discard(t, g, opp.ID, cardsch.ChainOfSmogDiscarded.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	opp, third := g.Seats[1], g.Seats[2]
	effectstest.Sweep01Discard(t, g, opp.ID, cardsch.ChainOfSmogDiscarded.Key())
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[3].InstanceID, opp.Hand.Cards[5].InstanceID)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	prompt := effectstest.LatestPickTarget(g, opp.ID)
	if prompt == nil {
		t.Fatalf("the copy's controller re-targets it: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(prompt.ID, opp.ID,
		game.TargetRef{Kind: game.TargetPlayer, ID: third.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	copies := effectstest.CopiesOnStack(g)
	if len(copies) != 1 || copies[0].Controller != opp.ID {
		t.Fatalf("one copy controlled by the discarding player, got %+v", copies)
	}
}

// Eerie Ultimatum: the pick's set rule ("different names") is registered
// with the key and the spell rides the prompt, so both survive a restore.
func TestSweep01EerieUltimatumPickKeepsItsSetRuleAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bearA := effectstest.SeedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	bearB := effectstest.SeedGraveyardCard(t, g, "Bear", "Creature — Bear", "")
	land := effectstest.SeedGraveyardCard(t, g, "Forest", "Basic Land — Forest", "")
	effectstest.CastCatalogSpell(t, g, "Eerie Ultimatum", "Sorcery", effectstest.B14EerieUltimatumOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	sweep01Pick(t, g, me.ID, effects.EerieUltimatumReturn)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	pick := sweep01Pick(t, g, me.ID, effects.EerieUltimatumReturn)
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bearA, bearB, land}); err == nil {
		t.Fatal("two Bears were accepted on the restored table")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bearA, land)
	for _, id := range []uuid.UUID{bearA, land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not return", id)
		}
	}
	if g.Battlefield.Contains(bearB) {
		t.Error("the second Bear returned")
	}
}

// Demonlord of Ashmouth: the choice and then the sacrifice pick are keyed,
// so a restore at either keeps the Demonlord or exiles it as answered.
func TestSweep01DemonlordChoiceAndSacrificeAreRestorePoints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.P2Push(g, me.ID, "Fodder", "Creature — Goblin", "", 1, 1)
	kept := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Demonlord of Ashmouth", TypeLine: "Creature — Demon", OracleID: effectstest.P2Demonlord, Power: 5, Toughness: 4})
	for i := 0; i < 8 && effectstest.LatestOptionPickFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if c := effectstest.LatestOptionPickFor(g, me.ID); c == nil || c.PickThen != cardsd.DemonlordChose.Key() {
		t.Fatalf("the choice is %+v, want one keyed %q", c, cardsd.DemonlordChose.Key())
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	sweep01Pick(t, g, me.ID, cardsd.DemonlordSacrificed)

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerChooseCards(t, g, me.ID, fodder)
	effectstest.SettleProwess(t, g)
	if !effects.OnBattlefield(g, kept) || effects.OnBattlefield(g, fodder) {
		t.Error("sacrificing another creature should keep the Demonlord")
	}
}

// Dwarven Armorer: the counter choice carries its target as data.
func TestSweep01DwarvenArmorerChoiceIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	armorer := effectstest.B12Push(g, me.ID, "Dwarven Armorer", "Creature — Dwarf", effectstest.DwarvenArmorerOracle, 0, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	discard := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "R")
	effectstest.B16Activate(t, g, me.ID, armorer, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear), DiscardIDs: []uuid.UUID{discard},
	})
	c := effectstest.SettleUntilOptionPick(t, g, me.ID, bear)
	if c == nil || c.PickThen != cardsdi.DwarvenArmorerChose.Key() {
		t.Fatalf("the counter choice is %+v, want one keyed %q", c, cardsdi.DwarvenArmorerChose.Key())
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, "+1/+0"); n != 1 {
		t.Fatalf("%d +1/+0 counters on the restored table, want 1", n)
	}
}

// Connive X: the discard carries the creature, so the restored answer puts
// one counter per nonland card discarded.
func TestSweep01ConniveXDiscardIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	nonland := effectstest.HandCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	var err error
	g.WithWriteLock(func() {
		err = effects.ConniveX{Controller: me.ID, Creature: bear, X: 1}.Apply(effects.NewContext(g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: me.ID, SourceCardID: bear,
		}))
	})
	if err != nil {
		t.Fatalf("ConniveX: %v", err)
	}
	// The hand is Spell A and the card drawn; the nonland one is discarded.
	effectstest.Sweep01Discard(t, g, me.ID, effects.ConniveXDiscarded.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerDiscard(t, g, me.ID, nonland)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters on the restored table, want 1 (one nonland discard)", got)
	}
}

// Connive each: the creatures still to connive ride the discard as its
// carry, so a restore between two of them loses neither the second creature
// nor its place in the chain.
func TestSweep01ConniveEachChainSurvivesARestoreBetweenCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	first := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	second := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	nonland := effectstest.HandCardFull(me, "Spell A", "Sorcery", "{1}", "", nil)
	land := effectstest.HandCardFull(me, "Forest", "Basic Land — Forest", "", "", nil)
	var err error
	g.WithWriteLock(func() {
		err = effects.ConniveEach{Controller: me.ID, Creatures: []uuid.UUID{first, second}}.Apply(effects.NewContext(g, &game.StackItem{
			Kind: game.StackItemTriggered, Controller: me.ID, SourceCardID: first,
		}))
	})
	if err != nil {
		t.Fatalf("ConniveEach: %v", err)
	}
	effectstest.Sweep01Discard(t, g, me.ID, effects.ConniveDiscarded.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerDiscard(t, g, me.ID, nonland)
	if got := effectstest.CounterCount(g, first, game.CounterPlusOne); got != 1 {
		t.Fatalf("the first creature has %d counters, want 1", got)
	}
	effectstest.Sweep01Discard(t, g, me.ID, effects.ConniveDiscarded.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerDiscard(t, g, me.ID, land)
	if got := effectstest.CounterCount(g, second, game.CounterPlusOne); got != 0 {
		t.Errorf("the second creature has %d counters, want 0 (a land was discarded)", got)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("a third discard is open")
	}
}

// Court of Ambition at six: the monarch half asks each opponent in turn. The
// discarder's choice and its discard are both restore points (one key pair
// per price), and the payer's question after them is asked on the restored
// table with the price intact.
func TestSweep01CourtOfAmbitionQuestionsAreRestorePoints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, discarder, payer, broke := g.Seats[1], g.Seats[2], g.Seats[3], g.Seats[0]
	effectstest.PushPermanentForTest(g, owner.ID, "Court of Ambition", effectstest.CourtOfAmbitionOracle, "Enchantment")
	effectstest.MonCrown(t, g, owner.ID)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	for broke.Hand.Size() > 1 {
		broke.Hand.Cards = broke.Hand.Cards[:len(broke.Hand.Cards)-1]
	}
	lives := map[uuid.UUID]int{discarder.ID: discarder.Life, payer.ID: payer.Life, broke.ID: broke.Life}
	discarderHand := discarder.Hand.Size()
	effectstest.MonSettle(t, g)
	six := cardsco.CourtOfAmbitionPrices[6]
	if c := effectstest.LatestOptionPickFor(g, discarder.ID); c == nil || c.PickThen != six.Answered.Key() {
		t.Fatalf("the first question is %+v, want one keyed %q", c, six.Answered.Key())
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	discarder = g.Seats[2]
	effectstest.AnswerOptionPick(t, g, discarder.ID, 1)
	effectstest.Sweep01Discard(t, g, discarder.ID, six.Discarded.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	discarder, payer = g.Seats[2], g.Seats[3]
	effectstest.AnswerDiscard(t, g, discarder.ID, discarder.Hand.Cards[0].InstanceID, discarder.Hand.Cards[1].InstanceID)
	if c := effectstest.LatestOptionPickFor(g, payer.ID); c == nil || c.PickThen != six.Answered.Key() {
		t.Fatalf("the second question is %+v, want one keyed %q", c, six.Answered.Key())
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	owner, discarder, payer, broke = g.Seats[1], g.Seats[2], g.Seats[3], g.Seats[0]
	effectstest.AnswerOptionPick(t, g, payer.ID, 0)
	effectstest.MonSettle(t, g)

	if got := discarderHand - discarder.Hand.Size(); got != 2 {
		t.Errorf("the discarding opponent pitched %d, want 2", got)
	}
	if got := lives[payer.ID] - payer.Life; got != 6 {
		t.Errorf("the paying opponent lost %d, want 6", got)
	}
	if got := lives[broke.ID] - broke.Life; got != 6 {
		t.Errorf("the one-card opponent lost %d, want 6 without being asked", got)
	}
	if discarder.Life != lives[discarder.ID] || owner.Life != 40 {
		t.Errorf("life moved for the discarder (%d) or the owner (%d)", discarder.Life, owner.Life)
	}
}

// Emergent Ultimatum: naming the opponent (keyed choose-a-player) and the
// opponent's pick (a keyed reveal pick) are restore points; the found cards
// ride them as data, so the restored answers shuffle the chosen card in and
// leave the rest to be cast.
func TestSweep01EmergentUltimatumNamingAndPickAreRestorePoints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SunbirdLibrary(me,
		game.Card{Name: "White Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{W}"},
		game.Card{Name: "Blue Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{U}"},
		game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"},
	)
	white, blue, green := ids[0], ids[1], ids[2]
	spell := effectstest.CastCatalogSpell(t, g, "Emergent Ultimatum", "Sorcery", effectstest.EmergentUltimatumOracle, nil)
	effectstest.PassPriorityUntilChoice(t, g)
	s := effectstest.SearchChoiceFor(g, me.ID)
	if s == nil {
		t.Fatal("no search prompt")
	}
	if err := g.ResolveSearchLibrary(s.ID, me.ID, []uuid.UUID{white, blue, green}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick); c == nil || c.PickThen != cardse.EmergentUltimatumOpponent.Key() {
		t.Fatalf("the naming prompt is %+v, want one keyed %q", c, cardse.EmergentUltimatumOpponent.Key())
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp := g.Seats[0], g.Seats[1]
	oc := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
	if oc == nil {
		t.Fatal("the restored table lost the naming prompt")
	}
	if err := g.ResolveOptionPick(oc.ID, me.ID, 0); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if rp := effectstest.LatestChoiceOfKind(g, game.PendingChoiceRevealPick); rp == nil || rp.Chooser != opp.ID || rp.PickThen != cardse.EmergentUltimatumPicked.Key() {
		t.Fatalf("the opponent's pick is %+v, want one keyed %q", rp, cardse.EmergentUltimatumPicked.Key())
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	rp := effectstest.LatestChoiceOfKind(g, game.PendingChoiceRevealPick)
	if rp == nil {
		t.Fatal("the restored table lost the reveal pick")
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{white}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	if effectstest.InExile(g, white) || !effectstest.LibraryHolds(me, white) {
		t.Error("the chosen card is shuffled into the library")
	}
	if !effectstest.InExile(g, blue) || !effectstest.InExile(g, green) {
		t.Error("the other two stay in exile to be cast")
	}
	if !g.Exile.Contains(spell) {
		t.Error("Emergent Ultimatum exiles itself")
	}
}
