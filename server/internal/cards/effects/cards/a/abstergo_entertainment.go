package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abstergo Entertainment — Legendary Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color.
//	 {3}, {T}, Exile Abstergo Entertainment: Return up to one target
//	 historic card from your graveyard to your hand, then exile all
//	 graveyards. (Artifacts, legendaries, and Sagas are historic.)"
//
// The mana abilities are Captivating Cave's. The exile ability's target
// is "up to one", so it is announced with no target as well, and the
// graveyard sweep still happens then (the ability resolves with no
// target). A target chosen is re-checked at resolution (CR 608.2b): when
// it has left the graveyard the ability does nothing at all. Historic is
// CR 700.6, read through the shared predicate.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d06a8026-1657-4404-8dff-64e44f1a14f8",
		Name:         "Abstergo Entertainment",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.PainlessColorless(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Exile Abstergo Entertainment: Return up to one target historic card from your graveyard to your hand, then exile all graveyards.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.ExileThis()),
			Purpose: game.Purpose{Answers: game.AnswerGraveyard},
			Targets: effects.TargetCardInGraveyard("up to one target historic card from your graveyard", effects.YouOwn(), effects.B14Historic()).WithCount(0, 1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ReturnTargetCardToHand(g, item); err != nil {
					return err
				}
				return effects.B02ExileAllGraveyards(g, item)
			},
		}},
	})
}
