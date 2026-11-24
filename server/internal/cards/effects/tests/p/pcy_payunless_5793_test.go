package testsp

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pcy_payunless_5793_test.go — the six Prophecy cards #5793 unblocked
// (#5617): Rhystic Lightning, Rhystic Syphon, Rhystic Deluge, Excise,
// Withdraw and Flay. Each is played through the engine: the opponent pays,
// the opponent declines, and nobody gets priority before the answer.

const (
	pcyRhysticLightning = "f6d371e6-e25e-4820-8cf7-f692a4d9be11"
	pcyRhysticSyphon    = "b1eea34a-74ef-4d7c-bf70-c64b3487f296"
	pcyRhysticDeluge    = "d59cfff9-2bce-439a-9456-ae68eabbf67d"
	pcyExcise           = "28e1a613-30ad-4013-be0b-b0365f627201"
	pcyWithdraw         = "626882c5-15a8-43ea-ab08-33a7e949c88c"
	pcyFlay             = "987b74f4-02cb-4d47-94e5-d58a3c8e92e4"
)

func pcyTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	return g, g.Seats[0], g.Seats[1]
}

func pcyOnBattlefield(g *game.Game, id uuid.UUID) bool { return g.Battlefield.Contains(id) }

// pcyHolds asserts the question is open and holds the table: the spell (or
// ability) has not finished, a pass and an advance are refused, and the
// step does not move.
func pcyHolds(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	if spell != uuid.Nil && !g.Stack.Contains(spell) {
		t.Error("the spell left the stack with its question open")
	}
	if !g.ResolutionPaused() {
		t.Error("the resolution is not paused on its question")
	}
	step := g.Turn.Step
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("a pass with the question open = %v, want ErrChoicePending", err)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("advance_step with the question open = %v, want ErrChoicePending", err)
	}
	if g.Turn.Step != step {
		t.Fatalf("the step moved from %v to %v with the question open", step, g.Turn.Step)
	}
}

func pcyAnswer(t *testing.T, g *game.Game, who uuid.UUID, pay bool) {
	t.Helper()
	effectstest.AnswerPayUnless(t, g, who, pay)
	g.SettleResolution()
	if effectstest.HasPayUnlessFor(g, who) {
		t.Fatal("the question is still open after the answer")
	}
}

// Rhystic Lightning -----------------------------------------------------

func pcyLightning(t *testing.T, target func(g *game.Game, opp *game.Player) game.TargetRef) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := pcyTable(t)
	spell := effectstest.CastCatalogSpell(t, g, "Rhystic Lightning", "Instant", pcyRhysticLightning,
		[]game.TargetRef{target(g, opp)})
	effectstest.PassUntilTaxed(t, g, opp.ID)
	return g, me, opp, spell
}

func pcyPlayerTarget(_ *game.Game, opp *game.Player) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
}

func TestRhysticLightningDeclinedDeals4ToAPlayer(t *testing.T) {
	g, _, opp, spell := pcyLightning(t, pcyPlayerTarget)
	before := opp.Life
	if opp.Life != before {
		t.Fatal("damage before the answer")
	}
	pcyHolds(t, g, spell)
	pcyAnswer(t, g, opp.ID, false)
	if opp.Life != before-4 {
		t.Errorf("life = %d, want %d", opp.Life, before-4)
	}
	if g.Stack.Contains(spell) {
		t.Error("the spell is still on the stack")
	}
}

func TestRhysticLightningPaidDeals2ToAPlayer(t *testing.T) {
	g, _, opp, _ := pcyLightning(t, pcyPlayerTarget)
	before := opp.Life
	effectstest.B06AddMana(opp, "C", "C")
	pcyAnswer(t, g, opp.ID, true)
	if opp.Life != before-2 {
		t.Errorf("life = %d, want %d", opp.Life, before-2)
	}
}

