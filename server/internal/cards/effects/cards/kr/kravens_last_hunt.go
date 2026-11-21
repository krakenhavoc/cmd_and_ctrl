package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kraven's Last Hunt — {3}{G} Enchantment — Saga (Marvel's Spider-Man
// promos, #4460):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Mill five cards. When you do, this Saga deals damage equal to
//	     the greatest power among creature cards in your graveyard to
//	     target creature.
//	 II — Target creature you control gets +2/+2 until end of turn.
//	 III — Return target creature card from your graveyard to your hand."
//
// Chapter I is Primal Witchstalker's mill-then-reflexive shape: with an
// empty library nothing is milled and there is no "when you do"; with
// cards to mill the CR 603.12 reflexive trigger picks its creature after
// the mill, so the graveyard it counts includes the five cards. The
// amount is read as the reflexive trigger resolves (CR 608.2h), each
// creature card's power as it is in the graveyard; a greatest power
// below 1 deals no damage. The Saga is the source, as it last existed if
// it has gone.
//
// Chapter II is Origin of the Hulk's until-end-of-turn pump; chapter III
// is The Binding of the Titans's return to hand.
//
// No simplification.
func init() {
	const name = "Kraven's Last Hunt"
	effects.Register(effects.Spec{
		OracleID:     "e74efcf2-d77f-4fb5-9027-77e462ec829e",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, effects.SagaChapterLabel(name, 1, "mill five cards; when you do, damage to target creature"),
				effects.Do(
					effects.Instruction(func(ctx *effects.Context) error {
						p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
						if p == nil || p.Library == nil || p.Library.Size() == 0 {
							return nil
						}
						ctx.RecordStep(kravensLastHuntMilled, nil, 1)
						return effects.MillCards{N: 5}.Apply(ctx)
					}),
					effects.Instruction(func(ctx *effects.Context) error {
						if ctx.StepRecord(kravensLastHuntMilled).N() == 0 {
							return nil
						}
						return effects.WhenYouDo(name+" — damage equal to the greatest power among creature cards in your graveyard to target creature",
							kravensLastHuntDamageBody).Apply(ctx)
					}),
				)),
			effects.ChapterTriggerTargeting(2, effects.SagaChapterLabel(name, 2, "target creature you control gets +2/+2 until end of turn"),
				effects.TargetCreature("target creature you control", effects.YouControl()),
				effects.Do(effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
					return effects.UntilEndOfTurn(ctx, id, nil, name+" — +2/+2", game.ModifyPTMod(2, 2))
				}))),
			effects.ChapterTriggerTargeting(3, effects.SagaChapterLabel(name, 3, "return target creature card from your graveyard to your hand"),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				effects.ReturnTargetCardToHand),
		},
	})
}

// kravensLastHuntMilled files that the library had a card to mill.
const kravensLastHuntMilled = "kravens-last-hunt/milled"

// kravensLastHuntDamageBody is chapter I's reflexive trigger. An on-disk
// key: never renamed, never reused.
var kravensLastHuntDamageBody = game.ReflexiveBody("kravens-last-hunt/damage",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		target, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		n := kravensGreatestGraveyardPower(g, item.Controller)
		if n <= 0 {
			return nil
		}
		return effects.DealDamage{Source: item.SourceCardID, Target: target, Amount: n}.Apply(ctx)
	},
	effects.ConstTargets(func() *game.TargetSpec { return effects.TargetCreature("target creature") }))

// kravensGreatestGraveyardPower is the greatest power among the creature
// cards in `player`'s graveyard, 0 with none.
func kravensGreatestGraveyardPower(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	best, found := 0, false
	for _, c := range p.Graveyard.Cards {
		if !c.IsCreature() {
			continue
		}
		if pw := c.PowerForComparison(); !found || pw > best {
			best, found = pw, true
		}
	}
	return best
}
