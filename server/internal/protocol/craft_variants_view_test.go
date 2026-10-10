package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// craft_variants_view_test.go — ADR 0137's 2026-10-10 amendment
// (#2709): the exile-a-permanent view carries craft's variants in the
// shapes the picker reads — an open count as min N / max 0, the
// one-of-each rule as each_of, and "share a card type" as shares.

func seatCraftMaterial(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Power: 1, Toughness: 1,
		Owner: owner, Controller: owner,
	})
	return id
}

func TestExilePermanentViewCarriesCraftVariants(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatCraftMaterial(g, owner, "Source", "Artifact")
	dino := seatCraftMaterial(g, owner, "Dino", "Creature — Dinosaur")
	merfolk := seatCraftMaterial(g, owner, "Merrow", "Creature — Merfolk")
	rock := seatCraftMaterial(g, owner, "Rock", "Artifact")

	open := &game.ExilePermanentsCost{Count: 1, OrMore: true, ExcludeSource: true, Label: "one or more"}
	eachOf := &game.ExilePermanentsCost{Count: 2, EachSubtype: []string{"Dinosaur", "Merfolk"}, ExcludeSource: true, Label: "a Dinosaur and a Merfolk"}
	shares := &game.ExilePermanentsCost{Count: 2, ShareCardType: true, ExcludeSource: true, Label: "two that share a card type"}

	g.ReadSnapshot(func() {
		v := exilePermanentCostOptions(g, owner, src, open)
		if v.Min != 1 || v.Max != 0 || len(v.Cards) != 3 {
			t.Errorf("open count: min %d max %d cards %v, want 1 / 0 over the three others", v.Min, v.Max, v.Cards)
		}

		v = exilePermanentCostOptions(g, owner, src, eachOf)
		if v.Min != 2 || v.Max != 2 || len(v.Cards) != 2 || len(v.EachOf) != 2 {
			t.Fatalf("each-of: %+v, want min/max 2 over the Dinosaur and the Merfolk, two groups", v)
		}
		if v.EachOf[0].Label != "a Dinosaur" || len(v.EachOf[0].Cards) != 1 || v.EachOf[0].Cards[0] != dino.String() {
			t.Errorf("first group = %+v, want the Dinosaur", v.EachOf[0])
		}
		if v.EachOf[1].Label != "a Merfolk" || len(v.EachOf[1].Cards) != 1 || v.EachOf[1].Cards[0] != merfolk.String() {
			t.Errorf("second group = %+v, want the Merfolk", v.EachOf[1])
		}

		v = exilePermanentCostOptions(g, owner, src, shares)
		if v.Shares == nil {
			t.Fatal("share rule: no shares on the view")
		}
		if got := v.Shares.Keys[dino.String()]; len(got) != 1 || got[0] != "creature" {
			t.Errorf("Dino's keys = %v, want [creature]", got)
		}
		if got := v.Shares.Keys[rock.String()]; len(got) != 0 {
			t.Errorf("Rock's keys = %v, want none: no other artifact shares with it", got)
		}
	})
}
