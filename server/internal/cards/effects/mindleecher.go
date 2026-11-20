package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mindleecher — Creature — Nightmare {4}{B}{B}, 5/5 (Commander 2020, #5513):
//
//	"Mutate {4}{B} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature
//	 on top plus all abilities from under it.)
//	 Flying
//	 Whenever this creature mutates, exile the top card of each opponent's
//	 library face down. You may look at and play those cards for as long as
//	 they remain exiled."
//
// Mutate is the ADR 0223 seam (mutate.go), and the trigger is Everquill
// Phoenix's, firing whether Mindleecher went over or under. Each card is
// Ixhel, Scion of Atraxa's face-down exile with a play permission for the
// trigger's controller that lasts while the card stays exiled: a land may be
// played (a land play), a spell cast for its own cost with the mana it
// normally takes, and the holder may look at the face-down card. An
// opponent with an empty library exiles nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "4864cc98-4fde-4010-81e9-ea4944347e7e",
		Name:             "Mindleecher",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Mutate("{4}{B}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventMutate, ThisCreatureMutates,
				"Mindleecher — exile the top card of each opponent's library face down; you may look at and play them while they remain exiled",
				Do(Instruction(mindleecherExile))),
		},
	})
}

// mindleecherExile exiles the top of each opponent's library face down
// with the play permission.
func mindleecherExile(ctx *Context) error {
	for _, opp := range ctx.Opponents() {
		if err := (ExileTopWithPermission{
			From:        opp,
			GrantTo:     ctx.Controller(),
			N:           1,
			FaceDown:    true,
			WhileExiled: true,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
