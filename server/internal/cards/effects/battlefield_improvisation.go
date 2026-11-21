package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Battlefield Improvisation — Instant {1}{W}:
//
//	"Target creature gets +2/+2 until end of turn. If that creature is
//	 attacking, you may attach any number of Equipment you control to it."
//
// The pump is Giant Growth's. The attacking check is read as the spell
// resolves, after the pump (CR 608.2c: the instructions are followed in
// order). Then the controller picks any number (a floor of zero) of the
// Equipment they control, and each one still on the battlefield is attached
// to the creature if it is still on the battlefield (CR 701.3), the pick
// shape Armed and Armored uses. A target that is no longer legal makes the
// whole spell do nothing (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "c5b85aa7-6a76-4fe2-a63e-4f376b6c9af7",
		Name:         "Battlefield Improvisation",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Battlefield Improvisation — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			host, found := ctx.Game.LookupCardForEffect(id)
			if !found || host.AttackingTarget == uuid.Nil {
				return nil
			}
			you := ctx.Controller()
			equipment := armedAndArmoredPermanents(ctx.Game, you, func(c game.Card) bool { return c.HasSubtype("Equipment") })
			if len(equipment) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  you,
				Source:   item.SourceCardID,
				Question: "Battlefield Improvisation — attach any number of Equipment you control to the attacking creature",
				Cards:    equipment,
				Min:      0,
				Max:      len(equipment),
				LiveZone: game.ZoneBattlefield,
				ThenKey:  battlefieldImprovisationEquipmentChosen,
				Carry:    []uuid.UUID{id},
			})
			return nil
		},
	})
}

// battlefieldImprovisationEquipmentChosen attaches each Equipment picked to
// the carried creature, if both are still on the battlefield. An on-disk
// key.
var battlefieldImprovisationEquipmentChosen = game.RegisterChooseCardsThen("choose-cards/battlefield-improvisation-equipment", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Carry) == 0 || !g.Battlefield.Contains(r.Carry[0]) {
		return nil
	}
	for _, id := range r.Picked {
		if !g.Battlefield.Contains(id) {
			continue
		}
		if err := g.AttachForEffect(id, game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]}); err != nil {
			return err
		}
	}
	return nil
})
