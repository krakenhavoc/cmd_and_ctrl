package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vulshok Battlemaster — Creature — Human Warrior {4}{R}, 2/2 (Mirrodin,
// S82 OLD, #5606):
//
//	"Haste
//	 When this creature enters, attach all Equipment on the battlefield to
//	 it. (Control of the Equipment doesn't change.)"
//
// Balan's attach-all (C17AttachAllEquipmentToSource) without its "you
// control": every Equipment on the battlefield, whoever controls it, is
// attached to the Battlemaster as the trigger resolves, if it is still the
// permanent that entered (CR 400.7). The controller of an Equipment does not
// change (CR 301.5c), and one that can't be attached to it (it can't equip a
// creature with protection from it, say) stays where it is (CR 701.3b).
//
// No simplification.
func init() {
	const name = "Vulshok Battlemaster"
	effects.Register(effects.Spec{
		OracleID:        "cf01312d-305a-4587-8f54-03ee471bdf8c",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(name+" — attach all Equipment on the battlefield to it", func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				var gear []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.HasSubtype("Equipment") {
						gear = append(gear, c.InstanceID)
					}
				}
				host := game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID}
				for _, id := range gear {
					if err := g.AttachForEffect(id, host); err != nil {
						return err
					}
				}
				return nil
			}),
		},
	})
}
