package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b653_cards_test.go — Invasion, batch 653 (S82, #5616). Each test
// plays the real catalog entry through the engine.

const (
	oldB653HeraldsOracle   = "09ee1332-741f-4c55-abc3-8bcff9031cd9"
	oldB653WinnowOracle    = "1b2ad355-3f1e-4c8b-93ee-848934646b23"
	oldB653StrikeOracle    = "1f0b3549-2d24-49e8-85bc-55e3a5e5f6b8"
	oldB653BindOracle      = "2b9557df-7158-4ada-be77-5c346851a568"
	oldB653MerchantOracle  = "3d62c8f8-7a1c-4771-891d-05fdaa56bf47"
	oldB653LairOracle      = "4dd0594d-5b11-4350-a926-294a67d9fc3b"
	oldB653FateOracle      = "541f13aa-a1ed-453a-b719-82daef1be607"
	oldB653ZapOracle       = "56115482-3fd4-45fb-b800-be68c5509cf2"
	oldB653SanctuaryOracle = "6f98f270-f6ff-4029-bcdd-3e7f6483eb5d"
)

func TestOldB653EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		oldB653HeraldsOracle: "Benalish Heralds", oldB653WinnowOracle: "Winnow", oldB653StrikeOracle: "Chaotic Strike",
		oldB653BindOracle: "Bind", oldB653MerchantOracle: "Vodalian Merchant", oldB653LairOracle: "Kavu Lair",
		oldB653FateOracle: "Manipulate Fate", oldB653ZapOracle: "Zap", oldB653SanctuaryOracle: "Elfhame Sanctuary",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB653Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, mana)
	params.Strict = true
	id := effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, oracle, params)
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB653Seed(me *game.Player, n int) {
	for i := 0; i < n; i++ {
		effectstest.SeedLibraryTop(me, "Filler", "Sorcery", "{1}")
	}
}

func TestOldB653BenalishHeraldsDrawsForABlueTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB653Seed(me, 3)
	h := effectstest.PushCatalogPermanent(g, me.ID, "Benalish Heralds", "Creature — Human Soldier", oldB653HeraldsOracle, false)
	before := effectstest.HandCount(g, me)
	effectstest.FloatForTest(g, me, "CCCU")
	effectstest.ApaActivate(t, g, me, h, 0, game.ActivateAbilityParams{})
	if got := effectstest.HandCount(g, me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, h); !tapped {
		t.Error("the Heralds did not tap")
	}
	effectstest.FloatForTest(g, me, "CCCU")
	if err := g.ActivateCatalogAbility(me.ID, h, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated while tapped")
	}
}

func TestOldB653WinnowDestroysOnlyWithADuplicateName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB653Seed(me, 4)
	a := effectstest.B12Creature(g, opp.ID, "Twin", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Twin", "Creature — Bear", 2, 2)
	solo := effectstest.B12Creature(g, opp.ID, "Solo", "Creature — Bear", 2, 2)
	before := effectstest.HandCount(g, me)
	oldB653Cast(t, g, me, "Winnow", "Instant", "{1}{W}", oldB653WinnowOracle, "CW", game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}}})
	if _, ok := effectstest.BattlefieldCard(g, a); ok {
		t.Error("a creature with a twin survived Winnow")
	}
	if _, ok := effectstest.BattlefieldCard(g, b); !ok {
		t.Error("the other twin was destroyed")
	}
	if got := effectstest.HandCount(g, me); got != before+1 {
		t.Errorf("hand %d, want %d (a draw)", got, before+1)
	}
	oldB653Cast(t, g, me, "Winnow", "Instant", "{1}{W}", oldB653WinnowOracle, "CW", game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: solo}}})
	if _, ok := effectstest.BattlefieldCard(g, solo); !ok {
		t.Error("a permanent with no twin was destroyed")
	}
	if got := effectstest.HandCount(g, me); got != before+2 {
		t.Errorf("hand %d, want %d (the draw happens either way)", got, before+2)
	}
}

func TestOldB653WinnowRefusesALandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CW")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Winnow", TypeLine: "Instant", ManaCost: "{1}{W}", OracleID: oldB653WinnowOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Error("Winnow targeted a land")
	}
}

