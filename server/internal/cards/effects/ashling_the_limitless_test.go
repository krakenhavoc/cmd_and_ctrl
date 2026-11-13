package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ashling_the_limitless_test.go — Ashling, the Limitless and the granted
// evoke seam (#5428).

const (
	ashlingLimitlessOracle   = "77a9e7cf-76e9-4323-8504-a5b831de00be"
	ashlingLimitlessTypeLine = "Legendary Creature — Elemental Sorcerer"
	ashlingElemName          = "Ashling Test Elemental"
)

func TestAshlingIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(ashlingLimitlessOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

// ashlingOnBattlefield walks the active seat to its main phase with an
// Ashling it controls.
func ashlingOnBattlefield(t *testing.T, g *game.Game) (me *game.Player, ashling uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ashling = pushCatalogPermanent(g, me.ID, "Ashling, the Limitless", ashlingLimitlessTypeLine, ashlingLimitlessOracle, false)
	return me, ashling
}

// ashlingCardInHand puts a card costing {5}{R} with this type line in
// `p`'s hand.
func ashlingCardInHand(g *game.Game, p *game.Player, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: ashlingElemName, TypeLine: typeLine, ManaCost: "{5}{R}",
			Power: 5, Toughness: 5, Colors: []string{"R"}, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// ashlingLimitlessSettle passes priority until the stack is empty, answering a
// trigger-order prompt in the order offered.
func ashlingLimitlessSettle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 24; i++ {
		if triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
}

// ashlingSacrifice has its controller sacrifice the permanent.
func ashlingSacrifice(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.SacrificePermanentForEffect(id); err != nil {
			t.Fatalf("sacrifice: %v", err)
		}
	})
}

// ashlingEndStep walks to the end step and resolves what triggers there,
// stopping at a pay-unless prompt for `me`.
func ashlingEndStep(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	advanceTo(t, g, game.StepEnd)
	for i := 0; i < 12; i++ {
		if hasPayUnlessFor(g, me.ID) || stackFullyEmpty(g) {
			return
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the end step did not settle")
}

// ashlingTriggersWaiting counts Ashling's triggers on the stack or waiting
// to be put there.
func ashlingTriggersWaiting(g *game.Game, ashling uuid.UUID) int {
	n := triggersOnStackFrom(g, ashling)
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == ashling {
			n++
		}
	}
	return n
}

func ashlingTokens(g *game.Game, name string) []game.Card {
	tokens, _ := copiesOnBattlefield(g, name)
	return tokens
}

// The whole line: an Elemental is cast from hand for the granted evoke
// {4}, is sacrificed as it enters, and Ashling copies it; the token has
// haste this turn and is sacrificed at the end step when {W}{U}{B}{R}{G}
// is not paid.
func TestAshlingGivesAnElementalEvokeAndCopiesItWhenItIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ashlingOnBattlefield(t, g)
	elem := ashlingCardInHand(g, me, "Creature — Elemental")
	floatMana(t, g, me, "{R}{R}{R}{R}")
	if err := g.CastSpell(me.ID, elem, game.CastSpellParams{AlternativeCost: GrantedAltCostEvoke, Strict: true}); err != nil {
		t.Fatalf("cast for the granted evoke: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want 0: evoke {4} is paid rather than {5}{R}", n)
	}
	ashlingLimitlessSettle(t, g)
	entered := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventETB && ev.CardID == elem {
			entered = true
		}
	}
	if !entered {
		t.Fatal("the evoked Elemental did not enter the battlefield")
	}
	if !me.Graveyard.Contains(elem) {
		t.Fatal("the evoked Elemental was not sacrificed")
	}
	toks := ashlingTokens(g, ashlingElemName)
	if len(toks) != 1 {
		t.Fatalf("%d token copies, want 1", len(toks))
	}
	tok := toks[0]
	if e := tok.Effective(); e.Power != 5 || e.Toughness != 5 || tok.Controller != me.ID {
		t.Errorf("the token is %d/%d under %s, want a 5/5 copy under its creator", e.Power, e.Toughness, tok.Controller)
	}
	if !hasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
		t.Error("the token has no haste this turn")
	}

	ashlingEndStep(t, g, me)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay-or-sacrifice prompt at the end step")
	}
	answerPayUnless(t, g, me.ID, false)
	ashlingLimitlessSettle(t, g)
	if n := len(ashlingTokens(g, ashlingElemName)); n != 0 {
		t.Errorf("%d tokens after declining to pay, want 0", n)
	}
}

