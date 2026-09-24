package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Departed Deckhand — Creature — Spirit Pirate {1}{U}, 2/2:
//
//	"When this creature becomes the target of a spell, sacrifice it.
//	 This creature can't be blocked except by Spirits.
//	 {3}{U}: Another target creature you control can't be blocked this
//	 turn except by Spirits."
//
// # The drawback
//
// "Becomes the target of A SPELL" is narrower than the "spell or
// ability" every other card in the family prints, and the difference
// is real: an equip, a Rhystic Study trigger or a fight ability that
// names this creature does NOT kill it. EventBecomesTarget is emitted
// for both and carries no discriminator, so the spell half is read off
// the stack (SelfTargetedByASpell) — a spell's stack item shares its ID
// with the source card, an ability's does not.
//
// The trigger fires at ANNOUNCE (CR 601.2c), so it goes on the stack
// above the spell that targeted and resolves first. The Deckhand is
// therefore already in its owner's graveyard when that spell tries to
// resolve, and the spell is countered on resolution for having no
// legal target (CR 608.2b) — which is the whole point of the card and
// falls out for free.
//
// # The evasion
//
// Both evasion clauses are "can't be blocked EXCEPT BY Spirits", a
// conditional block restriction (#750, ADR 0045 addendum Decision 11):
// a rule about the pair, parameterised by a predicate on the BLOCKER.
// The Deckhand's own is a Spec.BlockRules entry on itself. The {3}{U}
// grants the same rule to ANOTHER creature until end of turn, which a
// permanent's static cannot say, so it registers a turn-scoped rule
// (BlockRuleUntilEOT) pinned to the target at resolution (CR 611.2c):
// it survives the Deckhand leaving (CR 611.2b), and a target that
// leaves and returns is a new object and loses it.
//
// "Spirits" is an EFFECTIVE creature type, so a changeling can block.
// "Another" excludes the Deckhand by name, as every other "another
// target creature you control" on an activated ability does (#350).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a620a765-97ba-4687-acfd-4dec7da75d9f",
		Name:         "Departed Deckhand",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, SelfTargetedByASpell,
				"Departed Deckhand — sacrifice it",
				SacrificeThisIfStillOnBattlefield),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedExceptBy(OnSelf(), OfCreatureType("Spirit"), "Spirits"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}: Another target creature you control can't be blocked this turn except by Spirits.",
			Cost:    ManaCost("{3}{U}"),
			Targets: TargetCreature("another target creature you control", YouControl(), b03NotNamed("Departed Deckhand")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				// CR 608.2b: a target that left in response is
				// skipped, not an error.
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return BlockRuleUntilEOT{
						Target: t.ID,
						Rule: func(scope BlockScope) game.BlockRule {
							return CantBeBlockedExceptBy(scope, OfCreatureType("Spirit"), "Spirits")
						},
						Label: "Departed Deckhand — can't be blocked except by Spirits",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
