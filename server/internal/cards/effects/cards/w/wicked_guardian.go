package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wicked Guardian — {3}{B} Creature — Human Noble 4/2 (Jumpstart 2022, #5948):
//
//	"When this creature enters, you may have it deal 2 damage to another
//	 creature you control. If you do, draw a card."
//
// The creature is chosen as the trigger resolves, not targeted (a
// hexproof creature of yours can be picked), with "up to one" standing
// for the "may": picking none declines. Choosing is "doing": the card is
// drawn once the creature is picked, whether or not the damage is
// prevented.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dda70e25-b3c5-444d-9dff-1366771bf881",
		Name:         "Wicked Guardian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wicked Guardian — you may have it deal 2 damage to another creature you control; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ChoosePermanents{
						Question:   "Wicked Guardian — have it deal 2 damage to another creature you control?",
						Candidates: wickedGuardianCandidates(item.SourceCardID),
						Keyed:      wickedGuardianDamage,
						Carry:      []uuid.UUID{item.SourceCardID},
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

func wickedGuardianCandidates(self uuid.UUID) func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	return func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
		var ids []uuid.UUID
		for _, id := range effects.PermanentsControlledByMatching(g, of, effects.Creature()) {
			if id != self {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil, 0, 0
		}
		return ids, 0, 1
	}
}

var wickedGuardianDamage = effects.RegisterPermanentsPick("prompt-run/wicked-guardian-damage", effects.PermanentsDef{
	ThenRun: func(ctx *effects.Context, picked game.PromptedPicks, r game.RunDone) error {
		cards := picked.Cards()
		if len(cards) == 0 {
			return nil
		}
		var source uuid.UUID
		if len(r.Carry) > 0 {
			source = r.Carry[0]
		}
		if err := (effects.DealDamage{Source: source, Target: cards[0], Amount: 2}).Apply(ctx); err != nil {
			return err
		}
		return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	},
})
