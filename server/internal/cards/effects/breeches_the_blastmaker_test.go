package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const breechesBlastmakerOracle = "8514f20a-50c7-4319-84cb-2bf263548234"

// breechesRNGKey pins the keyed coin-flip stream (ADR 0054) so every
// run of breechesPlay draws the same sequence of "heads"/"tails"
// results from Breeches' own (player, source) stream. Without it,
// newCatalogGame leaves Game.rngKey unset, which mints a fresh
// crypto-random key on the first draw — so TestBreechesCopiesOnAWinAndBlastsOnALoss's
// scan over 12 unpinned flips had roughly a 1-in-4096 chance of
// landing all wins or all losses and failing (#1605). Any fixed,
// non-zero key works; this one has no meaning beyond "not all
// zeroes" and is not shared with internal/game's own known-answer
// tests.
func breechesRNGKey() [32]byte {
	return [32]byte{
		0xb2, 0xee, 0xc4, 0x1e, 0x05, 0xd6, 0x60, 0x5c,
		0x12, 0x87, 0xff, 0x3a, 0x9c, 0x44, 0x70, 0xd1,
		0x91, 0x2d, 0xa8, 0x63, 0x0f, 0xb7, 0x4a, 0x38,
		0x55, 0xc9, 0x6e, 0x21, 0xf4, 0x0a, 0xd3, 0x77,
	}
}

