package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blue_draw_deck_smoke_test.go — S22's exit criterion (issue #74):
//
//	"Theme-deck smoke test (blue draw deck plays 3 turns)"
//
// Built to the shape S27's theme_deck_smoke_test.go established, for
// the same three reasons, restated because they are the whole value
// of a test like this:
//
//  1. THE CARDS COME OFF THE IMPORT ROAD. Every row below is a
//     cards.Card — a Scryfall record — put through deck.ToGameCard,
//     the function the deck importer and the dev spawner use. The
//     mana cost the strict gate parses, the type line the land logic
//     reads and the oracle ID the catalog keys on all travel the
//     production path instead of being stamped onto a game.Card by
//     the test. That is the bug class a fixture-built test structurally
//     cannot catch (#274 and its battle twin were both invisible to
//     fixture tests for exactly this reason).
//
//  2. THE MANA IS REAL. Every cast goes through Strict + AutoTap, so
//     a spell this deck cannot actually pay for fails here rather than
//     quietly resolving for free — and a four-colour pile like this
//     one is where the auto-tapper gets to be wrong.
//
//  3. THE TURNS ARE REAL. The step machine runs three of the theme
//     seat's turns with the other three seats taking theirs in
//     between. Nothing is teleported: the Confidant flips because an
//     upkeep arrived, the Mine's extra card arrives because a draw
//     step began, the Sylvan Library chain opens in the draw step it
//     belongs to, and the cleanup step is what asks about hand size.
//
// WHAT ONLY THIS TEST CAN SEE
//
// Every card here has a focused test of its own. What none of them
// can assert is what happens when the deck is assembled:
//
//	Engine → payoff     four cards drawn in ONE draw step drain each
//	                    opponent four times through Psychosis Crawler
//	                    and leave four Insects behind through The
//	                    Locust God, because EventDrawCard fires per
//	                    card and both payoffs read it.
//	Reveal ≠ draw       Dark Confidant's upkeep flip grows the hand,
//	                    charges life, and is invisible to BOTH draw
//	                    payoffs and to Sylvan Library's "cards drawn
//	                    this turn" candidate list. Three cards
//	                    disagreeing about what a draw is would be a
//	                    silent, board-dependent bug.
//	Stale P/T           the Crawler is the right size immediately
//	                    after a flip that moves nothing on the
//	                    battlefield — no token, no counter, no tap.
//	                    That is #74's layer-invalidation fix seen
//	                    through cards rather than through an event.
//	The Top loop        the Top tucks itself onto the library and the
//	                    Confidant hands it straight back for 1 life.
//	                    Two cards, one loop, neither aware of the
//	                    other.
//	Reliquary Tower     an eighteen-card hand walks through cleanup,
//	                    and the same hand parks the cursor there the
//	                    moment the Tower is gone.
//	A chain mid-turn    Sylvan Library's three links resolve inside a
//	                    draw step that two other triggers are also
//	                    firing in.
//
// THE CONCESSIONS, STATED
//
// The lands start on the battlefield rather than being played one per
// turn, exactly as S27's test does and for the same reason: a land
// drop per turn caps turn three at three mana and tests the curve
// instead of the cards. Thirteen real imported lands, and the strict
// gate still has to pay every cost out of them.
//
// The spells all start in hand. A draw deck that had to find its own
// engine would be testing the shuffle.
//
// Niv-Mizzet, Parun is not in the catalog; Niv-Mizzet, the Firemind
// is, and it is left out on purpose — its per-draw trigger targets,
// so four draws in one step is four target prompts, and the deck
// already has two payoffs reading the same event. Mystic Remora
// (cumulative upkeep) is the one S22 card still unwritable; Bruvac was
// the other until #569 built the mill-amount replacement kind, and it
// is left out of this deck for the reason Niv-Mizzet is — the deck is
// a mana-and-timing smoke test, not a mill test, and Bruvac's own
// proof lives in mill_replacements_test.go.

// --- the deck -------------------------------------------------------

