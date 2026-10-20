package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jaya's Phoenix — Creature — Phoenix {4}{R}, 3/3:
//
//	"Flying, haste
//	 Whenever this creature deals combat damage to a player or
//	 planeswalker, copy the next loyalty ability you activate this turn
//	 when you activate it. You may choose new targets for the copy.
//	 Whenever you cast a planeswalker spell, you may return this card
//	 from your graveyard to the battlefield."
//
// The combat-damage trigger sets up a CR 603.7 delayed trigger watching
// your next activation of a loyalty ability (an ability with a loyalty
// symbol in its cost, CR 606.2) this turn. It fires on that activation,
// goes on the stack above the ability, and copies it while it is still
// there (CR 707.10), offering new targets. The copy is created, not
// activated, so it is not a loyalty activation of its own. Two combat
// damage triggers make two delayed triggers, and both copy the same next
// activation. If no loyalty ability is activated, it ends with the turn.
//
// The last ability works only from the graveyard (InGraveyard), on each
// planeswalker spell you cast, with its "you may" asked as it resolves;
// a Phoenix that has left the graveyard by then stays where it is. It
// returns under its owner's control (no controller is printed).
//
// No simplification. Guffster's Party deck request (#4184).
func init() {
	const name = "Jaya's Phoenix"
	Register(Spec{
		OracleID:        "c6a67820-5066-462f-a621-1ecab8bf4f97",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayerOrPlaneswalker,
				name+" — copy the next loyalty ability you activate this turn",
				func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:     name + " — copy that loyalty ability",
						On:        []game.EventKind{game.EventActivateAbility},
						Condition: youNextActivateALoyaltyAbilityCondition,
						Body:      copyTheAbilityBody,
					}.Apply(NewContext(g, item))
				}),
			InGraveyard(Optional(
				WheneverYouCast(Planeswalker(), name+" — return it from your graveyard to the battlefield",
					func(g *game.Game, item *game.StackItem) error {
						id := item.SourceCardID
						if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
							return nil
						}
						return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
					}),
				name+" — return it from your graveyard to the battlefield?")),
		},
	})
}
