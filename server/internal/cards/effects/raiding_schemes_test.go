package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// raiding_schemes_test.go — Raiding Schemes and the conspire seam
// (#4690, CR 702.78).

const raidingSchemesOracle = "854cac11-9a22-4dea-8c1b-c49a4858a754"

func TestRaidingSchemesIsRegistered(t *testing.T) {
	spec, ok := Lookup(raidingSchemesOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

// rsColoredCreature puts an untapped creature of the given colours onto
// `owner`'s battlefield.
func rsColoredCreature(g *game.Game, owner uuid.UUID, name string, colors ...string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Test",
			Colors: colors, Power: 2, Toughness: 2, Owner: owner, Controller: owner})
	})
	return id
}

// rsTable is a main-phase table with Raiding Schemes under the active
// seat's control and a Lightning Bolt in their hand.
func rsTable(t *testing.T) (g *game.Game, me *game.Player, bolt uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, bolt = sqBoltInHand(t, g)
	pushCatalogPermanent(g, me.ID, "Raiding Schemes", "Enchantment", raidingSchemesOracle, false)
	return g, me, bolt
}

func rsTapped(g *game.Game, id uuid.UUID) bool {
	tapped := false
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				tapped = c.Tapped
			}
		}
	})
	return tapped
}

// A red instant conspired with two red creatures is copied, and the copy
// may be given a new target (CR 702.78a, 707.10c).
func TestConspireCopiesARedInstant(t *testing.T) {
	g, me, bolt := rsTable(t)
	opp := g.Seats[1]
	a := rsColoredCreature(g, me.ID, "Goblin A", "R")
	b := rsColoredCreature(g, me.ID, "Goblin B", "R", "G")
	before := opp.Life

	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		TeamworkIDs:   []uuid.UUID{a, b},
	}); err != nil {
		t.Fatalf("cast with conspire: %v", err)
	}
	if !rsTapped(g, a) || !rsTapped(g, b) {
		t.Fatal("the two creatures were not tapped as the spell was cast")
	}
	if triggerOnStack(g, bolt) == nil {
		t.Fatal("no conspire trigger on the stack")
	}
	if prompts := sqResolveAll(t, g, me); prompts == 0 {
		t.Error("the copy never offered new targets")
	}
	if opp.Life != before-6 {
		t.Errorf("opponent lost %d, want 6 (the Bolt and its copy)", before-opp.Life)
	}
}

// Creatures that share no colour with the spell cannot pay.
func TestConspireRefusesCreaturesThatShareNoColor(t *testing.T) {
	g, me, bolt := rsTable(t)
	opp := g.Seats[1]
	red := rsColoredCreature(g, me.ID, "Goblin", "R")
	green := rsColoredCreature(g, me.ID, "Elf", "G")
	colorless := rsColoredCreature(g, me.ID, "Golem")
	for _, pair := range [][]uuid.UUID{{red, green}, {red, colorless}} {
		if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
			Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
			OptionalCosts: []int{0},
			TeamworkIDs:   pair,
		}); err == nil {
			t.Fatalf("conspire paid with a creature that shares no colour with the spell")
		}
	}
	if rsTapped(g, red) || rsTapped(g, green) {
		t.Error("a refused cast tapped a creature")
	}
	// One creature is not two.
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		TeamworkIDs:   []uuid.UUID{red},
	}); err == nil {
		t.Error("conspire paid with one creature")
	}
}

// No Raiding Schemes, no conspire; and a creature spell is not given it.
func TestConspireIsNotGrantedWithoutSchemesOrToACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, bolt := sqBoltInHand(t, g)
	opp := g.Seats[1]
	a := rsColoredCreature(g, me.ID, "Goblin A", "R")
	b := rsColoredCreature(g, me.ID, "Goblin B", "R")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		OptionalCosts: []int{0},
		TeamworkIDs:   []uuid.UUID{a, b},
	}); err == nil {
		t.Error("conspire was paid with no Raiding Schemes")
	}

	pushCatalogPermanent(g, me.ID, "Raiding Schemes", "Enchantment", raidingSchemesOracle, false)
	ogre := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: ogre, Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{2}{R}",
			Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, ogre, game.CastSpellParams{OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{a, b}}); err == nil {
		t.Error("a creature spell was given conspire")
	}
}

// A conspired permanent spell's copy becomes a token (CR 707.10f).
func TestConspiredEnchantmentMakesATokenCopy(t *testing.T) {
	g, me, _ := rsTable(t)
	a := rsColoredCreature(g, me.ID, "Goblin A", "R")
	b := rsColoredCreature(g, me.ID, "Goblin B", "R")
	ench := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: ench, Name: "Red Banner", TypeLine: "Enchantment", ManaCost: "{1}{R}",
			Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, ench, game.CastSpellParams{OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("cast with conspire: %v", err)
	}
	sqResolveAll(t, g, me)
	cards, tokens := 0, 0
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name != "Red Banner" {
				continue
			}
			if c.IsToken() {
				tokens++
			} else {
				cards++
			}
		}
	})
	if cards != 1 || tokens != 1 {
		t.Errorf("Red Banners on the battlefield: %d card(s), %d token(s); want 1 and 1", cards, tokens)
	}
}

// Unpaid, there is no trigger and no copy.
func TestBoltWithoutConspireIsNotCopied(t *testing.T) {
	g, me, bolt := rsTable(t)
	opp := g.Seats[1]
	rsColoredCreature(g, me.ID, "Goblin A", "R")
	rsColoredCreature(g, me.ID, "Goblin B", "R")
	before := opp.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if triggerOnStack(g, bolt) != nil {
		t.Error("a conspire trigger with no conspire paid")
	}
	sqResolveAll(t, g, me)
	if opp.Life != before-3 {
		t.Errorf("opponent lost %d, want 3", before-opp.Life)
	}
}

// The move enumerator offers the conspire cast, naming two creatures that
// share a colour with the spell, only when the seat has them.
func TestRaidingSchemesConspireIsEnumerated(t *testing.T) {
	g, me, bolt := rsTable(t)
	floatMana(t, g, me, "{R}")
	a := rsColoredCreature(g, me.ID, "Goblin", "R")
	rsColoredCreature(g, me.ID, "Elf", "G")
	conspireMoves := func() [][]uuid.UUID {
		var out [][]uuid.UUID
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != "cast_spell" || m.Source != bolt {
				continue
			}
			var p struct {
				OptionalCosts []int       `json:"optional_costs"`
				TeamworkIDs   []uuid.UUID `json:"teamwork_ids"`
				ConspireIDs   []uuid.UUID `json:"conspire_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			// #5688: the two creatures ride conspire_ids, and nothing is
			// on teamwork_ids for a spell with no tap cost of its own.
			if len(p.TeamworkIDs) > 0 {
				t.Errorf("a Bolt names %v on teamwork_ids", p.TeamworkIDs)
			}
			if len(p.OptionalCosts) > 0 {
				out = append(out, p.ConspireIDs)
			}
		}
		return out
	}
	if moves := conspireMoves(); len(moves) != 0 {
		t.Errorf("%d conspire moves with one red creature", len(moves))
	}
	b := rsColoredCreature(g, me.ID, "Goblin B", "R")
	moves := conspireMoves()
	if len(moves) == 0 {
		t.Fatal("no conspire move with two red creatures")
	}
	for _, ids := range moves {
		if len(ids) != 2 || !((ids[0] == a && ids[1] == b) || (ids[0] == b && ids[1] == a)) {
			t.Errorf("conspire move names %v, want the two red creatures", ids)
		}
	}
}
