package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// theme_deck_smoke_test.go — S27's exit criterion (issue #92):
//
//	"Theme-deck smoke test: a Superfriends + sagas + vehicles deck
//	 plays through 4 turns asserting expected interactions."
//
// Every other S27 test pins one primitive against a hand-built
// fixture. This one is the opposite shape on purpose, and the
// difference is the point:
//
//  1. THE CARDS COME OFF THE IMPORT ROAD. Each row below is a
//     cards.Card — a Scryfall record — passed through
//     deck.ToGameCard, the same function the deck importer and the
//     dev spawner use. So the printed loyalty that CR 306.5b turns
//     into counters, the printed defense CR 310.4 needs, the printed
//     keywords the combat gates read and the mana cost the strict
//     gate parses all travel the production path rather than being
//     stamped onto a game.Card by the test. #274 (planeswalkers
//     entering with zero loyalty) and its battle twin were both
//     invisible to fixture-built tests for exactly this reason: the
//     fixtures set the field the importer had no path to.
//
//  2. THE MANA IS REAL. Every cast goes through Strict + AutoTap, so
//     a card that the deck cannot actually pay for fails here
//     instead of quietly resolving for free.
//
//  3. THE TURNS ARE REAL. The game runs its own step machine across
//     four of the theme seat's turns (game turns 1, 5, 9 and 13 at a
//     four-player table) with the other three seats taking their
//     turns in between. Nothing is teleported into a step; the saga
//     advances because the precombat main phase arrived, the crew
//     effect expires because cleanup happened, and the loyalty gate
//     reopens because the turn changed.
//
// The one concession is the lands: eight Plains start on the
// battlefield rather than being played one per turn. A land drop per
// turn would cap the fourth turn at four mana and put Elspeth out of
// reach, which would test the mana curve instead of the card types.
// The Plains are still real imported Plains with the engine's derived
// basic-land mana ability, and the strict gate still has to pay every
// cost out of them.
//
// WHAT IT ASSERTS, BY CARD TYPE
//
//	Saga        entry lore counter (CR 714.2b), the precombat-main
//	            advance, chapters I / II / III firing in order, and
//	            the CR 704.5s sacrifice once the final chapter has
//	            RESOLVED.
//	Vehicle     a Vehicle is not a creature until crewed; a
//	            summoning-sick creature may crew (CR 702.122b); a
//	            Vehicle that has been on the battlefield since the
//	            turn began can attack the turn it is first crewed
//	            (CR 302.6 measures continuous control, not
//	            creature-hood); and the animation expires at cleanup.
//	Planeswalker printed loyalty becomes counters on entry; loyalty is
//	            paid at announce (CR 606.2); once per turn per
//	            planeswalker (CR 606.5) and the gate reopens next
//	            turn; combat damage from an attacker removes loyalty
//	            (#406); a walker taken to zero leaves (CR 704.5i).
//
// It also asserts the two things that only show up once several card
// types share a board: a Saga's tokens crewing a Vehicle, and a
// planeswalker sweeper that has to read post-layer power to spare the
// creatures its own controller made.
//
// WHAT IS DELIBERATELY NOT HERE
//
// Battles, the fourth card type of the sprint. The exit criterion
// names three, and the one battle in the catalog (Invasion of
// Innistrad) is {2}{B}{B} — putting it in this deck means adding an
// off-colour mana base to a mono-white list so that one cast can
// happen, which tests the auto-tapper rather than the card type. The
// battle lifecycle is pinned where it belongs: entry defense, the
// protector prompt and the zero-defense sweep in
// game/battle_test.go, and attacking one in
// game/attack_target_test.go (TestDeclareAttackerAcceptsABattleYouDoNotProtect,
// TestDamageRemovesDefenseFromABattle).
//
// Likewise the legend rule, the counter doublers and the Siege
// defeated trigger: all three are pinned by focused tests
// (game/legend_rule_test.go, superfriends_test.go, battles_test.go)
// and none of them need four turns of context to be believable.

// The oracle keys this file needs are already declared by the
// focused S27 tests — historyOfBenaliaOracle in saga_test.go,
// smugglersCopterOracle in vehicles_test.go, elspethSunsChampionOracle
// in superfriends_test.go and wanderingEmperorOracle in
// planeswalker_loyalty_test.go. They are reused rather than restated
// so a re-keyed card breaks in one place.

// themeDeckRow is a Scryfall record in the shape deck.ToGameCard
// consumes. Only the fields the engine reads are filled; art URIs,
// legalities and set metadata are irrelevant to a headless game.
func themeDeckRow(name, oracleID, typeLine, manaCost string) cards.Card {
	c := cards.Card{
		ID:       uuid.New(),
		Name:     name,
		TypeLine: typeLine,
		ManaCost: manaCost,
	}
	if oracleID != "" {
		c.OracleID = uuid.MustParse(oracleID)
	}
	return c
}