// castSpellWithCost is castCatalogSpell for a spell whose MANA COST
// the test needs — "damage equal to that spell's mana value" is
// unobservable on the zero-cost placeholder cards the harness seeds.
func castSpellWithCost(t *testing.T, g *game.Game, name, typeLine, oracleID, manaCost string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID,
		ManaCost: manaCost, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// passUntilPickTarget passes priority until `chooser` is asked to
// choose a target, or the stack empties.
func passUntilPickTarget(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16; i++ {
		if p := latestPickTarget(g, chooser); p != nil {
			return p
		}
		if stackFullyEmpty(g) {
			return nil
		}
		if err := g.PassPriority(); err != nil {
			return latestPickTarget(g, chooser)
		}
	}
	return latestPickTarget(g, chooser)
}

// breechesRun plays the whole card once: Breeches and one artifact on
// the battlefield, a first spell to set the tally, then a Lightning
// Bolt at an opponent as the SECOND spell of the turn. `burn` coin
// flips are drawn from Breeches' own keyed stream first, which is what
// lets one deterministic test reach both faces (ADR 0054 keys the
// stream on player + source, and the counter advances per draw).
//
// Returns whether the flip was won and how much life the opponent
// lost in total.
func breechesRun(t *testing.T, burn int) (won bool, lifeLost int) {
	t.Helper()
	return breechesPlay(t, burn, nil)
}

// breechesPlay is breechesRun with a hook that runs once the coin has
// been called and before either reflexive payoff resolves — the window
// in which the table can still answer "that spell". The hook is told
// the Bolt's ID and whether the flip was won. With a hook set, a payoff
// that asks for no target is not a failure: countering the spell is
// exactly how a copy effect can end up with nothing to ask about.
func breechesPlay(t *testing.T, burn int, afterFlip func(g *game.Game, bolt uuid.UUID, won bool)) (won bool, lifeLost int) {
	t.Helper()
	g := newCatalogGame(t)
	g.SetRNGKeyForTest(breechesRNGKey())
	me, opp := g.Seats[0], g.Seats[1]
	breeches := pushCatalogPermanent(g, me.ID, "Breeches, the Blastmaker",
		"Legendary Creature — Goblin Pirate", breechesBlastmakerOracle, false)
	relic := pushCatalogPermanent(g, me.ID, "Worn Relic", "Artifact", "", false)

	g.WithWriteLock(func() {
		for i := 0; i < burn; i++ {
			_, _ = g.FlipCoinsForEffect(game.RandomDraw{Player: me.ID, Source: breeches}, 1)
		}
	})

	// The first spell of the turn is silent.
	castCatalogSpell(t, g, "Opening Gambit", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the FIRST spell each turn does not trigger: %+v", g.PendingChoices)
	}

	before := opp.Life
	bolt := castSpellWithCost(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	// "You may sacrifice an artifact" — the resolution's own question,
	// so the controller picks WHICH artifact.
	answerMayChoice(t, g, me.ID, true)
	sac := latestChooseCardsFor(g, me.ID)
	if sac == nil {
		t.Fatalf("the controller picks which artifact to sacrifice: %+v", g.PendingChoices)
	}
	if err := g.ResolveChooseCards(sac.ID, me.ID, []uuid.UUID{relic}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(relic) {
		t.Fatal("the artifact was sacrificed")
	}

	call := latestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if call == nil {
		t.Fatalf("the sacrifice is followed by a coin flip: %+v", g.PendingChoices)
	}
	if err := g.ResolveCoinCall(call.ID, me.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	flips := randomEvents(g, game.EventFlipCoin)
	if len(flips) != burn+1 {
		t.Fatalf("one flip for the trigger on top of %d burned, got %d", burn, len(flips))
	}
	won = flips[len(flips)-1].Won
	if afterFlip != nil {
		afterFlip(g, bolt, won)
	}

	// Either payoff is a reflexive trigger that goes on the stack and
	// asks for a target: the copy's "you may choose new targets", or
	// the blast's "any target". Both point back at the same opponent.
	pick := passUntilPickTarget(t, g, me.ID)
	if pick == nil {
		if afterFlip != nil {
			passPriorityAroundTable(t, g)
			return won, before - opp.Life
		}
		t.Fatalf("the flip's payoff is a trigger that targets: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID,
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	return won, before - opp.Life
}

// TestBreechesCopiesOnAWinAndBlastsOnALoss is the whole card. The
// Bolt is {R} for 3 damage, so the two outcomes are arithmetically
// distinct: a copy deals another 3 (6 total), while a lost flip deals
// the spell's mana value of 1 (4 total).
func TestBreechesCopiesOnAWinAndBlastsOnALoss(t *testing.T) {
	sawWin, sawLoss := false, false
	for burn := 0; burn < 12 && !(sawWin && sawLoss); burn++ {
		won, lost := breechesRun(t, burn)
		if won {
			sawWin = true
			if lost != 6 {
				t.Errorf("a won flip copies the Bolt: opponent lost %d life, want 6", lost)
			}
			continue
		}
		sawLoss = true
		if lost != 4 {
			t.Errorf("a lost flip deals the spell's mana value of 1: opponent lost %d life, want 4", lost)
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("both faces must be exercised: saw a win=%v, saw a loss=%v", sawWin, sawLoss)
	}
}

// TestBreechesCopiesASpellCounteredBeforeTheCopyResolves — #1288,
// CR 608.2h. "When you win the flip, copy that spell" NAMES the spell;
// it does not target it. So when the Bolt is countered after the flip
// is won and before the reflexive trigger resolves, the copy is still
// made — from the Bolt as it last stood on the stack, with its target —
// the way storm and Doublecast copy a countered spell since #1255. The
// Bolt itself deals nothing; its copy deals 3.
func TestBreechesCopiesASpellCounteredBeforeTheCopyResolves(t *testing.T) {
	for burn := 0; burn < 12; burn++ {
		countered := false
		won, lost := breechesPlay(t, burn, func(g *game.Game, bolt uuid.UUID, won bool) {
			if won {
				counterNow(t, g, bolt)
				countered = true
			}
		})
		if !won {
			continue
		}
		if !countered {
			t.Fatal("the hook did not counter the Bolt — the fixture is wrong")
		}
		if lost != 3 {
			t.Errorf("opponent lost %d life, want 3: the countered Bolt deals nothing, its last-known copy deals 3 (CR 608.2h)", lost)
		}
		return
	}
	t.Fatal("no won flip in 12 burns — the fixture cannot reach the copy")
}

// Declining the sacrifice ends the ability: no flip, no payoff, and
// the artifact stays. The "you may" is real.
func TestBreechesDecliningTheSacrificeEndsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Breeches, the Blastmaker",
		"Legendary Creature — Goblin Pirate", breechesBlastmakerOracle, false)
	relic := pushCatalogPermanent(g, me.ID, "Worn Relic", "Artifact", "", false)

	castCatalogSpell(t, g, "Opening Gambit", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	before := opp.Life
	castSpellWithCost(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	answerMayChoice(t, g, me.ID, false)
	if len(randomEvents(g, game.EventFlipCoin)) != 0 {
		t.Error("the coin is gated on the sacrifice")
	}
	if !g.Battlefield.Contains(relic) {
		t.Error("nothing was sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 3 {
		t.Errorf("only the Bolt itself resolves: opponent lost %d life, want 3", got)
	}
}

// A controller with no artifact is asked nothing at all — "you may
// sacrifice an artifact" with no artifact is not a question, and
// without the sacrifice there is no flip.
func TestBreechesAsksNothingWithNoArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Breeches, the Blastmaker",
		"Legendary Creature — Goblin Pirate", breechesBlastmakerOracle, false)

	castCatalogSpell(t, g, "Opening Gambit", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	castSpellWithCost(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	if len(g.PendingChoices) != 0 {
		t.Errorf("no artifact, no question: %+v", g.PendingChoices)
	}
	if len(randomEvents(g, game.EventFlipCoin)) != 0 {
		t.Error("and no flip")
	}
}

// The printed keyword and the completeness declaration.
func TestBreechesIsCompleteAndHasMenace(t *testing.T) {
	spec, ok := Lookup(breechesBlastmakerOracle)
	if !ok {
		t.Fatal("Breeches, the Blastmaker is registered")
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("completeness = %v, want full", spec.Completeness)
	}
	if !hasAbility(spec.PrintedKeywords, "menace") {
		t.Errorf("Breeches prints menace, got %v", spec.PrintedKeywords)
	}
}
