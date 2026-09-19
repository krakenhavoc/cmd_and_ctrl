package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// double_major_test.go — CR 608.3f / CR 111.13, end to end. The copy
// of a permanent spell must become a TOKEN, not a second card, and
// it must be a token in every sense the rules use the word: a
// doubler doubles it, its enters-the-battlefield abilities fire, and
// it cannot come back as a card.

const (
	oracleDoubleMajor    = "ece44a82-dcf0-4439-bdd9-a09c99a6f159"
	oracleDoublingSeason = "01546b7d-a233-4176-8843-d732074dc5b6"
	oracleThrabenInsp    = "caa02547-66e3-4e27-a2d3-5e94f3e7a069"
)

// castCreatureSpellForCopy seeds a creature card in the active
// seat's hand with real printed stats and casts it, leaving it on
// the stack for something to copy.
func castCreatureSpellForCopy(t *testing.T, g *game.Game, name, typeLine, oracleID string, p, tough int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		ManaCost:   "{1}{G}",
		Colors:     []string{"G"},
		Power:      p,
		Toughness:  tough,
		Owner:      active.ID,
		Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// copiesOnBattlefield splits the battlefield cards named `name` into
// the tokens and the non-tokens.
func copiesOnBattlefield(g *game.Game, name string) (tokens, cards []game.Card) {
	for _, c := range g.Battlefield.Cards {
		if c.Name != name {
			continue
		}
		if c.IsToken() {
			tokens = append(tokens, c)
		} else {
			cards = append(cards, c)
		}
	}
	return tokens, cards
}

// TestDoubleMajorTurnsTheCopyIntoAToken is the headline: a token
// copy enters alongside the original, and the original still
// resolves as the card it is.
func TestDoubleMajorTurnsTheCopyIntoAToken(t *testing.T) {
	g := newCatalogGame(t)
	spell := castCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)

	tokens, cards := copiesOnBattlefield(g, "Grizzly Bears")
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
	if !stackFullyEmpty(g) {
		t.Error("the stack is not empty — the copy did not cease to exist")
	}
}

// TestDoubleMajorsCopyIsNotLegendary — the except clause (CR 707.10a)
// and the whole reason to play the card in Commander.
func TestDoubleMajorsCopyIsNotLegendary(t *testing.T) {
	g := newCatalogGame(t)
	spell := castCreatureSpellForCopy(t, g, "Tatyova, Benthic Druid",
		"Legendary Creature — Merfolk Druid", "oracle-tatyova", 3, 3)
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)

	tokens, cards := copiesOnBattlefield(g, "Tatyova, Benthic Druid")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("got %d tokens and %d cards, want one of each — the legend rule must not have eaten either",
			len(tokens), len(cards))
	}
	if hasCopySupertype(tokens[0], "Legendary") {
		t.Errorf("the token's type line is %q, want the except clause to have dropped legendary", tokens[0].TypeLine)
	}
	if !hasCopySupertype(cards[0], "Legendary") {
		t.Errorf("the ORIGINAL stopped being legendary (%q) — the except clause edited the wrong object",
			cards[0].TypeLine)
	}
}

// TestDoubleMajorsTokenRunsTheCreaturesEntersAbility — CR 111.13: the
// token is a copy of the spell, so it has the creature's abilities,
// and it entered the battlefield, so they trigger. Two Clues: one
// from the token, one from the card.
func TestDoubleMajorsTokenRunsTheCreaturesEntersAbility(t *testing.T) {
	g := newCatalogGame(t)
	spell := castCreatureSpellForCopy(t, g, "Thraben Inspector",
		"Creature — Human Soldier", oracleThrabenInsp, 1, 2)
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)

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

// TestDoubleMajorsTokenIsDoubledByDoublingSeason is the case that
// separates "a token" from "a permanent we called a token": CR
// 701.7b's creation window has to open on it, which only happens
// because the copy goes through the one token-creation path.
func TestDoubleMajorsTokenIsDoubledByDoublingSeason(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, active.ID, "Doubling Season", "Enchantment", oracleDoublingSeason, false)

	spell := castCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)

	tokens, cards := copiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 2 {
		t.Fatalf("token copies = %d, want 2 under Doubling Season", len(tokens))
	}
	if len(cards) != 1 {
		t.Errorf("real permanents = %d — the ORIGINAL creature spell is a card and is not doubled", len(cards))
	}
}

