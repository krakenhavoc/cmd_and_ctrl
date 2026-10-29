package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bilbo's Gambit — Instant {1}{W} (The Hobbit, HOB):
//
//	"Gift a Treasure (You may promise an opponent a gift as you cast this
//	 spell. If you do, they create a Treasure token before its other
//	 effects. It's an artifact with "{T}, Sacrifice this token: Add one
//	 mana of any color.")
//	 Return target spell to its owner's hand. If the gift was promised,
//	 players can't cast spells this turn."
//
// The gift is the keyword (CR 702.174, ADR 0089): GiftATreasure gives the
// chosen opponent a Treasure before anything else on the spell happens
// (CR 702.174j). The bounce is Reprieve's, NOT a counter (the card never
// prints the word), so a spell that can't be countered is returned all the
// same, and a spell that has left the stack in response leaves the Gambit
// with no legal target.
//
// The second sentence reads the promise back at resolution
// (ctx.GiftPromised): every seated player, the caster included, gets
// Mandate of Peace's outright cast ban until end of turn (CR 611.2).
// Abilities may still be activated; mana abilities, the Treasure's included,
// are untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef71ef07-8a35-41ec-a851-d95e9e3a221b",
		Name:         "Bilbo's Gambit",
		Completeness: CompletenessFull,
		Gift:         GiftATreasure(),
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (ReturnSpellToHand{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.GiftPromised() {
				return nil
			}
			duration := DurationUntilEndOfTurn(ctx)
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Eliminated {
					continue
				}
				if err := (RestrictCasting{
					Player:   p.ID,
					Rule:     game.CastBanRule{Kind: game.CastBanOutright},
					Label:    "Bilbo's Gambit — players can't cast spells this turn",
					Duration: duration,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