func TestRhysticLightningPayingWithoutTheManaIsADecline(t *testing.T) {
	g, _, opp, _ := pcyLightning(t, pcyPlayerTarget)
	before := opp.Life
	pcyAnswer(t, g, opp.ID, true) // "Pay" with an empty pool
	if opp.Life != before-4 {
		t.Errorf("life = %d, want %d", opp.Life, before-4)
	}
}

func TestRhysticLightningAsksAPermanentsController(t *testing.T) {
	var bear uuid.UUID
	g, _, opp, spell := pcyLightning(t, func(g *game.Game, opp *game.Player) game.TargetRef {
		bear = effectstest.PushCatalogPermanent(g, opp.ID, "Ogre", "Creature — Ogre", "", false)
		g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Toughness = 3
		return game.TargetRef{Kind: game.TargetCard, ID: bear}
	})
	pcyHolds(t, g, spell)
	pcyAnswer(t, g, opp.ID, false)
	if pcyOnBattlefield(g, bear) {
		t.Error("4 damage should have killed the 3-toughness creature")
	}
}

func TestRhysticLightningPaidDeals2ToAPermanent(t *testing.T) {
	var bear uuid.UUID
	g, _, opp, _ := pcyLightning(t, func(g *game.Game, opp *game.Player) game.TargetRef {
		bear = effectstest.PushCatalogPermanent(g, opp.ID, "Ogre", "Creature — Ogre", "", false)
		g.Battlefield.Cards[len(g.Battlefield.Cards)-1].Toughness = 3
		return game.TargetRef{Kind: game.TargetCard, ID: bear}
	})
	effectstest.B06AddMana(opp, "C", "C")
	pcyAnswer(t, g, opp.ID, true)
	if !pcyOnBattlefield(g, bear) {
		t.Error("2 damage should not have killed the 3-toughness creature")
	}
}

// Rhystic Syphon --------------------------------------------------------

func pcySyphon(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := pcyTable(t)
	spell := effectstest.CastCatalogSpell(t, g, "Rhystic Syphon", "Sorcery", pcyRhysticSyphon, []game.TargetRef{pcyPlayerTarget(g, opp)})
	effectstest.PassUntilTaxed(t, g, opp.ID)
	return g, me, opp, spell
}

func TestRhysticSyphonDeclinedDrains5(t *testing.T) {
	g, me, opp, spell := pcySyphon(t)
	meBefore, oppBefore := me.Life, opp.Life
	pcyHolds(t, g, spell)
	if me.Life != meBefore || opp.Life != oppBefore {
		t.Fatal("life changed before the answer")
	}
	pcyAnswer(t, g, opp.ID, false)
	if opp.Life != oppBefore-5 || me.Life != meBefore+5 {
		t.Errorf("life me=%d opp=%d, want %d and %d", me.Life, opp.Life, meBefore+5, oppBefore-5)
	}
	if g.Stack.Contains(spell) || !me.Graveyard.Contains(spell) {
		t.Error("the spell should have left the stack for the graveyard after the answer")
	}
}

func TestRhysticSyphonPaidChangesNothing(t *testing.T) {
	g, me, opp, _ := pcySyphon(t)
	meBefore, oppBefore := me.Life, opp.Life
	effectstest.B06AddMana(opp, "C", "C", "C")
	pcyAnswer(t, g, opp.ID, true)
	if opp.Life != oppBefore || me.Life != meBefore {
		t.Errorf("life me=%d opp=%d, want %d and %d", me.Life, opp.Life, meBefore, oppBefore)
	}
}

// Excise ----------------------------------------------------------------

func pcyExciseCast(t *testing.T, x int) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, opp := pcyTable(t)
	bear := effectstest.PushCatalogPermanent(g, opp.ID, "Bear", "Creature — Bear", "", false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == bear {
			g.Battlefield.Cards[i].AttackingTarget = me.ID
		}
	}
	c := game.NewCard("Excise", me.ID)
	c.TypeLine, c.OracleID = "Instant", pcyExcise
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Targets: effectstest.CardRefs(bear), XValue: x}); err != nil {
		t.Fatalf("cast Excise: %v", err)
	}
	return g, me, opp, c.InstanceID, bear
}

