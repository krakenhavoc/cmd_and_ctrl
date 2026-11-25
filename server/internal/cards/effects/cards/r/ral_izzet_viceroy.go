package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ral, Izzet Viceroy — Legendary Planeswalker — Ral {3}{U}{R}, loyalty 5
// (Mythic Edition):
//
//	"+1: Look at the top two cards of your library. Put one of them into your
//	 hand and the other into your graveyard.
//	 −3: Ral deals damage to target creature equal to the total number of
//	 instant and sorcery cards you own in exile and in your graveyard.
//	 −8: You get an emblem with "Whenever you cast an instant or sorcery spell,
//	 this emblem deals 4 damage to any target and you draw two cards.""
//
// The −3 counts as it resolves (Ral is the source; his last-known state is used
// if he has left). The emblem's trigger targets as it goes on the stack (CR
// 603.3d); the emblem is the source of the damage, and the two cards are drawn
// even if the target became illegal only when the whole trigger is countered by
// the rules (an illegal sole target removes the trigger, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f29f4c73-5e5c-4341-8582-2873fbe5c5a3",
		Name:            "Ral, Izzet Viceroy",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Ral, Izzet Viceroy emblem",
			Text:  "Whenever you cast an instant or sorcery spell, this emblem deals 4 damage to any target and you draw two cards.",
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()), "Ral, Izzet Viceroy emblem — 4 damage to any target and draw two cards", ralIzzetViceroyEmblem),
					effects.TargetAny()),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Look at the top two cards of your library. Put one of them into your hand and the other into your graveyard.",
				Cost:  effects.LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					player := ctx.Controller()
					return effects.TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
						Max:    1,
						Label:  "Ral, Izzet Viceroy — put one of them into your hand and the other into your graveyard",
						Key:    effects.TakeRestGraveyard,
					}.Apply(ctx)
				},
			},
			{
				Label:   "−3: Ral deals damage to target creature equal to the total number of instant and sorcery cards you own in exile and in your graveyard.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					n := effects.OldB41InstantsAndSorceriesOwnedInExileAndGraveyard(g, ctx.Controller())
					return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: n}.Apply(ctx)
				},
			},
			{
				Label:  "−8: You get an emblem with \"Whenever you cast an instant or sorcery spell, this emblem deals 4 damage to any target and you draw two cards.\"",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// ralIzzetViceroyEmblem is the emblem trigger's resolution: 4 damage to the
// chosen target, then its owner draws two cards. The draw does not depend on
// the damage (they are two clauses joined by "and"), but the trigger is
// removed from the stack unopposed if its only target is gone.
func ralIzzetViceroyEmblem(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}).Apply(ctx); err != nil {
		return err
	}
	return effects.DrawCards{N: 2}.Apply(ctx)
}
