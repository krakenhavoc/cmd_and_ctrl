package testsi

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsbu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/bu"
	cardsr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/r"
	cardssh "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/sh"
	cardsso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/so"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- 1. Outside-the-game picks ---------------------------------------

// Learn's Lesson half from a sideboard: the table waiting on the pick is
// a restore point, and the restored pick offers only the Lessons and
// reveals the one taken.
func TestLearnLessonPickIsARestorePoint(t *testing.T) {
	g, me := effectstest.LearnTable(t, effectstest.LessonCard("Environmental Sciences"), effectstest.OutsideCard("Lightning Bolt", "Instant"))
	if err := g.ResolveChooseCards(effectstest.LearnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp := g.Seats[0], g.Seats[1]
	lc := effectstest.LearnLessonPrompt(g, me)
	if lc == nil || lc.PickThen != effects.LearnLessonTake.Pick.Key() {
		t.Fatalf("the restored table's Lesson pick is %+v, want the keyed pick", lc)
	}
	if got := effectstest.CandidateNames(g, lc); !effectstest.EqualNames(got, []string{"Environmental Sciences"}) {
		t.Errorf("restored candidates %v, want only the Lesson", got)
	}
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Lightning Bolt")}); err == nil {
		t.Error("the restored pick accepted a card that is not a Lesson")
	}
	lesson := effectstest.SideCardID(me, "Environmental Sciences")
	if err := g.ResolveChooseCards(lc.ID, me.ID, []uuid.UUID{lesson}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	card, ok := effectstest.HandCardNamed(me, "Environmental Sciences")
	if !ok || !card.IsKnownTo(opp.ID) {
		t.Errorf("the Lesson came into the hand %v, revealed to the opponent %v", ok, ok && card.IsKnownTo(opp.ID))
	}
}

// Learn's Lesson half from the card database (ADR 0198): the filter is
// found by the prompt's key after the restore, so a typed card that is
// not a Lesson is still refused.
func TestLearnLessonPickFromTheDatabaseIsARestorePoint(t *testing.T) {
	prev := game.OutsideCollection
	game.OutsideCollection = effectstest.TestCollection{Cards: []game.Card{
		effectstest.DbCard("Lightning Bolt", "Instant", effectstest.LightningBoltOracle, "R"),
		effectstest.DbCard("Environmental Sciences", effectstest.LessonTypeLine, "", "G"),
	}}
	t.Cleanup(func() { game.OutsideCollection = prev })

	g, me := effectstest.LearnTable(t)
	if err := g.ResolveChooseCards(effectstest.LearnDiscardPrompt(g, me).ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.LatestCollectionPickFor(g, me.ID)
	if c == nil || c.PickThen != effects.LearnLessonTake.Pick.Key() {
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
	effectstest.PassPriorityAroundTable(t, g)
	if card, ok := effectstest.HandCardNamed(me, "Environmental Sciences"); !ok || card.Owner != me.ID {
		t.Errorf("the minted Lesson is in the hand %v", ok)
	}
}

// Burning Wish from a sideboard, restored strictly while the wish waits:
// the sorcery comes into the hand on the restored table, and the wish's
// next step (it exiles itself) still runs.
func TestBurningWishPickIsARestorePoint(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Rift Bolt", "Sorcery"), effectstest.OutsideCard("Lightning Bolt", "Instant"))
	wish := effectstest.CastCatalogSpell(t, g, "Burning Wish", "Sorcery", "13a884b5-8970-4d71-9082-a8b6c8db9a01", nil)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me := g.Seats[0]
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || c.PickThen != cardsbu.BurningWishTake.Pick.Key() {
		t.Fatalf("the restored wish is %+v, want the keyed pick", c)
	}
	if got := effectstest.CandidateNames(g, c); !effectstest.EqualNames(got, []string{"Rift Bolt"}) {
		t.Errorf("restored candidates %v, want the sorcery alone", got)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{effectstest.SideCardID(me, "Rift Bolt")}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Rift Bolt") {
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
	effectstest.InstallOutsideCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	src := effectstest.SpawnsireReady(t, g, me)
	if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.LatestCollectionPickFor(g, me.ID)
	if c == nil || c.PickThen != cardsso.SpawnsireOfUlamogCast.Pick.Key() || c.PickMatchArg != "Eldrazi" {
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
	if err := effectstest.CastFromOutside(g, me, cards[0].InstanceID, true); err != nil {
		t.Fatalf("the free cast on the restored table: %v", err)
	}
}

func TestOutsidePickWithAnUnknownKeyIsRefused(t *testing.T) {
	t.Run("sideboard", func(t *testing.T) {
		g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Rift Bolt", "Sorcery"))
		effectstest.CastCatalogSpell(t, g, "Burning Wish", "Sorcery", "13a884b5-8970-4d71-9082-a8b6c8db9a01", nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.RefuseUnknownKey(t, g, cardsbu.BurningWishTake.Pick.Key())
	})
	t.Run("database", func(t *testing.T) {
		effectstest.InstallOutsideCollection(t)
		g := effectstest.NewSideboardCatalogGame(t, 2)
		me := g.Seats[0]
		src := effectstest.SpawnsireReady(t, g, me)
		if err := g.ActivateCatalogAbility(me.ID, src, 1, game.ActivateAbilityParams{}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.RefuseUnknownKey(t, g, cardsso.SpawnsireOfUlamogCast.Pick.Key())
	})
}

// --- 2. Discard-then -------------------------------------------------

// Learn's discard half, restored strictly while it is open: discarding
// on the restored table draws a card, and declining offers the Lesson.
func TestLearnDiscardIsARestorePoint(t *testing.T) {
	for _, discard := range []bool{true, false} {
		g, _ := effectstest.LearnTable(t, effectstest.LessonCard("Environmental Sciences"))
		g = effectstest.RestoreRoundTrip(t, g, true)
		me := g.Seats[0]
		c := effectstest.LearnDiscardPrompt(g, me)
		if c == nil || c.PickThen != effects.LearnAfterDiscard.Key() {
			t.Fatalf("the restored discard is %+v, want the keyed prompt", c)
		}
		hand := me.Hand.Size()
		if !discard {
			if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
				t.Fatal(err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.LearnLessonPrompt(g, me) == nil {
				t.Errorf("declining on the restored table offered no Lesson: %+v", g.PendingChoices)
			}
			continue
		}
		pitched := c.ChooseCards[0]
		if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitched}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(pitched) || me.Hand.Size() != hand {
			t.Errorf("discarded %v, hand %d -> %d, want one discarded and one drawn",
				me.Graveyard.Contains(pitched), hand, me.Hand.Size())
		}
		if effectstest.LearnLessonPrompt(g, me) != nil {
			t.Error("a player who discarded on the restored table was offered a Lesson")
		}
	}
}

// Abandon Attachments' "you may discard a card. If you do, draw two",
// restored strictly while the discard is open.
func TestAbandonAttachmentsDiscardIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.HandCardFull(me, "Fodder", "Sorcery", "", "", nil)
	effectstest.CastCatalogSpell(t, g, "Abandon Attachments", "Instant — Lesson", effectstest.B35AbandonAttachmentsOracle, nil)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil || c.PickThen != effects.AbandonAttachmentsDraw.Key() {
		t.Fatalf("the restored discard is %+v, want the keyed prompt", c)
	}
	effectstest.AnswerDiscard(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) || me.Hand.Size() != hand+1 {
		t.Errorf("discarded %v, hand %d -> %d, want -1 +2", me.Graveyard.Contains(fodder), hand, me.Hand.Size())
	}
}

// Charming Scoundrel's rummage, restored strictly while the discard is
// open: the draw follows the discard on the restored table.
func TestCharmingScoundrelRummageIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastAndResolveCreature(t, g, "Charming Scoundrel", "Creature — Human Rogue", effectstest.ScoundrelOracle)
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{0}); err != nil {
		t.Fatal(err)
	}
	pitch := game.NewCard("Pitch", me.ID)
	me.Hand.PushTop(pitch)
	effectstest.SeedLibrary(me, "A", "B")
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	d := effectstest.DiscardChoiceFor(g, me.ID)
	if d == nil || d.PickThen != effects.CharmingScoundrelDraw.Key() {
		t.Fatalf("the restored discard is %+v, want the keyed prompt", d)
	}
	before := effectstest.HandCount(g, me)
	effectstest.AnswerDiscard(t, g, me.ID, pitch.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(pitch.InstanceID) || effectstest.HandCount(g, me) != before {
		t.Errorf("discarded %v, hand %d -> %d, want discard one, draw one",
			me.Graveyard.Contains(pitch.InstanceID), before, effectstest.HandCount(g, me))
	}
}

func TestDiscardThenWithAnUnknownKeyIsRefused(t *testing.T) {
	g, _ := effectstest.LearnTable(t, effectstest.LessonCard("Environmental Sciences"))
	effectstest.RefuseUnknownKey(t, g, effects.LearnAfterDiscard.Key())
}

// --- 3. Pay-unless ---------------------------------------------------

// smotheringTitheAsks is Smothering Tithe under seat 1 and seat 0
// drawing in their main phase: the Tithe's payer is asked.
func smotheringTitheAsks(t *testing.T) *game.Game {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	drawer, owner := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Smothering Tithe", effectstest.SmotheringTitheOracle, "Enchantment")
	effectstest.AdvanceToMain(t, g)
	if err := g.DrawCard(drawer.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, drawer.ID) {
		t.Fatalf("no pay_unless prompt for the drawer: %+v", g.PendingChoices)
	}
	return g
}

// Smothering Tithe's tax, restored strictly while the payer decides: a
// decline on the restored table makes the Treasure, and a payment does
// not.
func TestSmotheringTithePayUnlessIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.RestoreRoundTrip(t, smotheringTitheAsks(t), true)
		drawer, owner := g.Seats[0], g.Seats[1]
		if pay {
			drawer.ManaPool.AddMana(game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
		}
		prompt := effectstest.AnswerPayUnless(t, g, drawer.ID, pay)
		if prompt.PickThen != cardssh.SmotheringTitheDecline.Key() || prompt.PayCost != "{2}" {
			t.Fatalf("restored prompt: key %q cost %q", prompt.PickThen, prompt.PayCost)
		}
		want := 1
		if pay {
			want = 0
		}
		if n := effectstest.CountOnBattlefield(g, "Treasure", owner.ID); n != want {
			t.Errorf("pay %v: %d Treasures under the Tithe's controller, want %d", pay, n, want)
		}
	}
}

// Rhystic Study's tax, restored strictly while the caster decides: the
// decline on the restored table asks the Study's controller whether to
// draw, and "yes" draws.
func TestRhysticStudyPayUnlessIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, caster.ID) {
		t.Fatalf("no pay_unless prompt for the caster: %+v", g.PendingChoices)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	caster, owner = g.Seats[0], g.Seats[1]
	hand := owner.Hand.Size()
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PickThen != cardsr.RhysticStudyDecline.Key() {
		t.Fatalf("restored prompt key %q, want %q", prompt.PickThen, cardsr.RhysticStudyDecline.Key())
	}
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
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
	effectstest.RefuseUnknownKey(t, smotheringTitheAsks(t), cardssh.SmotheringTitheDecline.Key())
}

// A payer who leaves the game with the keyed prompt open has not paid
// (CR 800.4f's reading, #961): the decline still runs, found by its key.
func TestSmotheringTitheKeyedDeclineRunsWhenThePayerLeaves(t *testing.T) {
	g := effectstest.RestoreRoundTrip(t, smotheringTitheAsks(t), true)
	drawer, owner := g.Seats[0], g.Seats[1]
	if err := g.Concede(drawer.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := effectstest.CountOnBattlefield(g, "Treasure", owner.ID); n != 1 {
		t.Errorf("%d Treasures after the payer left, want 1", n)
	}
}
