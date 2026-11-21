package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

var joltedAwakePaid = effects.EnergyPaid("pay-unless/jolted-awake-return",
	effects.ReturnFirstLegalGraveyardTargetToBattlefield)

// Jolted Awake — Sorcery {W}:
//
//	"Choose up to one target artifact or creature card in your
//	 graveyard. You get {E}{E} (two energy counters). Then you may pay
//	 an amount of {E} equal to that card's mana value. If you do, return
//	 it from your graveyard to the battlefield.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// ADR 0129 §3 (#1995): with no target chosen, the caster still gets the
// energy and nothing is asked. With one, the amount is its mana value,
// paid through the energy prompt (CR 118.12); paid, it returns if it is
// still in the graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4306a9e5-6845-4fd4-96ad-8888147612aa",
		Name:         "Jolted Awake",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Targets: effects.TargetCardInGraveyard("up to one target artifact or creature card in your graveyard",
			effects.YouOwn(), effects.Or(effects.Artifact(), effects.Creature())).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GetEnergy{N: 2}).Apply(ctx); err != nil {
				return err
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.MayPayEnergy{
				N:        effects.FirstTargetManaValue(ctx.Game, item),
				Question: "Jolted Awake — pay {E} equal to its mana value to return it to the battlefield?",
				Then:     joltedAwakePaid,
			}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
