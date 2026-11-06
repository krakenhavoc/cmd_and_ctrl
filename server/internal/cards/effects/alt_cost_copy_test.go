package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// alt_cost_copy_test.go — #4783: a copy of a permanent spell copies the
// alternative cost paid for the original (CR 707.10), and becomes a token
// as it resolves (CR 707.10f, 608.3f). Evoke, dash and blitz attach their
// clauses to "the permanent this spell becomes" when that cost was paid
// (CR 702.74a, 702.109a, 702.152a, and warp's 702.185a), so the token
// gets them too. Escape
// does not: "escaped" means the spell was CAST from a graveyard (CR
// 702.138b), and a copy of a spell isn't cast (CR 707.10).
//
// Sneak, the first of these, has its own file (sneak_copy_test.go).

const (
	altCopyEvokeOracle  = "test-alt-copy-evoke-4783"
	altCopyDashOracle   = "test-alt-copy-dash-4783"
	altCopyBlitzOracle  = "test-alt-copy-blitz-4783"
	altCopyEscapeOracle = "test-alt-copy-escape-4783"
	altCopyWarpOracle   = "test-alt-copy-warp-4783"
	altCopyName         = "Alt Copy Elemental"
)

// registerAltCopyCreature registers a vanilla 2/2 whose one alternative
// cost is `alt`, so the token entering needs no target or trigger of its
// own. `zones` opens the graveyard for escape.
func registerAltCopyCreature(t *testing.T, oracle string, alt game.AlternativeCost, zones ...game.ZoneKind) {
	t.Helper()
	registerForTest(t, Spec{
		OracleID: oracle, Name: altCopyName, Completeness: CompletenessFull,
		CastableZones:    zones,
		AlternativeCosts: []game.AlternativeCost{alt},
	})
}

// altCopyCreature is the registered 2/2's card for `p`.
func altCopyCreature(p *game.Player, oracle string) game.Card {
	return game.Card{
		InstanceID: uuid.New(), Name: altCopyName, TypeLine: "Creature — Elemental", ManaCost: "{2}{R}",
		OracleID: oracle, Power: 2, Toughness: 2, Colors: []string{"R"},
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true},
	}
}

// castAltCopyCreature puts the 2/2 in `me`'s hand and casts it for the
// alternative cost `alt` ("" for its mana cost), leaving it on the stack.
func castAltCopyCreature(t *testing.T, g *game.Game, me *game.Player, oracle, alt string) uuid.UUID {
	t.Helper()
	toMainForCost(t, g)
	c := altCopyCreature(me, oracle)
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	floatForTest(g, me, "RRR")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{AlternativeCost: alt}); err != nil {
		t.Fatalf("cast for %q: %v", alt, err)
	}
	return c.InstanceID
}

// altCopyTokens are the token copies of the 2/2 on the battlefield.
func altCopyTokens(g *game.Game) []game.Card {
	tokens, _ := copiesOnBattlefield(g, altCopyName)
	return tokens
}

// endTheTurn walks to the end step and lets its delayed triggers resolve.
// The original's and the token's trigger at the same time, so the order
// prompt between them is answered as offered.
func endTheTurn(t *testing.T, g *game.Game) {
	t.Helper()
	advanceTo(t, g, game.StepEnd)
	for i := 0; i < 8; i++ {
		if triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if stackFullyEmpty(g) && len(g.PendingChoices) == 0 {
			return
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the end step did not settle")
}

// Evoke: "When this permanent enters, if its evoke cost was paid, its
// controller sacrifices it" (CR 702.74a). The copy's evoke cost was paid,
// so its token is sacrificed as the original is.
func TestACopyOfAnEvokedCreatureSpellIsSacrificedToo(t *testing.T) {
	registerAltCopyCreature(t, altCopyEvokeOracle, Evoke("{R}"))
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := castAltCopyCreature(t, g, me, altCopyEvokeOracle, "evoke")
	copyTheSpell(t, g, me, spell)

	// The copy resolves first: its token is on the battlefield, and the
	// evoke sacrifice is waiting to go on the stack or already there.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass: %v", err)
	}
	for i := 0; i < 8 && len(altCopyTokens(g)) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("pass: %v", err)
		}
	}
	if len(altCopyTokens(g)) != 1 {
		t.Fatalf("%d copy tokens after the copy resolved, want 1", len(altCopyTokens(g)))
	}
	assistCardSettle(t, g, me.ID, nil, true)

	if n := len(altCopyTokens(g)); n != 0 {
		t.Errorf("%d copy tokens still on the battlefield, want 0: the copy's evoke cost was paid (CR 702.74a)", n)
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the evoked original is sacrificed too")
	}
}

