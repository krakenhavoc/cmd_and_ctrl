package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orator of Ojutai — Creature — Bird Monk {1}{W}, 0/4 (Commander 2017, S82
// older sets):
//
//	"As an additional cost to cast this spell, you may reveal a Dragon card
//	 from your hand.
//	 Defender, flying
//	 When this creature enters, if you revealed a Dragon card or controlled a
//	 Dragon as you cast this spell, draw a card."
//
// The optional cost is Draconic Roar's behold a Dragon (CR 701.4: a Dragon
// revealed from hand or one you control). "Controlled a Dragon as you cast"
// is a number fixed at announcement (CastTimeValues, CR 601.2i) and carried
// onto the permanent (CastProvenance.CastValues, CR 400.7d), so a Dragon that
// dies in response, or arrives later, does not change it. The "if" is
// intervening (CR 603.4): checked as the creature enters and again as the
// trigger resolves, both against the cast record; a creature put onto the
// battlefield without being cast has none and draws nothing.
//
// No simplification.
func init() {
	const dragonKey = "controlled_a_dragon"
	revealedOrControlled := func(c game.Card) bool {
		return game.OptionalCostTimesPaid(c, c.Provenance.OptionalCosts, game.BeholdKey) > 0 || c.Provenance.CastValues[dragonKey] > 0
	}
	effects.Register(effects.Spec{
		OracleID:        "fbed6f1a-e224-413c-b13e-3555caaef02a",
		Name:            "Orator of Ojutai",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying"},
		OptionalCosts:   []game.AdditionalCost{effects.OptionalBehold("a", "Dragon")},
		CastTimeValues: []game.CastTimeValue{{
			Key:   dragonKey,
			Value: effects.ControlsADragonAsCast,
		}},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you revealed a Dragon card or controlled a Dragon as you cast this spell",
				Test: func(ctx *effects.Context) bool {
					c, ok := ctx.Game.LookupCardForEffect(ctx.Source())
					return ok && revealedOrControlled(c)
				},
			}, effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.Self(ev, source, lki, g) && revealedOrControlled(*source)
			}, "Orator of Ojutai — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
	})
}
