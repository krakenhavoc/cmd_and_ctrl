package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_prompts_restore2_test.go — #4264, the prompts #4256 left on the
// closure form: the "draw a card?" / "copy that spell?" confirm after a
// declined pay-unless, the one-seat discard-then callers, the exile half
// of Karn and Coax (a string carry), and Extrapolate the Impossible's
// pick (a set rule). Each test takes a STRICT restore (restoreRoundTrip
// with strict) while the prompt is open and answers it on the restored
// table; the *UnknownKey* tests are the rollback case.

// --- 1. The confirm after a declined pay-unless -----------------------

// Rhystic Study: the caster declines the tax, the Study's controller is
// asked "draw a card?", and the table waiting on that is a restore point.
func TestRhysticStudyDrawConfirmIsARestorePoint(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		caster, owner := g.Seats[0], g.Seats[1]
		pushPermanentForTest(g, owner.ID, "Rhystic Study", rhysticStudyOracle, "Enchantment")
		castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
		passPriorityAroundTable(t, g)
		answerPayUnless(t, g, caster.ID, false)
		if ask := latestChoiceOfKind(g, game.PendingChoiceConfirm); ask == nil || ask.PickThen != rhysticStudyDraw.Key() {
			t.Fatalf("the 'draw a card?' prompt is %+v, want the keyed confirm", ask)
		}
		g = restoreRoundTrip(t, g, true)
		owner = g.Seats[1]
		ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
		if ask == nil || ask.Chooser != owner.ID || ask.PickThen != rhysticStudyDraw.Key() {
			t.Fatalf("restored: %+v, want the keyed confirm for the Study's controller", ask)
		}
		hand := owner.Hand.Size()
		if err := g.ResolveConfirm(ask.ID, owner.ID, yes); err != nil {
			t.Fatal(err)
		}
		want := 0
		if yes {
			want = 1
		}
		if got := owner.Hand.Size() - hand; got != want {
			t.Errorf("yes=%v: drew %d on the restored table, want %d", yes, got, want)
		}
	}
}

// Mystic Remora's "draw a card?", restored strictly while it is open.
func TestMysticRemoraDrawConfirmIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Mystic Remora", oracleMysticRemora, "Enchantment")
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, caster.ID, false)
	g = restoreRoundTrip(t, g, true)
	owner = g.Seats[1]
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != owner.ID || ask.PickThen != mysticRemoraDraw.Key() {
		t.Fatalf("restored: %+v, want the keyed confirm for the Remora's controller", ask)
	}
	hand := owner.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, owner.ID, true); err != nil {
		t.Fatal(err)
	}
	if got := owner.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d on the restored table, want 1", got)
	}
}

// Wandering Archaic's "copy that spell?", restored strictly while it is
// open: the prompt keeps its key and the carried controller and spell,
// and the answer runs on the restored table without an error.
//
// The copy is made after the restore, and it is the Bolt's: the cast
// spell had already resolved when the question was answered, so the
// copy reads it through the resolving slot (Game.resolving), which the
// snapshot carries since #4270. Before that the answer copied nothing
// and the controller took 3 damage, not 6.
func TestWanderingArchaicCopyConfirmIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Wandering Archaic", "Creature — Avatar", wanderingArchaicOracle, false)
	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	lifeBefore := me.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, opp.ID, false)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.PickThen != wanderingArchaicCopy.Key() || len(ask.OptionCarry) != 2 {
		t.Fatalf("the 'copy that spell?' prompt is %+v, want the keyed confirm carrying the controller and the spell", ask)
	}
	carried := slices.Clone(ask.OptionCarry)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	ask = latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID || ask.PickThen != wanderingArchaicCopy.Key() || !slices.Equal(ask.OptionCarry, carried) {
		t.Fatalf("restored: %+v, want the keyed confirm with its carry %v", ask, carried)
	}
	var before int
	g.ReadSnapshot(func() { before = len(g.Events) })
	answerMayChoice(t, g, me.ID, true)
	settleThroughCopyRetarget(t, g, me.ID, me.ID)
	if n := countCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("the keyed answer on the restored table raised %d effect errors", n)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("the answered confirm left %d prompts open", len(g.PendingChoices))
	}
	if got := lifeBefore - me.Life; got != 6 {
		t.Errorf("restored at the confirm, then copied: took %d damage, want 6 (bolt + copy)", got)
	}
}

func TestConfirmWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Rhystic Study", rhysticStudyOracle, "Enchantment")
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, caster.ID, false)
	refuseUnknownKey(t, g, rhysticStudyDraw.Key())
}

// --- 2. The one-seat discard-then callers -----------------------------

