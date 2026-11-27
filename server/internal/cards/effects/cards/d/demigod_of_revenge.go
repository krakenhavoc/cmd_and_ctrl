package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Demigod of Revenge — Creature — Spirit Avatar {B/R}{B/R}{B/R}{B/R}{B/R},
// 5/4 (Shadowmoor, #5591):
//
//	"Flying, haste
//	 When you cast this spell, return all cards named Demigod of Revenge
//	 from your graveyard to the battlefield."
//
// The cast trigger is the FromStack shape (CR 603.2, 601.2i): it resolves
// above the spell, so the cards come back before the spell does and even if
// the spell is countered. Cards are matched by exact name in the
// controller's graveyard as it resolves; each returns under its owner's
// control, which is the controller's since it is their graveyard.
//
// No simplification.
func init() {
	const label = "Demigod of Revenge — return all cards named Demigod of Revenge from your graveyard to the battlefield"
	effects.Register(effects.Spec{
		OracleID:        "602ce676-380d-4c3c-b97a-9234ccdc2904",
		Name:            "Demigod of Revenge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range effects.GraveyardIDs(g, item.Controller, func(c game.Card) bool { return c.Name == "Demigod of Revenge" }) {
					if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: item.Controller}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
