package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wick, the Whorled Mind — Legendary Creature — Rat Warlock {3}{B}, 2/4
// (Bloomburrow, #4467):
//
//	"Whenever Wick or another Rat you control enters, create a 1/1 black
//	 Snail creature token if you don't control a Snail. Otherwise, put a
//	 +1/+1 counter on a Snail you control.
//	 {U}{B}{R}, Sacrifice a Snail: Wick deals damage equal to the
//	 sacrificed creature's power to each opponent. Then draw cards equal
//	 to the sacrificed creature's power."
//
// The trigger is Ganax's "this or another <type> you control" shape. The
// Snail question is asked as it resolves (it is not an intervening if):
// no Snail makes the token, one Snail takes the counter, and several ask
// the controller which gets it (a choose_cards prompt over the Snails, the
// answer checked against the battlefield). The sacrifice is a cost; the
// damage and the draw read the sacrificed creature's power as it last
// existed on the battlefield (CR 608.2h, Context.SacrificedPower), floored
// at zero, and Wick need not be on the battlefield for the damage to be
// his (the ruling). The Snail token is the printed 1/1 black Snail.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5af729e8-8a40-4991-8270-612e377451de",
		Name:         "Wick, the Whorled Mind",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, wickRatYouControlEntered,
				"Wick, the Whorled Mind — create a 1/1 black Snail token if you don't control a Snail; otherwise put a +1/+1 counter on a Snail you control",
				wickSnail),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}{B}{R}, Sacrifice a Snail: Wick deals damage equal to the sacrificed creature's power to each opponent. Then draw cards equal to the sacrificed creature's power.",
			Cost:    effects.Plus(effects.ManaCost("{U}{B}{R}"), effects.SacrificeN(1, "a Snail", effects.Subtype("Snail"))),
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Effect: effects.Do(
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.DamageEachOpponent{N: ctx.SacrificedPower()}.Apply(ctx)
				}),
				effects.Instruction(func(ctx *effects.Context) error {
					return effects.DrawCards{N: ctx.SacrificedPower()}.Apply(ctx)
				}),
			),
		}},
	})
}

// wickRatYouControlEntered is "Wick or another Rat you control enters":
// a Rat entered under the source's controller's control, the source itself
// included.
func wickRatYouControlEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
	return ok && c.HasSubtype(source.TextSubtype("Rat"))
}

// wickSnail is the trigger's body.
func wickSnail(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	snails := effects.PermanentsControlledByMatching(g, ctx.Controller(), effects.Subtype(ctx.TextSubtype("Snail")))
	switch len(snails) {
	case 0:
		return effects.CreateToken{Template: effects.TokenCard("1/1 black Snail"), N: 1}.Apply(ctx)
	case 1:
		return effects.AddCounter{Target: snails[0], Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Wick, the Whorled Mind — put a +1/+1 counter on a Snail you control",
		Cards:      snails,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    wickCounterPicked,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// wickCounterPicked puts the counter on the chosen Snail. An on-disk key.
var wickCounterPicked = effects.AskedPick("choose-cards/wick-the-whorled-mind-snail", func(ctx *effects.Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
			continue
		}
		return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return nil
})