// The negative for evoke: a copy of the same card cast for its mana cost
// becomes a token that stays.
func TestACopyOfAHardCastEvokeCreatureStays(t *testing.T) {
	registerAltCopyCreature(t, altCopyEvokeOracle, Evoke("{R}"))
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := castAltCopyCreature(t, g, me, altCopyEvokeOracle, "")
	copyTheSpell(t, g, me, spell)
	assistCardSettle(t, g, me.ID, nil, true)

	if n := len(altCopyTokens(g)); n != 1 {
		t.Errorf("%d copy tokens, want 1: a copy of a hard-cast spell is not evoked", n)
	}
	if !g.Battlefield.Contains(spell) {
		t.Error("the hard-cast original stays")
	}
}

// Dash: "If this spell's dash cost was paid, return the permanent this
// spell becomes to its owner's hand at the beginning of the next end
// step" and "as long as this permanent's dash cost was paid, it has
// haste" (CR 702.109a). The token has haste, and at the end step it is
// returned, where it ceases to exist (CR 704.5d).
func TestACopyOfADashedCreatureSpellHasHasteAndReturnsAtEndStep(t *testing.T) {
	registerAltCopyCreature(t, altCopyDashOracle, Dash("{R}"))
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := castAltCopyCreature(t, g, me, altCopyDashOracle, "dash")
	copyTheSpell(t, g, me, spell)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := altCopyTokens(g)
	if len(toks) != 1 {
		t.Fatalf("%d copy tokens, want 1", len(toks))
	}
	tok := toks[0].InstanceID
	if !hasEffectiveKeyword(t, g, tok, "haste") {
		t.Error("the copy's token has no haste: its dash cost was paid (CR 702.109a)")
	}
	endTheTurn(t, g)
	if g.Battlefield.Contains(tok) {
		t.Error("the copy's token is still on the battlefield after the end step")
	}
	if me.Hand.Contains(tok) {
		t.Error("the returned token is in hand: a token off the battlefield ceases to exist (CR 704.5d)")
	}
	if !me.Hand.Contains(spell) {
		t.Error("the dashed original returns to its owner's hand")
	}
}

// The negative for dash: a copy of a hard-cast dash creature has no haste
// and stays.
func TestACopyOfAHardCastDashCreatureHasNoHasteAndStays(t *testing.T) {
	registerAltCopyCreature(t, altCopyDashOracle, Dash("{R}"))
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := castAltCopyCreature(t, g, me, altCopyDashOracle, "")
	copyTheSpell(t, g, me, spell)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := altCopyTokens(g)
	if len(toks) != 1 {
		t.Fatalf("%d copy tokens, want 1", len(toks))
	}
	if hasEffectiveKeyword(t, g, toks[0].InstanceID, "haste") {
		t.Error("a copy of a hard-cast dash creature has haste")
	}
	endTheTurn(t, g)
	if !g.Battlefield.Contains(toks[0].InstanceID) {
		t.Error("a copy of a hard-cast dash creature is returned at the end step")
	}
}

// Blitz: "If this spell's blitz cost was paid, sacrifice the permanent
// this spell becomes at the beginning of the next end step" and "as long
// as this permanent's blitz cost was paid, it has haste and 'When this
// permanent is put into a graveyard from the battlefield, draw a card'"
// (CR 702.152a). The token gets all three.
func TestACopyOfABlitzedCreatureSpellHasHasteDrawsAndIsSacrificed(t *testing.T) {
	registerAltCopyCreature(t, altCopyBlitzOracle, Blitz("{R}"))
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := castAltCopyCreature(t, g, me, altCopyBlitzOracle, "blitz")
	copyTheSpell(t, g, me, spell)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := altCopyTokens(g)
	if len(toks) != 1 {
		t.Fatalf("%d copy tokens, want 1", len(toks))
	}
	tok := toks[0].InstanceID
	if !hasEffectiveKeyword(t, g, tok, "haste") {
		t.Error("the copy's token has no haste: its blitz cost was paid (CR 702.152a)")
	}
	if grantedTriggerCount(g, tok) != 1 {
		t.Error("the copy's token lacks \"When this creature dies, draw a card.\"")
	}
	hand := len(me.Hand.Cards)
	endTheTurn(t, g)
	if g.Battlefield.Contains(tok) {
		t.Error("the copy's token is not sacrificed at the end step")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the blitzed original is sacrificed at the end step")
	}
	if got := len(me.Hand.Cards); got != hand+2 {
		t.Errorf("hand %d after the end step, want %d: one draw for the original and one for the token", got, hand+2)
	}
}

