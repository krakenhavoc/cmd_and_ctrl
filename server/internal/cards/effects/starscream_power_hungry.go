package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Starscream, Power Hungry // Starscream, Seeker Leader (#3195, ADR
// 0152) — a More Than Meets the Eye card.
//
// Front face, Legendary Artifact Creature — Robot {3}{B}, 2/3:
//
//	"More Than Meets the Eye {2}{B}
//	 Flying
//	 Whenever you draw a card, if you're the monarch, target opponent
//	 loses 2 life.
//	 Whenever one or more creatures deal combat damage to you, convert
//	 Starscream."
//
// Back face, Legendary Artifact — Vehicle, 2/3:
//
//	"Living metal
//	 Flying, menace, haste
//	 Whenever Starscream deals combat damage to a player, if there is no
//	 monarch, that player becomes the monarch.
//	 Whenever you become the monarch, convert Starscream."
//
// Both "if" clauses are intervening ifs (CR 603.4): the trigger checks
// the crown as the event happens and the effect checks it again as the
// ability resolves. "One or more creatures deal combat damage to you" is
// a batch (CR 603.2c): one trigger per combat damage step, whoever
// controls the creatures. "That player" is the player Starscream hit,
// read off the triggering event.
//
// No simplification.
const starscreamOracleID = "3e924a35-0e08-4050-b33a-a143d2ab5aea"

func init() {
	Register(Spec{
		OracleID:         starscreamOracleID,
		Name:             "Starscream, Power Hungry",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{MoreThanMeetsTheEye("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventDrawCard, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ByYou(ev, source, ch, g) && YoureTheMonarch(g, source.Controller)
			}, "Starscream, Power Hungry — target opponent loses 2 life",
				func(g *game.Game, item *game.StackItem) error {
					if !YoureTheMonarch(g, item.Controller) {
						return nil // CR 603.4: still the monarch as it resolves
					}
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return GainLife{Player: t.ID, Amount: -2}.Apply(ctx)
						}
					}
					return nil
				}), TargetPlayer("target opponent", Opponent())),
			OncePerBatch(On(game.EventDealDamage, creatureDealtCombatDamageToYou,
				"Starscream, Power Hungry — convert Starscream", Do(ConvertThis{}))),
		},
	})
	Register(Spec{
		OracleID:        starscreamOracleID + "#1",
		Name:            "Starscream, Seeker Leader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "menace", "haste"},
		Static:          []game.StaticAbility{LivingMetal()},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ThisDealtCombatDamageToAPlayer(ev, source, ch, g) && g.Monarch == uuid.Nil
			}, "Starscream, Seeker Leader — that player becomes the monarch",
				func(g *game.Game, item *game.StackItem) error {
					if g.Monarch != uuid.Nil {
						return nil // CR 603.4: still no monarch as it resolves
					}
					ctx := NewContext(g, item)
					return BecomeTheMonarch{Player: ctx.Trigger().Event.Target}.Apply(ctx)
				}),
			WheneverYouBecomeTheMonarch("Starscream, Seeker Leader — convert Starscream", Do(ConvertThis{})),
		},
	})
}
