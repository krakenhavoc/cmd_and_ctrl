package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Summon: Yojimbo — Enchantment Creature — Saga Samurai {3}{W}, 3/4:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after IV.)
//	 I — Exile target artifact, enchantment, or tapped creature an
//	 opponent controls.
//	 II, III — Until your next turn, creatures can't attack you unless
//	 their controller pays {2} for each of those creatures.
//	 IV — Create X Treasure tokens, where X is the number of opponents
//	 who control a creature with power 4 or greater.
//	 Vigilance"
//
// Chapters II and III register a duration attack tax (#3398,
// game/attack_tax_duration.go): a ModAttackTax record naming the
// controller, timed by "until your next turn", read by the one attack
// pricer beside the battlefield statics. It defends the player only, as
// printed, not their planeswalkers.
//
// Chapter IV counts opponents, not creatures: two big creatures under
// one opponent are one Treasure.
//
// No simplification.
func init() {
	const name = "Summon: Yojimbo"
	const taxClause = "Until your next turn, creatures can't attack you unless their controller pays {2} for each of those creatures."
	const taxText = "until your next turn, creatures can't attack you unless their controller pays {2} for each of those creatures"
	tax := func(g *game.Game, item *game.StackItem) error {
		g.AttackTaxForEffect(item.SourceCardID, item.Controller, "{2}",
			g.UntilYourNextTurnDuration(item.Controller), name+" — "+taxClause)
		return nil
	}
	Register(Spec{
		OracleID:        "ff70d743-280e-476f-a3b8-92026419419d",
		Name:            name,
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, SagaChapterLabel(name, 1, "exile target artifact, enchantment, or tapped creature an opponent controls"),
				TargetPermanent("target artifact, enchantment, or tapped creature an opponent controls",
					OpponentControls(),
					Or(Artifact(), Enchantment(), And(Creature(), yojimboTapped()))),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					return ExileTarget{Target: t.ID}.Apply(ctx)
				}),
			ChapterTrigger(2, SagaChapterLabel(name, 2, taxText), tax),
			ChapterTrigger(3, SagaChapterLabel(name, 3, taxText), tax),
			ChapterTrigger(4, SagaChapterLabel(name, 4, "create a Treasure for each opponent who controls a creature with power 4 or greater"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					big := map[uuid.UUID]bool{}
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() && c.CurrentPower() >= 4 {
							big[c.Controller] = true
						}
					}
					n := 0
					for _, opp := range ctx.Opponents() {
						if big[opp] {
							n++
						}
					}
					if n == 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: n}.Apply(ctx)
				}),
		},
	})
}

// yojimboTapped is "tapped": the permanent is tapped now.
func yojimboTapped() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Tapped }
}
