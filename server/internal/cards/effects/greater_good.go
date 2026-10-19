package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greater Good — Enchantment {2}{G}{G} (issue #1117):
//
//	"Sacrifice a creature: Draw cards equal to the sacrificed
//	 creature's power, then discard three cards."
//
// A card-advantage engine for a deck that wants its creatures dead
// anyway: convert whatever is about to die into a fresh grip, minus
// the three worst cards in it. Greater Good is an enchantment, not a
// creature, so the cost never needs an "another" exclusion — nothing
// about the ability can eat itself.
//
// The sacrificed creature is read back off the event log at
// resolution, since the stack item does not carry the cost it paid
// (b17PermanentSacrificedToPay, the same read Jarad, Golgari Lich
// Lord uses). "Discard three cards" is the player's own choice
// (CR 701.9a), not a random discard, so it goes through the prompted
// discard path and self-clamps to whatever the hand holds after the
// draw.
//
// The sacrificed creature's power is its last-known information
// (CR 608.2h, departedCreaturePower): counters and an anthem's bonus
// both count, as they did while it was on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc0593c2-ccb4-4648-a592-c5bcd121dc72",
		Name:         "Greater Good",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a creature: Draw cards equal to the sacrificed creature's power, then discard three cards.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    SacrificeACreature(),
			Effect: Do(
				// The discard waits for a prompt the draw asks (CR 608.2c, #4155).
				Instruction(func(ctx *Context) error {
					fed, ok := b17PermanentSacrificedToPay(ctx.Game, ctx.Item)
					if !ok {
						return nil
					}
					return DrawCards{N: departedCreaturePower(ctx.Game, fed)}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					if _, ok := b17PermanentSacrificedToPay(ctx.Game, ctx.Item); !ok {
						return nil
					}
					ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   ctx.Controller(),
						Source:   ctx.Source(),
						N:        3,
						Question: "Greater Good — discard three cards",
					})
					return nil
				}),
			),
		}},
	})
}
