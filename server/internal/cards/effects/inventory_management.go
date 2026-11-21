package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inventory Management — Instant {R}{W} (Fallout):
//
//	"Split second (As long as this spell is on the stack, players can't cast
//	 spells or activate abilities that aren't mana abilities.)
//	 For each Aura and Equipment you control, you may attach it to a creature
//	 you control."
//
// Split second is the engine keyword (Krosan Grip). The set of Auras and
// Equipment is read once as the spell resolves (CR 608.2c), and for each the
// controller is offered the creatures they control with none a legal answer
// (a keyed choose_cards prompt each, carrying the attachment). Each answer
// re-checks both permanents and attaches through the engine, which refuses
// what the permanent can't legally be attached to (an Aura's enchant clause,
// CR 303.4; an Equipment on a non-creature, CR 301.5c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b20eae3d-d0ff-42c3-a060-d06629506781",
		Name:            "Inventory Management",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSplitSecond},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			g := ctx.Game
			var creatures []uuid.UUID
			var attachments []game.Card
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller != ctx.Controller() {
					continue
				}
				if c.IsCreature() {
					creatures = append(creatures, c.InstanceID)
				}
				if c.HasSubtype("Aura") || c.HasSubtype("Equipment") {
					attachments = append(attachments, c)
				}
			}
			if len(creatures) == 0 {
				return nil
			}
			for _, a := range attachments {
				g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  ctx.Controller(),
					Source:   ctx.Source(),
					Question: "Inventory Management — you may attach " + a.Name + " to a creature you control",
					Cards:    creatures,
					Min:      0,
					Max:      1,
					ThenKey:  inventoryManagementAttach,
					Carry:    []uuid.UUID{a.InstanceID},
					Asked:    game.NewAskedItem(ctx.Item),
				})
			}
			return nil
		},
	})
}

// inventoryManagementAttach attaches the carried Aura or Equipment to the
// creature picked, when both are still on the battlefield. An on-disk key:
// never renamed, never reused.
var inventoryManagementAttach = askedPick("choose-cards/inventory-management-attach", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	if !onBattlefield(ctx.Game, r.Carry[0]) || !onBattlefield(ctx.Game, r.Picked[0]) {
		return nil
	}
	return ctx.Game.AttachForEffect(r.Carry[0], game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]})
})
