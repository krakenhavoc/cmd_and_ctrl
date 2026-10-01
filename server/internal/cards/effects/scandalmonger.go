package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scandalmonger — Creature — Boar Monger {3}{B}, 3/3:
//
//	"{2}: Target player discards a card. Any player may activate this
//	 ability but only as a sorcery."
//
// An any-player row (CR 602.2, 602.1b) with a timing instruction: "only
// as a sorcery" (CR 602.5d) is the ACTIVATOR's sorcery timing, read for
// the player activating it, so a non-controller may use it in their own
// main phase with an empty stack and never in somebody else's turn.
// Whoever activates it pays the {2} out of their own pool (CR 602.1a)
// and chooses the target player, themselves included. The targeted
// player chooses the card (CR 701.9b) through the ordinary discard
// prompt; a player who has left the game or become an illegal target is
// asked nothing (CR 608.2b).
//
// No Purpose: ActivationPurpose has no "target opponent discards"
// field, so the bot does not reach across the table for it (ADR 0106
// owner decision 2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d98d8d8-e277-42c4-bf94-dd7d7f2047c3",
		Name:         "Scandalmonger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{2}: Target player discards a card. Any player may activate this ability but only as a sorcery.",
			Cost:         ManaCost("{2}"),
			Targets:      TargetPlayer("target player"),
			SorcerySpeed: true,
			AnyPlayer:    true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				for _, t := range NewContext(g, item).LegalTargets() {
					if t.Kind != game.TargetPlayer {
						continue
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
						Player:   t.ID,
						Source:   item.SourceCardID,
						N:        1,
						Question: "Scandalmonger — discard a card",
					})
					return nil
				}
				return nil
			},
		}},
	})
}
