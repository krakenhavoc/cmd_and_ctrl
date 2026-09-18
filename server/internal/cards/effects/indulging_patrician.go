package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Indulging Patrician — Creature — Vampire Noble {1}{W}{B}, 1/4
// (EDHREC rank 4300):
//
//	"Flying
//	 Lifelink
//	 At the beginning of your end step, if you gained 3 or more life
//	 this turn, each opponent loses 3 life."
//
// A three-mana flier that drains the whole table for 3 a turn once the
// lifegain deck is running — which in a four-player game is 9 life a
// turn, and the reason the Patrician is a Commander card rather than a
// Standard one. Its own lifelink is enough to turn the trigger on: a
// 1/4 flier connecting for 1 is not 3, but a single anthem or any
// other gain that turn finishes the job.
//
// The "if you gained 3 or more life this turn" is an INTERVENING-IF
// (CR 603.4), which is two checks and not one: the condition is tested
// when the end step begins, and again when the trigger resolves. A
// Sudden Spoiling in response cannot turn it off, but losing the life
// back can — gains and losses are tallied separately, and the tally
// only counts POSITIVE life changes, so a gain of 4 followed by a loss
// of 4 is still "gained 4 this turn". That is the rule: the card asks
// what you gained, not what you are at.
//
// "YOUR end step", so once a turn cycle, not once per player's turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "189131b7-ed59-4750-a4ed-453e4497a5c8",
		Name:            "Indulging Patrician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return isActivePlayer(g, source.Controller) &&
					b32EndStepAndYouGainedLifeThisTurnAtLeast(ev, source, g, 3)
			}, "Indulging Patrician — each opponent loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					// CR 603.4: the intervening-if is re-checked on
					// resolution, and the trigger does nothing if the
					// condition has gone away.
					if b15LifeGainedThisTurn(g, item.Controller) < 3 {
						return nil
					}
					ctx := NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if err := g.ChangePlayerLifeForEffect(item.SourceCardID, opp, -3); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
