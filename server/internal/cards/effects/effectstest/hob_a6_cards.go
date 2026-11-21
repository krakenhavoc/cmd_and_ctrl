package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Adventure plumbing -----------------------------------------------

func HobA6Face(name, typeLine, cost string, power, toughness int, colors ...string) game.Face {
	return game.Face{Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors, Power: power, Toughness: toughness}
}

// hobA6HandAdventure seats an Adventure card (creature face 0, Adventure
// face 1) in the active player's hand at a main phase.
func HobA6HandAdventure(t *testing.T, g *game.Game, oracle string, creature, adventure game.Face) (*game.Player, uuid.UUID) {
	t.Helper()
	AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	c := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutAdventure,
		Owner: me.ID, Controller: me.ID, Faces: []game.Face{creature, adventure},
	}
	c.SetFace(0)
	me.Hand.PushTop(c)
	return me, c.InstanceID
}

func HobA6Cast(g *game.Game, me *game.Player, id uuid.UUID, face int, targets []game.TargetRef, sacrifice ...uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{Face: face, Targets: targets, SacrificeIDs: sacrifice})
}

func HobA6MustCast(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, face int, targets []game.TargetRef, sacrifice ...uuid.UUID) {
	t.Helper()
	if err := HobA6Cast(g, me, id, face, targets, sacrifice...); err != nil {
		t.Fatalf("cast face %d: %v", face, err)
	}
	PassPriorityAroundTable(t, g)
}

// hobA6WentOnAnAdventure asserts CR 715.3d / 715.4: the card is in exile
// and its owner may cast the creature face from there.
func HobA6WentOnAnAdventure(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if !g.Exile.Contains(id) {
		t.Fatal("CR 715.3d: the resolved Adventure is not in exile")
	}
	perm := g.CastPermissionOnCardByIDForEffect(id)
	if face, ok := perm.NamedFace(); !ok || face != 0 {
		t.Errorf("grant faces = %v, want just the creature face [0]", perm.Faces)
	}
}

func HobA6Card(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// --- Gollum, Silent Slinker // Meager Meal ----------------------------

var (
	HobA6GollumCreature = HobA6Face("Gollum, Silent Slinker", "Legendary Creature — Halfling Horror", "{3}{B}", 4, 3, "B")
	HobA6MeagerMeal     = HobA6Face("Meager Meal", "Sorcery — Adventure", "{B}", 0, 0, "B")
)
