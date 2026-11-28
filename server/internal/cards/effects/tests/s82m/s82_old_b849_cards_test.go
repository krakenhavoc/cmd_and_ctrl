package testss82m

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b849_cards_test.go — Ice Age banding cards, S82 OLD batch 849 (#5636),
// played through the real catalog entries.

const (
	o849Wall      = "530514d9-7274-4835-a4b5-6dccef4da758"
	o849Knight    = "5d7747b4-061d-4459-ba80-51361526cbfa"
	o849Captain   = "7160ebce-ee3f-42ac-a3f7-b55f17f2c49f"
	o849Warrior   = "73b70a48-b62c-4fdf-9c2d-c6d6688f9ac8"
	o849Bearer    = "75decba6-d188-4144-8e74-58285be0fe9f"
	o849Phalanx   = "a8563a4f-a23b-4cb5-b6e5-795bf8d74ee6"
	o849Skyknight = "b06db20f-84c4-4e2e-a184-3cd70ea0d10d"
)

func TestS82OldB849CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o849Wall: "Wall of Shields", o849Knight: "Kjeldoran Knight", o849Captain: "Kjeldoran Skycaptain",
		o849Warrior: "Kjeldoran Warrior", o849Bearer: "Shield Bearer", o849Phalanx: "Kjeldoran Phalanx",
		o849Skyknight: "Kjeldoran Skyknight",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB849KeywordsAreRead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	for _, c := range []struct {
		name, tl, oracle string
		p, tough         int
		kws              []string
	}{
		{"Wall of Shields", "Artifact Creature — Wall", o849Wall, 0, 4, []string{"defender", game.KeywordBanding}},
		{"Kjeldoran Knight", "Creature — Human Knight", o849Knight, 1, 1, []string{game.KeywordBanding}},
		{"Kjeldoran Skycaptain", "Creature — Human Soldier", o849Captain, 2, 2, []string{"flying", "first strike", game.KeywordBanding}},
		{"Kjeldoran Warrior", "Creature — Human Warrior", o849Warrior, 1, 1, []string{game.KeywordBanding}},
		{"Shield Bearer", "Creature — Human Soldier", o849Bearer, 0, 3, []string{game.KeywordBanding}},
		{"Kjeldoran Phalanx", "Creature — Human Soldier", o849Phalanx, 2, 5, []string{"first strike", game.KeywordBanding}},
		{"Kjeldoran Skyknight", "Creature — Human Knight", o849Skyknight, 1, 1, []string{"flying", "first strike", game.KeywordBanding}},
	} {
		id := effectstest.B12Push(g, me.ID, c.name, c.tl, c.oracle, c.p, c.tough)
		for _, kw := range c.kws {
			if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
				t.Errorf("%s lacks %s", c.name, kw)
			}
		}
	}
}

func o849Mana(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("mana: %v", err)
		}
	})
}

func TestS82OldB849KjeldoranKnightPumpsPowerAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	k := effectstest.B12Push(g, me.ID, "Kjeldoran Knight", "Creature — Human Knight", o849Knight, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated {1}{W}: +1/+0 without paying")
	}
	o849Mana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate +1/+0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, k); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	o849Mana(t, g, me, "{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, k, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate +0/+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tough := effectstest.EffectiveToughness(t, g, k); tough != 3 {
		t.Errorf("toughness %d, want 3", tough)
	}
	// Repeatable: a second +1/+0 in the same turn.
	o849Mana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("second +1/+0: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, k); p != 3 {
		t.Errorf("power %d after two pumps, want 3", p)
	}
}