// Paying {W}{U}{B}{R}{G} keeps the token, and the haste was only until
// end of turn.
func TestAshlingTokenStaysWhenWUBRGIsPaid(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ashlingOnBattlefield(t, g)
	elem := pushDiesCreatureForTest(g, me.ID, ashlingElemName, "", "Creature — Elemental", 5, 5)
	ashlingSacrifice(t, g, elem)
	ashlingLimitlessSettle(t, g)
	toks := ashlingTokens(g, ashlingElemName)
	if len(toks) != 1 {
		t.Fatalf("%d token copies, want 1", len(toks))
	}
	tok := toks[0].InstanceID
	ashlingEndStep(t, g, me)
	floatMana(t, g, me, "{W}{U}{B}{R}{G}")
	answerPayUnless(t, g, me.ID, true)
	ashlingLimitlessSettle(t, g)
	if !g.Battlefield.Contains(tok) {
		t.Fatal("the token was sacrificed though {W}{U}{B}{R}{G} was paid")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want 0", n)
	}
	// Next turn: the haste has ended and nothing asks again.
	advanceTo(t, g, game.StepUpkeep)
	if hasEffectiveKeyword(t, g, tok, "haste") {
		t.Error("the token still has haste on a later turn")
	}
	if !g.Battlefield.Contains(tok) {
		t.Error("the token left on a later turn")
	}
}

// A token that has already left asks nothing at the end step.
func TestAshlingAsksNothingForATokenThatLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ashlingOnBattlefield(t, g)
	elem := pushDiesCreatureForTest(g, me.ID, ashlingElemName, "", "Creature — Elemental", 5, 5)
	ashlingSacrifice(t, g, elem)
	ashlingLimitlessSettle(t, g)
	toks := ashlingTokens(g, ashlingElemName)
	if len(toks) != 1 {
		t.Fatalf("%d token copies, want 1", len(toks))
	}
	// Sacrificing the token is not sacrificing a NONTOKEN Elemental.
	ashlingSacrifice(t, g, toks[0].InstanceID)
	ashlingLimitlessSettle(t, g)
	if n := len(ashlingTokens(g, ashlingElemName)); n != 0 {
		t.Fatalf("%d tokens after the token was sacrificed, want 0", n)
	}
	ashlingEndStep(t, g, me)
	ashlingLimitlessSettle(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a pay-or-sacrifice prompt for a token that is gone")
	}
}

func TestAshlingTriggersOnlyForANontokenElementalYouSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, ashling := ashlingOnBattlefield(t, g)
	opp := g.Seats[1]
	bear := pushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	theirs := pushDiesCreatureForTest(g, opp.ID, "Their Elemental", "", "Creature — Elemental", 3, 3)
	ashlingSacrifice(t, g, bear)
	ashlingSacrifice(t, g, theirs)
	if n := ashlingTriggersWaiting(g, ashling); n != 0 {
		t.Fatalf("%d Ashling triggers for a Bear and an opponent's Elemental, want 0", n)
	}
	// A noncreature Elemental counts: the text says "Elemental".
	kindred := pushDiesCreatureForTest(g, me.ID, "Kindred Thing", "", "Kindred Enchantment — Elemental", 0, 0)
	ashlingSacrifice(t, g, kindred)
	if n := ashlingTriggersWaiting(g, ashling); n != 1 {
		t.Fatalf("%d Ashling triggers for a kindred Elemental enchantment, want 1", n)
	}
	ashlingLimitlessSettle(t, g)
	if n := len(ashlingTokens(g, "Kindred Thing")); n != 1 {
		t.Errorf("%d token copies of the enchantment, want 1", n)
	}
}

