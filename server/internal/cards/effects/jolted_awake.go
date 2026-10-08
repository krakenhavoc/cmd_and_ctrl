package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4306a9e5-6845-4fd4-96ad-8888147612aa",
		Name:         "Jolted Awake",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Targets: TargetCardInGraveyard("up to one target artifact or creature card in your graveyard",
			YouOwn(), Or(Artifact(), Creature())).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (GetEnergy{N: 2}).Apply(ctx); err != nil {
				return err
			}
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return MayPayEnergy{
				N:        firstTargetManaValue(ctx.Game, item),
				Question: "Jolted Awake — pay {E} equal to its mana value to return it to the battlefield?",
				OnPay: func(ctx *Context) error {
					return returnFirstLegalGraveyardTargetToBattlefield(ctx.Game, ctx.Item)
				},
			}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
