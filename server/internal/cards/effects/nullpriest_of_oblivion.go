package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nullpriest of Oblivion — Creature — Vampire Cleric {1}{B}, 2/1
// (Foundations, #4465):
//
//	"Kicker {3}{B} (You may pay an additional {3}{B} as you cast this
//	 spell.)
//	 Lifelink
//	 Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 When this creature enters, if it was kicked, return target creature
//	 card from your graveyard to the battlefield."
//
// The kicked check is Gatekeeper of Malakir's intervening "if", read off
// the permanent's carried kicker record when the ability would trigger,
// so an unkicked Nullpriest puts nothing on the stack and asks for no
// target. The return is Karmic Guide's: a creature card you own in your
// graveyard, returned under your control if it is still there as the
// ability resolves. Lifelink and menace ride PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c1111228-ef88-4b64-91e0-66ee32a1f5e4",
		Name:            "Nullpriest of Oblivion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink", "menace"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{3}{B}")},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return game.CardKickedTimes(*source) > 0
			}),
			Key:     "Nullpriest of Oblivion — kicked, return target creature card from your graveyard to the battlefield",
			Targets: targetCreatureInYourGraveyard(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				reanimateSingleTarget(NewContext(g, item), item.Controller)
				return nil
			},
		}},
	})
}
