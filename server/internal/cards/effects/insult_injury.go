package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insult // Injury — split card, oracle 47543892-4d60-4c6b-a6a4-69b9172af01e:
//
//	Insult — Sorcery {2}{R}: "Damage can't be prevented this turn. If a
//	        source you control would deal damage this turn, it deals
//	        double that damage instead."
//	Injury — Sorcery {2}{R}, aftermath: "Injury deals 2 damage to target
//	        creature and 2 damage to target player or planeswalker."
//
// Both halves register as ADR 0034 faces (ADR 0103): Insult under the
// bare oracle ID, Injury under "#1". Aftermath is the engine's — Injury
// is cast only from the graveyard and exiled as it leaves the stack
// (CR 702.127a), found from the face's own text.
//
// INSULT. The turn grant (ADR 0107 §5) and the multiplier (ADR 0108 §3)
// are two records until cleanup. "A source you control" is read as the
// damage would be dealt (CR 611.2c), through the source's last-known
// information, so a creature you gain control of later in the turn is
// doubled too. The multiplier is a replacement, not a prevention effect,
// so its own "can't be prevented" leaves it alone. Two Insults are ×4
// (the ruling), with no order to ask.
//
// INJURY. Two target clauses, one damage instruction: the two hits are
// one instance (DamageInstanceForEffect), and each clause is rechecked
// on its own (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "47543892-4d60-4c6b-a6a4-69b9172af01e",
		Name:         "Insult",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (DamageCantBePreventedThisTurn{}).Apply(ctx); err != nil {
				return err
			}
			return MultiplyDamage{Factor: 2, Sources: game.DamageSourcesYours,
				Label: "Insult — your sources deal double damage"}.Apply(ctx)
		},
	})
	Register(Spec{
		OracleID:     "47543892-4d60-4c6b-a6a4-69b9172af01e#1",
		Name:         "Injury",
		Completeness: CompletenessFull,
		Targets:      Clauses(TargetCreature("target creature"), targetPlayerOrPlaneswalker()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.DamageInstanceForEffect(func() error {
				for slot := 0; slot < 2; slot++ {
					t, ok := ctx.ClauseTarget(slot)
					if !ok {
						continue
					}
					if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