// Ashling is an Elemental: sacrificed, she sees it and is copied as she
// last existed.
func TestAshlingCopiesHerselfWhenSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, ashling := ashlingOnBattlefield(t, g)
	ashlingSacrifice(t, g, ashling)
	ashlingLimitlessSettle(t, g)
	if !me.Graveyard.Contains(ashling) {
		t.Fatal("Ashling was not sacrificed")
	}
	toks := ashlingTokens(g, "Ashling, the Limitless")
	if len(toks) != 1 {
		t.Fatalf("%d token copies of Ashling, want 1", len(toks))
	}
	if toks[0].OracleID != ashlingLimitlessOracle {
		t.Error("the token is not a copy of Ashling")
	}
}

func TestAshlingEvokeIsRefusedWhereTheTextDoesNotReach(t *testing.T) {
	evoke := game.CastSpellParams{AlternativeCost: GrantedAltCostEvoke, Strict: true}
	t.Run("no Ashling", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		advanceToMain(t, g)
		elem := ashlingCardInHand(g, me, "Creature — Elemental")
		floatMana(t, g, me, "{R}{R}{R}{R}")
		if err := g.CastSpell(me.ID, elem, evoke); err == nil {
			t.Error("evoke with no Ashling")
		}
		// An opponent's Ashling gives this player's spells nothing.
		pushCatalogPermanent(g, g.Seats[1].ID, "Ashling, the Limitless", ashlingLimitlessTypeLine, ashlingLimitlessOracle, false)
		if err := g.CastSpell(me.ID, elem, evoke); err == nil {
			t.Error("evoke under an opponent's Ashling")
		}
	})
	t.Run("not an Elemental", func(t *testing.T) {
		g := newCatalogGame(t)
		me, _ := ashlingOnBattlefield(t, g)
		bear := ashlingCardInHand(g, me, "Creature — Bear")
		floatMana(t, g, me, "{R}{R}{R}{R}")
		if err := g.CastSpell(me.ID, bear, evoke); err == nil {
			t.Error("a Bear spell was given evoke")
		}
	})
	t.Run("not a permanent spell", func(t *testing.T) {
		g := newCatalogGame(t)
		me, _ := ashlingOnBattlefield(t, g)
		spell := ashlingCardInHand(g, me, "Kindred Instant — Elemental")
		floatMana(t, g, me, "{R}{R}{R}{R}")
		if err := g.CastSpell(me.ID, spell, evoke); err == nil {
			t.Error("an Elemental instant spell was given evoke")
		}
	})
	t.Run("a pool that cannot pay", func(t *testing.T) {
		g := newCatalogGame(t)
		me, _ := ashlingOnBattlefield(t, g)
		elem := ashlingCardInHand(g, me, "Creature — Elemental")
		floatMana(t, g, me, "{R}{R}{R}")
		if err := g.CastSpell(me.ID, elem, evoke); err == nil {
			t.Error("evoke {4} was paid with three mana")
		}
	})
	t.Run("not from the hand", func(t *testing.T) {
		g := newCatalogGame(t)
		me, _ := ashlingOnBattlefield(t, g)
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: ashlingElemName, TypeLine: "Creature — Elemental",
				ManaCost: "{5}{R}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
		})
		floatMana(t, g, me, "{R}{R}{R}{R}")
		if err := g.CastSpell(me.ID, id, evoke); err == nil {
			t.Error("evoke from the graveyard")
		}
	})
}

// A kindred Elemental enchantment is an Elemental permanent spell: it may
// be evoked, and is sacrificed as it enters.
func TestAshlingEvokesANoncreatureElementalPermanentSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ashlingOnBattlefield(t, g)
	ench := ashlingCardInHand(g, me, "Kindred Enchantment — Elemental")
	floatMana(t, g, me, "{R}{R}{R}{R}")
	if err := g.CastSpell(me.ID, ench, game.CastSpellParams{AlternativeCost: GrantedAltCostEvoke, Strict: true}); err != nil {
		t.Fatalf("cast for the granted evoke: %v", err)
	}
	ashlingLimitlessSettle(t, g)
	if !me.Graveyard.Contains(ench) {
		t.Error("the evoked enchantment was not sacrificed")
	}
	if n := len(ashlingTokens(g, ashlingElemName)); n != 1 {
		t.Errorf("%d token copies, want 1", n)
	}
}

