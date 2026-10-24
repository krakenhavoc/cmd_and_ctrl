package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Storm — Instant {1}{R}{R}:
//
//	"Lightning Storm deals X damage to any target, where X is 3 plus
//	 the number of charge counters on Lightning Storm.
//	 Discard a land card: Put two charge counters on Lightning Storm.
//	 You may choose a new target for it. Any player may activate this
//	 ability but only if Lightning Storm is on the stack."
//
// The card #1865 (ADR 0106 §1 amendment 2026-10-16) opened the stack to
// activated abilities for. Three engine pieces, none of them new to this
// file:
//
//   - The ability functions from the stack (CR 113.6): `Zones` names
//     game.ZoneStack, the way a cycling row names the hand. The
//     activation path finds the spell on the stack and asks
//     game.MayActivate, which answers for a spell as for a permanent: its
//     controller, or anyone for this AnyPlayer row (CR 602.2, 602.1b).
//   - "You" is the activator (CR 109.5): they discard the land
//     (CR 602.1a) and they choose the new target, not the spell's
//     controller. The new target must still be legal for Lightning
//     Storm, judged as its controller's spell (game.OfferRetargetForEffect).
//   - The counters go on the spell only while it is still the object the
//     ability came from (CR 400.7, game.AbilitySourceSpellForEffect):
//     countered or resolved in response, the ability does nothing.
//
// "Choose a new target" is CR 115.7d's "choose new targets", for a spell
// with one target: the activator may leave it where it is.
//
// The spell reads its counters as it resolves. They are on the spell
// object (CR 122.1), so a copy of Lightning Storm starts with none
// (CR 707.2) and counts only what is put on it.
func init() {
	Register(Spec{
		OracleID:     "2c0fceb6-d92d-4f4a-85a2-0af5be5bba81",
		Name:         "Lightning Storm",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			n := 3
			if c, ok := ctx.Game.LookupCardForEffect(item.ID); ok {
				n += c.Counters[game.CounterCharge]
			}
			return DealDamage{
				Source: ctx.Source(),
				Target: item.Targets[0].ID,
				Amount: n,
			}.Apply(ctx)
		},
		Activated: []ActivatedAbility{{
			Label: "Discard a land card: Put two charge counters on Lightning Storm. You may choose a new target for it. Any player may activate this ability but only if Lightning Storm is on the stack.",
			Cost: DiscardCardsMatching(1, "a land card", func(c game.Card) bool {
				return c.IsLand()
			}),
			Zones:     []game.ZoneKind{game.ZoneStack},
			AnyPlayer: true,
			// It moves the damage off whoever the storm is aimed at, which
			// is ADR 0142's "prevents or redirects damage". Threat-gated,
			// so smart autopass stops a player the storm targets and no
			// one else. The bot never reaches across for it: it declares
			// no Draws or ControllerLosesLife (ADR 0106 §1 decision 8).
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			// The new-target offer waits for a prompt the counters ask (a CR 616 order; CR 608.2c,
			// #4155).
			Effect: Do(
				Instruction(func(ctx *Context) error {
					spell, ok := ctx.Game.AbilitySourceSpellForEffect(ctx.Item)
					if !ok {
						return nil
					}
					return ctx.Game.AddCounterByForEffect(ctx.Controller(), spell, game.CounterCharge, 2)
				}),
				Instruction(func(ctx *Context) error {
					spell, ok := ctx.Game.AbilitySourceSpellForEffect(ctx.Item)
					if !ok {
						return nil
					}
					return ChangeTargets{
						StackID:  spell,
						Chooser:  ctx.Controller(),
						Policy:   game.RetargetChooseNew,
						Optional: true,
						Reason:   "Lightning Storm — you may choose a new target",
					}.Apply(ctx)
				}),
			),
		}},
	})
}
