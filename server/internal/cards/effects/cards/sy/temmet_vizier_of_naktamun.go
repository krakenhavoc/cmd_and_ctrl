package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Temmet, Vizier of Naktamun — {W}{U} Legendary Creature — Human Cleric 2/2
// (Amonkhet, S82 old batch 247):
//
//	"At the beginning of combat on your turn, target creature token you control
//	 gets +1/+1 until end of turn and can't be blocked this turn.
//	 Embalm {3}{W}{U}"
//
// Both halves apply to the one target as the trigger resolves; a target that
// is no longer a legal creature token you control gets neither (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Temmet, Vizier of Naktamun — target creature token you control gets +1/+1 until end of turn and can't be blocked this turn"
	effects.Register(effects.Spec{
		OracleID:     "e245d5c9-ae3a-4246-bd74-c6fed76b9fad",
		Name:         "Temmet, Vizier of Naktamun",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Embalm("{3}{W}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Temmet — +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked, Label: "Temmet — can't be blocked"}.Apply(ctx)
			}), effects.TargetCreature("target creature token you control", effects.YouControl(), effects.IsTokenPredicate())),
		},
	})
}
