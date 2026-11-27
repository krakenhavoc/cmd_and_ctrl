package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Legion Loyalist — Creature — Goblin Soldier {R}, 1/1 (Gatecrash, #5565):
//
//	"Haste
//	 Battalion — Whenever this creature and at least two other creatures
//	 attack, creatures you control gain first strike and trample until end of
//	 turn and can't be blocked by creature tokens this turn."
//
// Boros Strike-Captain's battalion condition (BattalionAttack). The set of
// creatures you control is locked as the ability resolves (CR 611.2c), so a
// creature that arrives later this turn gets none of it; one record gives
// that set first strike and trample (layer 6) and the block rule that bars
// creature tokens from blocking it (cantBeBlockedByTokens), all until end of
// turn. Whether a blocker is a token is read live as blockers are declared,
// and the attackers have been declared already, so the rule matters for the
// creatures that attack this combat.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "560a5452-0dc3-421d-bf16-8843a35dae6f",
		Name:            "Legion Loyalist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.BattalionAttack,
				"Legion Loyalist — battalion: creatures you control gain first strike and trample and can't be blocked by creature tokens this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Match: effects.And(effects.Creature(), effects.YouControl()),
						Mods: []game.Mod{
							game.AddKeywordsMod("first strike", "trample"),
							game.CantBeBlockedByTokensMod("creature tokens"),
						},
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Legion Loyalist — first strike, trample and can't be blocked by creature tokens until end of turn",
					}.Apply(ctx)
				}),
		},
	})
}
