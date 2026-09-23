package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rangers' Refueler — Artifact — Vehicle {1}{U}, 3/3:
//
//	"Whenever you activate an exhaust ability, draw a card.
//	 Exhaust — {4}: This Vehicle becomes an artifact creature. Put a
//	 +1/+1 counter on it. (Activate each exhaust ability only once.)
//	 Crew 2"
//
// The first card to WATCH an activation (#1184). Everything about it
// was impossible before the announcement had an event of its own: the
// announce emitted `EventTrigger`, the harvester returns immediately
// on that kind (a trigger firing further triggers must not re-enter
// the harvest), and even if it had not, the event carried no ability
// identity — nothing on it said which ability had been activated or
// whether it printed the keyword.
//
// So the trigger is an ordinary `On` over `EventActivateAbility` plus
// `EventManaAbilityActivated`, with the two stamps as its whole
// condition: your activation, and an exhaust one. Both kinds, because
// CR 605.1a makes a mana ability an activated ability and Loot, the
// Pathfinder prints an exhaust one.
//
// It triggers off its OWN exhaust ability. "An exhaust ability" does
// not say "another", so activating the {4} below draws a card as
// well; the trigger goes on the stack ABOVE the ability that made it
// (CR 603.3b) and resolves first, which is observable when the draw
// is what finds the answer to whatever happens next.
//
// The animation prints NO duration (CR 611.2a): this Vehicle is an
// artifact creature for the rest of the game, not until end of turn
// like a crewed one, which is why it is BecomeArtifactCreature rather
// than the crew primitive beside it. The +1/+1 counter goes on
// afterwards and is an ordinary placement, so Hardened Scales applies.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e7ac64d-42c3-4e3b-9a20-9f09ee9d8a7a",
		Name:         "Rangers' Refueler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouActivateAnExhaustAbility(
				"Rangers' Refueler — draw a card",
				Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "Exhaust — {4}: This Vehicle becomes an artifact creature. Put a +1/+1 counter on it.",
				Exhaust: true,
				Cost:    ManaCost("{4}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (BecomeArtifactCreature{
						Label: "Rangers' Refueler — becomes an artifact creature",
					}).Apply(ctx); err != nil {
						return err
					}
					return AddCounter{
						Target: item.SourceCardID,
						Kind:   game.CounterPlusOne,
						N:      1,
					}.Apply(ctx)
				},
			},
			{
				Label:  "Crew 2",
				Cost:   CrewCost(2),
				Effect: CrewEffect("Rangers' Refueler"),
			},
		},
	})
}
