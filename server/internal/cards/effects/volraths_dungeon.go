package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volrath's Dungeon — Enchantment {2}{B}{B}:
//
//	"Pay 5 life: Destroy this enchantment. Any player may activate
//	 this ability but only during their turn.
//	 Discard a card: Target player puts a card from their hand on top
//	 of their library. Activate only as a sorcery."
//
// ADR 0106 PR 6 (#1793).
//
//   - The first row is any-player (CR 602.2, 602.1b). "But only during
//     their turn" is an activation instruction (CR 602.1b), and "their"
//     is the player activating it, so it is DuringYourTurn() read with
//     the ACTIVATOR as "you" — any step of that player's own turn,
//     stack or no stack, and never a sorcery-speed gate. The 5 life is
//     the activator's (CR 602.1a, 119.4). The destruction is the
//     effect, so it can be answered.
//   - The second row is the controller's alone (CR 602.2). The discard
//     is a cost (DiscardACard), so it is paid before the target player
//     does anything. "Activate only as a sorcery" is CR 602.5d. The
//     TARGET player chooses which card goes back — "puts a card from
//     their hand" — through Brainstorm's put-back
//     (PutFromHandOnTopInAnyOrder, N = 1); a player with an empty hand
//     puts back nothing (CR 609.3).
//
// No purpose for the bot on the any-player row: destroying the Dungeon
// helps whoever is being locked by it, which no Purpose field
// describes.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8b968ea-76f3-4ee7-9d53-d251d8a9faf6",
		Name:         "Volrath's Dungeon",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:     "Pay 5 life: Destroy this enchantment. Any player may activate this ability but only during their turn.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      game.AbilityCost{Life: 5},
				AnyPlayer: true,
				Condition: DuringYourTurn(),
				Effect:    destroyThisPermanent(false),
			},
			{
				Label:        "Discard a card: Target player puts a card from their hand on top of their library. Activate only as a sorcery.",
				Cost:         DiscardACard(),
				Targets:      TargetPlayer("target player"),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						return PutFromHandOnTopInAnyOrder{
							Player: t.ID,
							N:      1,
							Label:  "Volrath's Dungeon — put a card from your hand on top of your library",
						}.Apply(ctx)
					}
					return nil
				},
			},
		},
	})
}
