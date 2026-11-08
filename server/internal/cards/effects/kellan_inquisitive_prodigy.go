package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kellan, Inquisitive Prodigy // Tail the Suspect — Legendary Creature —
// Human Faerie Detective {2}{G}{U}, 3/4 // Sorcery — Adventure {G}{U}
// (Murders at Karlov Manor, #4469):
//
//	Kellan — "Flying, vigilance
//	 Whenever Kellan attacks, destroy up to one target artifact. If you
//	 controlled that permanent, draw a card."
//	Tail the Suspect — "Investigate. You may play an additional land
//	 this turn. (Then exile this card. You may cast the creature later
//	 from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"), as Beanstalk Wurm's do; the Adventure lifecycle is the
// engine's (CR 715).
//
// The attack trigger is Azog, Moria's Ruin's: who controlled the
// artifact is read before the destroy, and the draw is the destroy's
// continuation (a commander's command-zone question comes first). The
// draw does not ask whether the artifact was destroyed: an indestructible
// artifact you controlled still draws you a card (the ruling). With no
// target, or a target gone by resolution, nothing happens.
//
// Tail the Suspect makes a Clue, then is Plant Beans' one-shot extra
// land drop, which stacks with others (the ruling).
//
// No simplification.
const kellanInquisitiveProdigyOracleID = "2f2d75a0-d751-4c74-b357-8f92a992447f"

func init() {
	Register(Spec{
		OracleID:        kellanInquisitiveProdigyOracleID,
		Name:            "Kellan, Inquisitive Prodigy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Kellan — destroy up to one target artifact; draw a card if you controlled it",
				kellanDestroyArtifact),
				TargetPermanent("up to one target artifact", Artifact()).WithCount(0, 1)),
		},
	})
	Register(Spec{
		OracleID:     kellanInquisitiveProdigyOracleID + "#1",
		Name:         "Tail the Suspect",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, ExtraLandDrops: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (CreateToken{Template: ClueToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
			ctx.Game.GrantAdditionalLandPlayForEffect(item.Controller, 1)
			return nil
		},
	})
}

// kellanDestroyArtifact destroys the target, then draws one card when its
// controller before the destroy was Kellan's controller.
func kellanDestroyArtifact(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	draws := 0
	if controller, ok := controllerOfTarget(ctx, id); ok && controller == ctx.Controller() {
		draws = 1
	}
	return g.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, withN(ctx.Tail(thenDrawCards), draws))
}
