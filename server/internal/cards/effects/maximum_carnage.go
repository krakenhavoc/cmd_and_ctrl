package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Maximum Carnage — {4}{R} Enchantment — Saga (Marvel's Spider-Man, #4460):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Until your next turn, each creature attacks each combat if able and
//	 attacks a player other than you if able.
//	 II — Add {R}{R}{R}.
//	 III — This Saga deals 5 damage to each opponent."
//
// I is Kardur, Doomscourge's requirement over every creature rather than the
// opponents' only: one live rule record (game.ScopeCreatures, CR 611.2c: a
// requirement modifies the rules, so it reaches a creature that enters
// later) carrying the two CR 508.1d requirements, "attacks each combat" and
// "attacks a player other than you", lasting until the chapter's controller's
// next turn. For the controller's own creatures the second requirement is
// always met, since they can only attack the opponents. II is plain mana
// (it empties with the main phase it is added in). III is one damage event
// from the Saga to each opponent.
//
// No simplification.
func init() {
	const name = "Maximum Carnage"
	Register(Spec{
		OracleID:     "7351fca5-8672-4ba3-a015-bd4b6b289efc",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, SagaChapterLabel(name, 1, "until your next turn, each creature attacks each combat if able and attacks a player other than you if able"),
				maximumCarnageRequirement),
			ChapterTrigger(2, SagaChapterLabel(name, 2, "add {R}{R}{R}"), Do(AddMana{Produced: "{R}{R}{R}"})),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "this Saga deals 5 damage to each opponent"), Do(DamageEachOpponent{N: 5})),
		},
	})
}

// maximumCarnageRequirement is chapter I: every creature must attack, and
// must attack a player other than the Saga's controller if able.
func maximumCarnageRequirement(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := ctx.Controller()
	if you == uuid.Nil {
		return nil
	}
	g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeCreatures, you,
		[]game.Mod{game.AddAttackRequirementMod(uuid.Nil), game.AddAttackRequirementMod(you)},
		DurationUntilYourNextTurn(ctx, you),
		"Maximum Carnage — each creature attacks each combat if able and attacks a player other than you if able")
	return nil
}
