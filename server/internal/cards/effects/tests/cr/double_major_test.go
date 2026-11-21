package testscr

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestDoubleMajorTurnsTheCopyIntoAToken is the headline: a token
// copy enters alongside the original, and the original still
// resolves as the card it is.
func TestDoubleMajorTurnsTheCopyIntoAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	tokens, cards := effectstest.CopiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 {
		t.Fatalf("token copies = %d, want 1 (CR 608.3f)", len(tokens))
	}
	if len(cards) != 1 || cards[0].InstanceID != spell {
		t.Fatalf("real permanents = %d, want the original card itself", len(cards))
	}
	tok := tokens[0]
	if tok.Power != 2 || tok.Toughness != 2 {
		t.Errorf("token P/T = %d/%d, want the copied 2/2", tok.Power, tok.Toughness)
	}
	if !tok.IsCreature() {
		t.Errorf("token type line = %q, want a creature", tok.TypeLine)
	}
	if tok.OracleID != "oracle-bears" {
		t.Errorf("token oracle ID = %q — the catalog identity is what makes a copy behave like one", tok.OracleID)
	}
	if tok.Controller != g.Seats[g.Turn.ActiveSeat].ID {
		t.Error("the token is not under the copy's controller")
	}
	// Nothing is left on the stack: the copy became the token and
	// ceased to exist.
	if !effectstest.StackFullyEmpty(g) {
		t.Error("the stack is not empty — the copy did not cease to exist")
	}
}

// TestDoubleMajorsCopyIsNotLegendary — the except clause (CR 707.10a)
// and the whole reason to play the card in Commander.
func TestDoubleMajorsCopyIsNotLegendary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCreatureSpellForCopy(t, g, "Tatyova, Benthic Druid",
		"Legendary Creature — Merfolk Druid", "oracle-tatyova", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	tokens, cards := effectstest.CopiesOnBattlefield(g, "Tatyova, Benthic Druid")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("got %d tokens and %d cards, want one of each — the legend rule must not have eaten either",
			len(tokens), len(cards))
	}
	if effectstest.HasCopySupertype(tokens[0], "Legendary") {
		t.Errorf("the token's type line is %q, want the except clause to have dropped legendary", tokens[0].TypeLine)
	}
	if !effectstest.HasCopySupertype(cards[0], "Legendary") {
		t.Errorf("the ORIGINAL stopped being legendary (%q) — the except clause edited the wrong object",
			cards[0].TypeLine)
	}
}

// TestDoubleMajorsTokenRunsTheCreaturesEntersAbility — CR 111.13: the
// token is a copy of the spell, so it has the creature's abilities,
// and it entered the battlefield, so they trigger. Two Clues: one
// from the token, one from the card.
func TestDoubleMajorsTokenRunsTheCreaturesEntersAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCreatureSpellForCopy(t, g, "Thraben Inspector",
		"Creature — Human Soldier", effectstest.OracleThrabenInsp, 1, 2)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	clues := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Clue" {
			clues++
		}
	}
	if clues != 2 {
		t.Fatalf("Clue tokens = %d, want 2 — the token copy's own ETB trigger must fire", clues)
	}
}

// TestDoubleMajorsTokenIsNotDoubledByDoublingSeason is #4994, the
// owner's 2026-11-19 decision: CR 111.13, "The token is not 'created'
// for the purposes of any replacement effects or triggered abilities
// that refer to creating a token." Doubling Season and Parallel Lives
// both replace a creation, so neither touches the token a copy of a
// permanent spell becomes: one token, not four.
func TestDoubleMajorsTokenIsNotDoubledByDoublingSeason(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, active.ID, "Doubling Season", "Enchantment", effectstest.OracleDoublingSeason, false)
	effectstest.PushCatalogPermanent(g, active.ID, "Parallel Lives", "Enchantment", effectstest.ParallelLivesOracle, false)

	spell := effectstest.CastCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("a prompt is open: %+v — the two doublers must not even be asked about the copy's token", g.PendingChoices)
	}
	tokens, cards := effectstest.CopiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 {
		t.Fatalf("token copies = %d, want 1: the copy's token is not created, so no doubler applies (CR 111.13)", len(tokens))
	}
	if len(cards) != 1 {
		t.Errorf("real permanents = %d — the ORIGINAL creature spell is a card and is not doubled", len(cards))
	}
	if n := effectstest.TokenCreatedEventsFor(g, tokens[0].InstanceID); n != 0 {
		t.Errorf("the token was announced as created %d times, want none (CR 111.13)", n)
	}
}

