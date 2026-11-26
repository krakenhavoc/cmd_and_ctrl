package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dimensional Infiltrator — {1}{U} Creature — Eldrazi 2/1 (Oath of the
// Gatewatch, S82 OLD b296):
//
//	"Devoid (This card has no color.)
//	 Flash
//	 Flying
//	 {1}{C}: Target opponent exiles the top card of their library. If it's a
//	 land card, you may return this creature to its owner's hand."
//
// An exile off the top of the target opponent's library whose tail reads the
// card that landed in exile; a land card opens Mistbreath Elder's "you may
// return this creature" question, which only returns the same object that
// activated (CR 400.7). The {C} in the cost is paid with colorless mana only.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b93732e4-7aeb-4c37-a415-2b0ff46e2056",
		Name:            "Dimensional Infiltrator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flash", "flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{C}: Target opponent exiles the top card of their library. If it's a land card, you may return this creature to its owner's hand.",
			Cost:    effects.ManaCost("{1}{C}"),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id := effects.TargetedPlayerOf(ctx)
				if id == uuid.Nil {
					return nil
				}
				return effects.MillToZone{Player: id, N: 1, To: game.ZoneExile, ThenTail: ctx.ItemTail(dimensionalInfiltratorExiled)}.Apply(ctx)
			},
		}},
	})
}

// dimensionalInfiltratorExiled is "if it's a land card, you may return this
// creature to its owner's hand". An on-disk key: never renamed, never reused.
var dimensionalInfiltratorExiled = effects.CardTail("then/dimensional-infiltrator-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || !c.IsLand() {
		return nil
	}
	return effects.MayChoice{
		Question: "Dimensional Infiltrator — return this creature to its owner's hand?",
		Then:     dimensionalInfiltratorReturn,
	}.Apply(ctx)
})

// dimensionalInfiltratorReturn is the yes answer: the Infiltrator that
// activated goes back to its owner's hand. An on-disk key.
var dimensionalInfiltratorReturn = effects.MayYes("confirm/dimensional-infiltrator-return", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	if !effects.OnBattlefield(ctx.Game, ctx.Source()) || ctx.Game.AbilitySourceIsNewObjectForEffect(ctx.Item) {
		return nil
	}
	return effects.BounceToHand{Target: ctx.Source()}.Apply(ctx)
})
