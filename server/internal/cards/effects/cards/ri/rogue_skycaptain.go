package cardsri

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rogue Skycaptain — Creature — Human Rogue Mercenary {2}{R} 3/4 (Alliances,
// S82 OLD batch 818):
//
//	"Flying
//	 At the beginning of your upkeep, put a wage counter on this creature.
//	 You may pay {2} for each wage counter on it. If you don't, remove all
//	 wage counters from this creature and an opponent gains control of it."
//
// The counter goes on first and the cost is counted from the counters that
// are there once it lands (the cumulative-upkeep order). Declining, or being
// unable to pay, removes every wage counter and the controller names the
// opponent who takes the creature for good.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e600136a-8b4f-4725-a29e-021e0cfc7d0a",
		Name:            "Rogue Skycaptain",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Rogue Skycaptain — put a wage counter on it; pay {2} for each wage counter or an opponent gains control of it",
				func(g *game.Game, item *game.StackItem) error {
					source := item.SourceCardID
					if !effects.OnBattlefield(g, source) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					return g.AddCounterByKeyedForEffect(uuid.Nil, source, oldB818WageCounter, 1, effects.NewContext(g, item).ItemTail(rogueSkycaptainPay))
				}),
		},
	})
}

const oldB818WageCounter = "wage"

var rogueSkycaptainPay = effects.CardTail("rogue-skycaptain/pay-wages", func(ctx *effects.Context, _ game.TailOutcome) error {
	source := ctx.Item.SourceCardID
	n := effects.CounterCountOn(ctx.Game, source, oldB818WageCounter)
	if n <= 0 {
		return nil
	}
	return effects.UpkeepPayUnless{
		Chooser:  ctx.Controller(),
		Cost:     strings.Repeat("{2}", n),
		Question: "Rogue Skycaptain — pay {2} for each wage counter on it, or an opponent gains control of it",
		Else:     rogueSkycaptainWages,
	}.Apply(ctx)
})

var rogueSkycaptainWages = effects.PayUnlessBranches("pay-unless/rogue-skycaptain-wages", nil,
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		source := r.Source
		if !effects.OnBattlefield(ctx.Game, source) {
			return nil
		}
		if n := effects.CounterCountOn(ctx.Game, source, oldB818WageCounter); n > 0 {
			if err := ctx.Game.AddCounterForEffect(source, oldB818WageCounter, -n); err != nil {
				return err
			}
		}
		var opponents []uuid.UUID
		for _, p := range ctx.Game.Seats {
			if p != nil && !p.Eliminated && p.ID != r.Chooser {
				opponents = append(opponents, p.ID)
			}
		}
		switch len(opponents) {
		case 0:
			return nil
		case 1:
			return oldB818OpponentTakesIt(ctx.Game, source, opponents[0])
		}
		ctx.Game.QueueChoosePlayerForEffect(game.ChoosePlayerPrompt{
			Chooser:  r.Chooser,
			Among:    opponents,
			Source:   source,
			Question: "Rogue Skycaptain — choose an opponent to gain control of it",
			ThenKey:  rogueSkycaptainOpponentChosen,
		})
		return nil
	})

var rogueSkycaptainOpponentChosen = effects.OptionPickThen("option-pick/rogue-skycaptain-opponent", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || r.Option.Player == uuid.Nil {
		return nil
	}
	return oldB818OpponentTakesIt(ctx.Game, r.Source, r.Option.Player)
})

func oldB818OpponentTakesIt(g *game.Game, source, to uuid.UUID) error {
	if g.GainControlForEffect(source, source, to, game.IndefiniteDuration(), "Rogue Skycaptain — an opponent gains control of it") {
		g.RecomputeLayersIfStaleLocked()
	}
	return nil
}