// blueDeckLand is one imported basic-or-utility land for the mana
// base. Basics get their mana ability derived from the type line; the
// Tower's comes from its catalog entry.
func blueDeckLand(t *testing.T, g *game.Game, p *game.Player, name, typeLine, oracleID, produces string) uuid.UUID {
	t.Helper()
	row := themeDeckRow(name, oracleID, typeLine, "")
	if produces != "" {
		row.ProducedMana = []string{produces}
	}
	land := importThemeCard(row, p.ID)
	land.InstanceID = uuid.New()
	return pushBattlefieldCardWithTimestamp(g, land)
}

// seedBlueManaBase puts thirteen lands onto the battlefield and
// returns the Reliquary Tower's instance ID — the one land the test
// later has to take away.
func seedBlueManaBase(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	for i := 0; i < 9; i++ {
		blueDeckLand(t, g, p, "Island", "Basic Land — Island", "", "U")
	}
	blueDeckLand(t, g, p, "Swamp", "Basic Land — Swamp", "", "B")
	blueDeckLand(t, g, p, "Forest", "Basic Land — Forest", "", "G")
	blueDeckLand(t, g, p, "Mountain", "Basic Land — Mountain", "", "R")
	return blueDeckLand(t, g, p, "Reliquary Tower", "Land", reliquaryTowerOracl, "")
}

// blueDeckSpell is one imported spell, by name, with its printed cost.
type blueDeckSpell struct {
	name     string
	oracleID string
	typeLine string
	manaCost string
	power    string
	tough    string
	keywords []string
}

// seedBlueDrawDeck puts the deck's spells into the theme seat's hand
// through deck.ToGameCard and returns their instance IDs by name.
func seedBlueDrawDeck(t *testing.T, g *game.Game, p *game.Player) map[string]uuid.UUID {
	t.Helper()
	spells := []blueDeckSpell{
		{"Sensei's Divining Top", senseisTopOracle, "Artifact", "{1}", "", "", nil},
		{"Howling Mine", howlingMineOracle, "Artifact", "{2}", "", "", nil},
		{"Dark Confidant", darkConfidantOracle, "Creature — Human Wizard", "{1}{B}", "2", "1", nil},
		{"Psychosis Crawler", psychosisCrawlerOracle, "Artifact Creature — Phyrexian Horror", "{5}", "*", "*", nil},
		{"Rhystic Study", rhysticStudyOracle, "Enchantment", "{2}{U}", "", "", nil},
		{"Sylvan Library", sylvanLibraryOracle, "Enchantment", "{1}{G}", "", "", nil},
		{"The Locust God", locustGodOracle, "Legendary Creature — God", "{4}{U}{R}", "4", "4", []string{"Flying"}},
		{"Consecrated Sphinx", consecratedSphinxOracle, "Creature — Sphinx", "{4}{U}{U}", "4", "6", []string{"Flying"}},
		{"Preordain", preordainOracle, "Sorcery", "{U}", "", "", nil},
		{"Ponder", ponderOracle, "Sorcery", "{U}", "", "", nil},
	}
	ids := make(map[string]uuid.UUID, len(spells))
	for _, s := range spells {
		row := themeDeckRow(s.name, s.oracleID, s.typeLine, s.manaCost)
		row.Power, row.Toughness = s.power, s.tough
		row.Keywords = s.keywords
		c := importThemeCard(row, p.ID)
		ids[c.Name] = c.InstanceID
		p.Hand.PushTop(c)
	}
	return ids
}

// seedBlueLibrary stacks the theme seat's library so every card the
// deck sees is named and priced. Mana value equals the number in the
// name, which is what makes "you lose life equal to its mana value"
// checkable rather than merely non-zero.
func seedBlueLibrary(t *testing.T, p *game.Player) {
	t.Helper()
	// Pushed bottom-first so Deck01 ends up on top.
	for i := 12; i >= 1; i-- {
		row := themeDeckRow(deckCardName(i), "", "Sorcery", deckCardCost(i))
		c := importThemeCard(row, p.ID)
		p.Library.PushTop(c)
	}
}

func deckCardName(i int) string {
	return "Deck" + string(rune('0'+i/10)) + string(rune('0'+i%10))
}

func deckCardCost(i int) string {
	return "{" + string(rune('0'+i/10)) + string(rune('0'+i%10)) + "}"
}

// --- driving the table ----------------------------------------------

// anyPendingChoiceFor reports whether the named seat owes an answer.
func anyPendingChoiceFor(g *game.Game, chooser uuid.UUID) bool {
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == chooser {
			return true
		}
	}
	return false
}

