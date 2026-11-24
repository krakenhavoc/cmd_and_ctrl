package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b569_helpers.go — Guildpact's haunt instants and sorceries (OLD
// batch 569, #5599; haunt #6083). Each prints its effect twice, once as the
// spell and once as "When the creature this card haunts dies, …", so the
// body is written once here and both halves call it.

// GainOneLifeForEachPlayer is "You gain 1 life for each player" (Benediction
// of Moons): every player still in the game counts, the controller among
// them, and one who has left the game does not. Counted as it resolves.
func GainOneLifeForEachPlayer(g *game.Game, item *game.StackItem) error {
	n := 0
	for _, p := range g.Seats {
		if p != nil && !p.Eliminated {
			n++
		}
	}
	return GainLife{Amount: n}.Apply(NewContext(g, item))
}

// TargetNonwhiteNonblackCreature is Seize the Soul's target clause, the
// spell's and the haunt trigger's.
func TargetNonwhiteNonblackCreature() *game.TargetSpec {
	return TargetCreature("target nonwhite, nonblack creature", And(Not(OfColor("W")), NonBlack()))
}

// DestroyTargetAndMakeAWhiteSpirit is "Destroy target nonwhite, nonblack
// creature. Create a 1/1 white Spirit creature token with flying." (Seize
// the Soul). With its one target illegal the spell or ability does nothing
// at all (CR 608.2b), so the token comes only with a legal target; one that
// is legal but survives (regeneration, indestructible) still gives it.
func DestroyTargetAndMakeAWhiteSpirit(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			break
		}
	}
	return CreateToken{Template: B28WhiteSpiritFlyingToken(), N: 1}.Apply(ctx)
}
