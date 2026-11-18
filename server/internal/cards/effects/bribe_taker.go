package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bribe Taker — Creature — Rhino Warrior {5}{G}, 6/6 (New Capenna
// Commander, #5491):
//
//	"Trample
//	 When this creature enters, for each kind of counter on permanents you
//	 control, you may put your choice of a +1/+1 counter or a counter of
//	 that kind on this creature."
//
// The kinds are fixed as the trigger resolves (CR 608.2h) and asked about
// in name order, so a counter Bribe Taker earns on the way does not add a
// kind. Each kind is a separate "may": none, a +1/+1 counter, or one of that
// kind (for +1/+1 itself, the two are the same counter, so it is one
// option). Bribe Taker's own counters count if it has any.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "482889f4-6cf5-457f-9f52-a3f841e038b5",
		Name:            "Bribe Taker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bribe Taker — for each kind of counter on permanents you control, put a +1/+1 counter or a counter of that kind on this creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
						return nil
					}
					return bribeTakerAsk(ctx, sncB07KindsAmongPermanentsYouControl(g, item.Controller), item.SourceCardID)
				}),
		},
	})
}

// bribeTakerAsk asks about the first of the kinds, then the next.
func bribeTakerAsk(ctx *Context, kinds []string, self uuid.UUID) error {
	if len(kinds) == 0 || !onBattlefield(ctx.Game, self) {
		return nil
	}
	options := []game.ChoiceOption{{Label: "No counter"}, {Label: "A +1/+1 counter"}}
	if kinds[0] != game.CounterPlusOne {
		options = append(options, game.ChoiceOption{Label: "A " + kinds[0] + " counter"})
	}
	return PickOption{
		Question: "Bribe Taker — put a +1/+1 counter or a " + kinds[0] + " counter on this creature?",
		Options:  options,
		ThenKey:  bribeTakerPicked,
		Carry:    []uuid.UUID{self},
		Words:    kinds,
	}.Apply(ctx)
}

// bribeTakerPicked puts the chosen counter and asks about the rest. An
// on-disk key.
var bribeTakerPicked game.OptionPickThen

func init() {
	bribeTakerPicked = OptionPickThen("option-pick/bribe-taker-counter", func(ctx *Context, r game.OptionPicked) error {
		if r.Option == nil || len(r.Words) == 0 || len(r.Carry) == 0 {
			return nil
		}
		self, kind := r.Carry[0], r.Words[0]
		var add string
		switch r.Index {
		case 1:
			add = game.CounterPlusOne
		case 2:
			add = kind
		}
		if add != "" && onBattlefield(ctx.Game, self) {
			if err := ctx.Game.AddCounterByForEffect(ctx.Controller(), self, add, 1); err != nil {
				return err
			}
		}
		return bribeTakerAsk(ctx, r.Words[1:], self)
	})
}