// Last Rites: the target rides the discard prompt as its Carry, so the
// hand that is revealed after the restore is the target's.
func TestLastRitesDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { caster.Hand.Cards = nil })
	a := handCardForTest(caster, "Card A", "Sorcery", "")
	b := handCardForTest(caster, "Card B", "Sorcery", "")
	handCardForTest(caster, "Card C", "Sorcery", "")
	revealHand(victim, rhForest(), rhBolt(), pbBear(), pbDivination())

	castAt(t, g, "Last Rites", "Sorcery", lastRitesOracle, pbPlayer(victim.ID))
	if ask := chooseCardsChoiceFor(g, caster.ID); ask == nil || ask.PickThen != lastRitesReveal.Key() {
		t.Fatalf("the discard is %+v, want the keyed prompt", ask)
	}
	g = restoreRoundTrip(t, g, true)
	caster = g.Seats[0]
	ask := chooseCardsChoiceFor(g, caster.ID)
	if ask == nil || ask.PickThen != lastRitesReveal.Key() || len(ask.OptionCarry) != 1 || ask.OptionCarry[0] != victim.ID {
		t.Fatalf("restored: %+v, want the keyed discard carrying the target", ask)
	}
	answerChooseCards(t, g, caster.ID, a, b)
	pick := openPick(t, g)
	if pick.Count != 2 || len(pick.DiscardOptions) != 3 || pick.FromPlayer != victim.ID {
		t.Fatalf("pick = count %d of %d options from %v, want 2 of the target's 3 nonland cards",
			pick.Count, len(pick.DiscardOptions), pick.FromPlayer)
	}
}

// Eriette's Whisper: the caster and the creature (read as the spell
// resolved) ride the prompt, so the Role goes on the same creature, under
// the caster's control, after the restore.
func TestEriettesWhisperDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.NewCard("Filler", opp.ID))
	}
	castCatalogSpell(t, g, "Eriette's Whisper", "Sorcery", erietteOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	c := discardChoiceFor(g, opp.ID)
	if c == nil || c.PickThen != eriettesWhisperRole.Key() || c.ChooseMax != 2 {
		t.Fatalf("restored: %+v, want the keyed discard of two", c)
	}
	wantNoRole(t, g, mine)
	discardFromHand(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	wantRole(t, g, mine, "Wicked Role")
}

// Pia, Aether Ascetic: the search follows the discard on the restored
// table, and only for a discard that was made.
func TestPiaDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ench := pushLibraryCardForTest(me, game.Card{
		InstanceID: uuid.New(), Name: "Test Enchantment", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	rfdCast(t, g, "Pia, Aether Ascetic", "Legendary Creature — Human Druid", rfdPiaAetherOracle, 2, 2)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := discardChoiceFor(g, me.ID)
	if c == nil || c.PickThen != piaSearch.Key() {
		t.Fatalf("restored: %+v, want the keyed discard", c)
	}
	answerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	passPriorityAroundTable(t, g)
	takeSearchPrompts(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment was not searched into the hand on the restored table")
	}
}

// Mishra's Command: the draw counts what was really discarded, by the
// player who discarded, after the restore.
func TestMishrasCommandDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	b19CastXModal(t, g, "Mishra's Command", "Sorcery", mishrasCommandOracle, "{X}{R}", 3,
		[]int{0, 3},
		[]game.TargetRef{
			modeRef(game.TargetPlayer, opp.ID, 0, 0),
			modeRef(game.TargetCard, mine, 1, 0),
		})
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	c := discardChoiceFor(g, opp.ID)
	if c == nil || c.PickThen != mishrasCommandDraw.Key() {
		t.Fatalf("restored: %+v, want the keyed discard", c)
	}
	hand := opp.Hand.Size()
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	g.SettleResolution()
	if got := opp.Hand.Size(); got != hand {
		t.Errorf("discarded 2, drew 2 back: hand %d -> %d", hand, got)
	}
}

// Thrilling Discovery's "discard two cards" is two or none, never one:
// the set rule is registered with the key, so it holds after a restore
// and the enumerator's window agrees.
func TestThrillingDiscoveryDiscardKeepsItsSetRuleAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Thrilling Discovery", "Sorcery", b39ThrillingDiscoveryOracle, nil)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	c := discardChoiceFor(g, me.ID)
	if c == nil || c.PickThen != drawThreeIfDiscardedTwo.then.Key() {
		t.Fatalf("restored: %+v, want the keyed discard", c)
	}
	hand := me.Hand.Size()
	one := b39HandIDs(me, 1)
	if err := g.ResolveChooseCards(c.ID, me.ID, one); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("a single card: err = %v, want ErrChoiceSetRejected", err)
	}
	var legalOne, legalTwo bool
	g.WithWriteLock(func() {
		legalOne = g.ChooseCardsPickLegalLocked(c, one)
		legalTwo = g.ChooseCardsPickLegalLocked(c, b39HandIDs(me, 2))
	})
	if legalOne || !legalTwo {
		t.Errorf("the enumerator's window offers one card %v and two cards %v, want false and true", legalOne, legalTwo)
	}
	answerDiscard(t, g, me.ID, b39HandIDs(me, 2)...)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("two for three: hand %d -> %d, want %d", hand, me.Hand.Size(), hand+1)
	}
}

