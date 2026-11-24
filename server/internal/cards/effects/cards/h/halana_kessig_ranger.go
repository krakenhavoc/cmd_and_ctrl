package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Halana, Kessig Ranger — Legendary Creature — Human Archer Ranger {3}{G},
// 3/4 (Commander Legends):
//
//	"Reach
//	 Whenever another creature you control enters, you may pay {2}. When
//	 you do, that creature deals damage equal to its power to target
//	 creature.
//	 Partner"
//
// Anointer of Valor's shape: the entry trigger's "may pay {2}" is the
// parent's prompt (MayPay) and the "when you do" is a CR 603.12
// reflexive trigger carrying the entered creature, with its own target
// (chosen as it goes on the stack, CR 603.3d). The damage is dealt by
// that creature, for its power as the reflexive trigger resolves, or as
// it last existed on the battlefield if it has left (CR 608.2h); a
// creature with no power deals none. Partner is a deck-construction rule
// (CR 702.124).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f5aa7385-12ea-4ecb-86eb-e8622ebb6908",
		Name:            "Halana, Kessig Ranger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverAnotherCreatureEntersUnderYourControl(halanaKessigRangerLabel,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: halanaKessigRangerLabel + "?",
						Then:     halanaKessigRangerPaid,
						Carry:    []uuid.UUID{ctx.Trigger().Event.CardID},
					}.Apply(ctx)
				}),
		},
	})
}

const halanaKessigRangerLabel = "Halana, Kessig Ranger — pay {2}; when you do, that creature deals damage equal to its power to target creature"

// halanaKessigRangerPaid is the paid branch: the reflexive "when you do".
// An on-disk key.
var halanaKessigRangerPaid = effects.PayUnlessBranches("pay-unless/halana-kessig-ranger-damage",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Halana, Kessig Ranger — that creature deals damage equal to its power to target creature",
			Cards: r.Carry,
			Body:  halanaKessigRangerDamageBody,
		}.Apply(ctx)
	}, nil)

// halanaKessigRangerDamageBody is the damage. An on-disk key.
var halanaKessigRangerDamageBody = game.ReflexiveBody("halana-kessig-ranger/damage",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		entered := ctx.PayloadCards()
		if len(entered) == 0 {
			return nil
		}
		target, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		power := effects.B13LastKnownPower(g, entered[0], game.Characteristic{})
		if c, onField := g.LookupCardForEffect(entered[0]); onField {
			if z := g.FindCardZoneForEffect(entered[0]); z != nil && z.Kind == game.ZoneBattlefield {
				g.RecomputeLayersIfStaleLocked()
				power = c.CurrentPower()
			}
		}
		if power <= 0 {
			return nil
		}
		return effects.DealDamage{Source: entered[0], Target: target, Amount: power}.Apply(ctx)
	}),
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))
