package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_p2b_restore_test.go — #4253's closure sweep, second pass, group B:
// the continuations that rode a primitive whose keyed form landed in slices
// 10a and 10b (SacrificeChoice.ThenTail, a mass verb's ThenTail, a single
// bounce's ThenTail, ExileThenIfItWas.Keyed, PutIntoLibrary.ThenTail) and
// the option picks that were left as closures. Each test opens the prompt,
// takes a STRICT restore through JSON (restoreRoundTrip with strict, which
// fails on a table that is not a restore point) and answers on the restored
// table.

const sacrificeChoiceKey = "choose-cards/sacrifice-choice"

// sweepP2BPick returns the open choose_cards prompt of `chooser`, failing
// the test when it names a key other than `want`.
func sweepP2BPick(t *testing.T, g *game.Game, chooser uuid.UUID, want string) *game.PendingChoice {
	t.Helper()
	c := effectstest.LatestChooseCardsFor(g, chooser)
	if c == nil {
		t.Fatalf("no choose-cards prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if c.PickThen != want {
		t.Fatalf("the prompt names key %q, want %q", c.PickThen, want)
	}
	return c
}

// Braids, Arisen Nightmare: an opponent's sacrifice-or-lose-life question is
// a keyed option pick carrying the opponents still to be asked, the
// permanents offered and the frozen card types; the sacrifice it opens is
// SacrificeChoice's keyed pick, whose tail asks the next opponent. The table
// is restored at each of the three prompts.
func TestSweepP2BBraidsOpponentsAreARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	meID := me.ID
	opp1, opp2 := g.Seats[(seat+1)%4].ID, g.Seats[(seat+2)%4].ID

	effectstest.SeedPermanentWithOracle(g, meID, "Braids, Arisen Nightmare", "Legendary Creature — Nightmare", effectstest.B02dBraidsOracle)
	myRock := effectstest.B12Permanent(g, meID, "My Rock", "Artifact")
	theirRock := effectstest.B12Permanent(g, opp1, "Their Rock", "Artifact")
	otherRock := effectstest.B12Permanent(g, opp2, "Other Rock", "Artifact")
	handBefore := me.Hand.Size()
	opp2Life := g.PlayerByIDForEffect(opp2).Life

	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != meID {
		t.Fatalf("Braids must ask its controller whether to sacrifice: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(ask.ID, meID, true); err != nil {
		t.Fatalf("ResolveConfirm(yes): %v", err)
	}
	pick := effectstest.LatestChooseCardsFor(g, meID)
	if pick == nil {
		t.Fatalf("Braids must ask which permanent to sacrifice: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(pick.ID, meID, []uuid.UUID{myRock}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}

	const optionKey = "option-pick/braids-arisen-nightmare-sacrifice-or-lose-life"
	sweep02OptionKey(t, g, opp1, optionKey)
	g = effectstest.RestoreRoundTrip(t, g, true)
	sweep02OptionKey(t, g, opp1, optionKey)
	effectstest.AnswerOptionPick(t, g, opp1, 1) // sacrifice

	sweepP2BPick(t, g, opp1, sacrificeChoiceKey)
	g = effectstest.RestoreRoundTrip(t, g, true)
	sac := sweepP2BPick(t, g, opp1, sacrificeChoiceKey)
	if !effectstest.HasID(sac.ChooseCards, theirRock) {
		t.Fatalf("the sacrifice offers %v, want the matching Artifact", sac.ChooseCards)
	}
	if err := g.ResolveChooseCards(sac.ID, opp1, []uuid.UUID{theirRock}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if g.Battlefield.Contains(theirRock) {
		t.Fatal("the first opponent's Artifact was not sacrificed on the restored table")
	}

	// The next opponent is asked off the tail, with its own question.
	sweep02OptionKey(t, g, opp2, optionKey)
	g = effectstest.RestoreRoundTrip(t, g, true)
	effectstest.AnswerOptionPick(t, g, opp2, 0) // lose 2 life; Braids's controller draws
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(otherRock) {
		t.Error("the second opponent declined and kept the Artifact")
	}
	if got := g.PlayerByIDForEffect(opp2).Life; got != opp2Life-2 {
		t.Errorf("the second opponent's life = %d, want %d", got, opp2Life-2)
	}
	// One draw for the opponent who declined, one for the third, who had
	// nothing to sacrifice.
	if got := g.PlayerByIDForEffect(meID).Hand.Size(); got != handBefore+2 {
		t.Errorf("Braids's controller holds %d cards, want %d", got, handBefore+2)
	}
}

// Fandaniel, Telophoroi Ascian: the sacrifice an opponent chooses is
// SacrificeChoice's keyed pick, whose tail carries the opponents still to be
// asked, so the restored sacrifice asks the next opponent.
func TestSweepP2BFandanielSacrificeIsARestorePoint(t *testing.T) {
	g := effectstest.FandanielGame(t, 2)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, g.Seats[2].ID, "Third Bear", 2, 2)
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, opp.ID, 1) // sacrifice
	sweepP2BPick(t, g, opp.ID, sacrificeChoiceKey)

	g = effectstest.RestoreRoundTrip(t, g, true)
	opp = g.Seats[1]
	pick := sweepP2BPick(t, g, opp.ID, sacrificeChoiceKey)
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveChooseCards on the restored table: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the Bear was not sacrificed on the restored table")
	}
	third := g.Seats[2]
	sweep02OptionKey(t, g, third.ID, "option-pick/fandaniel-telophoroi-ascian-sacrifice-or-lose-life")
	effectstest.AnswerOptionPick(t, g, third.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(theirs) {
		t.Error("the third seat declined and kept its creature")
	}
	if got := g.Seats[1].Life; got != 40 {
		t.Errorf("the sacrificing opponent's life = %d, want 40", got)
	}
	for _, i := range []int{2, 3} {
		if got := g.Seats[i].Life; got != 36 {
			t.Errorf("seat %d life = %d, want 36 (2 x 2)", i, got)
		}
	}
}

// Archfiend of Spite: the punished player's question is a keyed option pick
// carrying who was punished and how much, so the restored "sacrifice" answer
// takes that many permanents from that player.
func TestSweepP2BArchfiendOfSpiteChoiceIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fiend := effectstest.B12Push(g, me.ID, "Archfiend of Spite", "Creature — Demon", effectstest.ShArchfiendOracle, 6, 6)
	shooter := effectstest.B12Creature(g, opp.ID, "Shooter", "Creature — Human", 1, 1)
	x := effectstest.B12Creature(g, opp.ID, "X", "Creature — Bear", 2, 2)
	y := effectstest.B12Creature(g, opp.ID, "Y", "Creature — Bear", 2, 2)
	oppID, life := opp.ID, opp.Life

	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(shooter, fiend, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	const key = "option-pick/archfiend-of-spite-life-or-sacrifice"
	sweep02OptionKey(t, g, oppID, key)

	g = effectstest.RestoreRoundTrip(t, g, true)
	sweep02OptionKey(t, g, oppID, key)
	effectstest.AnswerOptionPick(t, g, oppID, 1)
	effectstest.AnswerSacrifice(t, g, oppID, x)
	effectstest.AnswerSacrifice(t, g, oppID, y)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(x) || g.Battlefield.Contains(y) {
		t.Error("the two chosen permanents were not sacrificed on the restored table")
	}
	if got := g.PlayerByIDForEffect(oppID).Life; got != life {
		t.Errorf("life %d -> %d, want unchanged: they sacrificed instead", life, got)
	}
}

// Deadly Tempest over a Bear and a Needle (DestroyAllMatching.ThenTail): the
// Needle's destruction asks, and the restored answer charges its controller
// for both creatures.
func TestSweepP2BDeadlyTempestIsARestorePointMidSweep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.TableSeats(g)
	effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	needle := effectstest.B12Creature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	effectstest.SeedOrderProbes(g, opp.ID, effectstest.GraveyardOrderProbeOracles, "Graveyard Order Probe")
	myLife, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Deadly Tempest", "Sorcery", effectstest.DeadlyTempestOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me, opp = effectstest.SeatByID(g, me.ID), effectstest.SeatByID(g, opp.ID)
	effectstest.AnswerExileOrder(t, g)
	if !opp.Graveyard.Contains(needle) {
		t.Fatal("the Needle was not destroyed after the restored answer")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("their life %d -> %d, want -2 for their two creatures destroyed", oppLife, opp.Life)
	}
	if me.Life != myLife {
		t.Errorf("the caster lost life %d -> %d with no creature destroyed", myLife, me.Life)
	}
}

// Chain of Vapor on a Needle (BounceToHand.ThenTail): the bounce asks, and
// the restored answer puts the Needle in its owner's hand and offers its
// controller the chain, read off the tail.
func TestSweepP2BChainOfVaporBounceIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	needle := effectstest.PushTribalCreature(g, opp.ID, "Needle", "Creature — Test", 1, 1)
	effectstest.PushLand(g, opp.ID, "Their Island")
	effectstest.SeedOrderProbes(g, opp.ID, effectstest.LibraryHandOrderProbeOracles, "Library Hand Order Probe")
	effectstest.CastCatalogSpell(t, g, "Chain of Vapor", "Instant", effectstest.ChainOfVaporOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: needle}})
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	opp = effectstest.SeatByID(g, opp.ID)
	effectstest.AnswerExileOrder(t, g)
	if !opp.Hand.Contains(needle) {
		t.Fatal("the Needle is not in its owner's hand after the restored answer")
	}
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != opp.ID {
		t.Fatalf("the bounced permanent's controller must be offered the chain: %+v", g.PendingChoices)
	}
}

// Cling to Dust on a creature card named Needle (ExileThenIfItWas.Keyed):
// the exile asks, and the restored answer gains its caster 3 life.
func TestSweepP2BClingToDustExileIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	needle := effectstest.Batch01GraveyardCard(opp, "Needle", "Creature — Test")
	effectstest.SeedExileOrderProbes(g, opp.ID)
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Cling to Dust", "Instant", effectstest.ClingToDustOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: needle}})
	effectstest.PassPriorityAroundTable(t, g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveBatch)
	me = effectstest.SeatByID(g, me.ID)
	effectstest.AnswerExileOrder(t, g)
	if _, ok := effectstest.FindInZone(g.Exile, needle); !ok {
		t.Fatal("the Needle is not in exile after the restored answer")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d -> %d, want +3: it was a creature card", life, me.Life)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want unchanged: the Otherwise branch must not run", hand, me.Hand.Size())
	}
}