// Tersa Lightshatter's "discard up to two, then draw that many" shares
// one registered draw rule with the rest of its family.
func TestTersaDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Tersa Lightshatter", "Legendary Creature — Orc Wizard", tersaOracle, nil)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	ch := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ch == nil || ch.PickThen != drawAsManyAsDiscarded.then.Key() {
		t.Fatalf("restored: %+v, want the keyed discard", ch)
	}
	hand := me.Hand.Size()
	two := append([]uuid.UUID(nil), ch.ChooseCards[:2]...)
	if err := g.ResolveChooseCards(ch.ID, me.ID, two); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want unchanged: discard two, draw two", hand, got)
	}
}

func TestDiscardThenCallersWithAnUnknownKeyAreRefused(t *testing.T) {
	t.Run("last rites", func(t *testing.T) {
		g := newCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		handCardForTest(caster, "Card A", "Sorcery", "")
		revealHand(victim, rhForest())
		castAt(t, g, "Last Rites", "Sorcery", lastRitesOracle, pbPlayer(victim.ID))
		refuseUnknownKey(t, g, lastRitesReveal.Key())
	})
	t.Run("draw rule", func(t *testing.T) {
		g := newCatalogGame(t)
		castCatalogSpell(t, g, "Thrilling Discovery", "Sorcery", b39ThrillingDiscoveryOracle, nil)
		passPriorityAroundTable(t, g)
		refuseUnknownKey(t, g, drawThreeIfDiscardedTwo.then.Key())
	})
}

// --- 3. Karn's and Coax's exile half (a string carry) ----------------