// Cast for its mana cost, the Elemental stays: only the evoke cost
// carries the sacrifice.
func TestAshlingElementalCastForItsManaCostStays(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ashlingOnBattlefield(t, g)
	elem := ashlingCardInHand(g, me, "Creature — Elemental")
	floatMana(t, g, me, "{R}{R}{R}{R}{R}{R}")
	if err := g.CastSpell(me.ID, elem, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	ashlingLimitlessSettle(t, g)
	if !g.Battlefield.Contains(elem) {
		t.Error("the hard-cast Elemental did not stay")
	}
}

// The spell gained evoke as it was cast (CR 610.5): it is sacrificed
// though Ashling left before it resolved. With Ashling gone, nothing
// copies it.
func TestAshlingEvokedSpellIsSacrificedAfterAshlingLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, ashling := ashlingOnBattlefield(t, g)
	elem := ashlingCardInHand(g, me, "Creature — Elemental")
	if err := g.CastSpell(me.ID, elem, game.CastSpellParams{AlternativeCost: GrantedAltCostEvoke}); err != nil {
		t.Fatalf("cast for the granted evoke: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}, ashling); err != nil {
		t.Fatalf("exile Ashling: %v", err)
	}
	ashlingLimitlessSettle(t, g)
	if !me.Graveyard.Contains(elem) {
		t.Error("the evoked Elemental was not sacrificed after Ashling left")
	}
	if n := len(ashlingTokens(g, ashlingElemName)); n != 0 {
		t.Errorf("%d token copies with no Ashling, want 0", n)
	}
}

// An Elemental that prints evoke keeps its own cost beside Ashling's, and
// either is evoke: Mulldrifter cast for the granted {4} draws two, is
// sacrificed, and its token copy draws two more.
func TestAshlingEvokeBesideAPrintedEvoke(t *testing.T) {
	const mulldrifter = "24d0f5e7-0d9e-4b76-900e-a7274e80312d"
	if !PaidEvoke(GrantedAltCostEvoke) || !PaidEvoke(AltCostKeyEvoke) || PaidEvoke("") {
		t.Fatal("PaidEvoke disagrees about the evoke keys")
	}
	g := newCatalogGame(t)
	me, _ := ashlingOnBattlefield(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mulldrifter", TypeLine: "Creature — Elemental", ManaCost: "{4}{U}",
			OracleID: mulldrifter, Power: 2, Toughness: 2, Colors: []string{"U"}, Owner: me.ID, Controller: me.ID})
	})
	offered := map[string]bool{}
	g.WithWriteLock(func() {
		card, _ := g.LookupCardForEffect(id)
		for _, ac := range g.CastOffersForLocked(me.ID, card, game.ZoneHand, nil) {
			if ac != nil {
				offered[ac.Key] = true
			}
		}
	})
	if !offered[AltCostKeyEvoke] || !offered[GrantedAltCostEvoke] {
		t.Fatalf("offers = %v, want the printed evoke and the granted one", offered)
	}
	hand := me.Hand.Size()
	floatMana(t, g, me, "{R}{R}{R}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: GrantedAltCostEvoke, Strict: true}); err != nil {
		t.Fatalf("cast for the granted evoke: %v", err)
	}
	ashlingLimitlessSettle(t, g)
	if !me.Graveyard.Contains(id) {
		t.Fatal("the evoked Mulldrifter was not sacrificed")
	}
	if n := len(ashlingTokens(g, "Mulldrifter")); n != 1 {
		t.Fatalf("%d Mulldrifter tokens, want 1", n)
	}
	// Minus the Mulldrifter cast, plus two for it and two for the token.
	if got := me.Hand.Size(); got != hand-1+4 {
		t.Errorf("hand is %d, want %d: two cards for Mulldrifter and two for its token copy", got, hand-1+4)
	}
}