// TestDoubleMajorsTokenCeasesToExistWhenItLeaves — CR 111.8 /
// CR 704.5d. Bouncing the token must not put a card in a hand.
func TestDoubleMajorsTokenCeasesToExistWhenItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	spell := castCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)

	tokens, _ := copiesOnBattlefield(g, "Grizzly Bears")
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
	if _, still := findBattlefieldByID(g, tokens[0].InstanceID); still {
		t.Error("the token is still on the battlefield")
	}
}

// TestUndoAcrossADoubleMajorToken — the token is made inside the
// resolution, so an undo has to unmake it, and a replay has to make
// exactly one again rather than two or none.
func TestUndoAcrossADoubleMajorToken(t *testing.T) {
	g := newCatalogGame(t)
	spell := castCreatureSpellForCopy(t, g, "Grizzly Bears", "Creature — Bear", "oracle-bears", 2, 2)
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	snap := g.Clone()

	passPriorityAroundTable(t, g)
	if tokens, _ := copiesOnBattlefield(g, "Grizzly Bears"); len(tokens) != 1 {
		t.Fatalf("setup: token copies = %d, want 1", len(tokens))
	}

	g.WithWriteLock(func() { g.RestoreFrom(snap) })
	if tokens, _ := copiesOnBattlefield(g, "Grizzly Bears"); len(tokens) != 0 {
		t.Fatalf("the token survived the undo: %d left", len(tokens))
	}

	passPriorityAroundTable(t, g)
	tokens, cards := copiesOnBattlefield(g, "Grizzly Bears")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("after the replay: %d tokens and %d cards, want one of each", len(tokens), len(cards))
	}
}

// TestADoubleMajorTokenSurvivesASnapshotRoundTrip — a token copy is
// ordinary card-shaped state with no closures in it, so a deploy in
// the middle of a game keeps it.
func TestADoubleMajorTokenSurvivesASnapshotRoundTrip(t *testing.T) {
	g := newCatalogGame(t)
	spell := castCreatureSpellForCopy(t, g, "Tatyova, Benthic Druid",
		"Legendary Creature — Merfolk Druid", "oracle-tatyova", 3, 3)
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)

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

	tokens, cards := copiesOnBattlefield(restored, "Tatyova, Benthic Druid")
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("after the round trip: %d tokens and %d cards, want one of each", len(tokens), len(cards))
	}
	if tokens[0].Power != 3 || tokens[0].Toughness != 3 {
		t.Errorf("restored token P/T = %d/%d, want 3/3", tokens[0].Power, tokens[0].Toughness)
	}
	if hasCopySupertype(tokens[0], "Legendary") {
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
	g := newCatalogGame(t)

	spell, err := castWithOptionalCosts(t, g, "Wolfbriar Elemental",
		"Creature — Elemental", wolfbriarOracle, nil, []int{0, 0}, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	castCatalogSpell(t, g, "Double Major", "Instant", oracleDoubleMajor,
		[]game.TargetRef{{Kind: game.TargetCard, ID: spell}})
	passPriorityAroundTable(t, g)

	tokens, cards := copiesOnBattlefield(g, "Wolfbriar Elemental")
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
	g := newCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID

	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: victim}})
	castCatalogSpell(t, g, "Reverberate", "Instant", reverberateOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bolt}})

	for i := 0; i < 8 && latestPickTarget(g, me) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	prompt := latestPickTarget(g, me)
	if prompt == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g.ResolvePickTarget(prompt.ID, me,
		game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	if len(g.Battlefield.Cards) != 0 {
		t.Errorf("the battlefield holds %d cards — a copied instant must not enter anything", len(g.Battlefield.Cards))
	}
	if got := graveyardSize(g, me); got != 2 {
		t.Errorf("graveyard = %d, want 2 (the Bolt and the Reverberate) — the copy is not a card (CR 707.10)", got)
	}
	if got := lifeOf(g, victim); got != 34 {
		t.Errorf("victim life = %d, want 34 — both the Bolt and its copy should have resolved", got)
	}
}
