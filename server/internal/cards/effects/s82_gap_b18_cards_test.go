package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b18_cards_test.go — Mystery Booster 2 artifacts and creatures, GAP batch 18 (#5948).

const (
	gapThopterFoundry = "88bef744-550e-4f33-b1ff-a8ee990ec754"
	gapEaterOfTheDead = "3c414ace-fd04-4fac-9379-24f5910f20e4"
	gapDakmorSorc     = "be5d9522-40b4-4f45-978b-d78bd00f9491"
	gapDeathsShadow   = "e08722b3-0f79-4c82-b298-603e04a37857"
)

func TestS82GapB18EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapThopterFoundry: "Thopter Foundry", gapEaterOfTheDead: "Eater of the Dead",
		gapDakmorSorc: "Dakmor Sorceress", gapDeathsShadow: "Death's Shadow",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82GapB18ThopterFoundrySacrificesANontokenArtifactForAThopter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	foundry := b12Push(g, me.ID, "Thopter Foundry", "Artifact", gapThopterFoundry, 0, 0)
	trinket := b12Permanent(g, me.ID, "Trinket", "Artifact")
	tokenArt := b12Permanent(g, me.ID, "Clue", "Token Artifact — Clue")
	advanceToMain(t, g)

	b06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{tokenArt}}); err == nil {
		t.Fatal("sacrificed a token artifact to a nontoken cost")
	}
	if !g.Battlefield.Contains(tokenArt) {
		t.Fatal("the refused activation still sacrificed the token")
	}
	life := me.Life
	b16Activate(t, g, me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{trinket}})
	if g.Battlefield.Contains(trinket) {
		t.Error("the Trinket was not sacrificed")
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Thopter" && c.IsToken() {
			n++
			if !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Error("the Thopter lacks flying")
			}
		}
	}
	if n != 1 {
		t.Errorf("Thopter tokens = %d, want 1", n)
	}
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestS82GapB18EaterOfTheDeadExilesOnlyWhileTappedThenUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eater := b12Push(g, me.ID, "Eater of the Dead", "Creature — Horror", gapEaterOfTheDead, 3, 4)
	dead := b38Graveyard(g, opp.ID, "Dead Bear", "Creature — Bear", "{1}{G}", 2, 2)
	land := b38Graveyard(g, opp.ID, "Dead Land", "Land", "", 0, 0)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, eater, 0, game.ActivateAbilityParams{Targets: b16TargetCard(land), Strict: true}); err == nil {
		t.Fatal("a land card is not a creature card")
	}
	// Untapped: the ability resolves and does nothing.
	b16Activate(t, g, me.ID, eater, 0, game.ActivateAbilityParams{Targets: b16TargetCard(dead), Strict: true})
	if !opp.Graveyard.Contains(dead) {
		t.Error("an untapped Eater exiled the card")
	}
	// Tapped: exile the creature card and untap.
	b16Tap(g, eater)
	b16Activate(t, g, me.ID, eater, 0, game.ActivateAbilityParams{Targets: b16TargetCard(dead), Strict: true})
	if opp.Graveyard.Contains(dead) || !inExile(g, dead) {
		t.Error("the tapped Eater should exile the creature card")
	}
	if b16Tapped(t, g, eater) {
		t.Error("the Eater should untap")
	}
}

func TestS82GapB18DakmorSorceressCountsSwamps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s := b12Push(g, me.ID, "Dakmor Sorceress", "Creature — Human Wizard Sorcerer", gapDakmorSorc, 0, 4)
	if p, to := effectivePTOf(t, g, s); p != 0 || to != 4 {
		t.Errorf("no Swamps: %d/%d, want 0/4", p, to)
	}
	b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, opp.ID, "Swamp", "Basic Land — Swamp")
	if p, to := effectivePTOf(t, g, s); p != 2 || to != 4 {
		t.Errorf("two own Swamps: %d/%d, want 2/4", p, to)
	}
}

func TestS82GapB18DeathsShadowShrinksWithYourLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	d := b12Push(g, me.ID, "Death's Shadow", "Creature — Avatar", gapDeathsShadow, 13, 13)
	gapSetLife(g, me, 5)
	if p, to := effectivePTOf(t, g, d); p != 8 || to != 8 {
		t.Errorf("at 5 life: %d/%d, want 8/8", p, to)
	}
	gapSetLife(g, me, 20)
	if p, to := effectivePTOf(t, g, d); p != -7 || to != -7 {
		t.Errorf("at 20 life: %d/%d, want -7/-7", p, to)
	}
}

// gapSetLife moves a life total through the engine, so the layer cache is dropped.
func gapSetLife(g *game.Game, p *game.Player, to int) {
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, to-p.Life)
	})
}