// CR 707.10: a copy of a spell cast for the granted evoke cost copied
// that cost, so the token it becomes is sacrificed too. That token is a
// token, so only the original is copied by Ashling.
func TestACopyOfASpellCastForGrantedEvokeIsSacrificedToo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ashlingOnBattlefield(t, g)
	elem := ashlingCardInHand(g, me, "Creature — Elemental")
	if err := g.CastSpell(me.ID, elem, game.CastSpellParams{AlternativeCost: GrantedAltCostEvoke}); err != nil {
		t.Fatalf("cast for the granted evoke: %v", err)
	}
	copyTheSpell(t, g, me, elem)
	// The copy resolves first and its token enters.
	for i := 0; i < 8 && len(ashlingTokens(g, ashlingElemName)) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	toks := ashlingTokens(g, ashlingElemName)
	if len(toks) != 1 {
		t.Fatalf("%d copy tokens after the copy resolved, want 1", len(toks))
	}
	spellCopy := toks[0].InstanceID
	ashlingLimitlessSettle(t, g)
	if g.Battlefield.Contains(spellCopy) {
		t.Error("the copy's token stayed: its evoke cost was paid (CR 702.74a)")
	}
	if !me.Graveyard.Contains(elem) {
		t.Error("the evoked original was not sacrificed")
	}
	if n := len(ashlingTokens(g, ashlingElemName)); n != 1 {
		t.Errorf("%d tokens at the end, want 1: Ashling's copy of the nontoken original", n)
	}
}

// --- #5803: the end-step question holds the end step ------------------

// ashlingQuestionOpen makes a token copy, walks to its controller's end
// step and resolves the delayed trigger up to its question. It returns
// the table with the question open, the token and the turn it was asked
// in.
func ashlingQuestionOpen(t *testing.T) (g *game.Game, tok uuid.UUID, turn int) {
	t.Helper()
	g = newCatalogGame(t)
	me, _ := ashlingOnBattlefield(t, g)
	elem := pushDiesCreatureForTest(g, me.ID, ashlingElemName, "", "Creature — Elemental", 5, 5)
	ashlingSacrifice(t, g, elem)
	ashlingLimitlessSettle(t, g)
	toks := ashlingTokens(g, ashlingElemName)
	if len(toks) != 1 {
		t.Fatalf("%d token copies, want 1", len(toks))
	}
	ashlingEndStep(t, g, me)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay-or-sacrifice prompt at the end step")
	}
	if g.Turn.Step != game.StepEnd {
		t.Fatalf("the question was asked in %v, want the end step", g.Turn.Step)
	}
	return g, toks[0].InstanceID, g.Turn.Seq
}

// ashlingQuestionFor returns the open pay-or-sacrifice prompt for p.
func ashlingQuestionFor(t *testing.T, g *game.Game, p uuid.UUID) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == p {
			return c
		}
	}
	t.Fatal("no pay-or-sacrifice prompt")
	return nil
}

// assertAshlingHoldsTheEndStep checks that nothing moves the table out of
// the end step while the question is open.
func assertAshlingHoldsTheEndStep(t *testing.T, g *game.Game, turn int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	if c := ashlingQuestionFor(t, g, me.ID); !g.ChoicePromptBlocksTable(c) {
		t.Error("the question does not block the table")
	}
	if !g.ResolutionPaused() {
		t.Error("the delayed trigger's resolution is not paused on its question")
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("a pass with the question open = %v, want ErrChoicePending", err)
	}
	if _, err := g.AdvanceStep(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("advance_step with the question open = %v, want ErrChoicePending", err)
	}
	if err := g.EndTurnByPassing(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("end_turn with the question open = %v, want ErrChoicePending", err)
	}
	if err := g.PassTurn(); !errors.Is(err, game.ErrChoicePending) {
		t.Errorf("pass_turn with the question open = %v, want ErrChoicePending", err)
	}
	if g.Turn.Step != game.StepEnd || g.Turn.Seq != turn {
		t.Fatalf("the cursor left the end step with the question open (turn %d, %v)", g.Turn.Seq, g.Turn.Step)
	}
}

// assertAshlingDeclinedInTheEndStep answers "don't pay" and checks that
// the token is sacrificed in the end step the question was asked in.
func assertAshlingDeclinedInTheEndStep(t *testing.T, g *game.Game, tok uuid.UUID, turn int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	answerPayUnless(t, g, me.ID, false)
	g.SettleResolution()
	if g.Battlefield.Contains(tok) {
		t.Error("the token is still on the battlefield after its controller declined")
	}
	if hasPayUnlessFor(g, me.ID) {
		t.Error("the question is still open after it was answered")
	}
	if g.Turn.Step != game.StepEnd || g.Turn.Seq != turn {
		t.Errorf("the sacrifice landed in turn %d, %v; want turn %d's end step", g.Turn.Seq, g.Turn.Step, turn)
	}
	if !stackFullyEmpty(g) {
		t.Error("the delayed trigger did not finish resolving after the answer")
	}
}

