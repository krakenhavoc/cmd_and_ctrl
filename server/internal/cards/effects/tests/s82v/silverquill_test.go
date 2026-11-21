package testss82v

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// silverquill_test.go — Silverquill, the Disputant and the granted
// casualty seam (#3245).

const mcSilverquillOracle = "66deb3f3-0314-427e-9cb8-6befa365ae94"

func TestSilverquillIsRegisteredAndWhole(t *testing.T) {
	spec, ok := effects.Lookup(mcSilverquillOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

func TestSilverquillGivesAnInstantCasualtyAndCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bolt := effectstest.SqBoltInHand(t, g)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	before := opp.Life

	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{bear},
	}); err != nil {
		t.Fatalf("cast with casualty: %v", err)
	}
	if g.Battlefield.Contains(bear) {
		t.Fatal("the creature was not sacrificed as the spell was cast")
	}
	if it := effectstest.TriggerOnStack(g, bolt); it == nil {
		t.Fatal("no casualty trigger on the stack")
	}
	effectstest.SqResolveAll(t, g, me)
	if opp.Life != before-6 {
		t.Errorf("opponent lost %d, want 6 (the Bolt and its copy)", before-opp.Life)
	}
}

func TestSilverquillCanPayItsOwnCasualty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bolt := effectstest.SqBoltInHand(t, g)
	opp := g.Seats[1]
	sq := effectstest.PushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	before := opp.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{sq},
	}); err != nil {
		t.Fatalf("cast sacrificing Silverquill: %v", err)
	}
	if g.Battlefield.Contains(sq) {
		t.Fatal("Silverquill was not sacrificed")
	}
	effectstest.SqResolveAll(t, g, me)
	if opp.Life != before-6 {
		t.Errorf("opponent lost %d, want 6: the spell is copied though Silverquill left", before-opp.Life)
	}
}

func TestCasualtyIsRefusedWithoutSilverquillOrOnACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bolt := effectstest.SqBoltInHand(t, g)
	opp := g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	params := game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{bear},
	}
	if err := g.CastSpell(me.ID, bolt, params); err == nil {
		t.Error("casualty was paid with no Silverquill")
	}

	effectstest.PushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	ogre := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: ogre, Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}",
			Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, ogre, game.CastSpellParams{OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a creature spell was given casualty")
	}
}

func TestCasualtyOneRefusesAPowerZeroCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bolt := effectstest.SqBoltInHand(t, g)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	wall := effectstest.PushVanillaCreature(g, me.ID, "Wall", 0, 4)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{wall},
	}); err == nil {
		t.Error("a 0-power creature paid casualty 1")
	}
}

func TestBoltWithoutCasualtyIsNotCopied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bolt := effectstest.SqBoltInHand(t, g)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	before := opp.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if effectstest.TriggerOnStack(g, bolt) != nil {
		t.Error("a casualty trigger with no casualty paid")
	}
	effectstest.SqResolveAll(t, g, me)
	if opp.Life != before-3 {
		t.Errorf("opponent lost %d, want 3", before-opp.Life)
	}
}

// The move enumerator offers the casualty cast only under Silverquill.
func TestSilverquillCasualtyIsEnumerated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bolt := effectstest.SqBoltInHand(t, g)
	effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}")
	casualtyMoves := func() int {
		n := 0
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != "cast_spell" || m.Source != bolt {
				continue
			}
			var p struct {
				OptionalCosts []int `json:"optional_costs"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			if len(p.OptionalCosts) > 0 {
				n++
			}
		}
		return n
	}
	if n := casualtyMoves(); n != 0 {
		t.Errorf("%d casualty moves with no Silverquill", n)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	if n := casualtyMoves(); n == 0 {
		t.Error("no casualty move under Silverquill")
	}
}
