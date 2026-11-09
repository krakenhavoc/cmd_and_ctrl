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
// The flash clause is Bard's Company's shape: the card's own timing
// permission (FlashIfYou, Spec.FlashIf, #5207), read as the cast begins
// (CR 601.3), and not an alternative cost, so a flash cast may still
// claim one (CR 118.9a). The three instructions are done in the printed order,
// each waiting for the one before it (CR 608.2c): the opponent picks the
// creature (their prompt, not a target), then the card to discard, then
// loses the life. A target opponent with no creature or no cards in hand
// skips that part and still does the rest.
//
// No simplification.
func init() {
	const name = "Scarring Memories"
	Register(Spec{
		OracleID:     "63957ec0-89e6-4be2-8740-846ca56ee48c",
		Name:         name,
		Completeness: CompletenessFull,
		FlashIf:      FlashIfYou("if you control an attacking legendary creature", avB30aYouControlAnAttackingLegendaryCreature),
		Targets:      TargetPlayer("target opponent", Opponent()),
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
