package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_prompts_restore_test.go — #4256: the outside-the-game picks, the
// discard-then prompts and the pay-unless prompts are keyed, so a table
// waiting on any of them is a restore point. Each test takes a STRICT
// restore (restoreRoundTrip with strict, which fails on a table that is
// not a restore point) while the prompt is open, then answers it on the
// restored table. The *UnknownKeyIsRefused tests are the rollback case:
// a binary that does not know the key refuses the file.

// refuseUnknownKey rewrites every snapshot prompt naming `key` to an
// unregistered key and checks the restore is refused.
func refuseUnknownKey(t *testing.T, g *game.Game, key string) {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].PickThen == key {
			snap.PendingChoices[i].PickThen = "choose-cards/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatalf("the snapshot holds no prompt naming %q", key)
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// --- 1. Outside-the-game picks ---------------------------------------

// Learn's Lesson half from a sideboard: the table waiting on the pick is
// a restore point, and the restored pick offers only the Lessons and
// reveals the one taken.
func TestLearnLessonPickIsARestorePoint(t *testing.T) {
	g, me := learnTable(t, lessonCard("Environmental Sciences"), outsideCard("Lightning Bolt", "Instant"))
	if err := g.ResolveChooseCards(learnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me, opp := g.Seats[0], g.Seats[1]
	lc := learnLessonPrompt(g, me)
	if lc == nil || lc.PickThen != learnLessonTake.pick.Key() {
		t.Fatalf("the restored table's Lesson pick is %+v, want the keyed pick", lc)
	}
	if got := candidateNames(g, lc); !equalNames(got, []string{"Environmental Sciences"}) {
		t.Errorf("restored candidates %v, want only the Lesson", got)
	}
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{sideCardID(me, "Lightning Bolt")}); err == nil {
		t.Error("the restored pick accepted a card that is not a Lesson")
	}
	lesson := sideCardID(me, "Environmental Sciences")
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{lesson}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	card, ok := handCardNamed(me, "Environmental Sciences")
	if !ok || !card.IsKnownTo(opp.ID) {
		t.Errorf("the Lesson came into the hand %v, revealed to the opponent %v", ok, ok && card.IsKnownTo(opp.ID))
	}
}

// Learn's Lesson half from the card database (ADR 0198): the filter is
// found by the prompt's key after the restore, so a typed card that is
// not a Lesson is still refused.
func TestLearnLessonPickFromTheDatabaseIsARestorePoint(t *testing.T) {
	prev := game.OutsideCollection
	game.OutsideCollection = testCollection{cards: []game.Card{
		dbCard("Lightning Bolt", "Instant", lightningBoltOracle, "R"),
		dbCard("Environmental Sciences", lessonTypeLine, "", "G"),
	}}
	t.Cleanup(func() { game.OutsideCollection = prev })

	g, me := learnTable(t)
	if err := g.ResolveChooseCards(learnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := latestCollectionPickFor(g, me.ID)
	if c == nil || c.PickThen != learnLessonTake.pick.Key() {
		t.Fatalf("the restored table's collection pick is %+v, want the keyed pick", c)
	}
	if !slices.Equal(c.CollectionSuggestions, []string{"Environmental Sciences"}) {
		t.Errorf("restored suggestions %v", c.CollectionSuggestions)
	}
	if found, err := g.SearchOutsideCollection(c.ID, me.ID, "Lightning", 10); err != nil || len(found) != 0 {
		t.Errorf("the restored search offers %v (err %v), want no Bolt", found, err)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Lightning Bolt"}); err == nil {
		t.Error("the restored pick took a card that is not a Lesson")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Environmental Sciences"}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if card, ok := handCardNamed(me, "Environmental Sciences"); !ok || card.Owner != me.ID {
		t.Errorf("the minted Lesson is in the hand %v", ok)
	}
}

// Burning Wish from a sideboard, restored strictly while the wish waits:
// the sorcery comes into the hand on the restored table, and the wish's
// next step (it exiles itself) still runs.
func TestBurningWishPickIsARestorePoint(t *testing.T) {
	g := newSideboardCatalogGame(t, 2, outsideCard("Rift Bolt", "Sorcery"), outsideCard("Lightning Bolt", "Instant"))
	wish := castCatalogSpell(t, g, "Burning Wish", "Sorcery", "13a884b5-8970-4d71-9082-a8b6c8db9a01", nil)
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me := g.Seats[0]
	c := latestChooseCardsFor(g, me.ID)
	if c == nil || c.PickThen != burningWishTake.pick.Key() {
		t.Fatalf("the restored wish is %+v, want the keyed pick", c)
	}
	if got := candidateNames(g, c); !equalNames(got, []string{"Rift Bolt"}) {
		t.Errorf("restored candidates %v, want the sorcery alone", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{sideCardID(me, "Rift Bolt")}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !outsideHandHas(me, "Rift Bolt") {
		t.Error("the sorcery did not come into the hand on the restored table")
	}
	if !g.Exile.Contains(wish) {
		t.Error("Burning Wish's next step did not exile it on the restored table")
	}
}

// Spawnsire of Ulamog from the card database: the subtype is the
// ability's word, carried by the prompt (PickMatchArg), so the restored
// filter still reads "Eldrazi".
func TestSpawnsireDatabasePickKeepsItsWordAcrossARestore(t *testing.T) {
	installOutsideCollection(t)
	g := newSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	src := spawnsireReady(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := latestCollectionPickFor(g, me.ID)
	if c == nil || c.PickThen != spawnsireOfUlamogCast.pick.Key() || c.PickMatchArg != "Eldrazi" {
		t.Fatalf("the restored pick is %+v, want the keyed pick with its word", c)
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Grizzly Bears"}); err == nil {
		t.Fatal("the restored pick accepted a card that is not an Eldrazi")
	}
	if err := g.ResolveOutsideCollection(c.ID, me.ID, []string{"Ulamog's Crusher"}); err != nil {
		t.Fatal(err)
	}
	cards := me.SideboardCards()
	if len(cards) != 1 {
		t.Fatalf("minted %d cards, want 1", len(cards))
	}
	if err := castFromOutside(g, me, cards[0].InstanceID, true); err != nil {
		t.Fatalf("the free cast on the restored table: %v", err)
	}
}

func TestOutsidePickWithAnUnknownKeyIsRefused(t *testing.T) {
	t.Run("sideboard", func(t *testing.T) {
		g := newSideboardCatalogGame(t, 2, outsideCard("Rift Bolt", "Sorcery"))
		castCatalogSpell(t, g, "Burning Wish", "Sorcery", "13a884b5-8970-4d71-9082-a8b6c8db9a01", nil)
		passPriorityAroundTable(t, g)
		refuseUnknownKey(t, g, burningWishTake.pick.Key())
	})
	t.Run("database", func(t *testing.T) {
		installOutsideCollection(t)
		g := newSideboardCatalogGame(t, 2)
		me := g.Seats[0]
		src := spawnsireReady(t, g, me)
		if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		refuseUnknownKey(t, g, spawnsireOfUlamogCast.pick.Key())
	})
}

// --- 2. Discard-then -------------------------------------------------

// Learn's discard half, restored strictly while it is open: discarding
// on the restored table draws a card, and declining offers the Lesson.
func TestLearnDiscardIsARestorePoint(t *testing.T) {
	for _, discard := range []bool{true, false} {
		g, _ := learnTable(t, lessonCard("Environmental Sciences"))
		g = restoreRoundTrip(t, g, true)
		me := g.Seats[0]
		c := learnDiscardPrompt(g, me)
		if c == nil || c.PickThen != learnAfterDiscard.Key() {
			t.Fatalf("the restored discard is %+v, want the keyed prompt", c)
		}
		hand := me.Hand.Size()
		if !discard {
			if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
			if learnLessonPrompt(g, me) == nil {
				t.Errorf("declining on the restored table offered no Lesson: %+v", g.PendingChoices)
			}
			continue
		}
		pitched := c.ChooseCards[0]
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitched}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(pitched) || me.Hand.Size() != hand {
			t.Errorf("discarded %v, hand %d -> %d, want one discarded and one drawn",
				me.Graveyard.Contains(pitched), hand, me.Hand.Size())
		}
		if learnLessonPrompt(g, me) != nil {
			t.Error("a player who discarded on the restored table was offered a Lesson")
		}
	}
}

// Abandon Attachments' "you may discard a card. If you do, draw two",
// restored strictly while the discard is open.
func TestAbandonAttachmentsDiscardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := handCardFull(me, "Fodder", "Sorcery", "", "", nil)
	castCatalogSpell(t, g, "Abandon Attachments", "Instant — Lesson", b35AbandonAttachmentsOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := discardChoiceFor(g, me.ID)
	if c == nil || c.PickThen != abandonAttachmentsDraw.Key() {
		t.Fatalf("the restored discard is %+v, want the keyed prompt", c)
	}
	answerDiscard(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) || me.Hand.Size() != hand+1 {
		t.Errorf("discarded %v, hand %d -> %d, want -1 +2", me.Graveyard.Contains(fodder), hand, me.Hand.Size())
	}
}

// Charming Scoundrel's rummage, restored strictly while the discard is
// open: the draw follows the discard on the restored table.
func TestCharmingScoundrelRummageIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castAndResolveCreature(t, g, "Charming Scoundrel", "Creature — Human Rogue", scoundrelOracle)
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatal(err)
	}
	pitch := game.NewCard("Pitch", me.ID)
	me.Hand.PushTop(pitch)
	seedLibrary(me, "A", "B")
	passPriorityAroundTable(t, g)
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	d := discardChoiceFor(g, me.ID)
	if d == nil || d.PickThen != charmingScoundrelDraw.Key() {
		t.Fatalf("the restored discard is %+v, want the keyed prompt", d)
	}
	before := handCount(g, me)
	answerDiscard(t, g, me.ID, pitch.InstanceID)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch.InstanceID) || handCount(g, me) != before {
		t.Errorf("discarded %v, hand %d -> %d, want discard one, draw one",
			me.Graveyard.Contains(pitch.InstanceID), before, handCount(g, me))
	}
}

func TestDiscardThenWithAnUnknownKeyIsRefused(t *testing.T) {
	g, _ := learnTable(t, lessonCard("Environmental Sciences"))
	refuseUnknownKey(t, g, learnAfterDiscard.Key())
}

// --- 3. Pay-unless ---------------------------------------------------

// smotheringTitheAsks is Smothering Tithe under seat 1 and seat 0
// drawing in their main phase: the Tithe's payer is asked.
func smotheringTitheAsks(t *testing.T) *game.Game {
	t.Helper()
	g := newCatalogGame(t)
	drawer, owner := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Smothering Tithe", smotheringTitheOracle, "Enchantment")
	advanceToMain(t, g)
	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, drawer.ID) {
		t.Fatalf("no pay_unless prompt for the drawer: %+v", g.PendingChoices)
	}
	return g
}

