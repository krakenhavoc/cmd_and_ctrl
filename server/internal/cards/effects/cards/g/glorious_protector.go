package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glorious Protector — Creature — Angel Cleric {2}{W}{W}, 3/4 (Kaldheim,
// reprinted in Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"Flash
//	 Flying
//	 When this creature enters, you may exile any number of non-Angel creatures
//	 you control until this creature leaves the battlefield.
//	 Foretell {2}{W}"
//
// Foretell is the keyword's special action (CR 702.143). The enters trigger
// asks as it resolves for any number (including none) of the non-Angel
// creatures the controller controls, Teferi Akosa's choose_cards prompt with a
// floor of zero, and exiles each pick that is still a non-Angel creature they
// control with ExileUntil{ThisLeaves} (Werefox Bodyguard's duration, CR
// 610.3): the cards return when this creature leaves the battlefield, and
// nothing is exiled if it has already gone (CR 610.3c).
//
// No simplification.
func init() {
	const label = "Glorious Protector — you may exile any number of non-Angel creatures you control until this creature leaves the battlefield"
	pick := effects.AskedPick("choose-cards/glorious-protector-exile", func(ctx *effects.Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !c.IsCreature() || c.HasSubtypeWord("Angel") || c.Controller != ctx.Controller() {
				continue
			}
			if err := (effects.ExileUntil{Target: id, ThisLeaves: true,
				Label: "Glorious Protector — the exiled creature returns when Glorious Protector leaves the battlefield"}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
	effects.Register(effects.Spec{
		OracleID:        "6bdc4996-f7d9-4dc4-b70a-aa63f6816b1d",
		Name:            "Glorious Protector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		SpecialActions:  []game.SpecialAction{effects.Foretell("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				var candidates []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.IsCreature() && c.Controller == item.Controller && !c.HasSubtypeWord("Angel") {
						candidates = append(candidates, c.InstanceID)
					}
				}
				if len(candidates) == 0 {
					return nil
				}
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Glorious Protector — exile any number of non-Angel creatures you control until this creature leaves the battlefield",
					Cards:    candidates,
					Min:      0,
					Max:      0,
					LiveZone: game.ZoneBattlefield,
					ThenKey:  pick,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			}),
		},
	})
}