func TestExciseDeclinedExilesTheAttacker(t *testing.T) {
	g, _, opp, spell, bear := pcyExciseCast(t, 3)
	effectstest.PassUntilTaxed(t, g, opp.ID)
	pcyHolds(t, g, spell)
	if !pcyOnBattlefield(g, bear) {
		t.Fatal("the creature left before the answer")
	}
	pcyAnswer(t, g, opp.ID, false)
	if pcyOnBattlefield(g, bear) || !g.Exile.Contains(bear) {
		t.Error("declining: the creature should be exiled")
	}
}

func TestExcisePaidKeepsTheAttacker(t *testing.T) {
	g, _, opp, _, bear := pcyExciseCast(t, 3)
	effectstest.PassUntilTaxed(t, g, opp.ID)
	effectstest.B06AddMana(opp, "C", "C", "C")
	pcyAnswer(t, g, opp.ID, true)
	if !pcyOnBattlefield(g, bear) {
		t.Error("paying {3}: the creature should stay")
	}
}

func TestExcisePayingLessThanXIsADecline(t *testing.T) {
	g, _, opp, _, bear := pcyExciseCast(t, 3)
	effectstest.PassUntilTaxed(t, g, opp.ID)
	effectstest.B06AddMana(opp, "C", "C")
	pcyAnswer(t, g, opp.ID, true)
	if pcyOnBattlefield(g, bear) {
		t.Error("two of three mana: the creature should be exiled")
	}
}

func TestExciseForXZeroAsksNothing(t *testing.T) {
	g, _, opp, _, bear := pcyExciseCast(t, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("X = 0 asked for a payment of {0}")
	}
	if !pcyOnBattlefield(g, bear) {
		t.Error("X = 0: the creature is paid for and stays")
	}
}

// Flay ------------------------------------------------------------------

func pcyFlayCast(t *testing.T, hand int) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := pcyTable(t)
	for opp.Hand.Size() > 0 {
		opp.Hand.PopTop()
	}
	for i := 0; i < hand; i++ {
		opp.Hand.PushTop(game.NewCard("Filler", opp.ID))
	}
	spell := effectstest.CastCatalogSpell(t, g, "Flay", "Sorcery", pcyFlay, []game.TargetRef{pcyPlayerTarget(g, opp)})
	return g, me, opp, spell
}

func TestFlayDeclinedDiscardsTwo(t *testing.T) {
	g, _, opp, spell := pcyFlayCast(t, 4)
	effectstest.PassUntilTaxed(t, g, opp.ID)
	if got := opp.Hand.Size(); got != 3 {
		t.Fatalf("hand = %d with the question open, want 3 (the first discard is done)", got)
	}
	pcyHolds(t, g, spell)
	pcyAnswer(t, g, opp.ID, false)
	if got := opp.Hand.Size(); got != 2 {
		t.Errorf("hand = %d after declining, want 2", got)
	}
	if g.Stack.Contains(spell) {
		t.Error("the spell is still on the stack")
	}
}

func TestFlayPaidDiscardsOne(t *testing.T) {
	g, _, opp, _ := pcyFlayCast(t, 4)
	effectstest.PassUntilTaxed(t, g, opp.ID)
	effectstest.B06AddMana(opp, "C")
	pcyAnswer(t, g, opp.ID, true)
	if got := opp.Hand.Size(); got != 3 {
		t.Errorf("hand = %d after paying, want 3", got)
	}
}

func TestFlayAnEmptiedHandIsNotAsked(t *testing.T) {
	g, _, opp, _ := pcyFlayCast(t, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("a player with no cards left was asked to pay")
	}
	if opp.Hand.Size() != 0 {
		t.Errorf("hand = %d, want 0", opp.Hand.Size())
	}
}

// Withdraw --------------------------------------------------------------

