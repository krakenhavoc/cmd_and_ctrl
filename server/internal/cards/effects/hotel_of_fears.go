package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hotel of Fears — Plane — Spacecraft (Doctor Who, #4870):
//
//	"At the beginning of your upkeep, exile the top card of your library.
//	 You lose life equal to its mana value. You may play that card this
//	 turn.
//	 Praise Him — Whenever chaos ensues, choose a color. Put X +1/+1
//	 counters on target creature you control, where X is your devotion to
//	 that color. Then sacrifice another creature. (Your devotion to a
//	 color is the number of mana symbols of that color in the mana costs
//	 of permanents you control.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6),
// so "your upkeep" is the planar controller's. The exiled card may be
// played until end of turn (an impulse permission), and the life lost is
// its mana value. The chaos ability (CR 311.7) asks the colour as it
// resolves (a keyed tail, so the pause is a restore point), counts
// devotion then (CR 700.5), and the planar controller then sacrifices a
// creature other than the target, chosen as the sacrifice happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79d3e356-9eb9-49bb-b4ba-f4f71876869e",
		Name:         "Hotel of Fears",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Hotel of Fears — exile the top card of your library, lose life equal to its mana value; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					ids, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, 1, game.CastPermission{})
					if err != nil {
						return err
					}
					for _, id := range ids {
						c, ok := g.LookupCardForEffect(id)
						if !ok || c.ManaValue() <= 0 {
							continue
						}
						if err := g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -c.ManaValue()); err != nil {
							return err
						}
					}
					return nil
				}),
			Targeting(WheneverChaosEnsues("Praise Him — Hotel of Fears: choose a color; X +1/+1 counters on target creature you control (X = devotion); then sacrifice another creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if _, ok := b16FirstLegalTargetCard(ctx); !ok {
						return nil
					}
					ChooseColorTail(game.ColorForBenefit, g, ctx.Controller(), ctx.Source(),
						"Praise Him — choose a color for your devotion", ctx.ItemTail(hotelOfFearsPraiseHim))
					return nil
				}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

// hotelOfFearsPraiseHim is the rest of the chaos ability once the colour
// is chosen: the counters, then the sacrifice. An on-disk key: never
// renamed, never reused.
var hotelOfFearsPraiseHim = CardTail("hotel-of-fears/praise-him", func(ctx *Context, o game.TailOutcome) error {
	t, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if x := devotionTo(ctx.Game, ctx.Controller(), o.Tail.Word); x > 0 {
		if err := (AddCounter{Target: t, Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
			return err
		}
	}
	var others []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && c.InstanceID != t {
			others = append(others, c.InstanceID)
		}
	}
	return SacrificeChoice{
		Player:     ctx.Controller(),
		Candidates: others,
		Question:   "Praise Him — sacrifice another creature",
	}.Apply(ctx)
})
