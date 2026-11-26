package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stone Idol Trap — Instant — Trap {5}{R} (Worldwake, S82 OLD b488):
//
//	"This spell costs {1} less to cast for each attacking creature.
//	 Create a 6/12 colorless Construct artifact creature token with trample.
//	 Exile it at the beginning of your next end step."
//
// Ancient Stone Idol's reduction: a self cost modifier counting every attacking creature,
// whoever controls it, judged as the spell is cast (CR 601.2f). The token is created by
// the spell and exiled by a delayed trigger of YOURS at the beginning of your next end step
// (CR 603.7): cast on an opponent's turn it waits for your own turn's end step, cast on
// yours before the end step it fires on this turn's. A token that has already left the
// battlefield is simply not there to be exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47cd98e2-0eaf-4292-a840-c5e8d035c83c",
		Name:         "Stone Idol Trap",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.PermanentsOnBattlefield(effects.AttackingCreature()),
				"This spell costs {1} less to cast for each attacking creature."),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: ctx.Controller(),
				Source:     ctx.Source(),
				Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("6/12 colorless Construct artifact with trample")), Count: 1}},
			}, ctx.ItemTail(stoneIdolTrapExile))
		},
	})
}

// stoneIdolTrapExile exiles the Construct at the beginning of your next end step. An
// on-disk key: never renamed, never reused.
var stoneIdolTrapExile = effects.S82OldB488ExileTokensAtYourNextEndStep("then/stone-idol-trap-exile",
	"Stone Idol Trap — exile the Construct")
