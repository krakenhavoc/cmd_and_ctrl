package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lumbering Worldwagon — Artifact — Vehicle {2}{G}, */4 (EDHREC rank
// 4193):
//
//	"This Vehicle's power is equal to the number of lands you control.
//	 Whenever this Vehicle enters or attacks, you may search your
//	 library for a basic land card, put it onto the battlefield
//	 tapped, then shuffle.
//	 Crew 4"
//
// A three-mana Vehicle that ramps on the way in and again on every
// attack, and whose power is the ramp. In a landfall deck the search
// is the point twice over: the land that enters is a trigger as well
// as a mana.
//
// Three mechanics, all of them already in the engine:
//
//   - The power is a CHARACTERISTIC-DEFINING ability (CR 604.3), so
//     it is Layer 7a and not a modification: it SETS power rather than
//     adding to it, and anything that pumps the Worldwagon in 7c
//     stacks on top. Recomputed every pass, so the land that the
//     attack trigger fetches makes the Worldwagon bigger before combat
//     damage — which matters, because the trigger resolves during the
//     declare-attackers step.
//   - "Enters OR attacks" is one printed ability with two trigger
//     conditions, so it is one TriggeredAbility watching both
//     (WhenThisEntersOrAttacks, Sun Titan's shape).
//   - Crew 4 reads EFFECTIVE power across the creatures tapped to pay,
//     which is the engine's business.
//
// "YOU MAY search" is a real optional search (CR 701.23b): the prompt
// lets the controller decline both the card and the shuffle, which is
// the printed wording and matters to anyone tracking their library
// order.
//
// A Vehicle is not a creature until it is crewed, so the entry
// trigger fires off an artifact entering rather than a creature —
// nothing here depends on that, but it is why the Worldwagon's own
// ETB does not feed a "whenever a creature you control enters"
// payoff.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e60a641-1f2d-45ff-b3f2-d389be938b22",
		Name:         "Lumbering Worldwagon",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = b10LandsControlled(g, source.Controller)
			},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Lumbering Worldwagon — you may search for a basic land", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:        item.Controller,
					Predicate:     b30IsBasicLandCard,
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Optional:      true,
					TappedOnEntry: true,
					Shuffle:       true,
					Reason:        "Lumbering Worldwagon — a basic land card, onto the battlefield tapped",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:  "Crew 4",
			Cost:   CrewCost(4),
			Effect: CrewEffect("Lumbering Worldwagon"),
		}},
	})
}
