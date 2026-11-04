package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magma Opus — Instant {6}{U}{R} (Secrets of Strixhaven Commander,
// #4834):
//
//	"Magma Opus deals 4 damage divided as you choose among any number of
//	 targets. Tap two target permanents. Create a 4/4 blue and red
//	 Elemental creature token. Draw two cards.
//	 {U/R}{U/R}, Discard this card: Create a Treasure token."
//
// Two target clauses, chosen together as the spell is cast (the ruling):
// up to four targets that divide the 4 damage, each given at least 1 (CR
// 601.2d), and two different permanents to tap, tapped or not. The same
// permanent may be chosen once for each instance of "target" (CR 115.3).
// Each clause is read for its own picks at resolution, so a permanent
// that is only being tapped takes no damage. The sentences run in
// printed order (ResolveSteps), each waiting for any prompt the one
// before it asked.
//
// The discard ability is activated from the hand, its cost discarding
// the card (Boseiju's channel shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "985f9dd7-e693-48ff-9dd6-a1a3cbaa236b",
		Name:         "Magma Opus",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetAny().WithCount(0, 4).Dividing(Divide(4)),
			TargetPermanent("two target permanents").WithCount(2, 2),
		),
		Purpose: game.Purpose{Draws: 2, Tokens: 1},
		OnResolve: ResolveSteps(
			Instruction(magmaOpusDamage),
			Instruction(func(ctx *Context) error {
				for _, t := range ctx.ClauseTargets(1) {
					if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
						continue
					}
					if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			CreateToken{Template: TokenCard("4/4 blue and red Elemental"), N: 1},
			DrawCards{N: 2},
		),
		Activated: []ActivatedAbility{{
			Label:   "{U/R}{U/R}, Discard this card: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{U/R}{U/R}", DiscardSelf: true},
			Zones:   []game.ZoneKind{game.ZoneHand},
			Effect:  Do(CreateToken{Template: TreasureToken(), N: 1}),
		}},
	})
}

// magmaOpusDamage deals each still-legal pick of the first clause its
// announced share of the 4 (CR 608.2b: a share is not redistributed).
func magmaOpusDamage(ctx *Context) error {
	shares := ctx.Item.Distribution
	return ctx.Game.DamageInstanceForEffect(func() error {
		for _, t := range ctx.ClauseTargets(0) {
			if !ctx.IsTargetLegal(t) || shares[t.ID] <= 0 {
				continue
			}
			if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: shares[t.ID]}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}
