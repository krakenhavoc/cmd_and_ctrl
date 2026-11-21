package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// bro_prototype_cards_test.go — the eleven Brothers' War prototype cards
// (#5656), each cast both ways through the engine, and each clause that
// reads "this creature's power" checked against both sizes.

const ProtoReminder = " (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)"

// protoRow is one card as printed, with what its inset says.
type ProtoRow struct {
	Oracle, Name, TypeLine, Cost string
	Power, Toughness             int
	Text                         string

	InsetCost, InsetColor      string
	InsetPower, InsetToughness int
	Keywords                   []string
}

func (r ProtoRow) Card(t *testing.T, owner uuid.UUID) game.Card {
	return PrototypeCard(t, owner, r.Oracle, r.Name, r.TypeLine, r.Cost, r.Power, r.Toughness, r.Text)
}

// protoManaFor is exactly the mana a cost asks for, generic paid with {C}.
func ProtoManaFor(t *testing.T, cost string) string {
	t.Helper()
	parsed, err := game.ParseCost(cost)
	if err != nil {
		t.Fatalf("ParseCost(%s): %v", cost, err)
	}
	rest := cost
	if i := strings.Index(cost, "}"); parsed.Generic > 0 && i >= 0 {
		rest = cost[i+1:]
	}
	return strings.Repeat("{C}", parsed.Generic) + rest
}

// castProtoRow casts the card from the active player's hand at their
// main phase, prototyped or not, paying exactly its cost, and resolves
// it. It leaves any prompt the entry opened unanswered.
func CastProtoRow(t *testing.T, g *game.Game, r ProtoRow, prototyped bool) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	c := r.Card(t, me.ID)
	me.Hand.PushTop(c)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	params := game.CastSpellParams{Strict: true}
	cost := r.Cost
	if prototyped {
		params.Face = game.PrototypeFace
		cost = r.InsetCost
	}
	me.ManaPool = nil
	FloatMana(t, g, me, ProtoManaFor(t, cost))
	if err := g.CastSpell(me.ID, c.InstanceID, params); err != nil {
		t.Fatalf("cast %s (prototyped=%v): %v", r.Name, prototyped, err)
	}
	if left := me.ManaPool.Total(); left != 0 {
		t.Fatalf("%s (prototyped=%v) left %d mana floating, want its whole cost spent", r.Name, prototyped, left)
	}
	PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(c.InstanceID) {
		t.Fatalf("%s did not resolve onto the battlefield", r.Name)
	}
	return c.InstanceID
}

// pushProtoRow puts the card on the battlefield as the permanent a cast
// would have made, untapped and able to attack and tap.
func PushProtoRow(t *testing.T, g *game.Game, owner uuid.UUID, r ProtoRow, prototyped bool) uuid.UUID {
	t.Helper()
	c := r.Card(t, owner)
	if prototyped {
		c.SetFace(game.PrototypeFace)
	}
	return PushBattlefieldCardWithTimestamp(g, c)
}