// importThemeCard runs one row down the production import road and
// hands the result to `owner`. Mirrors what deck.Resolve does for a
// player-uploaded list: build the game.Card, then attribute it.
func importThemeCard(row cards.Card, owner uuid.UUID) game.Card {
	c := deck.ToGameCard(row, false)
	c.Owner = owner
	c.Controller = owner
	return c
}

// seedThemeDeck puts the eight Plains onto the battlefield and the
// four spells into the theme seat's hand, all through
// deck.ToGameCard. Returns the hand instance IDs by card name.
func seedThemeDeck(t *testing.T, g *game.Game, p *game.Player) map[string]uuid.UUID {
	t.Helper()

	plains := themeDeckRow("Plains", "", "Basic Land — Plains", "")
	plains.ProducedMana = []string{"W"}
	for i := 0; i < 8; i++ {
		land := importThemeCard(plains, p.ID)
		land.InstanceID = uuid.New()
		pushBattlefieldCardWithTimestamp(g, land)
	}

	benalia := themeDeckRow("History of Benalia", historyOfBenaliaOracle,
		"Enchantment — Saga", "{1}{W}{W}")

	copter := themeDeckRow("Smuggler's Copter", smugglersCopterOracle,
		"Artifact — Vehicle", "{2}")
	copter.Power, copter.Toughness = "3", "3"
	copter.Keywords = []string{"Flying"}

	elspeth := themeDeckRow("Elspeth, Sun's Champion", elspethSunsChampionOracle,
		"Legendary Planeswalker — Elspeth", "{4}{W}{W}")
	elspeth.Loyalty = "4"

	emperor := themeDeckRow("The Wandering Emperor", wanderingEmperorOracle,
		"Legendary Planeswalker — Wanderer", "{1}{W}{W}")
	emperor.Loyalty = "3"
	emperor.Keywords = []string{"Flash"}

	ids := make(map[string]uuid.UUID, 4)
	for _, row := range []cards.Card{benalia, copter, elspeth, emperor} {
		c := importThemeCard(row, p.ID)
		ids[c.Name] = c.InstanceID
		p.Hand.PushTop(c)
	}
	return ids
}

// castThemeSpell casts from the theme seat's hand with the strict
// mana gate and the auto-tapper engaged, then settles the stack.
func castThemeSpell(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// counterOn reads one counter kind off a battlefield permanent, or
// -1 when the permanent has left the battlefield.
func counterOn(g *game.Game, id uuid.UUID, kind string) int {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Counters[kind]
		}
	}
	return -1
}

// isCreatureNow reports whether a permanent is a creature after
// continuous effects — the question a crewed Vehicle's whole
// lifecycle turns on.
func isCreatureNow(g *game.Game, id uuid.UUID) bool {
	c, ok := battlefieldCardByID(g, id)
	return ok && c.IsCreature()
}

