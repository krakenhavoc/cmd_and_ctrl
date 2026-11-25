package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Basri, Devoted Paladin — Legendary Planeswalker — Basri {4}{W}{W}, loyalty 4
// (Core Set 2021, #5511):
//
//	"+1: Put a +1/+1 counter on up to one target creature. It gains vigilance
//	 until end of turn.
//	 −1: Whenever a creature attacks this turn, put a +1/+1 counter on it.
//	 −6: Creatures you control get +2/+2 and gain flying until end of turn."
//
// The +1 is Basri Ket's counter-then-keyword shape on the one target, with
// vigilance; with no target it does nothing. The −1 is a repeating delayed
// trigger (CR 603.7) for the rest of the turn, one per creature that attacks,
// any player's (Ondu Rising's shape); it is not tied to Basri, so it still
// works if he has left, and a creature gone from the battlefield by then gets
// nothing. The −6 locks the set of creatures you control when it resolves
// (CR 611.2c) and gives them both effects until end of turn.
//
// No simplification.
func init() {
	mine := effects.And(effects.Creature(), effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:        "1601e553-31c2-4e04-9f0c-bc17e49eeff7",
		Name:            "Basri, Devoted Paladin",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Put a +1/+1 counter on up to one target creature. It gains vigilance until end of turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect:  effects.Do(effects.PlusOneCounterThenKeywordOnTheTarget("vigilance", "Basri, Devoted Paladin — vigilance until end of turn")...),
			},
			{
				Label: "−1: Whenever a creature attacks this turn, put a +1/+1 counter on it.",
				Cost:  effects.LoyaltyCost(-1),
				Effect: effects.Do(effects.DelayedOnEvent{
					Label:   "Basri, Devoted Paladin — put a +1/+1 counter on the attacking creature",
					On:      []game.EventKind{game.EventAttack},
					Body:    basriDevotedPaladinCounterBody,
					Repeats: true,
				}),
			},
			{
				Label: "−6: Creatures you control get +2/+2 and gain flying until end of turn.",
				Cost:  effects.LoyaltyCost(-6),
				Effect: effects.Do(
					effects.BoostUntilEOT{Match: mine, Power: 2, Toughness: 2, Label: "Basri, Devoted Paladin — +2/+2"},
					effects.GrantKeywordUntilEOT{Match: mine, Keywords: []string{"flying"}, Label: "Basri, Devoted Paladin — flying"}),
			},
		},
	})
}

// basriDevotedPaladinCounterBody puts a +1/+1 counter on the creature that
// attacked (the trigger's payload). On-disk key: never renamed, never reused.
var basriDevotedPaladinCounterBody = game.SimpleDelayedBody("s82-old-b107/basri-devoted-paladin-attacker-counter",
	func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, id := range ctx.PayloadCards() {
			if _, ok := g.PermanentRefForEffect(id); !ok {
				continue
			}
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