// Smothering Tithe's tax, restored strictly while the payer decides: a
// decline on the restored table makes the Treasure, and a payment does
// not.
func TestSmotheringTithePayUnlessIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := restoreRoundTrip(t, smotheringTitheAsks(t), true)
		drawer, owner := g.Seats[0], g.Seats[1]
		if pay {
			drawer.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		}
		prompt := answerPayUnless(t, g, drawer.ID, pay)
		if prompt.PickThen != smotheringTitheDecline.Key() || prompt.PayCost != "{2}" {
			t.Fatalf("restored prompt: key %q cost %q", prompt.PickThen, prompt.PayCost)
		}
		want := 1
		if pay {
			want = 0
		}
		if n := countOnBattlefield(g, "Treasure", owner.ID); n != want {
			t.Errorf("pay %v: %d Treasures under the Tithe's controller, want %d", pay, n, want)
		}
	}
}

// Rhystic Study's tax, restored strictly while the caster decides: the
// decline on the restored table asks the Study's controller whether to
// draw, and "yes" draws.
func TestRhysticStudyPayUnlessIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Rhystic Study", rhysticStudyOracle, "Enchantment")
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	passPriorityAroundTable(t, g)
	if !hasPayUnlessFor(g, caster.ID) {
		t.Fatalf("no pay_unless prompt for the caster: %+v", g.PendingChoices)
	}
	g = restoreRoundTrip(t, g, true)
	caster, owner = g.Seats[0], g.Seats[1]
	hand := owner.Hand.Size()
	prompt := answerPayUnless(t, g, caster.ID, false)
	if prompt.PickThen != rhysticStudyDecline.Key() {
		t.Fatalf("restored prompt key %q, want %q", prompt.PickThen, rhysticStudyDecline.Key())
	}
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != owner.ID {
		t.Fatalf("no 'draw a card?' for the Study's controller on the restored table: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, owner.ID, true); err != nil {
		t.Fatal(err)
	}
	if got := owner.Hand.Size() - hand; got != 1 {
		t.Errorf("the Study's controller drew %d on the restored table, want 1", got)
	}
}

func TestPayUnlessWithAnUnknownKeyIsRefused(t *testing.T) {
	refuseUnknownKey(t, smotheringTitheAsks(t), smotheringTitheDecline.Key())
}

// A payer who leaves the game with the keyed prompt open has not paid
// (CR 800.4f's reading, #961): the decline still runs, found by its key.
func TestSmotheringTitheKeyedDeclineRunsWhenThePayerLeaves(t *testing.T) {
	g := restoreRoundTrip(t, smotheringTitheAsks(t), true)
	drawer, owner := g.Seats[0], g.Seats[1]
	if err := g.Concede(drawer.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := countOnBattlefield(g, "Treasure", owner.ID); n != 1 {
		t.Errorf("%d Treasures after the payer left, want 1", n)
	}
}
