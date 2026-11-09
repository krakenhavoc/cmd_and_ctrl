package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scarring Memories — Sorcery — Lesson {3}{B} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"You may cast this spell as though it had flash if you control an
//	 attacking legendary creature.
//	 Target opponent sacrifices a creature of their choice, discards a
//	 card, and loses 3 life."
//
// The flash clause is Bard's Company's shape: an offer at the printed cost
// that opens the instant-speed window for its own claim
// (AlternativeCost.AsThoughFlash), gated by Condition, read as the cast
// begins (CR 601.3). The three instructions are done in the printed order,
// each waiting for the one before it (CR 608.2c): the opponent picks the
// creature (their prompt, not a target), then the card to discard, then
// loses the life. A target opponent with no creature or no cards in hand
// skips that part and still does the rest.
//
// Simplification (#5207): the offer is an alternative cost, so the flash
// cast can't also take another alternative cost (CR 118.9a). On the
// printed card the flash is a timing permission, so it could. Declared
// as the caveat.
func init() {
	const name = "Scarring Memories"
	Register(Spec{
		OracleID:     "63957ec0-89e6-4be2-8740-846ca56ee48c",
		Name:         name,
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Casting this at instant speed while you control an attacking legendary creature can't be combined with another alternative cost.",
		},
		AlternativeCosts: []game.AlternativeCost{{
			Key:           "flash-if-you-control-an-attacking-legendary-creature",
			Label:         "Cast with flash — you control an attacking legendary creature ({3}{B})",
			ManaCost:      "{3}{B}",
			AsThoughFlash: true,
			Condition:     avB30aYouControlAnAttackingLegendaryCreature,
		}},
		Targets: TargetPlayer("target opponent", Opponent()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return targetPlayerSacrificesACreature(name+" — sacrifice a creature")(ctx.Game, ctx.Item)
			}),
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetPlayer {
						ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
							Player: t.ID, Source: ctx.Source(), N: 1, Question: name + " — discard a card",
						})
					}
				}
				return nil
			}),
			TargetPlayerLosesLife{N: 3},
		),
	})
}
