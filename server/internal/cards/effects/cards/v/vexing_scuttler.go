package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vexing Scuttler — Creature — Eldrazi Crab {8}, 4/5:
//
//	"Emerge {6}{U} (You may cast this spell by sacrificing a creature and
//	 paying the emerge cost reduced by that creature's mana value.)
//	 When you cast this spell, you may return target instant or sorcery
//	 card from your graveyard to your hand."
//
// Emerge is the shared alternative cost (ADR 0135 §4). The cast trigger
// is a "you may", asked as it triggers; on a yes the instant or sorcery
// card returns above the spell, even if the Scuttler is countered. A card
// that left the graveyard in response is not returned (CR 608.2b).
//
// No simplification.
func init() {
	cast := effects.WhenYouCastThisSpell("Vexing Scuttler — return target instant or sorcery card from your graveyard to your hand",
		effects.ReturnFirstLegalGraveyardTargetToHand)
	cast.Targets = effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn())
	effects.Register(effects.Spec{
		OracleID:         "2a5605a8-606e-4866-9473-7a2240ac6b8c",
		Name:             "Vexing Scuttler",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Emerge("{6}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(cast, "Vexing Scuttler — return an instant or sorcery card from your graveyard to your hand?"),
		},
	})
}