// TestDoubleMajorsTokenDoesNotTriggerAWheneverYouCreateAToken is the
// other half of CR 111.13: a triggered ability that refers to creating a
// token does not see the copy's token (Mirkwood Bats' "whenever you
// create … a token"), while one that refers to entering does, so the
// token's own enters ability still fires (the Thraben Inspector case
// above). The tokens-created tally does not count it either (CR 608.3f:
// "any rules or effects that refer to creating a token").
func TestDoubleMajorsTokenDoesNotTriggerAWheneverYouCreateAToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, active.ID, "Mirkwood Bats", "Creature — Bat", effectstest.MirkwoodBatsOracle, false)
	lifeBefore := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lifeBefore[p.ID] = p.Life
	}

	spell := effectstest.CastCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	tokens, _ := effectstest.CopiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 {
		t.Fatalf("token copies = %d, want 1", len(tokens))
	}
	for _, p := range g.Seats {
		if p.Life != lifeBefore[p.ID] {
			t.Errorf("seat %s life %d -> %d: Mirkwood Bats triggered on a token that was not created (CR 111.13)", p.ID, lifeBefore[p.ID], p.Life)
		}
	}
	var created int
	g.ReadSnapshot(func() { created = g.TurnTally.Players[active.ID].TokensCreated })
	if created != 0 {
		t.Errorf("tokens created this turn = %d, want 0: the copy's token is not created (CR 608.3f)", created)
	}

	// The control: a token that IS created does drain.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(active.ID, effects.TreasureToken(), 1) })
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.ID == active.ID {
			continue
		}
		if p.Life != lifeBefore[p.ID]-1 {
			t.Errorf("opponent %s life %d -> %d after a created Treasure, want -1", p.ID, lifeBefore[p.ID], p.Life)
		}
	}
}

// TestDoubleMajorsTokenCeasesToExistWhenItLeaves — CR 111.8 /
// CR 704.5d. Bouncing the token must not put a card in a hand.
func TestDoubleMajorsTokenCeasesToExistWhenItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	spell := effectstest.CastCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	tokens, _ := effectstest.CopiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 {
		t.Fatalf("token copies = %d, want 1", len(tokens))
	}
	handBefore := len(active.Hand.Cards)
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(tokens[0].InstanceID); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
	})
	// CR 111.7: it really does reach the hand. The CR 704.5d
	// state-based check at the next boundary is what takes it away,
	// and an effect-driven move deliberately does not run that inline.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if got := len(active.Hand.Cards); got != handBefore {
		t.Fatalf("hand went from %d to %d — a bounced token must not become a card (CR 111.8)", handBefore, got)
	}
	if _, still := effectstest.FindBattlefieldByID(g, tokens[0].InstanceID); still {
		t.Error("the token is still on the battlefield")
	}
}

// TestUndoAcrossADoubleMajorToken — the token is made inside the
// resolution, so an undo has to unmake it, and a replay has to make
// exactly one again rather than two or none.
func TestUndoAcrossADoubleMajorToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	snap := g.Clone()

	effectstest.PassPriorityAroundTable(t, g)
	if tokens, _ := effectstest.CopiesOnBattlefield(g, "Grizzly Bears"); len(tokens) != 1 {
		t.Fatalf("setup: token copies = %d, want 1", len(tokens))
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if tokens, _ := effectstest.CopiesOnBattlefield(g, "Grizzly Bears"); len(tokens) != 0 {
		t.Fatalf("the token survived the undo: %d left", len(tokens))
	}

	effectstest.PassPriorityAroundTable(t, g)
	tokens, cards := effectstest.CopiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("after the replay: %d tokens and %d cards, want one of each", len(tokens), len(cards))
	}
}

