package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Subira, Tulzidi Caravanner — Legendary Creature — Human Shaman {2}{R},
// 2/3 (Commander Masters):
//
//	"Haste
//	 {1}: Another target creature with power 2 or less can't be blocked
//	 this turn.
//	 {1}{R}, {T}, Discard your hand: Until end of turn, whenever a creature
//	 you control with power 2 or less deals combat damage to a player, draw
//	 a card."
//
// The second ability is Access Tunnel's body, one power lower and without
// the tap, aimed at another creature. The third is Dovin, Grand Arbiter's
// delayed "whenever a creature you control deals combat damage to a player"
// (until end of turn, repeating), narrowed to a creature whose power is 2
// or less as it dealt the damage. The hand is discarded as a cost, so the
// delayed trigger exists even with an empty hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e2e3272-d4e1-4329-b1d2-f393b3511b82",
		Name:            "Subira, Tulzidi Caravanner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}: Another target creature with power 2 or less can't be blocked this turn.",
				Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    ManaCost("{1}"),
				Targets: Another(TargetCreature("another target creature with power 2 or less", PowerLE(2))),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Subira, Tulzidi Caravanner — can't be blocked",
						}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "{1}{R}, {T}, Discard your hand: Until end of turn, whenever a creature you control with power 2 or less deals combat damage to a player, draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{1}{R}"), TapCost(), DiscardYourHand()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DelayedOnEvent{
						Label:     "Subira, Tulzidi Caravanner — draw a card",
						On:        []game.EventKind{game.EventDealDamage},
						Condition: subiraSmallCreatureConnects,
						Body:      subiraDrawBody,
						Repeats:   true,
					}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

var (
	// subiraSmallCreatureConnects is "a creature you control with power 2 or
	// less deals combat damage to a player", one match per creature dealing it.
	subiraSmallCreatureConnects = game.DelayedCondition("combat-damage/subira-small-creature-connects",
		func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
			if !combatDamageToPlayerBy(ev, dt.Controller, g) {
				return false
			}
			src, ok := g.LookupCardForEffect(ev.Source)
			return ok && src.CurrentPower() <= 2
		})

	subiraDrawBody = game.SimpleDelayedBody("subira-tulzidi-caravanner/draw", func(g *game.Game, item *game.StackItem) error {
		return DrawCards{N: 1}.Apply(NewContext(g, item))
	})
)
