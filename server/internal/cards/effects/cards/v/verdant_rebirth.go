package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verdant Rebirth — Instant {1}{G} (Ixalan):
//
//	"Until end of turn, target creature gains "When this creature dies, return
//	 it to its owner's hand."
//	 Draw a card."
//
// Family's Favor's duration grant (ADR 0093): a death trigger on the target
// that lasts the turn and returns it from the graveyard to its OWNER's hand,
// as Immortal Phoenix's does. The draw is not tied to the target: a target that
// became illegal gets no grant, and the spell still draws only if it resolves
// (it resolves while any target is legal; with none it is countered).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c35540ad-9ab1-4be9-b97d-603813394a26",
		Name:         "Verdant Rebirth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key: verdantRebirthGrant,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Verdant Rebirth — return it to its owner's hand", effects.ReturnSelfFromGraveyardToHand),
			},
			Text: "When this creature dies, return it to its owner's hand.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.GrantAbilitiesFor{
					Target: id,
					Keys:   []string{verdantRebirthGrant},
					Label:  "Verdant Rebirth — gains \"When this creature dies, return it to its owner's hand\" until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}

// verdantRebirthGrant is the granted bundle's key. An on-disk identity: never
// renamed, never reused.
const verdantRebirthGrant = "verdant-rebirth/dies-return"
