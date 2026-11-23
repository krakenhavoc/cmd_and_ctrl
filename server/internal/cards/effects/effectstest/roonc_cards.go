package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roonc_cards_test.go — #4128 slice c, the Roon of the Hidden Realm
// deck request: Arctic Treeline, Yavimaya Hollow, Extraordinary Journey,
// Lonis Cryptozoologist, Moon-Blessed Cleric, Palinchron, Surprise
// Deployment, Titan of Industry, and the audit of Wood Elves.

const (
	RooncTreelineOracle   = "32fb0096-546c-4633-b6e3-ba3f2d9fa49c"
	RooncHollowOracle     = "53d6113d-acdb-4754-9641-f7991a96c7b9"
	RooncJourneyOracle    = "69288566-3015-45de-b391-6fe079865f0c"
	RooncLonisOracle      = "0ee06ed0-717a-4bfa-b634-d00e013c1f16"
	RooncClericOracle     = "89ed0598-a3ee-4d14-9a35-25e0dd80d9fb"
	RooncPalinchronOracle = "0023888e-7bec-43e0-8dee-d1a4eb94b372"
	RooncDeployOracle     = "61a1748c-e772-4338-9edd-720cd5d2fc12"
	RooncTitanOracle      = "5d8acafe-c13e-43ca-a89e-0cfd12d95662"
	RooncWoodElvesOracle  = "8973bd99-20f8-4867-90ef-50392147ee1b"
)

// --- helpers ---------------------------------------------------------

// rooncHand puts a card in a player's hand and returns its ID.
func RooncHand(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

// rooncCastNow casts a card from the active seat's hand at the CURRENT
// step, without walking to a main phase (Surprise Deployment is cast in
// combat).
func RooncCastNow(g *game.Game, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	me := g.Seats[g.Turn.ActiveSeat]
	id := RooncHand(me, c)
	return id, g.CastSpell(me.ID, id, params)
}

// rooncCastMain casts a card from the active seat's hand in a main phase.
func RooncCastMain(t *testing.T, g *game.Game, c game.Card, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	AdvanceToMain(t, g)
	id, err := RooncCastNow(g, c, params)
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}

func RooncRef(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// rooncDestroy is a bare destroy effect from outside the catalog.
func RooncDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
}

// --- Moon-Blessed Cleric ---------------------------------------------

func RooncClericGame(t *testing.T) (*game.Game, *game.Player, *game.Player, []uuid.UUID) {
	t.Helper()
	g := NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ids := SeedSearchLibrary(me,
		SearchTestLand("Island", "Basic Land — Island"),
		SearchTestLand("Pacifism", "Enchantment — Aura"),
		SearchTestLand("Grizzly Bears", "Creature — Bear"),
		SearchTestLand("Rhystic Study", "Enchantment"),
		SearchTestLand("Sol Ring", "Artifact"),
	)
	RooncCastMain(t, g, game.Card{Name: "Moon-Blessed Cleric", TypeLine: "Creature — Human Elf Cleric",
		OracleID: RooncClericOracle, Power: 3, Toughness: 2}, game.CastSpellParams{})
	PassPriorityAroundTable(t, g)
	return g, me, opp, ids
}

// --- Surprise Deployment ----------------------------------------------

func RooncDeploymentCard() game.Card {
	return game.Card{Name: "Surprise Deployment", TypeLine: "Instant", OracleID: RooncDeployOracle, ManaCost: "{3}{W}"}
}