func TestOldB653ChaoticStrikeTimingAndFlip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CR")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Chaotic Strike", TypeLine: "Instant", ManaCost: "{1}{R}", OracleID: oldB653StrikeOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Fatal("cast in a main phase")
	}
	won, lost := 0, 0
	for table := 0; table < 40 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		oldB653Seed(me, 3)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(bear, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		effectstest.FinishBlockDeclarations(t, g)
		before := effectstest.HandCount(g, me)
		effectstest.FloatForTest(g, me, "CR")
		effectstest.CastFromHandForTest(t, g, me, "Chaotic Strike", "Instant", "{1}{R}", oldB653StrikeOracle, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RandomEvents(g, game.EventFlipCoin)
		if len(evs) != flips+1 {
			t.Fatalf("%d coins flipped, want 1", len(evs)-flips)
		}
		if got := effectstest.HandCount(g, me); got != before+1 {
			t.Errorf("hand %d, want %d: the draw happens win or lose", got, before+1)
		}
		power := effectstest.EffectivePower(t, g, bear)
		if evs[len(evs)-1].Won {
			won++
			if power != 3 {
				t.Errorf("won the flip: power %d, want 3", power)
			}
		} else {
			lost++
			if power != 2 {
				t.Errorf("lost the flip: power %d, want 2", power)
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("never saw both outcomes (won %d, lost %d)", won, lost)
	}
}

func TestOldB653BindCountersAnActivatedAbilityAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB653Seed(me, 3)
	effectstest.AdvanceToMain(t, g)
	// An opposing Benalish Heralds activates; the active player answers with Bind.
	h := effectstest.PushCatalogPermanent(g, g.Seats[g.Turn.ActiveSeat].ID, "Benalish Heralds", "Creature — Human Soldier", oldB653HeraldsOracle, false)
	owner := g.Seats[g.Turn.ActiveSeat]
	effectstest.FloatForTest(g, owner, "CCCU")
	if err := g.ActivateCatalogAbility(owner.ID, h, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	var abil uuid.UUID
	for id, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemActivated {
			abil = id
		}
	}
	if abil == uuid.Nil {
		t.Fatal("no activated ability on the stack")
	}
	_ = opp
	caster := me
	if owner.ID == me.ID {
		caster = g.Seats[1]
	}
	oldB653Seed(caster, 3)
	before, ownerBefore := effectstest.HandCount(g, caster), effectstest.HandCount(g, owner)
	effectstest.FloatForTest(g, caster, "CG")
	effectstest.CastFromHandForTest(t, g, caster, "Bind", "Instant", "{1}{G}", oldB653BindOracle, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: abil}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, owner); got != ownerBefore {
		t.Errorf("the countered ability still drew: hand %d, want %d", got, ownerBefore)
	}
	if got := effectstest.HandCount(g, caster); got != before+1 {
		t.Errorf("Bind's controller hand %d, want %d (the cast nets zero, one draw)", got, before+1)
	}
}

func TestOldB653VodalianMerchantLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB653Seed(me, 3)
	before := effectstest.HandCount(g, me)
	effectstest.CastHoldCreature(t, g, "Vodalian Merchant", "Creature — Merfolk", oldB653MerchantOracle, 1, 2)
	// The Merchant left the hand, a card was drawn and one must be discarded.
	effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HandCount(g, me); got != before {
		t.Errorf("hand %d, want %d (cast one, drew one, discarded one)", got, before)
	}
}

func TestOldB653KavuLairDrawsForTheEntrantsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB653Seed(me, 3)
	oldB653Seed(opp, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Kavu Lair", "Enchantment", oldB653LairOracle, false)
	for _, tc := range []struct {
		controller *game.Player
		power      int
		draws      bool
	}{{opp, 4, true}, {opp, 3, false}, {me, 5, true}} {
		before := effectstest.HandCount(g, tc.controller)
		beforeOther := effectstest.HandCount(g, g.Seats[0])
		effectstest.EnterCreature(t, g, tc.controller.ID, "Entrant", tc.power, tc.power)
		effectstest.PassPriorityAroundTable(t, g)
		want := before
		if tc.draws {
			want++
		}
		if got := effectstest.HandCount(g, tc.controller); got != want {
			t.Errorf("power %d for %s: hand %d, want %d", tc.power, tc.controller.Name, got, want)
		}
		_ = beforeOther
	}
}

func TestOldB653ManipulateFateExilesThreeAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB653Seed(me, 6)
	lib := me.Library.Size()
	hand := effectstest.HandCount(g, me)
	oldB653Cast(t, g, me, "Manipulate Fate", "Sorcery", "{1}{U}", oldB653FateOracle, "CU", game.CastSpellParams{})
	var picks []uuid.UUID
	for _, c := range me.Library.Cards[:3] {
		picks = append(picks, c.InstanceID)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, picks...)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != lib-4 {
		t.Errorf("library %d, want %d (three exiled, one drawn)", got, lib-4)
	}
	if got := effectstest.HandCount(g, me); got != hand+1 {
		t.Errorf("hand %d, want %d (the spell left, one drawn)", got, hand+1)
	}
}

func TestOldB653ZapDamagesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB653Seed(me, 3)
	before := effectstest.HandCount(g, me)
	life := opp.Life
	oldB653Cast(t, g, me, "Zap", "Instant", "{2}{R}", oldB653ZapOracle, "CCR", game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if got := effectstest.HandCount(g, me); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestOldB653ElfhameSanctuarySearchesAndSkipsTheDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMainOf(t, g, 0)
	oldB653Seed(me, 4)
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.PushCatalogPermanent(g, me.ID, "Elfhame Sanctuary", "Enchantment", oldB653SanctuaryOracle, false)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.TfSettle(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.TfSettle(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.PassPriorityAroundTable(t, g)
	before := effectstest.HandCount(g, me)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	if effectstest.ZoneOf(g, forest) != game.ZoneHand {
		t.Errorf("the Forest is in %v, want the hand", effectstest.ZoneOf(g, forest))
	}
	if got := effectstest.HandCount(g, me); got != before {
		t.Errorf("hand %d, want %d: the draw step is skipped", got, before)
	}
}
