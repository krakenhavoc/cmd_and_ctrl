package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bow of Nylea — Legendary Enchantment Artifact {1}{G}{G} (Secret Lair Drop):
//
//	"Attacking creatures you control have deathtouch.
//	 {1}{G}, {T}: Choose one —
//	 • Put a +1/+1 counter on target creature.
//	 • Bow of Nylea deals 2 damage to target creature with flying.
//	 • You gain 3 life.
//	 • Put up to four target cards from your graveyard on the bottom of your
//	   library in any order."
//
// The fourth mode's cards are placed together, you ordering the pile, each in
// its owner's library (they are all yours); a target that has left the
// graveyard is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "361b965f-2ce7-49f8-84e7-7325ea0c948d",
		Name:         "Bow of Nylea",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.AttackingCreaturesHave("Attacking creatures you control have deathtouch.", nil, "deathtouch"),
		},
		Activated: []effects.ActivatedAbility{{
			Label: "{1}{G}, {T}: Choose one — put a +1/+1 counter on target creature; or 2 damage to target creature with flying; or you gain 3 life; or put up to four target cards from your graveyard on the bottom of your library in any order.",
			Cost:  effects.Plus(effects.ManaCost("{1}{G}"), effects.TapCost()),
			Modes: effects.ChooseOne(
				effects.ModeDoing("Put a +1/+1 counter on target creature.", effects.TargetCreature("target creature"),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						for _, t := range ctx.ModeTargets(occ) {
							return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
						}
						return nil
					}),
				effects.ModeDoing("Bow of Nylea deals 2 damage to target creature with flying.",
					effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						for _, t := range ctx.ModeTargets(occ) {
							return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 2}.Apply(ctx)
						}
						return nil
					}),
				effects.ModeDoing("You gain 3 life.", nil,
					func(_ *game.StackItem, ctx *effects.Context, _ int) error {
						return effects.GainLife{Amount: 3}.Apply(ctx)
					}),
				effects.ModeDoing("Put up to four target cards from your graveyard on the bottom of your library in any order.",
					effects.TargetCardInGraveyard("up to four target cards from your graveyard", effects.YouOwn()).WithCount(0, 4),
					func(_ *game.StackItem, ctx *effects.Context, occ int) error {
						var ids []uuid.UUID
						for _, t := range ctx.ModeTargets(occ) {
							ids = append(ids, t.ID)
						}
						if len(ids) == 0 {
							return nil
						}
						return effects.PutInLibraryInAnyOrder{
							Cards: ids, From: game.ZoneGraveyard, Placement: game.LibraryPlaceBottom,
							Label: "Bow of Nylea — put the cards on the bottom of your library in any order",
						}.Apply(ctx)
					}),
			),
		}},
	})
}