// Warp shares the seam: "If this spell's warp cost was paid, exile the
// permanent this spell becomes at the beginning of the next end step"
// (CR 702.185a). The token is exiled, where it ceases to exist.
func TestACopyOfAWarpedCreatureSpellIsExiledAtEndStep(t *testing.T) {
	registerAltCopyCreature(t, altCopyWarpOracle, Warp("{R}"))
	g := newCatalogGame(t)
	me := g.Seats[0]
	spell := castAltCopyCreature(t, g, me, altCopyWarpOracle, "warp")
	copyTheSpell(t, g, me, spell)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := altCopyTokens(g)
	if len(toks) != 1 {
		t.Fatalf("%d copy tokens, want 1", len(toks))
	}
	endTheTurn(t, g)
	if g.Battlefield.Contains(toks[0].InstanceID) {
		t.Error("the copy's token is not exiled at the end step: its warp cost was paid (CR 702.185a)")
	}
	if !g.Exile.Contains(spell) {
		t.Error("the warped original is exiled at the end step")
	}
}

// Escape: the copy was not cast (CR 707.10), so it did not escape (CR
// 702.138b) and its token enters with none of the "escapes with" counters
// (CR 702.138c) the original gets.
func TestACopyOfAnEscapedCreatureSpellDidNotEscape(t *testing.T) {
	registerAltCopyCreature(t, altCopyEscapeOracle, EscapeWithCounters("{R}", 1, 2), game.ZoneGraveyard)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	c := altCopyCreature(me, altCopyEscapeOracle)
	g.WithWriteLock(func() { me.Graveyard.PushTop(c) })
	pay := seedGraveyardFodder(t, g, 1)
	floatForTest(g, me, "R")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay,
	}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	copyTheSpell(t, g, me, c.InstanceID)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := altCopyTokens(g)
	if len(toks) != 1 {
		t.Fatalf("%d copy tokens, want 1", len(toks))
	}
	if got := toks[0].Counters["+1/+1"]; got != 0 {
		t.Errorf("the copy's token has %d +1/+1 counters, want 0: a copy isn't cast, so it didn't escape", got)
	}
	if toks[0].Escaped() {
		t.Error("the copy's token reports it escaped (CR 702.138b)")
	}
	orig, ok := g.LookupCardForEffect(c.InstanceID)
	if !ok || !g.Battlefield.Contains(c.InstanceID) || orig.Counters["+1/+1"] != 2 || !orig.Escaped() {
		t.Errorf("the escaped original: on battlefield=%v counters=%d escaped=%v, want 2 counters and escaped",
			g.Battlefield.Contains(c.InstanceID), orig.Counters["+1/+1"], orig.Escaped())
	}
}

// Through a real copier: Double Major on a dashed creature spell under
// Doubling Season and Parallel Lives, so the token creation stops on the
// CR 616 order prompt. The game is restored there, and each token the
// restored answer makes still has the dash riders (the chained
// engine/spell-copy-token-alt-cost tail). Then a second round trip with
// the return pending: the tokens and the original still go at the end
// step.
func TestADashedSpellCopyKeepsItsRidersAcrossTheOrderPromptAndRestore(t *testing.T) {
	registerAltCopyCreature(t, altCopyDashOracle, Dash("{R}"))
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ds, pl := tokenOrderBoard(g, me.ID)
	spell := castAltCopyCreature(t, g, me, altCopyDashOracle, "dash")
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)
	if !orderPromptOpen(g) {
		t.Fatalf("no order prompt is open: %+v", g.PendingChoices)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	me = g.Seats[g.Turn.ActiveSeat]
	answerOrderBySource(t, g, ds, pl)
	assistCardSettle(t, g, me.ID, nil, true)

	toks := altCopyTokens(g)
	if len(toks) == 0 {
		t.Fatal("no copy tokens after the restored answer")
	}
	for _, tok := range toks {
		if !hasEffectiveKeyword(t, g, tok.InstanceID, "haste") {
			t.Errorf("token %s has no haste after the restored creation", tok.InstanceID)
		}
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	endTheTurn(t, g)
	if n := len(altCopyTokens(g)); n != 0 {
		t.Errorf("%d copy tokens after the restored end step, want 0", n)
	}
	if !me.Hand.Contains(spell) {
		t.Error("the dashed original returns to hand after the restore")
	}
}
