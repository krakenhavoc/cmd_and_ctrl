package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// permission_per_face_test.go — #4696, the enumerator half: two live
// permissions over one adventurer card in the graveyard, one opening the
// creature ("creature spells") and one the Adventure ("instant and sorcery
// spells"). The bot is offered both halves (CR 601.3, 715.3a), each under
// the permission that opens it, whichever is found first, and CastSpell
// accepts every move offered (dispatchAll).
func TestEnumeratorOffersEachFaceUnderThePermissionThatOpensIt(t *testing.T) {
	for _, creatureFirst := range []bool{true, false} {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		clearHand(seat)
		advanceTo(t, g, game.StepPrecombatMain)
		grant := func(f game.PermissionFilter) {
			g.WithWriteLock(func() {
				g.GrantCastPermissionForEffect(game.CastPermission{
					Player: seat.ID, Zone: game.ZoneGraveyard, Scope: game.ScopeStanding, Filter: f, Label: "test permission",
				})
			})
		}
		if creatureFirst {
			grant(game.PermissionFilter{CreatureOnly: true})
			grant(game.PermissionFilter{InstantOrSorceryOnly: true})
		} else {
			grant(game.PermissionFilter{InstantOrSorceryOnly: true})
			grant(game.PermissionFilter{CreatureOnly: true})
		}
		c := game.Card{
			InstanceID: uuid.New(), OracleID: "adventure-enum-4696", Layout: game.LayoutAdventure,
			Owner: seat.ID, Controller: seat.ID,
			Faces: []game.Face{
				{Name: "Enum Knight", TypeLine: "Creature — Knight", ManaCost: "{0}", Power: 1, Toughness: 1},
				{Name: "Enum Insight", TypeLine: "Instant — Adventure", ManaCost: "{0}"},
			},
		}
		c.SetFace(0)
		seat.Graveyard.PushTop(c)

		moves := legal.EnumerateFor(g, seat.ID)
		seen := map[int]bool{}
		for _, m := range castMovesFor(moves, c.InstanceID) {
			seen[faceOf(t, m)] = true
		}
		if !seen[0] || !seen[1] {
			t.Fatalf("creature permission first %v: faces offered = %v, want the creature (0) and the Adventure (1)", creatureFirst, seen)
		}
		dispatchAll(t, g, seat.ID, moves)
	}
}