func pcyWithdrawCast(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g, me, opp := pcyTable(t)
	first := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
	second := effectstest.PushCatalogPermanent(g, opp.ID, "Theirs", "Creature — Bear", "", false)
	spell := effectstest.CastCatalogSpell(t, g, "Withdraw", "Instant", pcyWithdraw, effectstest.CardRefs(first, second))
	effectstest.PassUntilTaxed(t, g, opp.ID)
	return g, me, opp, spell, first, second
}

func TestWithdrawDeclinedReturnsBoth(t *testing.T) {
	g, me, opp, spell, first, second := pcyWithdrawCast(t)
	if pcyOnBattlefield(g, first) || !me.Hand.Contains(first) {
		t.Error("the first creature is returned before the question")
	}
	pcyHolds(t, g, spell)
	if !pcyOnBattlefield(g, second) {
		t.Fatal("the second creature left before the answer")
	}
	pcyAnswer(t, g, opp.ID, false)
	if pcyOnBattlefield(g, second) || !opp.Hand.Contains(second) {
		t.Error("declining: the second creature should return to its owner's hand")
	}
}

func TestWithdrawPaidReturnsOnlyTheFirst(t *testing.T) {
	g, me, opp, _, first, second := pcyWithdrawCast(t)
	effectstest.B06AddMana(opp, "C")
	pcyAnswer(t, g, opp.ID, true)
	if !me.Hand.Contains(first) {
		t.Error("the first creature should be in hand")
	}
	if !pcyOnBattlefield(g, second) {
		t.Error("paying {1}: the second creature should stay")
	}
}

func TestWithdrawAsksTheSecondCreaturesController(t *testing.T) {
	g, me, opp := pcyTable(t)
	first := effectstest.PushCatalogPermanent(g, opp.ID, "Theirs", "Creature — Bear", "", false)
	second := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
	effectstest.CastCatalogSpell(t, g, "Withdraw", "Instant", pcyWithdraw, effectstest.CardRefs(first, second))
	effectstest.PassUntilTaxed(t, g, me.ID)
	if effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("the first creature's controller was asked")
	}
	pcyAnswer(t, g, me.ID, false)
	if !me.Hand.Contains(second) {
		t.Error("declining: the second creature returns")
	}
}

// Rhystic Deluge --------------------------------------------------------

func pcyDelugeActivate(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := pcyTable(t)
	deluge := effectstest.Mrd2Catalog(g, me.ID, "Rhystic Deluge", "Enchantment", pcyRhysticDeluge, "{2}{U}", 0, 0)
	bear := effectstest.SeedBear(g, opp.ID)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, deluge, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate Rhystic Deluge: %v", err)
	}
	effectstest.PassUntilTaxed(t, g, opp.ID)
	return g, me, opp, bear
}

func pcyTapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatal("creature gone")
	}
	return c.Tapped
}

func TestRhysticDelugeDeclinedTapsTheCreature(t *testing.T) {
	g, _, opp, bear := pcyDelugeActivate(t)
	pcyHolds(t, g, uuid.Nil)
	if pcyTapped(t, g, bear) {
		t.Fatal("tapped before the answer")
	}
	pcyAnswer(t, g, opp.ID, false)
	if !pcyTapped(t, g, bear) {
		t.Error("declining: the creature should be tapped")
	}
}

func TestRhysticDelugePaidLeavesItUntapped(t *testing.T) {
	g, _, opp, bear := pcyDelugeActivate(t)
	effectstest.B06AddMana(opp, "C")
	pcyAnswer(t, g, opp.ID, true)
	if pcyTapped(t, g, bear) {
		t.Error("paying {1}: the creature should stay untapped")
	}
}

// A restore with the question open ---------------------------------------

func TestWithdrawQuestionIsARestorePoint(t *testing.T) {
	g, _, _, _, _, second := pcyWithdrawCast(t)
	g = effectstest.RestoreRoundTrip(t, g, true)
	opp := g.Seats[1]
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("the restored table lost the question")
	}
	pcyAnswer(t, g, opp.ID, false)
	if pcyOnBattlefield(g, second) {
		t.Error("declining after a restore: the second creature should return")
	}
}