// Karn's −2, exile half, restored strictly while the pick is open: the
// chosen card goes to the hand, and a card that left exile meanwhile does
// not.
func TestKarnExilePickIsARestorePoint(t *testing.T) {
	g, karn := karnGame(t)
	me := g.Seats[0]
	relic := coaxExile(g, me.ID, "Exiled Relic", "Artifact", false)
	coaxExile(g, me.ID, "Exiled Bear", "Creature — Bear", false)
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := latestChooseCardsFor(g, me.ID)
	if c == nil || c.PickThen != karnTheGreatCreatorExile.then.Key() || len(c.ChooseCards) != 1 || c.ChooseCards[0] != relic {
		t.Fatalf("restored: %+v, want the keyed exile pick of the artifact", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{relic}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !handHasID(me, relic) {
		t.Error("the artifact did not come into the hand on the restored table")
	}
}

// Karn's "none" on the restored exile pick asks the pick from outside
// the game, over artifacts only.
func TestKarnExilePickNoneLooksOutsideAfterARestore(t *testing.T) {
	installTestCollection(t)
	g, karn := karnGame(t)
	me := g.Seats[0]
	coaxExile(g, me.ID, "Exiled Relic", "Artifact", false)
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := latestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no exile pick after the restore: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	// This table's deck seated no sideboard, so "outside" is the database.
	pick := latestCollectionPickFor(g, me.ID)
	if pick == nil || !equalNames(pick.CollectionSuggestions, []string{"Sol Ring"}) {
		t.Fatalf("\"none\" on the restored exile pick asked %+v, want the database's artifact", pick)
	}
}

// Coax's exile pick, restored strictly: "none" asks the outside pick
// with the spell's word for "Eldrazi", which rode the prompt.
func TestCoaxExilePickKeepsItsWordAcrossARestore(t *testing.T) {
	installTestCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	coaxExile(g, me.ID, "Eldrazi Devastator", "Creature — Eldrazi", false)
	castCatalogSpell(t, g, "Coax from the Blind Eternities", "Sorcery", coaxOracle, nil)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := latestChooseCardsFor(g, me.ID)
	if c == nil || c.PickThen != coaxFromTheBlindEternitiesExile.then.Key() || c.PickMatchArg != "Eldrazi" {
		t.Fatalf("restored: %+v, want the keyed exile pick with its word", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	pick := latestCollectionPickFor(g, me.ID)
	if pick == nil || pick.PickMatchArg != "Eldrazi" || !equalNames(pick.CollectionSuggestions, []string{"Ulamog's Crusher"}) {
		t.Fatalf("the outside pick after \"none\" is %+v, want the Eldrazi from the database with its word", pick)
	}
}

func TestExilePickWithAnUnknownKeyIsRefused(t *testing.T) {
	g, karn := karnGame(t)
	me := g.Seats[0]
	coaxExile(g, me.ID, "Exiled Relic", "Artifact", false)
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	refuseUnknownKey(t, g, karnTheGreatCreatorExile.then.Key())
}

// --- 4. Extrapolate the Impossible (a set rule) -----------------------

// The first prompt, from a sideboard, restored strictly: the set rule
// "exactly two with different names" still refuses one card, a pair of
// copies and three cards, and the enumerator's window offers the pairs
// the engine accepts. The reveal then follows on the restored table.
func TestExtrapolateTheImpossiblePickIsARestorePoint(t *testing.T) {
	g := newSideboardCatalogGame(t, 2,
		outsideCard("Alpha", "Sorcery"), outsideCard("Alpha", "Sorcery"), outsideCard("Beta", "Instant"))
	castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me, opp := g.Seats[0], g.Seats[1]
	c := latestChooseCardsFor(g, me.ID)
	if c == nil || c.PickThen != extrapolatePick.Key() || c.ChooseMin != 0 || c.ChooseMax != 2 {
		t.Fatalf("restored: %+v, want the keyed pick of up to two", c)
	}
	cards := me.SideboardCards()
	one, copies, pair := []uuid.UUID{cards[0].InstanceID},
		[]uuid.UUID{cards[0].InstanceID, cards[1].InstanceID},
		[]uuid.UUID{cards[0].InstanceID, cards[2].InstanceID}
	for name, picks := range map[string][]uuid.UUID{"one card": one, "two copies": copies} {
		if err := g.ResolveChooseCards(c.ID, me.ID, picks); !errors.Is(err, game.ErrChoiceSetRejected) {
			t.Errorf("%s: err = %v, want ErrChoiceSetRejected", name, err)
		}
	}
	var legalCopies, legalPair bool
	g.WithWriteLock(func() {
		legalCopies = g.ChooseCardsPickLegalLocked(c, copies)
		legalPair = g.ChooseCardsPickLegalLocked(c, pair)
	})
	if legalCopies || !legalPair {
		t.Errorf("the enumerator's window offers the copies %v and the pair %v, want false and true", legalCopies, legalPair)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, pair); err != nil {
		t.Fatalf("a pair with different names: %v", err)
	}
	if pendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		answerChosenPlayer(t, g, me.ID, opp.ID)
	}
	rp := latestRevealPickFor(g, opp.ID)
	if rp == nil || len(rp.ChooseCards) != 2 {
		t.Fatalf("the opponent was not shown the pair on the restored table: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{cards[2].InstanceID}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Beta") {
		t.Error("the card the opponent chose did not come into the hand")
	}
}

// The database branch (a deck with no sideboard): the set rule is found
// by the prompt's key after the restore, so two copies of one name are
// still refused.
func TestExtrapolateTheImpossibleDatabasePickIsARestorePoint(t *testing.T) {
	installTestCollection(t)
	g := newSideboardCatalogGame(t, 2)
	castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	c := latestCollectionPickFor(g, me.ID)
	if c == nil || c.PickThen != extrapolatePick.Key() {
		t.Fatalf("restored: %+v, want the keyed collection pick", c)
	}
	for name, names := range map[string][]string{
		"one card":   {"Lightning Bolt"},
		"two copies": {"Lightning Bolt", "Lightning Bolt"},
	} {
		if err := g.ResolveOutsideCollection(c.ID, me.ID, names); err == nil {
			t.Errorf("%s was accepted: the set rule is lost", name)
		}
	}
	var legal bool
	g.WithWriteLock(func() {
		legal = g.CollectionPickLegalLocked(c, []string{"Lightning Bolt", "Sol Ring"})
	})
	if !legal {
		t.Error("the enumerator's window refuses a pair with different names")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt", "Sol Ring"}); err != nil {
		t.Fatalf("a pair with different names: %v", err)
	}
	if !slices.ContainsFunc(g.Seats[0].SideboardCards(), func(card game.Card) bool { return card.Name == "Sol Ring" }) {
		t.Error("the pair was not minted outside the game")
	}
}

func TestExtrapolatePickWithAnUnknownKeyIsRefused(t *testing.T) {
	t.Run("sideboard", func(t *testing.T) {
		g := newSideboardCatalogGame(t, 2, outsideCard("Alpha", "Sorcery"), outsideCard("Beta", "Instant"))
		castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
		passPriorityAroundTable(t, g)
		refuseUnknownKey(t, g, extrapolatePick.Key())
	})
	t.Run("database", func(t *testing.T) {
		installTestCollection(t)
		g := newSideboardCatalogGame(t, 2)
		castCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", extrapolateOracle, nil)
		passPriorityAroundTable(t, g)
		refuseUnknownKey(t, g, extrapolatePick.Key())
	})
}
