package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// any_player_test.go — ADR 0106 §1 decision 8 (owner decision 2): the
// heuristic activates an "Any player may activate this ability" row on
// a permanent it does not control only when the row declares a purpose
// for the activator, and never otherwise.

// anyPlayerCreature is a permanent controlled by `controller` with one
// any-player row carrying `purpose` (nil: none declared).
func anyPlayerCreature(id string, controller int, name string, purpose *protocol.ActivationPurposeView) protocol.CardView {
	c := creature(id, controller, name, 5, 5)
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: "{3}: something. Any player may activate this ability.",
		ManaCost: "{3}", AnyPlayer: true, Purpose: purpose,
	}}
	return c
}

func TestBotReachesAcrossOnlyForADeclaredPurpose(t *testing.T) {
	const xantchaMove = "Xantcha, Sleeper Agent (controlled by Seat1): {3}: Xantcha's controller loses 2 life and you draw a card."
	const ogreMove = "Flailing Ogre (controlled by Seat1): {1}: This creature gets +1/+1 until end of turn."

	t.Run("a declared purpose is taken", func(t *testing.T) {
		xantcha := anyPlayerCreature(cardID(1), 1, "Xantcha, Sleeper Agent",
			&protocol.ActivationPurposeView{Draws: 1, ControllerLosesLife: 2})
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(xantcha, land(cardID(10), 0), land(cardID(11), 0), land(cardID(12), 0)),
			withTurn(3, 0, "postcombat_main"))
		in := input(0, v, passMove(0), activateMove(t, 0, cardID(1), xantchaMove, nil))
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != xantchaMove {
			t.Errorf("chose %q, want the declared-purpose activation on the opponent's Xantcha", got)
		}
	})

	t.Run("no purpose is never chosen", func(t *testing.T) {
		ogre := anyPlayerCreature(cardID(2), 1, "Flailing Ogre", nil)
		// Tapped, so nothing but the purpose rule stands between the
		// generic ActivateBase and the pass threshold.
		ogre.Tapped = true
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
			withBattlefield(ogre, land(cardID(10), 0)),
			withTurn(3, 0, "postcombat_main"))
		in := input(0, v, passMove(0), activateMove(t, 0, cardID(2), ogreMove, nil))
		if got := chose(t, in, decide(t, heuristic.New(), in)); got == ogreMove {
			t.Errorf("pumped an opponent's Flailing Ogre: a row with no purpose must never be chosen")
		}
	})
}

// ADR 0106 §1 amendment 2026-10-07 (#1947): an "Only your opponents may
// activate" row on an opponent's permanent (Clergy of the Holy Nimbus) is
// a legal move for this seat, and the bot still never takes it: it is not
// an any-player row that declares a purpose.
func TestBotNeverTakesAnOpponentsOnlyRow(t *testing.T) {
	const clergyMove = "Clergy of the Holy Nimbus (controlled by Seat1): {1}: This creature can't be regenerated this turn."
	clergy := creature(cardID(3), 1, "Clergy of the Holy Nimbus", 1, 1)
	clergy.Tapped = true
	clergy.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: "{1}: This creature can't be regenerated this turn.",
		ManaCost: "{1}", OpponentsOnly: true,
	}}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(clergy, land(cardID(10), 0)),
		withTurn(3, 0, "postcombat_main"))
	in := input(0, v, passMove(0), activateMove(t, 0, cardID(3), clergyMove, nil))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == clergyMove {
		t.Errorf("activated an opponent's opponents-only row with no declared purpose")
	}
}
