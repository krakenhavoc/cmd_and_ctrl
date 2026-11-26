package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Talus Paladin — Creature — Human Knight Ally {3}{W}, 2/3 (Worldwake, S82 OLD b487):
//
//	"Whenever this creature or another Ally you control enters, you may have
//	 Allies you control gain lifelink until end of turn, and you may put a +1/+1
//	 counter on this creature."
//
// Two independent "you may"s inside one resolution (CR 608.2d), asked in order: the
// lifelink for the Allies you control as it resolves (the set is locked then, CR
// 611.2c, so an Ally that enters later does not gain it, and "Ally" is read through the
// card's text), then the counter, which is asked whatever the first answer was and
// goes on the Paladin only while it is still the same object on the battlefield
// (CR 400.7). The trigger fires for the Paladin's own entry and for each other Ally
// entering under your control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98914c0e-967f-4c8b-bd74-e2ce15c7fe95",
		Name:         "Talus Paladin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.WwkThisOrAnotherAllyEntered,
				"Talus Paladin — you may have Allies you control gain lifelink until end of turn, and you may put a +1/+1 counter on this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: "Talus Paladin — have Allies you control gain lifelink until end of turn?",
						Then:     talusPaladinLifelink,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// talusPaladinLifelink answers the first "you may": on a yes the Allies you control
// gain lifelink; either way the second "you may" is asked. An on-disk key: never
// renamed, never reused.
var talusPaladinLifelink = effects.MayAnswer("confirm/talus-paladin-lifelink", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		err := effects.GrantKeywordUntilEOT{
			Match:    effects.And(effects.Creature(), effects.YouControl(), effects.Subtype(ctx.TextSubtype("Ally"))),
			Keywords: []string{"lifelink"},
			Label:    "Talus Paladin — Allies gain lifelink until end of turn",
		}.Apply(ctx)
		if err != nil {
			return err
		}
	}
	return effects.MayChoice{
		Player:   r.Chooser,
		Question: "Talus Paladin — put a +1/+1 counter on this creature?",
		Then:     talusPaladinCounter,
		Carry:    []uuid.UUID{},
	}.Apply(ctx)
})

// talusPaladinCounter answers the second "you may". An on-disk key: never renamed,
// never reused.
var talusPaladinCounter = effects.MayYes("confirm/talus-paladin-counter", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	return effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
