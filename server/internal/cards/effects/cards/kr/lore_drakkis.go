package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lore Drakkis — Creature — Lizard Beast {1}{U}{R}, 2/3 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {U/R}{U/R} (If you cast this spell for its mutate cost, put it over
//	 or under target non-Human creature you own. They mutate into the creature
//	 on top plus all abilities from under it.)
//	 Whenever this creature mutates, return target instant or sorcery card from
//	 your graveyard to your hand."
//
// Mutate is the engine's (ADR 0223). The target is chosen as the trigger goes
// on the stack (CR 603.3d) and re-checked as it resolves (CR 608.2b); a merge
// with no instant or sorcery card in your graveyard puts nothing on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "3ec271d8-8264-4574-8c4c-e57b35872b1e",
		Name:             "Lore Drakkis",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{U/R}{U/R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Lore Drakkis — return target instant or sorcery card from your graveyard to your hand", effects.ReturnTargetedCardToHand),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery()))),
		},
	})
}
