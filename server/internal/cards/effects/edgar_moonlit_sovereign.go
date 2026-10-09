package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Edgar, Moonlit Sovereign — Legendary Creature — Werewolf Noble
// {3}{G}{G}, 4/4:
//
//	"Flash
//	 At the beginning of your end step, if you didn't cast a spell this
//	 turn, put two +1/+1 counters on Edgar.
//	 {4}{G}: Put a +1/+1 counter on each creature you control with
//	 a +1/+1 counter on it."
//
// The end-step trigger is an intervening "if" (CR 603.4) in the shape
// AtYourEndStepIfRevolt uses: the cast tally is read when the step
// begins and again as the trigger resolves, so a spell cast in response
// (flash) turns it off. The tally counts every spell the controller
// cast this turn, Edgar's own included, so a flashed-in Edgar does not
// grow that turn. The activated ability picks its creatures as it
// resolves (CR 611.2c), all at once, including Edgar himself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1917bda6-c0e1-4c80-8009-74c28cf6b8e9",
		Name:            "Edgar, Moonlit Sovereign",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, AllOf(ByYou, rfCreatureBYouCastNoSpellThisTurn),
				"Edgar, Moonlit Sovereign — put two +1/+1 counters on Edgar",
				func(g *game.Game, item *game.StackItem) error {
					if g.CastTallyFor(item.Controller).Total > 0 {
						return nil
					}
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{{
			Label: "{4}{G}: Put a +1/+1 counter on each creature you control with a +1/+1 counter on it.",
			Cost:  ManaCost("{4}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				var ids []uuid.UUID
				for _, c := range g.Battlefield.Cards {
					if c.Controller == item.Controller && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 {
						ids = append(ids, c.InstanceID)
					}
				}
				for _, id := range ids {
					if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
