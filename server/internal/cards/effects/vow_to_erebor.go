package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vow to Erebor — Instant {1}{W} (The Hobbit):
//
//	"Untap target creature you control. It gets +2/+2 until end of turn. If
//	 it's a Dwarf, you may attach an Equipment you control to it."
//
// The untap and the pump are plain primitives on the target. The Dwarf check
// reads the creature as it is at resolution (it is still on the battlefield:
// the target was just untapped), through ctx.TextSubtype so a text-changing
// effect on the spell is honoured. The attach is part of the same resolution,
// so it is asked inline, Yuffie's way: a choose_cards prompt over the
// Equipment the controller has, with zero a legal answer. Every Equipment is
// offered, including one already attached to something else, which moves
// (CR 301.5c). The pick is re-checked on submit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f8961618-ae68-4d13-84eb-8b5464ce4971",
		Name:         "Vow to Erebor",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Vow to Erebor — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			g := ctx.Game
			host, ok := g.LookupCardForEffect(id)
			if !ok || !host.HasSubtype(ctx.TextSubtype("Dwarf")) || !g.Battlefield.Contains(id) {
				return nil
			}
			var equipment []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && c.HasSubtype("Equipment") {
					equipment = append(equipment, c.InstanceID)
				}
			}
			if len(equipment) == 0 {
				return nil
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  ctx.Controller(),
				Source:   item.SourceCardID,
				Question: "Vow to Erebor — you may attach an Equipment you control to the Dwarf",
				Cards:    equipment,
				Min:      0,
				Max:      1,
				ThenKey:  vowToEreborAttach,
				LiveZone: game.ZoneBattlefield,
				Carry:    []uuid.UUID{id},
			})
			return nil
		},
	})
}

// vowToEreborAttach is the Equipment attaching to the Dwarf, keyed (#4253):
// the Dwarf rides as the carry and the controller is the chooser.
var vowToEreborAttach = game.RegisterChooseCardsThen("choose-cards/vow-to-erebor-attach",
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
