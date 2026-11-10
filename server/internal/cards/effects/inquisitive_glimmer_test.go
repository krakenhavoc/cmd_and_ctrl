package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const inquisitiveGlimmerOracle = "4abb227a-534e-4300-9009-065a10f10e8a"

func TestInquisitiveGlimmerIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(inquisitiveGlimmerOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Inquisitive Glimmer: %+v %v", spec, ok)
	}
}

func TestInquisitiveGlimmerDiscountsOnlyYourEnchantmentSpells(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost, pool string
		wantErr                    bool
	}{
		{"enchantment is {1} cheaper", "Enchantment", "{1}{W}", "W", false},
		{"enchantment at full price", "Enchantment", "{1}{W}", "", true},
		{"a creature is not discounted", "Creature — Bear", "{1}{G}", "G", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			pushCatalogPermanent(g, me.ID, "Inquisitive Glimmer", "Enchantment Creature — Fox Glimmer", inquisitiveGlimmerOracle, false)
			advanceToMain(t, g)
			floatForTest(g, me, tc.pool)
			id := uuid.New()
			g.WithWriteLock(func() {
				me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test spell", TypeLine: tc.typeLine, ManaCost: tc.cost, Owner: me.ID, Controller: me.ID})
			})
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
			if (err != nil) != tc.wantErr {
				t.Fatalf("cast err = %v, want error %v", err, tc.wantErr)
			}
		})
	}
}

func inquisitiveGlimmerRoom(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Owner: owner, Controller: owner, Layout: game.LayoutSplit,
		Faces: []game.Face{
			{Name: "Left", TypeLine: "Enchantment — Room", ManaCost: "{W}"},
			{Name: "Right", TypeLine: "Enchantment — Room", ManaCost: "{2}{U}"},
		},
		Unlocked: game.DoorBit(game.DoorLeft),
	}
	c.SettleImported()
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func TestInquisitiveGlimmerMakesUnlockingCostOneLess(t *testing.T) {
	for _, tc := range []struct {
		name    string
		glimmer bool
		pool    string
		wantErr bool
	}{
		{"with the Glimmer {1}{U} unlocks", true, "UU", false},
		{"with the Glimmer {U} alone is not enough", true, "U", true},
		{"without it {2}{U} is the price", false, "UU", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			if tc.glimmer {
				pushCatalogPermanent(g, me.ID, "Inquisitive Glimmer", "Enchantment Creature — Fox Glimmer", inquisitiveGlimmerOracle, false)
			}
			room := inquisitiveGlimmerRoom(g, me.ID)
			advanceToMain(t, g)
			floatForTest(g, me, tc.pool)
			err := g.PerformSpecialAction(me.ID, room, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorRight, Strict: true})
			if (err != nil) != tc.wantErr {
				t.Fatalf("unlock err = %v, want error %v", err, tc.wantErr)
			}
		})
	}
}
