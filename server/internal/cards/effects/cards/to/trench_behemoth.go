package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Trench Behemoth — Creature — Kraken {5}{U}{U}, 7/7 (Commander Legends):
//
//	"Return a land you control to its owner's hand: Untap this creature. It
//	 gains hexproof until end of turn.
//	 Landfall — Whenever a land you control enters, target creature an opponent
//	 controls attacks during its controller's next combat phase if able."
//
// The activated ability's cost is the return of a land you control (any, a
// tapped one included), paid as it is announced; the Behemoth untaps and gains
// hexproof until end of turn, as the ability resolves. The landfall trigger
// targets a creature an opponent controls and pins the attack requirement to
// that creature (CR 400.7), read live against its controller: Sizzling Soloist's
// requirement, lasting through that controller's next turn's combat.
//
// No simplification.
func init() {
	const landfall = "Trench Behemoth — landfall: target creature an opponent controls attacks during its controller's next combat phase if able"
	effects.Register(effects.Spec{
		OracleID:     "dc2b1863-b45c-4213-ad89-5e39432daca7",
		Name:         "Trench Behemoth",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Return a land you control to its owner's hand: Untap this creature. It gains hexproof until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ReturnAPermanentToHand("a land you control", effects.Land()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.UntapTarget{Target: item.SourceCardID}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"hexproof"},
					Label:    "Trench Behemoth — hexproof until end of turn",
				}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Landfall(landfall, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.SncCreatureAttacksDuringItsControllersNextCombat(ctx, id,
					"Trench Behemoth — attacks during its controller's next combat phase if able")
			}), effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
