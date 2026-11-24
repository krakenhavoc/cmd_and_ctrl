package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Othelm, Sigardian Outcast — Legendary Creature — Human {1}{G}{W}, 2/2
// (Secret Lair Drop):
//
//	"{2}, {T}: Choose target creature card in your graveyard that was put
//	 there from the battlefield this turn. Return it to the battlefield
//	 tapped.
//	 Partner—Friends forever (You can have two commanders if both have this
//	 ability.)"
//
// Brought Back's target clause (PutThereFromTheBattlefieldThisTurn, so a
// countered or discarded creature never qualifies) narrowed to creature cards,
// and its simultaneous tapped return under your control. A creature gone from
// the graveyard by resolution is dropped (CR 608.2b). Partner—Friends forever
// is a deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ee401bf-dc86-41a9-84f9-57aa0a314d08",
		Name:         "Othelm, Sigardian Outcast",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, {T}: Choose target creature card in your graveyard that was put there from the battlefield this turn. Return it to the battlefield tapped.",
			Cost:  effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Targets: effects.TargetCardInGraveyard("target creature card in your graveyard that was put there from the battlefield this turn",
				effects.YouOwn(), effects.Creature(), effects.PutThereFromTheBattlefieldThisTurn),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.ReturnFromGraveyardTogether{Targets: effects.LegalTargetCardIDs(ctx), Tapped: true}.Apply(ctx)
			},
		}},
	})
}
