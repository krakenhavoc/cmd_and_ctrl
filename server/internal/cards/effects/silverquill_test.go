package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// silverquill_test.go — Silverquill, the Disputant and the granted
// casualty seam (#3245).

const mcSilverquillOracle = "66deb3f3-0314-427e-9cb8-6befa365ae94"

func TestSilverquillIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(mcSilverquillOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

// sqBoltInHand puts a Lightning Bolt in the active seat's hand at its
// main phase and returns it.
func sqBoltInHand(t *testing.T, g *game.Game) (me *game.Player, bolt uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bolt = uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
			OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID})
	})
	return me, bolt
}

// sqResolveAll passes until the stack is empty, keeping each copy's
// target when the copy asks.
func sqResolveAll(t *testing.T, g *game.Game, me *game.Player) (prompts int) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if stackFullyEmpty(g) && latestPickTarget(g, me.ID) == nil {
			return prompts
		}
		if p := latestPickTarget(g, me.ID); p != nil {
			prompts++
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: g.Seats[1].ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
	return prompts
}

func TestSilverquillGivesAnInstantCasualtyAndCopiesIt(t *testing.T) {
	g := newCatalogGame(t)
	me, bolt := sqBoltInHand(t, g)
	opp := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
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
	if it := triggerOnStack(g, bolt); it == nil {
		t.Fatal("no casualty trigger on the stack")
	}
	sqResolveAll(t, g, me)
	if opp.Life != before-6 {
		t.Errorf("opponent lost %d, want 6 (the Bolt and its copy)", before-opp.Life)
	}
}

func TestSilverquillCanPayItsOwnCasualty(t *testing.T) {
	g := newCatalogGame(t)
	me, bolt := sqBoltInHand(t, g)
	opp := g.Seats[1]
	sq := pushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
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
	sqResolveAll(t, g, me)
	if opp.Life != before-6 {
		t.Errorf("opponent lost %d, want 6: the spell is copied though Silverquill left", before-opp.Life)
	}
}

func TestCasualtyIsRefusedWithoutSilverquillOrOnACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, bolt := sqBoltInHand(t, g)
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	params := game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{bear},
	}
	if err := g.CastSpell(me.ID, bolt, params); err == nil {
		t.Error("casualty was paid with no Silverquill")
	}

	pushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
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
	g := newCatalogGame(t)
	me, bolt := sqBoltInHand(t, g)
	opp := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	wall := pushVanillaCreature(g, me.ID, "Wall", 0, 4)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{wall},
	}); err == nil {
		t.Error("a 0-power creature paid casualty 1")
	}
}

func TestBoltWithoutCasualtyIsNotCopied(t *testing.T) {
	g := newCatalogGame(t)
	me, bolt := sqBoltInHand(t, g)
	opp := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	before := opp.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if triggerOnStack(g, bolt) != nil {
		t.Error("a casualty trigger with no casualty paid")
	}
	sqResolveAll(t, g, me)
	if opp.Life != before-3 {
		t.Errorf("opponent lost %d, want 3", before-opp.Life)
	}
}

// The move enumerator offers the casualty cast only under Silverquill.
func TestSilverquillCasualtyIsEnumerated(t *testing.T) {
	g := newCatalogGame(t)
	me, bolt := sqBoltInHand(t, g)
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	floatMana(t, g, me, "{R}")
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
	pushCatalogPermanent(g, me.ID, "Silverquill, the Disputant", "Legendary Creature — Elder Dragon", mcSilverquillOracle, false)
	if n := casualtyMoves(); n == 0 {
		t.Error("no casualty move under Silverquill")
	}
}
