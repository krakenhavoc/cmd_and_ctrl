package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdSeatFaceDown puts a disguised Museum Nightwatch onto the battlefield
// under owner's control with no entry announced: a face-down permanent
// that has been there since an earlier turn, so the turn tally has not
// seen it enter. Only its controller knows what it is.
func FdSeatFaceDown(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: "Museum Nightwatch", TypeLine: "Creature — Centaur Soldier",
		OracleID: MkmMPNightwatchOracle, Power: 3, Toughness: 2, Owner: owner, Controller: owner,
	}
	c.SetFaceDown(game.FaceDownDisguised)
	// CR 708.5: its controller may look at it, and nobody else.
	c.KnownBy = map[uuid.UUID]bool{owner: true}
	return PushBattlefieldCardWithTimestamp(g, c)
}

// fdTurnUpTiming is the turn_face_up window for `player` on `id`.
func FdTurnUpTiming(g *game.Game, player, id uuid.UUID) bool {
	var ok bool
	g.WithWriteLock(func() {
		c, found := g.LookupCardForEffect(id)
		ok = found && g.SpecialActionTimingOKLocked(player, c, game.SpecialActionTurnFaceUp)
	})
	return ok
}
