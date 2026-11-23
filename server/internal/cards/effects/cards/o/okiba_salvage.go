package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Okiba Salvage — Sorcery {4}{B} (Kamigawa: Neon Dynasty, #5493):
//
//	"Return target creature or Vehicle card from your graveyard to the
//	 battlefield. Then put two +1/+1 counters on that permanent if you control
//	 an artifact and an enchantment."
//
// The counters are not "enters with": "then" reads the board after the return
// (a returned Vehicle is itself an artifact), and puts the counters on the
// permanent that entered, through the entry's keyed tail so a Clone's copy
// choice or a pay-life entry that pauses is waited for. If the card did not
// enter, nothing is counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "142409ce-1187-4554-bc07-6432c4d22dbc",
		Name:         "Okiba Salvage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature or Vehicle card from your graveyard", effects.YouOwn(), effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return ctx.Game.ReturnFromGraveyardToBattlefieldKeyedForEffect(t.ID, ctx.Controller(), false, ctx.ItemTail(okibaSalvageCounters))
			}
			return nil
		},
	})
}

// okibaSalvageCounters is "then put two +1/+1 counters on that permanent if
// you control an artifact and an enchantment". An on-disk key.
var okibaSalvageCounters = effects.EnteredTail("okiba-salvage/counters", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil || !effects.NeoB08YouControlAnArtifactAndAnEnchantment(ctx.Game, ctx.Controller()) {
		return nil
	}
	return effects.AddCounter{Target: entered, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
})