// advanceToStepOf walks the step machine to `step` on the named
// seat's turn. Deliberately drives the real machine rather than
// assigning Turn.Step: the saga advance, the untap and the cleanup
// all hang off step entry, and a test that skipped them would be
// asserting against a board no game could produce.
func advanceToStepOf(t *testing.T, g *game.Game, seat int, step game.Step) {
	t.Helper()
	for i := 0; i < 600; i++ {
		if g.Turn.Step == step && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatalf("never reached %s on seat %d", step, seat)
}

// knightIDs returns the Knight tokens History of Benalia has made so
// far, in battlefield order. The test only needs "a Knight" and "a
// second, different Knight", so the order is not load-bearing.
func knightIDs(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Knight" {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// TestThemeDeckPlaysFourTurns is S27's exit criterion. See the file
// comment for what each turn proves and why the harness is shaped
// the way it is.
func TestThemeDeckPlaysFourTurns(t *testing.T) {
	g := newCatalogGame(t)
	themeSeat := g.Turn.ActiveSeat
	me := g.Seats[themeSeat]
	opponent := g.Seats[(themeSeat+1)%len(g.Seats)]

	hand := seedThemeDeck(t, g, me)

	// --- turn 1 ------------------------------------------------------
	//
	// The Vehicle lands as a bare artifact and the Saga lands with its
	// first lore counter already on it (CR 714.2b), which fires
	// chapter I before anybody gets priority back.
	advanceToMainOf(t, g, themeSeat)

	castThemeSpell(t, g, me, hand["Smuggler's Copter"])
	copter := hand["Smuggler's Copter"]
	if !onBattlefield(g, copter) {
		t.Fatal("Smuggler's Copter never reached the battlefield")
	}
	if isCreatureNow(g, copter) {
		t.Error("an uncrewed Vehicle is a creature (CR 301.7 — it is not, until a crew ability resolves)")
	}
	// Flying rode in on the Scryfall record's `keywords` array rather
	// than the catalog entry: #317 / #319 / #320 put printed keywords
	// on the import road precisely so a card the catalog has never
	// heard of still flies.
	if c, _ := battlefieldCardByID(g, copter); !game.HasKeyword(&c, "flying") {
		t.Error("the Copter's printed flying did not survive deck.ToGameCard")
	}

	castThemeSpell(t, g, me, hand["History of Benalia"])
	saga := hand["History of Benalia"]
	if got := counterOn(g, saga, game.CounterLore); got != 1 {
		t.Fatalf("lore counters on the turn it entered = %d, want 1 (CR 714.2b)", got)
	}
	if got := len(knightIDs(g)); got != 1 {
		t.Fatalf("Knights after chapter I = %d, want 1", got)
	}

	// A creature that arrived this turn may still crew: tapping to
	// crew is not paying a {T} cost (CR 702.122b). This is the rule
	// implementations most often get wrong in the strict direction.
	firstKnight := knightIDs(g)[0]
	if err := g.ActivateCatalogAbility(me.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{firstKnight},
	}); err != nil {
		t.Fatalf("crew with a summoning-sick Knight (CR 702.122b): %v", err)
	}
	passPriorityAroundTable(t, g)
	if !isCreatureNow(g, copter) {
		t.Error("the Copter did not become a creature when its crew ability resolved")
	}

	// --- turn 2 ------------------------------------------------------
	//
	// Cleanup took the animation with it; the saga ticks to II on the
	// way through precombat main; and the Copter — on the battlefield
	// since last turn — can attack the moment it is crewed again.
	advanceToStepOf(t, g, themeSeat, game.StepUpkeep)
	if isCreatureNow(g, copter) {
		t.Error("the crew animation survived cleanup (ADR 0035: until-EOT effects expire at CR 514.2)")
	}

	advanceToMainOf(t, g, themeSeat)
	// The chapter ability goes on the stack as the lore counter
	// lands (CR 714.2b); it needs a trip round the table to resolve.
	passPriorityAroundTable(t, g)
	if got := counterOn(g, saga, game.CounterLore); got != 2 {
		t.Fatalf("lore counters after the second precombat main = %d, want 2 (CR 714.2b)", got)
	}
	if got := len(knightIDs(g)); got != 2 {
		t.Fatalf("Knights after chapter II = %d, want 2", got)
	}

	secondKnight := knightIDs(g)[1]
	if err := g.ActivateCatalogAbility(me.ID, copter, 0, game.ActivateAbilityParams{
		CrewIDs: []uuid.UUID{secondKnight},
	}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)

	advanceToStepOf(t, g, themeSeat, game.StepDeclareAttackers)
	beforeLife := opponent.Life
	beforeHand := len(me.Hand.Cards)
	beforeYard := len(me.Graveyard.Cards)
	if err := g.DeclareAttacker(copter, opponent.ID); err != nil {
		t.Fatalf("a Vehicle crewed the turn after it entered could not attack (CR 302.6): %v", err)
	}
	// "Whenever this Vehicle attacks ... you may draw a card. If you
	// do, discard a card." The draw is immediate; the discard is an
	// obligation the controller discharges by hand (S13.4's
	// DiscardPending pause), so what the trigger leaves behind is one
	// extra card and one card owed.
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != beforeHand+1 {
		t.Errorf("hand size after the Copter's loot = %d, want %d (the draw half)", got, beforeHand+1)
	}
	if got := g.DiscardPending[me.ID]; got != 1 {
		t.Errorf("discards owed after the Copter's loot = %d, want 1", got)
	}
	if got := len(me.Graveyard.Cards); got != beforeYard {
		t.Errorf("graveyard grew to %d before the owed discard was made", got)
	}

	advanceToStepOf(t, g, themeSeat, game.StepCombatDamage)
	if loss := beforeLife - opponent.Life; loss != 3 {
		t.Errorf("Copter combat damage = %d, want 3", loss)
	}

	// --- turn 3 ------------------------------------------------------
	//
	// Chapter III pumps the Knights and, once it has RESOLVED, the
	// CR 704.5s state-based action sacrifices the Saga. Elspeth then
	// enters with the loyalty her Scryfall record printed.
	advanceToMainOf(t, g, themeSeat)
	passPriorityAroundTable(t, g)
	// Chapter III both fired and finished: the pump is the evidence
	// it resolved, and the empty battlefield slot is the evidence
	// CR 704.5s waited for that resolution before sacrificing.
	if got := effectivePower(t, g, firstKnight); got != 4 {
		t.Errorf("Knight power after chapter III = %d, want 4 (2 printed + 2)", got)
	}
	if onBattlefield(g, saga) {
		t.Error("a Saga at its final chapter with nothing on the stack survived (CR 704.5s)")
	}
	if !inGraveyardOf(g, me.ID, saga) {
		t.Error("the sacrificed Saga is not in its owner's graveyard")
	}

	castThemeSpell(t, g, me, hand["Elspeth, Sun's Champion"])
	elspeth := hand["Elspeth, Sun's Champion"]
	if got := counterOn(g, elspeth, game.CounterLoyalty); got != 4 {
		t.Fatalf("Elspeth's loyalty counters on entry = %d, want 4 — printed loyalty did not "+
			"survive deck.ToGameCard (#274)", got)
	}

	if err := g.ActivateCatalogAbility(me.ID, elspeth, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Elspeth +1: %v", err)
	}
	if got := counterOn(g, elspeth, game.CounterLoyalty); got != 5 {
		t.Errorf("loyalty after +1 = %d, want 5 — the cost is paid at announce (CR 606.2)", got)
	}
	// CR 606.5: one loyalty ability per planeswalker per turn.
	err := g.ActivateCatalogAbility(me.ID, elspeth, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrLoyaltyAlreadyActivated) {
		t.Errorf("second loyalty activation in one turn returned %v, want ErrLoyaltyAlreadyActivated", err)
	}
	passPriorityAroundTable(t, g)
	if got := countNamed(g, "Soldier"); got != 3 {
		t.Errorf("Soldiers after Elspeth's +1 = %d, want 3", got)
	}

	// --- between turns: an opponent attacks the planeswalker ---------
	//
	// CR 508.1d — a creature may be declared attacking a planeswalker
	// its controller's opponent controls, and the damage comes off as
	// loyalty counters (CR 120.3c). This is the half of #406 that made
	// planeswalkers unkillable by damage.
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Raging Bear",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	oppSeat := (themeSeat + 1) % len(g.Seats)
	advanceToStepOf(t, g, oppSeat, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, elspeth); err != nil {
		t.Fatalf("declare an attacker against a planeswalker (CR 508.1d): %v", err)
	}
	passPriorityAroundTable(t, g)
	advanceToStepOf(t, g, oppSeat, game.StepCombatDamage)
	if got := counterOn(g, elspeth, game.CounterLoyalty); got != 3 {
		t.Fatalf("Elspeth's loyalty after taking 2 combat damage = %d, want 3 (#406)", got)
	}

	// --- turn 4 ------------------------------------------------------
	//
	// The chapter-III pump is long gone, the loyalty gate has
	// reopened, and Elspeth's −3 takes her to zero: the sweep resolves
	// and then CR 704.5i puts her in the graveyard.
	advanceToMainOf(t, g, themeSeat)
	if got := effectivePower(t, g, firstKnight); got != 2 {
		t.Errorf("Knight power a turn after chapter III = %d, want 2 (the pump was until end of turn)", got)
	}

	castThemeSpell(t, g, me, hand["The Wandering Emperor"])
	emperor := hand["The Wandering Emperor"]
	if got := counterOn(g, emperor, game.CounterLoyalty); got != 3 {
		t.Errorf("The Wandering Emperor's loyalty on entry = %d, want 3", got)
	}

	ogre := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Hulking Ogre",
		TypeLine:   "Creature — Ogre",
		Power:      5,
		Toughness:  5,
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})

	if err := g.ActivateCatalogAbility(me.ID, elspeth, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Elspeth −3 on a new turn (the CR 606.5 gate should have reopened): %v", err)
	}
	passPriorityAroundTable(t, g)

	if onBattlefield(g, ogre) {
		t.Error("a 5/5 survived 'destroy all creatures with power 4 or greater'")
	}
	if !onBattlefield(g, firstKnight) {
		t.Error("a 2/2 Knight was swept by a 'power 4 or greater' sweeper")
	}
	if countNamed(g, "Soldier") != 3 {
		t.Error("the 1/1 Soldiers were swept by a 'power 4 or greater' sweeper")
	}
	if onBattlefield(g, elspeth) {
		t.Error("a planeswalker at zero loyalty survived (CR 704.5i)")
	}
	if !inGraveyardOf(g, me.ID, elspeth) {
		t.Error("the dead planeswalker is not in its owner's graveyard")
	}
}
