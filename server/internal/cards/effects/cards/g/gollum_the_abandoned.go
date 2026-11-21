package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gollum, the Abandoned — Legendary Creature — Halfling Horror {1}{B}, 2/2:
//
//	"Gollum can't block.
//	 When Gollum enters, exile up to one target card from an opponent's
//	 graveyard. Each opponent loses 2 life.
//	 {2}, Sacrifice an artifact or creature: Return this card from your
//	 graveyard to your hand. Activate only as a sorcery."
//
// Bloodghast's can't-block rule and Gollum, Patient Plotter's graveyard
// ability with the Ravenous Squirrel sacrifice. The enter trigger's "up to
// one" is a zero-to-one target clause, chosen as the trigger goes on the
// stack and dropped from the stack-time prompt when no opponent has a card
// in a graveyard; the life loss does not depend on a card being exiled.
// "An opponent's graveyard" is a card its owner is not the caster.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d88facd-cf7e-498e-ab6b-6bd021316162",
		Name:         "Gollum, the Abandoned",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: effects.TargetCardInGraveyard("up to one target card from an opponent's graveyard", effects.Not(effects.YouOwn())).WithCount(0, 1),
			Key:     "Gollum, the Abandoned — exile up to one target card from an opponent's graveyard; each opponent loses 2 life",
			Effect: func(g *game.Game, item *game.StackItem) error {
				if err := effects.ExileFirstTarget(g, item); err != nil {
					return err
				}
				return effects.EachOpponentLosesLife{N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}, Sacrifice an artifact or creature: Return this card from your graveyard to your hand. Activate only as a sorcery.",
			Cost: effects.Plus(effects.ManaCost("{2}"), game.AbilityCost{
				SacrificeOther: effects.SacrificeSpec("an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			}),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.ReturnThisCardFromYourGraveyardToHand,
		}},
	})
}
