package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Along the Crooked Way — Enchantment {2}{B}:
//
//	"When this enchantment enters, return target creature card from
//	 your graveyard to your hand.
//	 Whenever a creature card leaves your graveyard, amass Goblins 1.
//	 {1}{B}: Goblins and Orcs you control gain menace until end of
//	 turn."
//
// Three separate abilities, as printed.
//
//   - The return is the ordinary targeted graveyard-to-hand. With no
//     creature card to name the trigger is removed (CR 603.3d).
//   - The leave trigger is Syr Konrad's condition,
//     creatureCardLeftYourGraveyard: one trigger per card, so two
//     cards leaving together amass twice (unlike Desecrated Tomb's "one
//     or more"). The card the enters trigger just returned leaves the
//     graveyard, so the enters trigger returning a creature
//     amasses once as well. "Your graveyard" is the one you own.
//   - The menace grant locks in its set of Goblins and Orcs at
//     resolution (CR 611.2c), reading effective subtypes, so an Army
//     that amass made a Goblin counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4437cb3-85f7-49b5-bb41-6a948549d987",
		Name:         "Along the Crooked Way",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Along the Crooked Way — return target creature card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
					}),
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.Creature(), effects.YouOwn()),
			),
			effects.OnAny(effects.LeftGraveyardEvents, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CreatureCardLeftYourGraveyard(ev, source, g)
			}, "Along the Crooked Way — amass Goblins 1", effects.Do(effects.Amass{Subtype: "Goblin", N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Goblins and Orcs you control gain menace until end of turn.",
			Cost:    effects.ManaCost("{1}{B}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				you := item.Controller
				return effects.GrantKeywordUntilEOT{
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.Controller == you && (c.HasSubtypeWord("Goblin") || c.HasSubtypeWord("Orc"))
					},
					Keywords: []string{"menace"},
					Label:    "Along the Crooked Way — menace",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
