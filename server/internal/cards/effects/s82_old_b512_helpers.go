package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b512_helpers.go — Shards of Alara, "Life gain and loss" (S82 OLD
// batch 512, #5588). Append-only.

// S82OldB512ArtifactCardInYourGraveyard is "an artifact card is in your
// graveyard" (Windwright Mage), as a WhileYouHave condition.
func S82OldB512ArtifactCardInYourGraveyard(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Graveyard == nil {
		return false
	}
	for _, c := range p.Graveyard.Cards {
		if c.IsArtifact() {
			return true
		}
	}
	return false
}

// S82OldB512TargetPlayerLosesOne is "Target player loses 1 life." as an
// activated ability's body (Onyx Goblet, Jund Battlemage): the target is the
// row's only one, and one that has left the game does nothing (CR 608.2b).
func S82OldB512TargetPlayerLosesOne(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	p := TargetedPlayerOf(ctx)
	if p == uuid.Nil {
		return nil
	}
	return GainLife{Player: p, Amount: -1}.Apply(ctx)
}
