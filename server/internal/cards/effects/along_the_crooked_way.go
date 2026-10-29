package effects

import (
	"github.com/google/uuid"

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
//     b02CreatureCardLeftYourGraveyard: one trigger per card, so two
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
	Register(Spec{
		OracleID:     "f4437cb3-85f7-49b5-bb41-6a948549d987",
		Name:         "Along the Crooked Way",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Along the Crooked Way — return target creature card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						return returnLegalGraveyardTargetsToHand(NewContext(g, item))
					}),
				TargetCardInGraveyard("target creature card in your graveyard", Creature(), YouOwn()),
			),
			On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b02CreatureCardLeftYourGraveyard(ev, source, g)
			}, "Along the Crooked Way — amass Goblins 1", Do(Amass{Subtype: "Goblin", N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: Goblins and Orcs you control gain menace until end of turn.",
			Cost:    ManaCost("{1}{B}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: func(g *game.Game, item *game.StackItem) error {
				you := item.Controller
				return GrantKeywordUntilEOT{
					Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
						return c.Controller == you && (c.HasSubtypeWord("Goblin") || c.HasSubtypeWord("Orc"))
					},
					Keywords: []string{"menace"},
					Label:    "Along the Crooked Way — menace",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
