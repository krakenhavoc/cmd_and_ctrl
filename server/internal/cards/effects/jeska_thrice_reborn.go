package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeska, Thrice Reborn — Legendary Planeswalker — Jeska {2}{R},
// printed loyalty 0:
//
//	"Jeska enters with a loyalty counter on her for each time you've
//	 cast a commander from the command zone this game.
//	 0: Choose target creature. Until your next turn, if that creature
//	    would deal combat damage to one of your opponents, it deals
//	    triple that damage to that player instead.
//	 −X: Jeska deals X damage to each of up to three targets.
//	 Jeska, Thrice Reborn can be your commander.
//	 Partner"
//
// The entry count is a CR 614.1c replacement reading the controller's
// commander-cast tally (Player.CommanderCasts, which only a cast from the
// command zone bumps), summed over every commander. Cast before any
// commander, she enters with none and CR 704.5i takes her. The 0 is
// ADR 0108 §3's damage multiplier pinned to the target (CR 400.7),
// combat damage to an opponent only, until her controller's next turn.
// The −X is a loyalty cost of X (LoyaltyMinusX, #1944), X damage to each
// of up to three targets at once (CR 120.2). Partner is deck
// construction, not battlefield behaviour.
func init() {
	upToThree := TargetAny()
	upToThree.Label = "up to three targets"
	upToThree.Min, upToThree.Max = 0, 3
	Register(Spec{
		OracleID:     "b1fbfcf3-6921-4417-a58e-0f5e5d34a105",
		Name:         "Jeska, Thrice Reborn",
		Completeness: CompletenessFull,
		XMatters:     true,
		Replacements: []game.ReplacementEffect{
			b19EntersWithCountersCounted(game.CounterLoyalty, func(g *game.Game, src *game.Card) int {
				p := g.PlayerByIDForEffect(src.Controller)
				if p == nil {
					return 0
				}
				n := 0
				for _, casts := range p.CommanderCasts {
					n += casts
				}
				return n
			}, "Jeska, Thrice Reborn: enters with a loyalty counter for each time you've cast a commander from the command zone"),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "0: Choose target creature. Until your next turn, if that creature would deal combat damage to one of your opponents, it deals triple that damage to that player instead.",
				Cost:    LoyaltyCost(0),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return MultiplyDamage{
						Factor: 3, From: t.ID, Recipients: game.DamageRecipientsOpponents,
						CombatOnly: true, UntilYourNextTurn: true,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−X: Jeska deals X damage to each of up to three targets.",
				Cost:    LoyaltyMinusX(),
				Targets: upToThree,
				Purpose: ForTargets(DamageXToTarget(0)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, t := range ctx.LegalTargets() {
						ids = append(ids, t.ID)
					}
					return DealDamageToEachThen(ctx, ids, ctx.X(), nil)
				},
			},
		},
	})
}