// TestAshlingEndStepQuestionHoldsTheEndStep is #5803: the delayed
// trigger asks its controller in their own end step, so nobody can pass
// priority or leave the step until they answer, and a decline
// sacrifices the token in that end step (CR 118.12, CR 117.3b, CR
// 500.2).
func TestAshlingEndStepQuestionHoldsTheEndStep(t *testing.T) {
	g, tok, turn := ashlingQuestionOpen(t)
	assertAshlingHoldsTheEndStep(t, g, turn)
	if !g.Battlefield.Contains(tok) {
		t.Fatal("the token left before its controller answered")
	}
	assertAshlingDeclinedInTheEndStep(t, g, tok, turn)
}

// Paying keeps the token, and the table is free once it is paid.
func TestAshlingEndStepQuestionPaidKeepsTheToken(t *testing.T) {
	g, tok, turn := ashlingQuestionOpen(t)
	me := g.Seats[g.Turn.ActiveSeat]
	assertAshlingHoldsTheEndStep(t, g, turn)
	floatMana(t, g, me, "{W}{U}{B}{R}{G}")
	answerPayUnless(t, g, me.ID, true)
	g.SettleResolution()
	if !g.Battlefield.Contains(tok) {
		t.Fatal("the token was sacrificed though {W}{U}{B}{R}{G} was paid")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want 0", n)
	}
	if g.Turn.Step != game.StepEnd || g.Turn.Seq != turn {
		t.Errorf("paying moved the cursor to turn %d, %v", g.Turn.Seq, g.Turn.Step)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Errorf("advance_step after the answer = %v, want it to go through", err)
	}
}

// A strict restore taken with the question open resumes it: the restored
// table still holds the end step, and the answer still lands there.
func TestAshlingEndStepQuestionSurvivesAStrictRestore(t *testing.T) {
	t.Run("decline", func(t *testing.T) {
		g, tok, turn := ashlingQuestionOpen(t)
		g = restoreRoundTrip(t, g, true)
		assertAshlingHoldsTheEndStep(t, g, turn)
		assertAshlingDeclinedInTheEndStep(t, g, tok, turn)
	})
	t.Run("pay", func(t *testing.T) {
		g, tok, turn := ashlingQuestionOpen(t)
		g = restoreRoundTrip(t, g, true)
		assertAshlingHoldsTheEndStep(t, g, turn)
		me := g.Seats[g.Turn.ActiveSeat]
		floatMana(t, g, me, "{W}{U}{B}{R}{G}")
		answerPayUnless(t, g, me.ID, true)
		g.SettleResolution()
		if !g.Battlefield.Contains(tok) {
			t.Error("the token was sacrificed though {W}{U}{B}{R}{G} was paid after a restore")
		}
	})
}

// A snapshot taken mid-question before #5803 has the same prompt and the
// same decline key, only without the step anchor. It still restores
// strictly, and declining still sacrifices the token.
func TestAshlingQuestionFromBeforeTheStepAnchorStillRestores(t *testing.T) {
	g, tok, _ := ashlingQuestionOpen(t)
	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("the game is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	choices, _ := doc["pendingChoices"].([]any)
	stripped := 0
	for _, c := range choices {
		if m, ok := c.(map[string]any); ok {
			if _, had := m["owedInStep"]; had {
				delete(m, "owedInStep")
				stripped++
			}
		}
	}
	if stripped != 1 {
		t.Fatalf("stripped the step anchor from %d prompts, want 1", stripped)
	}
	if raw, err = json.Marshal(doc); err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	var old game.GameSnapshot
	if err := json.Unmarshal(raw, &old); err != nil {
		t.Fatalf("decode the old snapshot: %v", err)
	}
	restored, err := old.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict of the old snapshot: %v", err)
	}
	me := restored.Seats[restored.Turn.ActiveSeat]
	if !hasPayUnlessFor(restored, me.ID) {
		t.Fatal("the restored table lost the question")
	}
	answerPayUnless(t, restored, me.ID, false)
	restored.SettleResolution()
	if restored.Battlefield.Contains(tok) {
		t.Error("declining after restoring the old snapshot left the token on the battlefield")
	}
}
