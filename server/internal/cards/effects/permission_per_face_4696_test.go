package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// permission_per_face_4696_test.go — #4696 at the card level: Vizier of the
// Menagerie's "creature spells from the top of your library" and an
// "instant and sorcery spells from the top of your library" permission
// over Bonecrusher Giant on top. Both halves are castable (CR 601.3), the
// Giant through Vizier and Stomp through the other (CR 715.3a), whichever
// of the two the lookup finds first. The engine half is
// game/permission_per_face_test.go.

// vizierAndSpellsFromTop puts Vizier of the Menagerie on the battlefield,
// stores an instant-and-sorcery permission over the top of the library,
// and puts Bonecrusher Giant on top. The stored permission is the one the
// lookup finds first (a stored permission is asked before one a permanent
// gives), so before #4696 the Giant was neither offered nor castable.
func vizierAndSpellsFromTop(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMainForCost(t, g)
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(game.CastPermission{
			Player: me.ID, Zone: game.ZoneLibrary, Scope: game.ScopeStanding, TopOfLibraryOnly: true,
			Filter: game.PermissionFilter{InstantOrSorceryOnly: true},
			Label:  "Cast instant and sorcery spells from the top of your library",
		})
	})
	pushCatalogPermanent(g, me.ID, "Vizier of the Menagerie", "Creature — Snake Cleric",
		"4f890d42-e4f1-4eed-aa6c-718865de8927", false)
	giant := bonecrusherGiant(me.ID)
	g.WithWriteLock(func() { me.Library.PushTop(giant) })
	return g, me, opp, giant.InstanceID
}

func TestVizierAndAnInstantPermissionOfferBothHalvesOfBonecrusherGiant(t *testing.T) {
	g, me, opp, giant := vizierAndSpellsFromTop(t)
	floatForTest(g, me, "RR")
	if faces := castMoveFaces(g, me.ID, giant); !slices.Equal(faces, []int{0, 1}) {
		t.Fatalf("the bot is offered faces %v, want both [0 1]", faces)
	}
	if err := g.CastSpell(me.ID, giant, game.CastSpellParams{
		FromZone: "library", Strict: true, Face: 1,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast Stomp from the top of the library: %v", err)
	}
	if spell, ok := stackCardOf(g, giant); !ok || spell.ActiveFace != 1 || !spell.IsInstant() {
		t.Fatalf("on the stack %+v, want Stomp", spell)
	}
}

// The face only Vizier opens, beside the permission found first.
func TestVizierCastsTheGiantBesideAnInstantPermissionFoundFirst(t *testing.T) {
	g, me, _, giant := vizierAndSpellsFromTop(t)
	floatForTest(g, me, "RRR")
	if err := g.CastSpell(me.ID, giant, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("cast the Giant from the top of the library: %v", err)
	}
	if spell, ok := stackCardOf(g, giant); !ok || spell.ActiveFace != 0 || !spell.IsCreature() {
		t.Fatalf("on the stack %+v, want the Giant", spell)
	}
}
