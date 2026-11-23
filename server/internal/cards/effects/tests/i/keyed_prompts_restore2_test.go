package testsi

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsco "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/co"
	cardse "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/e"
	cardsex "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/ex"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	cardsmi "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/mi"
	cardsp "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/p"
	cardsr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/r"
	cardsw "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/w"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		g := effectstest.NewCatalogGame(t)
		caster, owner := g.Seats[0], g.Seats[1]
		effectstest.PushPermanentForTest(g, owner.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
			[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPayUnless(t, g, caster.ID, false)
		if ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm); ask == nil || ask.PickThen != cardsr.RhysticStudyDraw.Key() {
			t.Fatalf("the 'draw a card?' prompt is %+v, want the keyed confirm", ask)
		}
		g = effectstest.RestoreRoundTrip(t, g, true)
		owner = g.Seats[1]
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
		if ask == nil || ask.Chooser != owner.ID || ask.PickThen != cardsr.RhysticStudyDraw.Key() {
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
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Mystic Remora", effectstest.OracleMysticRemora, "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, caster.ID, false)
	g = effectstest.RestoreRoundTrip(t, g, true)
	owner = g.Seats[1]
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != owner.ID || ask.PickThen != cardsmi.MysticRemoraDraw.Key() {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Wandering Archaic", "Creature — Avatar", effectstest.WanderingArchaicOracle, false)
	for g.Turn.ActiveSeat != 1 || (g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	lifeBefore := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.PickThen != cardsw.WanderingArchaicCopy.Key() || len(ask.OptionCarry) != 2 {
		t.Fatalf("the 'copy that spell?' prompt is %+v, want the keyed confirm carrying the controller and the spell", ask)
	}
	carried := slices.Clone(ask.OptionCarry)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	ask = effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID || ask.PickThen != cardsw.WanderingArchaicCopy.Key() || !slices.Equal(ask.OptionCarry, carried) {
		t.Fatalf("restored: %+v, want the keyed confirm with its carry %v", ask, carried)
	}
	var before int
	g.ReadSnapshot(func() { before = len(g.Events) })
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.SettleThroughCopyRetarget(t, g, me.ID, me.ID)
	if n := effectstest.CountCatalogEvents(g, game.EventEffectError, before); n != 0 {
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
	g := effectstest.NewCatalogGame(t)
	caster, owner := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Rhystic Study", effectstest.RhysticStudyOracle, "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: owner.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, caster.ID, false)
	effectstest.RefuseUnknownKey(t, g, cardsr.RhysticStudyDraw.Key())
}

// --- 2. The one-seat discard-then callers -----------------------------

// Last Rites: the target rides the discard prompt as its Carry, so the
// hand that is revealed after the restore is the target's.
func TestLastRitesDiscardIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { caster.Hand.Cards = nil })
	a := effectstest.HandCardForTest(caster, "Card A", "Sorcery", "")
	b := effectstest.HandCardForTest(caster, "Card B", "Sorcery", "")
	effectstest.HandCardForTest(caster, "Card C", "Sorcery", "")
	effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.RhBolt(), effectstest.PbBear(), effectstest.PbDivination())

	effectstest.CastAt(t, g, "Last Rites", "Sorcery", effectstest.LastRitesOracle, effectstest.PbPlayer(victim.ID))
	if ask := effectstest.ChooseCardsChoiceFor(g, caster.ID); ask == nil || ask.PickThen != cardskr.LastRitesReveal.Key() {
		t.Fatalf("the discard is %+v, want the keyed prompt", ask)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	caster = g.Seats[0]
	ask := effectstest.ChooseCardsChoiceFor(g, caster.ID)
	if ask == nil || ask.PickThen != cardskr.LastRitesReveal.Key() || len(ask.OptionCarry) != 1 || ask.OptionCarry[0] != victim.ID {
		t.Fatalf("restored: %+v, want the keyed discard carrying the target", ask)
	}
	effectstest.AnswerChooseCards(t, g, caster.ID, a, b)
	pick := effectstest.OpenPick(t, g)
	if pick.Count != 2 || len(pick.DiscardOptions) != 3 || pick.FromPlayer != victim.ID {
		t.Fatalf("pick = count %d of %d options from %v, want 2 of the target's 3 nonland cards",
			pick.Count, len(pick.DiscardOptions), pick.FromPlayer)
	}
}

