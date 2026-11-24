package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const meTheImmortalOracleID = "26ad696e-9f17-4110-9c59-b9211e662900"

// Me, the Immortal — Legendary Creature — Human Rogue {2}{G}{U}{R}, 3/3:
//
//	"At the beginning of combat on your turn, put your choice of a +1/+1,
//	 first strike, vigilance, or menace counter on Me.
//	 Counters remain on Me as it moves to any zone other than a player's
//	 hand or library.
//	 You may cast this card from your graveyard by discarding two cards in
//	 addition to paying its other costs."
//
// The first ability is Aragorn, Company Leader's four-way counter pick; the
// engine reads keyword counters itself (CR 122.1b). The second is
// Skullbriar's (game.RegisterCountersRemain, read by MoveCard): the counters
// stay on the card through the graveyard, exile, the stack and back onto the
// battlefield, and are lost the moment it goes to a hand or a library. The
// third is Rona, Sheoldred's Faithful's graveyard cast: the printed mana cost
// plus a discard of two cards (CastFromGraveyardByDiscarding).
func init() {
	game.RegisterCountersRemain(meTheImmortalOracleID)
	effects.Register(effects.Spec{
		OracleID:         meTheImmortalOracleID,
		Name:             "Me, the Immortal",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.CastFromGraveyardByDiscarding("{2}{G}{U}{R}", 2)},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Me, the Immortal — put your choice of a +1/+1, first strike, vigilance, or menace counter on Me", meTheImmortalPick),
		},
	})
}

var meTheImmortalCounterKinds = []string{game.CounterPlusOne, game.CounterFirstStrike, game.CounterVigilance, game.CounterMenace}

func meTheImmortalPick(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
		return nil
	}
	options := make([]game.ChoiceOption, len(meTheImmortalCounterKinds))
	for i, kind := range meTheImmortalCounterKinds {
		options[i] = game.ChoiceOption{Label: "A " + kind + " counter"}
	}
	return effects.PickOption{
		Question: "Me, the Immortal — put your choice of a counter on Me",
		Options:  options,
		ThenKey:  meTheImmortalCounterChosen,
		WithItem: true,
	}.Apply(ctx)
}

// meTheImmortalCounterChosen puts the chosen counter on Me, if Me is still
// the object that triggered. An on-disk key: never renamed, never reused.
var meTheImmortalCounterChosen = effects.OptionPickThen("option-pick/me-the-immortal-counter", func(ctx *effects.Context, r game.OptionPicked) error {
	index, item := r.Index, ctx.Item
	if index < 0 || index >= len(meTheImmortalCounterKinds) {
		return nil
	}
	if !effects.OnBattlefield(ctx.Game, item.SourceCardID) || effects.SourceIsNewObject(ctx.Game, item) {
		return nil
	}
	return ctx.Game.AddCounterByForEffect(item.Controller, item.SourceCardID, meTheImmortalCounterKinds[index], 1)
})