// TestADoubleMajorTokenSurvivesASnapshotRoundTrip — a token copy is
// ordinary card-shaped state with no closures in it, so a deploy in
// the middle of a game keeps it.
func TestADoubleMajorTokenSurvivesASnapshotRoundTrip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := effectstest.CastCreatureSpellForCopy(t, g, "Tatyova, Benthic Druid",
		"Legendary Creature — Merfolk Druid", "oracle-tatyova", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}

	tokens, cards := effectstest.CopiesOnBattlefield(restored, "Tatyova, Benthic Druid")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("after the round trip: %d tokens and %d cards, want one of each", len(tokens), len(cards))
	}
	if tokens[0].Power != 3 || tokens[0].Toughness != 3 {
		t.Errorf("restored token P/T = %d/%d, want 3/3", tokens[0].Power, tokens[0].Toughness)
	}
	if effectstest.HasCopySupertype(tokens[0], "Legendary") {
		t.Errorf("the restored token is legendary again: %q", tokens[0].TypeLine)
	}
}

// TestDoubleMajorsTokenCountsTheKicksTheCopyInherited is the CR
// 707.10b / CR 400.7d half of "the copy keeps the choices made when
// casting it": a Double Major on a Wolfbriar Elemental kicked twice
// makes a token that creates its own two Wolves.
//
// The kick count is NOT a copiable value and is deliberately not on
// PrintedValues — it is cast-time state, copied onto the copy's STACK
// ITEM by #988 and stamped onto the permanent at entry, exactly as
// the ordinary permanent-resolution branch stamps it. The token path
// has to do the same or the copy silently makes nothing.
func TestDoubleMajorsTokenCountsTheKicksTheCopyInherited(t *testing.T) {
	g := effectstest.NewCatalogGame(t)

	spell, err := effectstest.CastWithOptionalCosts(t, g, "Wolfbriar Elemental",
		"Creature — Elemental", effectstest.WolfbriarOracle, nil, []int{0, 0}, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	tokens, cards := effectstest.CopiesOnBattlefield(g, "Wolfbriar Elemental")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("got %d token copies and %d cards, want one of each", len(tokens), len(cards))
	}
	if got := game.CardKickedTimes(tokens[0]); got != 2 {
		t.Errorf("the token was kicked %d times, want 2 (CR 707.10b)", got)
	}
	wolves := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wolf" {
			wolves++
		}
	}
	if wolves != 4 {
		t.Errorf("Wolves = %d, want 4 — two from the card and two from its token copy", wolves)
	}
	// …and it is still not a CHARACTERISTIC: a later Clone of the
	// token copies a 4/4 Elemental with no cost record (CR 707.2).
	if v := game.CopiableValuesOf(tokens[0]); v.Name != "Wolfbriar Elemental" {
		t.Errorf("copiable values off the token: %+v", v)
	}
}

// TestACopyOfANonpermanentSpellStillCeasesToExist — the branch this
// change did NOT touch. Reverberate on a Lightning Bolt still leaves
// one Bolt in the graveyard and nothing on the battlefield.
func TestACopyOfANonpermanentSpellStillCeasesToExist(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	effectstest.CastCatalogSpell(t, g, "Reverberate", "Instant", effectstest.ReverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})

	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := effectstest.LatestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(prompt.ID, me,
		game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	if len(g.Battlefield.Cards) != 0 {
		t.Errorf("the battlefield holds %d cards — a copied instant must not enter anything", len(g.Battlefield.Cards))
	}
	if got := effectstest.GraveyardSize(g, me); got != 2 {
		t.Errorf("graveyard = %d, want 2 (the Bolt and the Reverberate) — the copy is not a card (CR 707.10)", got)
	}
	if got := effectstest.LifeOf(g, victim); got != 34 {
		t.Errorf("victim life = %d, want 34 — both the Bolt and its copy should have resolved", got)
	}
}

// oracleBigBertha is Big Bertha's catalog key ({X}{G}, "enters with X
// +1/+1 counters on her").
const oracleBigBertha = "32654ffa-b869-4f8b-a153-4242e8f9d92e"

// castXCreatureSpellForCopy casts an {X} creature card for `x` from the
// active seat's hand and leaves it on the stack for something to copy.
func castXCreatureSpellForCopy(t *testing.T, g *game.Game, name, typeLine, oracleID, manaCost string, p, tough, x int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracleID, ManaCost: manaCost,
		Power: p, Toughness: tough, Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{XValue: x}); err != nil {
		t.Fatalf("cast %s for X=%d: %v", name, x, err)
	}
	return id
}