// Eriette's Whisper: the caster and the creature (read as the spell
// resolved) ride the prompt, so the Role goes on the same creature, under
// the caster's control, after the restore.
func TestEriettesWhisperDiscardIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	for i := 0; i < 3; i++ {
		opp.Hand.PushTop(game.NewCard("Filler", opp.ID))
	}
	effectstest.CastCatalogSpell(t, g, "Eriette's Whisper", "Sorcery", effectstest.ErietteOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil || c.PickThen != cardse.EriettesWhisperRole.Key() || c.ChooseMax != 2 {
		t.Fatalf("restored: %+v, want the keyed discard of two", c)
	}
	effectstest.WantNoRole(t, g, mine)
	effectstest.DiscardFromHand(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WantRole(t, g, mine, "Wicked Role")
}

// Pia, Aether Ascetic: the search follows the discard on the restored
// table, and only for a discard that was made.
func TestPiaDiscardIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ench := effectstest.PushLibraryCardForTest(me, game.Card{
		InstanceID: uuid.New(), Name: "Test Enchantment", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID,
	})
	effectstest.RfdCast(t, g, "Pia, Aether Ascetic", "Legendary Creature — Human Druid", effectstest.RfdPiaAetherOracle, 2, 2)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil || c.PickThen != cardsp.PiaSearch.Key() {
		t.Fatalf("restored: %+v, want the keyed discard", c)
	}
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(ench) {
		t.Error("the enchantment was not searched into the hand on the restored table")
	}
}

// Mishra's Command: the draw counts what was really discarded, by the
// player who discarded, after the restore.
func TestMishrasCommandDiscardIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.B19CastXModal(t, g, "Mishra's Command", "Sorcery", effectstest.MishrasCommandOracle, "{X}{R}", 3,
		[]int{0, 3},
		[]game.TargetRef{
			effectstest.ModeRef(game.TargetPlayer, opp.ID, 0, 0),
			effectstest.ModeRef(game.TargetCard, mine, 1, 0),
		})
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	c := effectstest.DiscardChoiceFor(g, opp.ID)
	if c == nil || c.PickThen != effects.MishrasCommandDraw.Key() {
		t.Fatalf("restored: %+v, want the keyed discard", c)
	}
	hand := opp.Hand.Size()
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	g.SettleResolution()
	if got := opp.Hand.Size(); got != hand {
		t.Errorf("discarded 2, drew 2 back: hand %d -> %d", hand, got)
	}
}

// Thrilling Discovery's "discard two cards" is two or none, never one:
// the set rule is registered with the key, so it holds after a restore
// and the enumerator's window agrees.
func TestThrillingDiscoveryDiscardKeepsItsSetRuleAcrossARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Thrilling Discovery", "Sorcery", effectstest.B39ThrillingDiscoveryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me := g.Seats[0]
	c := effectstest.DiscardChoiceFor(g, me.ID)
	if c == nil || c.PickThen != effects.DrawThreeIfDiscardedTwo.Then.Key() {
		t.Fatalf("restored: %+v, want the keyed discard", c)
	}
	hand := me.Hand.Size()
	one := effectstest.B39HandIDs(me, 1)
	if err := g.ResolveChooseCards(c.ID, me.ID, one); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("a single card: err = %v, want ErrChoiceSetRejected", err)
	}
	var legalOne, legalTwo bool
	g.WithWriteLock(func() {
		legalOne = g.ChooseCardsPickLegalLocked(c, one)
		legalTwo = g.ChooseCardsPickLegalLocked(c, effectstest.B39HandIDs(me, 2))
	})
	if legalOne || !legalTwo {
		t.Errorf("the enumerator's window offers one card %v and two cards %v, want false and true", legalOne, legalTwo)
	}
	effectstest.AnswerDiscard(t, g, me.ID, effectstest.B39HandIDs(me, 2)...)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("two for three: hand %d -> %d, want %d", hand, me.Hand.Size(), hand+1)
	}
}

