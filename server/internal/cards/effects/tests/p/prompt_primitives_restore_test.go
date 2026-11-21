package testsp

import (
	"testing"

	"github.com/google/uuid"

	cardskr "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/kr"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prompt_primitives_restore_test.go — #4253 slice 10b: the prompt-level
// primitives' keyed forms. Each test takes a STRICT restore while the
// primitive's prompt is open (restoreRoundTrip with strict, the boot path)
// and answers it on the restored table.

// Krark, the Thumbless: the coin call is keyed (coin-flip/krark-the-thumbless,
// the spell in its Carry, the trigger as its asked item), and the restored
// call copies the spell or returns it.
func TestKrarksCoinCallIsARestorePoint(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushDiesCreatureForTest(g, me.ID, "Krark, the Thumbless", effectstest.KrarkThumblessOrcl, "Legendary Creature — Goblin Wizard", 2, 2)
		div := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.SpellControlDivinationOracle, nil)
		before := me.Hand.Size()
		effectstest.PassPriorityAroundTable(t, g)
		flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil {
			t.Fatal("Krark asked for no coin call")
		}
		if flip.PickThen != cardskr.KrarkFlipped.Key() {
			t.Fatalf("the coin call's key is %q, want Krark's keyed continuation", flip.PickThen)
		}

		g = effectstest.RestoreRoundTrip(t, g, true)
		me = g.Seats[0]
		flip = effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil {
			t.Fatal("the restored table has no coin call")
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		won := false
		for _, ev := range effectstest.RandomEvents(g, game.EventFlipCoin) {
			won = ev.Won
		}
		effectstest.PassPriorityAroundTable(t, g)
		if won {
			sawWin = true
			if got := me.Hand.Size(); got != before+4 {
				t.Errorf("won flip: hand %d -> %d, want +4 (Divination and its copy)", before, got)
			}
			continue
		}
		sawLoss = true
		if !me.Hand.Contains(div) {
			t.Error("lost flip: the spell should be back in its owner's hand")
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("64 games never showed both outcomes (win=%v loss=%v)", sawWin, sawLoss)
	}
}

// A bare FlipACoin (no continuation) is keyed under the engine's own
// coin-flip/nothing, so its call is a restore point too.
func TestABareCoinFlipIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.FlipCoinForEffect(game.CoinFlipSpec{Flipper: me.ID, Question: "flip"})
	})
	g = effectstest.RestoreRoundTrip(t, g, true)
	flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if flip == nil {
		t.Fatal("the restored table has no coin call")
	}
	if err := g.ResolveCoinCall(flip.ID, g.Seats[0].ID, "tails"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	if n := len(effectstest.RandomEvents(g, game.EventFlipCoin)); n != 1 {
		t.Errorf("%d flips logged, want 1", n)
	}
}

// Windshaper Planetar: each reselect question is keyed
// (option-pick/reselect-attack), the attackers still to ask about in its
// Carry, so the table is a restore point at the first question and at the
// second.
func TestWindshaperPlanetarsQuestionsAreRestorePoints(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attackerSeat, me, second, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	a := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger A", 2, 2)
	b := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger B", 2, 2)
	effectstest.AttackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{a: me.ID, b: me.ID})

	effectstest.CastFlashPermanentInPlace(t, g, me.ID, "Windshaper Planetar", "Creature — Angel", effectstest.WindshaperOracle)
	if effectstest.SettleUntilOptionPick(t, g, me.ID, uuid.Nil) == nil {
		t.Fatalf("no first reselect prompt")
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	first := effectstest.LatestOptionPickFor(g, me.ID)
	if first == nil {
		t.Fatal("the restored table has no reselect prompt")
	}
	if err := g.ResolveOptionPick(first.ID, me.ID, effectstest.OptionFor(t, first, second.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	next := effectstest.LatestOptionPickFor(g, me.ID)
	if next == nil {
		t.Fatalf("no second prompt for the second attacker")
	}
	if err := g.ResolveOptionPick(next.ID, me.ID, effectstest.OptionFor(t, next, third.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Errorf("a third prompt opened for two attackers")
	}
	ca, _ := effectstest.AangCardOnBF(g, a)
	cb, _ := effectstest.AangCardOnBF(g, b)
	got := map[uuid.UUID]bool{ca.AttackingTarget: true, cb.AttackingTarget: true}
	if !got[second.ID] || !got[third.ID] {
		t.Errorf("attacks ended on %v and %v, want seats 2 and 3 one each", ca.AttackingTarget, cb.AttackingTarget)
	}
}

// Starving Revenant: the surveil's "then" is a keyed tail told how many cards
// went back on top, so the surveil prompt is a restore point and the restored
// answer draws and loses life for each card kept.
func TestStarvingRevenantsSurveilIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.S58Library(me, "Deep", "Next", "Top")
	effectstest.S58CastRevenant(t, g, me)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("no surveil prompt")
	}
	hand, life := me.Hand.Size(), me.Life

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the restored table has no surveil prompt")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards[:1], c.ScryCards[1:]); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("cards drawn: %d, want 1", got)
	}
	if me.Life != life-3 {
		t.Errorf("life %d→%d, want a loss of 3", life, me.Life)
	}
}

// Talara's Bane under Rhox Faithmender and Alhammarret's Archive: the gain
// asks for the order, and the discard is the gain's keyed tail
// (RevealedPick.DoneTail), so the order prompt is a restore point and the
// restored answer still discards the card, after the gain.
func TestTalarasBaneIsARestorePointAtTheLifeGainOrderPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, victim := g.Seats[0], g.Seats[1]
	faith := effectstest.SeedReplacementPermanent(g, effectstest.RhoxFaithmenderOracle, "Rhox Faithmender", caster.ID)
	archive := effectstest.SeedReplacementPermanent(g, effectstest.AlhammarretsArchiveOracle, "Alhammarret's Archive", caster.ID)
	ids := effectstest.RevealHand(victim, effectstest.RhForest(), effectstest.PbBear())
	life := caster.Life

	effectstest.CastAt(t, g, "Talara's Bane", "Sorcery", effectstest.TalarasBaneOracle, effectstest.PbPlayer(victim.ID))
	pick := effectstest.OpenVariant(t, g)
	if err := g.ResolvePendingChoice(pick.ID, caster.ID, ids[1:]); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !effectstest.OrderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeLifeTail)
	caster, victim = g.Seats[0], g.Seats[1]
	if victim.Graveyard.Contains(ids[1]) {
		t.Fatal("the restored table discarded before the gain")
	}
	effectstest.AnswerOrderBySource(t, g, archive, faith)
	if got := caster.Life - life; got != 8 {
		t.Errorf("the caster gained %d, want 8 (2, doubled twice)", got)
	}
	if !victim.Graveyard.Contains(ids[1]) {
		t.Error("the restored gain's tail did not discard the chosen card")
	}
}
