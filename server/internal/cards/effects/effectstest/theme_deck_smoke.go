package effectstest

import (
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
//	Saga        entry lore counter (CR 714.3), the precombat-main
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
//	            paid at announce (CR 606.4); once per turn per
//	            planeswalker (CR 606.3) and the gate reopens next
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
func ThemeDeckRow(name, oracleID, typeLine, manaCost string) cards.Card {
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
func ImportThemeCard(row cards.Card, owner uuid.UUID) game.Card {
	c := deck.ToGameCard(row, false)
	c.Owner = owner
	c.Controller = owner
	return c
}

// castThemeSpell casts from the theme seat's hand with the strict
// mana gate and the auto-tapper engaged, then settles the stack.
func CastThemeSpell(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

// counterOn reads one counter kind off a battlefield permanent, or
// -1 when the permanent has left the battlefield.
func CounterOn(g *game.Game, id uuid.UUID, kind string) int {
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
func IsCreatureNow(g *game.Game, id uuid.UUID) bool {
	c, ok := BattlefieldCardByID(g, id)
	return ok && c.IsCreature()
}

// advanceToStepOf walks the step machine to `step` on the named
// seat's turn. Deliberately drives the real machine rather than
// assigning Turn.Step: the saga advance, the untap and the cleanup
// all hang off step entry, and a test that skipped them would be
// asserting against a board no game could produce.
func AdvanceToStepOf(t *testing.T, g *game.Game, seat int, step game.Step) {
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
