package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Titan of Industry — Creature — Elemental {4}{G}{G}{G}, 7/7:
//
//	"Reach, trample
//	 When this creature enters, choose two —
//	 • Destroy target artifact or enchantment.
//	 • Target player gains 5 life.
//	 • Create a 4/4 green Rhino Warrior creature token.
//	 • Put a shield counter on a creature you control."
//
// An entering trigger that chooses two of four, the Craftwork Crusher
// shape (a Modes clause on a WhenThisEnters ability). Two of the four
// bullets target, so the engine's per-bullet target slots carry them:
// the artifact-or-enchantment pick and the player pick each read their
// own group with ModeTarget (#764), and a trigger that chose the
// Rhino and the counter asks for no target at all.
//
// The shield-counter bullet does not target — "a creature you control"
// is a choice made as the bullet resolves (CR 608.2d), so hexproof and
// shroud are irrelevant and a creature that only just arrived qualifies.
// It is the LAST bullet, so printed order (CR 700.2c) puts it after the
// Rhino: choosing the Rhino and the counter lets the new token be the
// one shielded, and the prompt simply has nothing else to wait on. A
// player with no creature is asked nothing. The counter itself is the
// engine's (#2000): the first damage or destroy effect that would hit
// the creature removes the counter instead.
//
// The token is the table's "4/4 green Rhino Warrior" (Workshop Warchief
// makes the same one).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Titan of Industry — choose two", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseN("Choose two", 2, 2,
		ModeDoing("Destroy target artifact or enchantment.",
			TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			titanOfIndustryDestroy),
		ModeWithPurpose(ModeDoing("Target player gains 5 life.",
			TargetPlayer("target player"),
			titanOfIndustryGainLife), ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 5})),
		ModeWithPurpose(ModeDoing("Create a 4/4 green Rhino Warrior creature token.",
			nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CreateToken{Controller: ctx.Controller(), Template: TokenCard("4/4 green Rhino Warrior"), N: 1}.Apply(ctx)
			}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		ModeDoing("Put a shield counter on a creature you control.",
			nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return ChoosePermanents{
					Question: "Titan of Industry — put a shield counter on a creature you control",
					Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
						var ids []uuid.UUID
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == of && c.IsCreature() {
								ids = append(ids, c.InstanceID)
							}
						}
						return ids, 1, 1
					},
					Then: titanOfIndustryShield,
				}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "5d8acafe-c13e-43ca-a89e-0cfd12d95662",
		Name:            "Titan of Industry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}

// titanOfIndustryDestroy is the first bullet: destroy the target artifact
// or enchantment, if it is still a legal target (CR 608.2b).
func titanOfIndustryDestroy(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return DestroyTarget{Target: t.ID}.Apply(ctx)
}

// titanOfIndustryGainLife is the second bullet: the target player, who
// may be anyone, gains 5.
func titanOfIndustryGainLife(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	return GainLife{Player: t.ID, Amount: 5}.Apply(ctx)
}

// titanOfIndustryShield is the last bullet's answer: one shield counter
// on the creature picked. A package-level function so the continuation
// captures nothing.
func titanOfIndustryShield(ctx *Context, picked game.PromptedPicks) error {
	cards := picked.Cards()
	if len(cards) != 1 {
		return nil
	}
	return AddCounter{Target: cards[0], Kind: game.CounterShield, N: 1}.Apply(ctx)
}
