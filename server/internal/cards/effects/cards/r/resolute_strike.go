package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Resolute Strike — Instant {W} (Zendikar Rising, #5508):
//
//	"Target creature gets +2/+2 until end of turn. If it's a Warrior, you
//	 may attach an Equipment you control to it."
//
// Vow to Erebor's shape: the pump, then, if the target is a Warrior as the
// spell resolves (through ctx.TextSubtype, so a text-changing effect is
// honoured), a choose_cards prompt over the Equipment its controller has
// that could legally be attached to it (CR 701.3b), with zero a legal answer.
// An Equipment already attached elsewhere is offered and moves (CR 301.5c).
// The pick is re-checked when submitted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "563e9574-3f60-4558-88cb-aaf97caa7515",
		Name:         "Resolute Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Resolute Strike — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			g := ctx.Game
			host, ok := g.LookupCardForEffect(id)
			if !ok || !host.HasSubtype(ctx.TextSubtype("Warrior")) || !g.Battlefield.Contains(id) {
				return nil
			}
			var equipment []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.HasSubtype("Equipment") && !c.IsAttachedTo(id) && g.AttachLegalForEffect(c.InstanceID, id) {
					equipment = append(equipment, c.InstanceID)
				}
			}
			if len(equipment) == 0 {
				return nil
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   item.SourceCardID,
				Question: "Resolute Strike — you may attach an Equipment you control to the Warrior",
				Cards:    equipment,
				Min:      0,
				Max:      1,
				ThenKey:  resoluteStrikeAttach,
				LiveZone: game.ZoneBattlefield,
				Carry:    []uuid.UUID{id},
			})
			return nil
		},
	})
}

// resoluteStrikeAttach is the Equipment attaching to the Warrior, keyed: the
// Warrior rides as the carry and the controller is the chooser. An on-disk key.
var resoluteStrikeAttach = game.RegisterChooseCardsThen("choose-cards/resolute-strike-attach",
	func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 || len(r.Carry) == 0 || !g.Battlefield.Contains(r.Carry[0]) || !g.Battlefield.Contains(r.Picked[0]) {
			return nil
		}
		eq, ok := g.LookupCardForEffect(r.Picked[0])
		if !ok || eq.Controller != r.Chooser || !eq.HasSubtype("Equipment") {
			return nil
		}
		return g.AttachForEffect(r.Picked[0], game.TargetRef{Kind: game.TargetCard, ID: r.Carry[0]})
	})
