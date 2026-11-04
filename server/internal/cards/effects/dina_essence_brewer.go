package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dina, Essence Brewer — Legendary Creature — Dryad Druid {1}{B}{G}, 2/3
// (Secrets of Strixhaven Commander, #4834):
//
//	"Whenever you sacrifice a creature, draw a card. This ability
//	 triggers only once each turn.
//	 {2}, {T}, Sacrifice another creature: You gain X life and put X
//	 +1/+1 counters on target creature you control, where X is the
//	 sacrificed creature's power."
//
// The trigger watches EventSacrifice, which fires before the zone move,
// so the sacrificed permanent is still there to be read as a creature.
// "Only once each turn" is the turn's trigger tally for this source and
// row (b11TriggeredThisTurn), and OncePerBatch keeps two creatures
// sacrificed at once to one trigger.
//
// X is the sacrificed creature's power as it last existed on the
// battlefield (CR 608.2h, Context.SacrificedPower), floored at zero.
// The counters wait for a prompt the life gain asks (a CR 616 order).
// A target that is gone by resolution leaves the ability with no legal
// target, and it does nothing at all (CR 608.2b).
//
// No simplification.
const dinaEssenceBrewerDrawLabel = "Dina, Essence Brewer — draw a card (you sacrificed a creature)"

func init() {
	Register(Spec{
		OracleID:     "f61c1dc4-2f09-4b50-957f-ee656c659072",
		Name:         "Dina, Essence Brewer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && !b11TriggeredThisTurn(g, source.InstanceID, dinaEssenceBrewerDrawLabel)
			}, dinaEssenceBrewerDrawLabel, Do(DrawCards{N: 1}))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice another creature: You gain X life and put X +1/+1 counters on target creature you control, where X is the sacrificed creature's power",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: Do(
				Instruction(func(ctx *Context) error {
					return GainLife{Player: ctx.Controller(), Amount: ctx.SacrificedPower()}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return AddCounter{Target: id, Kind: game.CounterPlusOne, N: ctx.SacrificedPower()}.Apply(ctx)
				}),
			),
		}},
	})
}
