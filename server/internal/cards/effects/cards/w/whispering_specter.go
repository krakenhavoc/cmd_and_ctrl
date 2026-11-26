package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Whispering Specter — Creature — Phyrexian Specter {1}{B}{B}, 1/1 (New
// Phyrexia, #5575):
//
//	"Flying
//	 Infect (This creature deals damage to creatures in the form of -1/-1
//	 counters and to players in the form of poison counters.)
//	 Whenever this creature deals combat damage to a player, you may sacrifice
//	 it. If you do, that player discards a card for each poison counter they
//	 have."
//
// The "you may" is the controller's choice as the trigger resolves; "if you do"
// hangs off the Specter really leaving the battlefield. The number of cards is
// the poison counters the damaged player has by then (the infect damage that
// triggered this has already put its counters on them), and the player chooses
// what to discard (CR 701.9a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "18d270bf-d0d3-4fac-a8d2-5f920e789bb5",
		Name:            "Whispering Specter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "infect"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Whispering Specter — you may sacrifice it; if you do, that player discards a card for each poison counter they have",
				effects.Do(effects.MayChoice{
					Question: "Whispering Specter — sacrifice it to make that player discard a card for each poison counter they have?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     whisperingSpecterAnswer,
				})),
		},
	})
}

// whisperingSpecterAnswer is the "yes": sacrifice the Specter, then the
// discards once it has gone. An on-disk key.
var whisperingSpecterAnswer = effects.MayYes("confirm/whispering-specter-sacrifice", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	return effects.SacrificePermanent{Target: ctx.Source(), ThenTail: ctx.ItemTail(whisperingSpecterDiscard)}.Apply(ctx)
})

// whisperingSpecterDiscard is "that player discards a card for each poison
// counter they have", run once the Specter is gone. An on-disk key.
var whisperingSpecterDiscard = effects.CardTail("then/whispering-specter-discard", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) || ctx.Item == nil || ctx.Item.Trigger == nil {
		return nil
	}
	player := ctx.Item.Trigger.Event.Target
	p := ctx.Game.PlayerByIDForEffect(player)
	if p == nil || p.Counters[game.CounterPoison] <= 0 {
		return nil
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: ctx.Source(), N: p.Counters[game.CounterPoison]})
	return nil
})