// settleBlueStack empties the stack, answering the two prompts the
// TRIGGER HARVESTER itself raises on the way — an optional trigger's
// yes/no (answered yes: this is a draw deck and it wants its draws)
// and the CR 603.3b ordering prompt when two of the seat's triggers
// go on the stack together. The ordering prompt is answered so the
// stack comes out as the drain would build it unasked, last-queued on
// top: since #1529 an optional trigger (Sylvan Library) joins its
// batch before the drain instead of landing alone above it, and this
// keeps the stack the turn-three walk below was written against.
//
// It answers nothing else. A choice queued by a card's own resolution
// — the scry, the look-at-top, the Sylvan Library chain — is what the
// test is here to watch, so the helper stops and lets the test take
// over.
func settleBlueStack(t *testing.T, g *game.Game, me uuid.UUID) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if answerTriggerOrderLastQueuedFirst(t, g) {
			continue
		}
		if ch := latestTriggerPromptFor(g, me); ch != nil {
			if err := g.ResolveTriggerPrompt(ch.ID, me, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt %q: %v", ch.Reason, err)
			}
			continue
		}
		if anyPendingChoiceFor(g, me) {
			return
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack would not settle in 64 passes")
}

// walkBlueTableTo advances the step machine to `step` on `seat`,
// settling whatever the steps in between put on the stack. The other
// three seats take real turns; their draw steps fire the Mine.
func walkBlueTableTo(t *testing.T, g *game.Game, me uuid.UUID, seat int, step game.Step) {
	t.Helper()
	for i := 0; i < 600; i++ {
		if g.Turn.Step == step && g.Turn.ActiveSeat == seat {
			return
		}
		if !stackFullyEmpty(g) || len(g.PendingTriggers) > 0 {
			settleBlueStack(t, g, me)
			if !stackFullyEmpty(g) {
				// A prompt the test owns is open and the stack
				// cannot drain past it; step forward anyway rather
				// than spinning.
				if _, err := g.AdvanceStep(); err != nil {
					t.Fatalf("AdvanceStep: %v", err)
				}
			}
			continue
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached %s on seat %d", step, seat)
}

// insectCount is how many of The Locust God's tokens are out.
func insectCount(g *game.Game) int { return countNamed(g, "Insect") }

// lifeOfOpponents snapshots every seat but the theme seat.
func blueDeckOpponentLife(g *game.Game, me uuid.UUID) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		if p.ID != me {
			out[p.ID] = p.Life
		}
	}
	return out
}

// assertEachOpponentLost checks the drain fired once per card for
// every opponent, which is the claim a single-opponent assertion
// cannot make.
func assertEachOpponentLost(t *testing.T, g *game.Game, before map[uuid.UUID]int, want int, what string) {
	t.Helper()
	for _, p := range g.Seats {
		was, ok := before[p.ID]
		if !ok {
			continue
		}
		if got := was - p.Life; got != want {
			t.Errorf("%s: opponent %s lost %d life, want %d", what, p.Name, got, want)
		}
	}
}

// --- the test -------------------------------------------------------

