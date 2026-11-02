package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alibou, Ancient Witness — Legendary Artifact Creature — Golem
// {3}{R}{W}, 4/5:
//
//	"Other artifact creatures you control have haste.
//	 Whenever one or more artifact creatures you control attack, Alibou
//	 deals X damage to any target and you scry X, where X is the number
//	 of tapped artifacts you control."
//
// The haste is a layer 6 grant to every other artifact creature its
// controller controls, read from effective types. The trigger fires once
// per declaration (a batch, CR 603.2c) when at least one declared
// attacker you control is an artifact creature — Alibou itself counts,
// being one. It targets any target as it goes on the stack.
//
// X is the number of tapped artifacts you control, counting every
// artifact (creature or not) tapped as the ability resolves, so attackers
// without vigilance and tapped Treasures and Clues both count. It is
// worked out once and used for both the damage and the scry; the scry
// waits for any prompt the damage asks (CR 608.2c). With its target gone
// the ability does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5e8e116-10fe-48b0-b3d8-6edb39bd5f90",
		Name:         "Alibou, Ancient Witness",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID != source.InstanceID && target.Controller == source.Controller &&
				target.IsCreature() && target.IsArtifact()
		}, "haste")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if !attackDeclaredByYou(ev, source.Controller) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsCreature() && c.IsArtifact()
				}, "Alibou, Ancient Witness — X damage to any target and scry X, X = tapped artifacts you control",
					Do(
						Instruction(func(ctx *Context) error {
							legal := ctx.LegalTargets()
							if len(legal) == 0 {
								return nil
							}
							x := countControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.IsArtifact() && c.Tapped })
							ctx.RecordStep(alibouX, nil, x)
							if x <= 0 {
								return nil
							}
							return DealDamage{Source: ctx.Item.SourceCardID, Target: legal[0].ID, Amount: x}.Apply(ctx)
						}),
						Instruction(func(ctx *Context) error {
							if x := ctx.StepRecord(alibouX).N(); x > 0 {
								return Scry{Player: ctx.Controller(), N: x}.Apply(ctx)
							}
							return nil
						}),
					))),
				TargetAny()),
		},
	})
}

const alibouX = "alibou-ancient-witness/x"
