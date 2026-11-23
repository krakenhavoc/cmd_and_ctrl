package effectstest

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// retraceTable is a main phase on seat 0's turn with two Forests on the
// battlefield and a hand holding one land and one spell.
func RetraceTable(t *testing.T) (g *game.Game, me *game.Player, land, spare uuid.UUID) {
	t.Helper()
	g = NewCatalogGame(t)
	AdvanceToMain(t, g)
	me = g.Seats[0]
	me.Hand.Cards = nil
	land = uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Hand Forest", TypeLine: "Basic Land — Forest",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	spare = uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spare, Name: "Spare Spell", TypeLine: "Sorcery", ManaCost: "{2}",
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	return
}

func AddForests(g *game.Game, me *game.Player, n int) {
	for i := 0; i < n; i++ {
		PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	}
}

func InGraveyard(p *game.Player, id uuid.UUID) bool { return p.Graveyard.Contains(id) }

func GraveyardCardOf(g *game.Game, p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Owner: p.ID, Controller: p.ID,
		KnownBy: map[uuid.UUID]bool{p.ID: true}})
	return id
}

func PayMana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
}

func RetraceCast(g *game.Game, p *game.Player, card uuid.UUID, discard ...uuid.UUID) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "retrace", AltCostIDs: discard, Strict: true,
	})
}

type retraceMove struct {
	AlternativeCost string   `json:"alternative_cost"`
	FromZone        string   `json:"from_zone"`
	AltCostIDs      []string `json:"alt_cost_ids"`
}

// retraceMoves is every enumerated cast of `card` out of the graveyard.
func RetraceMoves(g *game.Game, seat, card uuid.UUID) []retraceMove {
	var out []retraceMove
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeCastSpell || m.Source != card {
			continue
		}
		var p retraceMove
		if err := json.Unmarshal(m.Params, &p); err != nil {
			continue
		}
		if p.FromZone == "graveyard" {
			out = append(out, p)
		}
	}
	return out
}

func GraveyardViewOf(t *testing.T, g *game.Game, viewer, card uuid.UUID) protocol.CardView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, viewer.String())
	for _, s := range v.Seats {
		for _, c := range s.Graveyard.Cards {
			if c.InstanceID == card.String() {
				return c
			}
		}
	}
	t.Fatalf("card %s is not in any graveyard view", card)
	return protocol.CardView{}
}

// ---------------------------------------------------------------- Six

func SixTable(t *testing.T) (g *game.Game, me *game.Player, six, land uuid.UUID) {
	t.Helper()
	g, me, land, _ = RetraceTable(t)
	AddForests(g, me, 4)
	six = PushCatalogPermanent(g, me.ID, "Six", "Legendary Creature — Treefolk", SixOracle, false)
	return
}