// Tersa Lightshatter's "discard up to two, then draw that many" shares
// one registered draw rule with the rest of its family.
func TestTersaDiscardIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Tersa Lightshatter", "Legendary Creature — Orc Wizard", effectstest.TersaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me := g.Seats[0]
	ch := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if ch == nil || ch.PickThen != effects.DrawAsManyAsDiscarded.Then.Key() {
		t.Fatalf("restored: %+v, want the keyed discard", ch)
	}
	hand := me.Hand.Size()
	two := append([]uuid.UUID(nil), ch.ChooseCards[:2]...)
	if err := g.ResolveChooseCards(ch.ID, me.ID, two); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d -> %d, want unchanged: discard two, draw two", hand, got)
	}
}

func TestDiscardThenCallersWithAnUnknownKeyAreRefused(t *testing.T) {
	t.Run("last rites", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		caster, victim := g.Seats[0], g.Seats[1]
		effectstest.HandCardForTest(caster, "Card A", "Sorcery", "")
		effectstest.RevealHand(victim, effectstest.RhForest())
		effectstest.CastAt(t, g, "Last Rites", "Sorcery", effectstest.LastRitesOracle, effectstest.PbPlayer(victim.ID))
		effectstest.RefuseUnknownKey(t, g, cardskr.LastRitesReveal.Key())
	})
	t.Run("draw rule", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		effectstest.CastCatalogSpell(t, g, "Thrilling Discovery", "Sorcery", effectstest.B39ThrillingDiscoveryOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.RefuseUnknownKey(t, g, effects.DrawThreeIfDiscardedTwo.Then.Key())
	})
}

// --- 3. Karn's and Coax's exile half (a string carry) ----------------