// TestBlueDrawDeckPlaysThreeTurns is S22's exit criterion. See the
// file comment for what each turn proves and why the harness is
// shaped the way it is.
func TestBlueDrawDeckPlaysThreeTurns(t *testing.T) {
	g := newCatalogGame(t)
	themeSeat := g.Turn.ActiveSeat
	me := g.Seats[themeSeat]
	oppSeat := (themeSeat + 1) % len(g.Seats)

	tower := seedBlueManaBase(t, g, me)
	hand := seedBlueDrawDeck(t, g, me)
	seedBlueLibrary(t, me)

	// ================= turn one: the engine comes down =================

	walkBlueTableTo(t, g, me.ID, themeSeat, game.StepPrecombatMain)

	castThemeSpell(t, g, me, hand["Howling Mine"])
	castThemeSpell(t, g, me, hand["Sensei's Divining Top"])
	castThemeSpell(t, g, me, hand["Dark Confidant"])
	castThemeSpell(t, g, me, hand["Psychosis Crawler"])

	crawler := hand["Psychosis Crawler"]
	if !onBattlefield(g, crawler) {
		t.Fatal("Psychosis Crawler never reached the battlefield")
	}
	// The CDA, off the import road: the Scryfall record prints "*/*",
	// so a Crawler that is 0/0 here would have died to CR 704.5f the
	// moment it landed.
	if got := effectivePower(t, g, crawler); got != me.Hand.Size() {
		t.Fatalf("Crawler power = %d with %d cards in hand, want %d", got, me.Hand.Size(), me.Hand.Size())
	}

	// The Top's two abilities are a loop. Look at three...
	top := hand["Sensei's Divining Top"]
	if err := g.ActivateCatalogAbility(me.ID, top, 0, game.ActivateAbilityParams{
		Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("Top's look-at-three: %v", err)
	}
	settleBlueStack(t, g, me.ID)
	look := lookAtTopChoiceFor(g, me.ID)
	if look == nil {
		t.Fatal("the Top queued no look-at prompt")
	}
	if len(look.ScryCards) != 3 {
		t.Fatalf("the Top showed %d cards, want 3", len(look.ScryCards))
	}
	// ...put the third one on top...
	if err := g.ResolveLookAtTop(look.ID, me.ID, []uuid.UUID{
		look.ScryCards[2], look.ScryCards[0], look.ScryCards[1],
	}); err != nil {
		t.Fatalf("ResolveLookAtTop: %v", err)
	}
	if got := libraryTopNames(me, 1)[0]; got != "Deck03" {
		t.Fatalf("top of library after the reorder is %q, want Deck03", got)
	}

	// ...and tap to draw exactly that card. One draw, one drain per
	// opponent: the payoff reads the engine's event, not the spell.
	lifeBefore := blueDeckOpponentLife(g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, top, 1, game.ActivateAbilityParams{
		Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("Top's tap-to-draw: %v", err)
	}
	settleBlueStack(t, g, me.ID)
	if !me.Hand.Contains(look.ScryCards[2]) {
		t.Error("the Top drew a card other than the one arranged on top")
	}
	assertEachOpponentLost(t, g, lifeBefore, 1, "one card drawn off the Top")
	// The Top put ITSELF on top of the library, which is the half of
	// the loop that makes it repeatable — and sets up the next turn.
	if onBattlefield(g, top) {
		t.Error("the Top did not tuck itself")
	}
	if got := libraryTopNames(me, 1)[0]; got != "Sensei's Divining Top" {
		t.Fatalf("top of library after the tuck is %q, want Sensei's Divining Top", got)
	}

	// ================= turn two: Bob, and the Mine =====================

	walkBlueTableTo(t, g, me.ID, themeSeat, game.StepUpkeep)

	handBefore := me.Hand.Size()
	myLifeBefore := me.Life
	lifeBefore = blueDeckOpponentLife(g, me.ID)
	settleBlueStack(t, g, me.ID)

	// The Confidant hands the Top straight back, for its mana value
	// of 1. Neither card knows the other exists.
	if !me.Hand.Contains(top) {
		t.Error("the Confidant did not flip the Top the deck had just tucked")
	}
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand after the flip = %d, want %d", got, handBefore+1)
	}
	if got := myLifeBefore - me.Life; got != 1 {
		t.Errorf("life paid for flipping the Top = %d, want 1 (its mana value)", got)
	}
	// REVEALING IS NOT DRAWING. The card reached hand and neither
	// draw payoff noticed.
	assertEachOpponentLost(t, g, lifeBefore, 0, "the Confidant's flip")
	// And the Crawler is the RIGHT SIZE, immediately, off a flip that
	// moved nothing on the battlefield: no token, no counter, no tap
	// to refresh the layer cache. This is the assertion that fails
	// without #74's conditional layer invalidation.
	if got := effectivePower(t, g, crawler); got != me.Hand.Size() {
		t.Errorf("Crawler power = %d after the flip, want %d — the CDA is reading a stale hand",
			got, me.Hand.Size())
	}

	// The draw step: the turn-based draw plus the Mine's additional
	// card. Two draws, two drains each.
	lifeBefore = blueDeckOpponentLife(g, me.ID)
	handBefore = me.Hand.Size()
	walkBlueTableTo(t, g, me.ID, themeSeat, game.StepPrecombatMain)
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand after the draw step = %d, want %d (the turn draw plus Howling Mine)",
			got, handBefore+2)
	}
	assertEachOpponentLost(t, g, lifeBefore, 2, "the draw step with a Howling Mine out")

	castThemeSpell(t, g, me, hand["Rhystic Study"])
	castThemeSpell(t, g, me, hand["Sylvan Library"])
	castThemeSpell(t, g, me, hand["The Locust God"])
	for _, name := range []string{"Rhystic Study", "Sylvan Library", "The Locust God"} {
		if !onBattlefield(g, hand[name]) {
			t.Fatalf("%s never reached the battlefield", name)
		}
	}

	// Cleanup with an eighteen-card hand and a Reliquary Tower: the
	// cursor walks straight through, because CR 402.2's cap is
	// derived from the battlefield rather than written on the player.
	walkBlueTableTo(t, g, me.ID, themeSeat, game.StepEnd)
	settleBlueStack(t, g, me.ID)
	if me.Hand.Size() <= 7 {
		t.Fatalf("hand is %d at the end step; the Tower assertion needs a hand over seven", me.Hand.Size())
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into cleanup: %v", err)
	}
	if g.Turn.Step == game.StepCleanup {
		t.Error("the cursor parked at cleanup with a Reliquary Tower on the battlefield")
	}
	if len(g.DiscardPending) != 0 {
		t.Errorf("discards owed with a Reliquary Tower out: %v", g.DiscardPending)
	}

	// ================ turn three: four draws in one step ===============

	walkBlueTableTo(t, g, me.ID, themeSeat, game.StepUpkeep)

	myLifeBefore = me.Life
	settleBlueStack(t, g, me.ID)
	if got := myLifeBefore - me.Life; got != 4 {
		t.Errorf("life paid for the Confidant's second flip = %d, want 4 (Deck04's mana value)", got)
	}
	// The card the flip put into hand. MoveCard pushes to the top, so
	// it is the last one — and Sylvan Library must not offer it.
	flipped := me.Hand.Cards[me.Hand.Size()-1].InstanceID

	// The draw step, which three separate things want: the turn-based
	// draw (CR 504.1, not a trigger at all), Howling Mine's additional
	// card, and Sylvan Library's two. Every one of the four is a drain
	// and an Insect.
	lifeBefore = blueDeckOpponentLife(g, me.ID)
	handBefore = me.Hand.Size()
	insectsBefore := insectCount(g)
	walkBlueTableTo(t, g, me.ID, themeSeat, game.StepDraw)
	settleBlueStack(t, g, me.ID)

	// Sylvan Library's trigger was harvested last, so it is on top of
	// the stack and resolves FIRST — the Mine's card is still
	// underneath while the chain is being answered. That ordering is
	// why the candidate list below is three and not four, and it is
	// the real stack, not a quirk of the harness.
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Sylvan Library queued no choose-cards prompt")
	}
	if n := len(g.PendingChoices); n != 1 {
		t.Errorf("%d prompts open at once, want 1 — the chain must not front-load", n)
	}
	// THE CONFIDANT'S CARD IS NOT A CARD DRAWN THIS TURN. It arrived
	// in this hand, this turn, from the top of this library — and
	// Sylvan Library must not offer it, because it was revealed and
	// put into hand rather than drawn. A third card agreeing about
	// what a draw is.
	for _, id := range pick.ChooseCards {
		if id == flipped {
			t.Error("Sylvan Library offered the card Dark Confidant revealed; a flip is not a draw")
		}
	}
	if len(pick.ChooseCards) != 3 {
		t.Errorf("Sylvan Library offered %d cards drawn this turn, want 3 (the turn draw plus its own two)",
			len(pick.ChooseCards))
	}

	chosen := []uuid.UUID{pick.ChooseCards[0], pick.ChooseCards[1]}
	myLifeBefore = me.Life
	libBefore := me.Library.Size()
	if err := g.ResolveChooseCards(pick.ID, me.ID, chosen); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	// Link three, card one: pay.
	first := confirmChoiceFor(g, me.ID)
	if first == nil {
		t.Fatal("the pick queued no pay-or-put-back prompt")
	}
	if err := g.ResolveConfirm(first.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm(pay): %v", err)
	}
	// Link three, card two — queued by the FIRST card's answer, and
	// priced against a life total that answer already moved.
	second := confirmChoiceFor(g, me.ID)
	if second == nil {
		t.Fatal("paying for the first card did not queue the second card's question")
	}
	if err := g.ResolveConfirm(second.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm(put back): %v", err)
	}
	if got := myLifeBefore - me.Life; got != 4 {
		t.Errorf("life paid across the chain = %d, want 4 (one paid, one put back)", got)
	}
	if me.Library.Size() != libBefore+1 {
		t.Errorf("library = %d after the put-back, want %d", me.Library.Size(), libBefore+1)
	}
	if !me.Hand.Contains(chosen[0]) {
		t.Error("the paid-for card left the hand")
	}
	if me.Hand.Contains(chosen[1]) {
		t.Error("the declined card is still in hand")
	}

	// Now the rest of the step: Howling Mine's card, waiting under the
	// chain the whole time, and every drain and Insect the four draws
	// owe.
	settleBlueStack(t, g, me.ID)
	if got := len(g.CardsDrawnThisTurnFor(me.ID)); got != 4 {
		t.Errorf("cards drawn this turn = %d, want 4 (turn draw + Sylvan's two + the Mine's)", got)
	}
	assertEachOpponentLost(t, g, lifeBefore, 4, "four cards drawn in one draw step")
	if got := insectCount(g) - insectsBefore; got != 4 {
		t.Errorf("Insects made in the draw step = %d, want 4 — The Locust God fires per CARD", got)
	}
	// Four cards in, one put back on top and drawn again by the Mine
	// — which is a real and slightly absurd consequence of the two
	// cards sharing a step, and worth pinning rather than smoothing.
	if got := me.Hand.Size(); got != handBefore+3 {
		t.Errorf("hand after the whole draw step = %d, want %d", got, handBefore+3)
	}

	// The main phase: the last two cantrips, each one a drain and an
	// Insect on its own.
	walkBlueTableTo(t, g, me.ID, themeSeat, game.StepPrecombatMain)
	castThemeSpell(t, g, me, hand["Consecrated Sphinx"])

	lifeBefore = blueDeckOpponentLife(g, me.ID)
	insectsBefore = insectCount(g)
	castThemeSpell(t, g, me, hand["Preordain"])
	scry := scryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("Preordain queued no scry prompt")
	}
	if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	settleBlueStack(t, g, me.ID)
	assertEachOpponentLost(t, g, lifeBefore, 1, "Preordain's draw")
	if got := insectCount(g) - insectsBefore; got != 1 {
		t.Errorf("Insects from Preordain's draw = %d, want 1", got)
	}

	// Ponder is the chained-prompt cantrip: look at three, THEN
	// answer "you may shuffle", THEN draw — and the draw is on both
	// branches.
	lifeBefore = blueDeckOpponentLife(g, me.ID)
	castThemeSpell(t, g, me, hand["Ponder"])
	ponderLook := lookAtTopChoiceFor(g, me.ID)
	if ponderLook == nil {
		t.Fatal("Ponder queued no look-at prompt")
	}
	if err := g.ResolveLookAtTop(ponderLook.ID, me.ID, ponderLook.ScryCards); err != nil {
		t.Fatalf("ResolveLookAtTop (Ponder): %v", err)
	}
	shuffle := confirmChoiceFor(g, me.ID)
	if shuffle == nil {
		t.Fatal("Ponder's reorder did not chain into the shuffle question")
	}
	if err := g.ResolveConfirm(shuffle.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm (Ponder, keep the order): %v", err)
	}
	settleBlueStack(t, g, me.ID)
	assertEachOpponentLost(t, g, lifeBefore, 1, "Ponder's draw")

	// Take the Tower away and ask the same question again. An
	// eighteen-card hand that walked through cleanup last turn parks
	// the cursor there now — which is what makes the earlier
	// assertion a fact about the Tower rather than about the cap.
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneBattlefield},
		game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID},
		tower,
	); err != nil {
		t.Fatalf("destroying the Tower: %v", err)
	}
	walkBlueTableTo(t, g, me.ID, themeSeat, game.StepEnd)
	settleBlueStack(t, g, me.ID)
	handAtEnd := me.Hand.Size()
	if handAtEnd <= 7 {
		t.Fatalf("hand is %d at the end step; the cap assertion needs a hand over seven", handAtEnd)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep into cleanup: %v", err)
	}
	if g.Turn.Step != game.StepCleanup {
		t.Errorf("the cursor is at %s, want cleanup — an over-max hand with no Tower must pause", g.Turn.Step)
	}
	if got := g.DiscardPending[me.ID]; got != handAtEnd-7 {
		t.Errorf("discards owed with the Tower gone = %d, want %d", got, handAtEnd-7)
	}

	// ========== the coda: what the deck does on someone else's turn =====
	//
	// Rhystic Study and Consecrated Sphinx are the two cards here
	// that never do anything on their controller's turn. They are
	// also the reason a draw deck taxes a whole table rather than
	// just drawing quietly, and neither has ever been seen next to a
	// draw payoff before this test.

	// Discharge the owed discards so the cursor can leave cleanup —
	// which is also the assertion that the pause is a pause and not a
	// wedge.
	owed := g.DiscardPending[me.ID]
	toss := make([]uuid.UUID, 0, owed)
	for i := 0; i < owed; i++ {
		toss = append(toss, me.Hand.Cards[i].InstanceID)
	}
	if err := g.DiscardSelection(me.ID, toss); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	if g.Turn.Step == game.StepCleanup {
		t.Error("the cursor stayed at cleanup after the discard was made")
	}
	if got := me.Hand.Size(); got != 7 {
		t.Errorf("hand after the cleanup discard = %d, want 7", got)
	}

	// Give the next seat two Islands and a spell to cast with them.
	opponent := g.Seats[oppSeat]
	for i := 0; i < 2; i++ {
		blueDeckLand(t, g, opponent, "Island", "Basic Land — Island", "", "U")
	}
	oppSpell := importThemeCard(themeDeckRow("Opponent's Cantrip", "", "Sorcery", "{1}"), opponent.ID)
	opponent.Hand.PushTop(oppSpell)

	// Their draw step. Two cards — the turn-based draw and Howling
	// Mine's — and the Sphinx offers a two-card draw for each one,
	// because EventDrawCard fires per card on an OPPONENT'S draw too.
	lifeBefore = blueDeckOpponentLife(g, me.ID)
	insectsBefore = insectCount(g)
	handBefore = me.Hand.Size()
	walkBlueTableTo(t, g, me.ID, oppSeat, game.StepPrecombatMain)
	settleBlueStack(t, g, me.ID)

	if got := me.Hand.Size() - handBefore; got != 4 {
		t.Errorf("cards drawn off the opponent's draw step = %d, want 4 (two Sphinx triggers, two cards each)", got)
	}
	// Every card the Sphinx gave me is a drain and an Insect. The
	// opponent's own two draws are not: the payoffs read Actor.
	assertEachOpponentLost(t, g, lifeBefore, 4, "the Sphinx's four cards")
	if got := insectCount(g) - insectsBefore; got != 4 {
		t.Errorf("Insects from the Sphinx's cards = %d, want 4", got)
	}

	// And the tax. The opponent casts; Rhystic Study asks for {1};
	// they decline; the card I draw for it drains them again.
	lifeBefore = blueDeckOpponentLife(g, me.ID)
	insectsBefore = insectCount(g)
	handBefore = me.Hand.Size()
	if err := g.CastSpell(opponent.ID, oppSpell.InstanceID, game.CastSpellParams{
		Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("the opponent could not cast into a Rhystic Study: %v", err)
	}
	settleBlueStack(t, g, me.ID)
	if !hasPayUnlessFor(g, opponent.ID) {
		t.Fatal("Rhystic Study did not tax the opponent's spell")
	}
	answerPayUnless(t, g, opponent.ID, false)
	settleBlueStack(t, g, me.ID)

	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("cards drawn off the declined Rhystic tax = %d, want 1", got)
	}
	assertEachOpponentLost(t, g, lifeBefore, 1, "the card drawn off Rhystic Study")
	if got := insectCount(g) - insectsBefore; got != 1 {
		t.Errorf("Insects from the Rhystic draw = %d, want 1", got)
	}
}