// TestDoubleMajorsTokenOfAnXCreatureKeepsTheX is #4673: CR 707.10
// copies the X chosen for the spell (CR 107.3m), so a Double Major on
// a Big Bertha cast for X = 3 makes a token that enters with three
// +1/+1 counters and remembers X = 3, not 0.
func TestDoubleMajorsTokenOfAnXCreatureKeepsTheX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	spell := castXCreatureSpellForCopy(t, g, "Big Bertha", "Legendary Creature — Mutant Hero",
		oracleBigBertha, "{X}{G}", 1, 1, 3)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	tokens, cards := effectstest.CopiesOnBattlefield(g, "Big Bertha")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("got %d token copies and %d cards, want one of each", len(tokens), len(cards))
	}
	if n := effectstest.CountersOn(g, cards[0].InstanceID, game.CounterPlusOne); n != 3 {
		t.Errorf("the card has %d counters, want 3", n)
	}
	if n := effectstest.CountersOn(g, tokens[0].InstanceID, game.CounterPlusOne); n != 3 {
		t.Errorf("the token copy has %d counters, want 3 (CR 707.10: the copy keeps X)", n)
	}
	if x := tokens[0].CastX(); x != 3 {
		t.Errorf("the token's X = %d, want 3", x)
	}
}

// TestDoubleMajorsTokenOfAnXCreatureDoublesItsCounters: the copy's X
// counters are an ordinary "enters with" replacement, so Doubling
// Season doubles them on the token as it does on the card. The token
// itself is not doubled (CR 111.13, #4994: it is not created): one
// token, six counters.
func TestDoubleMajorsTokenOfAnXCreatureDoublesItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, active.ID, "Doubling Season", "Enchantment", effectstest.OracleDoublingSeason, false)
	spell := castXCreatureSpellForCopy(t, g, "Big Bertha", "Legendary Creature — Mutant Hero",
		oracleBigBertha, "{X}{G}", 1, 1, 3)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	tokens, _ := effectstest.CopiesOnBattlefield(g, "Big Bertha")
	if len(tokens) != 1 {
		t.Fatalf("token copies = %d, want 1: Doubling Season doubles the counters, not the token (CR 111.13)", len(tokens))
	}
	for _, tok := range tokens {
		if n := effectstest.CountersOn(g, tok.InstanceID, game.CounterPlusOne); n != 6 {
			t.Errorf("a token copy has %d counters, want 6 (3 doubled)", n)
		}
	}
}

// TestDoubleMajorsTokenOfTheRuinousWreckingCrewChoosesUpToTheSameX is
// #4673's mode half: the token's own "choose up to X" reads the X its
// spell copied (#4617), so a copy of a Crew cast for X = 2 asks for up
// to two bullets rather than removing its trigger as X = 0 would.
func TestDoubleMajorsTokenOfTheRuinousWreckingCrewChoosesUpToTheSameX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := castXCreatureSpellForCopy(t, g, "The Ruinous Wrecking Crew", "Legendary Creature — Human Villain",
		effects.TheRuinousWreckingCrewOracle, "{X}{B}{R}", 2, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Double Major", "Instant", effectstest.OracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	effectstest.PassPriorityAroundTable(t, g)

	tokens, _ := effectstest.CopiesOnBattlefield(g, "The Ruinous Wrecking Crew")
	if len(tokens) != 1 {
		t.Fatalf("token copies = %d, want 1", len(tokens))
	}
	tok := tokens[0]
	if n := effectstest.CountersOn(g, tok.InstanceID, game.CounterPlusOne); n != 2 {
		t.Errorf("the token copy has %d counters, want 2", n)
	}
	c := effectstest.ModePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the token's trigger asked nothing: it read X = 0")
	}
	if c.Source != tok.InstanceID {
		t.Fatalf("the mode prompt is for %v, want the token %v", c.Source, tok.InstanceID)
	}
	if c.ModeMin != 0 || c.ModeMax != 2 {
		t.Errorf("prompt bounds %d..%d, want 0..2 (the copied X)", c.ModeMin, c.ModeMax)
	}
}
