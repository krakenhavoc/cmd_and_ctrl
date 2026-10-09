package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trumpeting Carnosaur — Creature — Dinosaur {4}{R}{R}, 7/6:
//
//	"Trample
//	 When this creature enters, discover 5.
//	 {2}{R}, Discard this card: It deals 3 damage to target creature or
//	 planeswalker."
//
// The last ability works from the HAND (CR 113.6, ADR 0062): the card is
// discarded as the cost, and "it" — the card, now in the graveyard —
// deals the damage. It is not cycling, so it fires no cycle event. The
// discover is ADR 0099's (game/discover.go).
func init() {
	Register(Spec{
		OracleID:        "f2ef8bda-373d-4387-900d-0f1b6ccf72e9",
		Name:            "Trumpeting Carnosaur",
		Completeness:    CompletenessFull,
		Discovers:       true,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Trumpeting Carnosaur — discover 5", DiscoverN(5)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, Discard this card: It deals 3 damage to target creature or planeswalker",
			Cost:    Plus(ManaCost("{2}{R}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Targets: TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