// Karn's −2, exile half, restored strictly while the pick is open: the
// chosen card goes to the hand, and a card that left exile meanwhile does
// not.
func TestKarnExilePickIsARestorePoint(t *testing.T) {
	g, karn := effectstest.KarnGame(t)
	me := g.Seats[0]
	relic := effectstest.CoaxExile(g, me.ID, "Exiled Relic", "Artifact", false)
	effectstest.CoaxExile(g, me.ID, "Exiled Bear", "Creature — Bear", false)
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || c.PickThen != cardsj.KarnTheGreatCreatorExile.Then.Key() || len(c.ChooseCards) != 1 || c.ChooseCards[0] != relic {
		t.Fatalf("restored: %+v, want the keyed exile pick of the artifact", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{relic}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, relic) {
		t.Error("the artifact did not come into the hand on the restored table")
	}
}

// Karn's "none" on the restored exile pick asks the pick from outside
// the game, over artifacts only.
func TestKarnExilePickNoneLooksOutsideAfterARestore(t *testing.T) {
	effectstest.InstallTestCollection(t)
	g, karn := effectstest.KarnGame(t)
	me := g.Seats[0]
	effectstest.CoaxExile(g, me.ID, "Exiled Relic", "Artifact", false)
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatalf("no exile pick after the restore: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	// This table's deck seated no sideboard, so "outside" is the database.
	pick := effectstest.LatestCollectionPickFor(g, me.ID)
	if pick == nil || !effectstest.EqualNames(pick.CollectionSuggestions, []string{"Sol Ring"}) {
		t.Fatalf("\"none\" on the restored exile pick asked %+v, want the database's artifact", pick)
	}
}

// Coax's exile pick, restored strictly: "none" asks the outside pick
// with the spell's word for "Eldrazi", which rode the prompt.
func TestCoaxExilePickKeepsItsWordAcrossARestore(t *testing.T) {
	effectstest.InstallTestCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	me := g.Seats[0]
	effectstest.CoaxExile(g, me.ID, "Eldrazi Devastator", "Creature — Eldrazi", false)
	effectstest.CastCatalogSpell(t, g, "Coax from the Blind Eternities", "Sorcery", effectstest.CoaxOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || c.PickThen != cardsco.CoaxFromTheBlindEternitiesExile.Then.Key() || c.PickMatchArg != "Eldrazi" {
		t.Fatalf("restored: %+v, want the keyed exile pick with its word", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	pick := effectstest.LatestCollectionPickFor(g, me.ID)
	if pick == nil || pick.PickMatchArg != "Eldrazi" || !effectstest.EqualNames(pick.CollectionSuggestions, []string{"Ulamog's Crusher"}) {
		t.Fatalf("the outside pick after \"none\" is %+v, want the Eldrazi from the database with its word", pick)
	}
}

func TestExilePickWithAnUnknownKeyIsRefused(t *testing.T) {
	g, karn := effectstest.KarnGame(t)
	me := g.Seats[0]
	effectstest.CoaxExile(g, me.ID, "Exiled Relic", "Artifact", false)
	if err := g.ActivateCatalogAbility(me.ID, karn, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.RefuseUnknownKey(t, g, cardsj.KarnTheGreatCreatorExile.Then.Key())
}

// --- 4. Extrapolate the Impossible (a set rule) -----------------------

// The first prompt, from a sideboard, restored strictly: the set rule
// "exactly two with different names" still refuses one card, a pair of
// copies and three cards, and the enumerator's window offers the pairs
// the engine accepts. The reveal then follows on the restored table.
func TestExtrapolateTheImpossiblePickIsARestorePoint(t *testing.T) {
	g := effectstest.NewSideboardCatalogGame(t, 2,
		effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Beta", "Instant"))
	effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp := g.Seats[0], g.Seats[1]
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil || c.PickThen != cardsex.ExtrapolatePick.Key() || c.ChooseMin != 0 || c.ChooseMax != 2 {
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
	if effectstest.PendingOfKind(g, game.PendingChoiceOptionPick) != nil {
		effectstest.AnswerChosenPlayer(t, g, me.ID, opp.ID)
	}
	rp := effectstest.LatestRevealPickFor(g, opp.ID)
	if rp == nil || len(rp.ChooseCards) != 2 {
		t.Fatalf("the opponent was not shown the pair on the restored table: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(rp.ID, opp.ID, []uuid.UUID{cards[2].InstanceID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OutsideHandHas(me, "Beta") {
		t.Error("the card the opponent chose did not come into the hand")
	}
}

// The database branch (a deck with no sideboard): the set rule is found
// by the prompt's key after the restore, so two copies of one name are
// still refused.
func TestExtrapolateTheImpossibleDatabasePickIsARestorePoint(t *testing.T) {
	effectstest.InstallTestCollection(t)
	g := effectstest.NewSideboardCatalogGame(t, 2)
	effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.RestoreRoundTrip(t, g, true)
	me := g.Seats[0]
	c := effectstest.LatestCollectionPickFor(g, me.ID)
	if c == nil || c.PickThen != cardsex.ExtrapolatePick.Key() {
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
		g := effectstest.NewSideboardCatalogGame(t, 2, effectstest.OutsideCard("Alpha", "Sorcery"), effectstest.OutsideCard("Beta", "Instant"))
		effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.RefuseUnknownKey(t, g, cardsex.ExtrapolatePick.Key())
	})
	t.Run("database", func(t *testing.T) {
		effectstest.InstallTestCollection(t)
		g := effectstest.NewSideboardCatalogGame(t, 2)
		effectstest.CastCatalogSpell(t, g, "Extrapolate the Impossible", "Sorcery", effectstest.ExtrapolateOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.RefuseUnknownKey(t, g, cardsex.ExtrapolatePick.Key())
	})
}
