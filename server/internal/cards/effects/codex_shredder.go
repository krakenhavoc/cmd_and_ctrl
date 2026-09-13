package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Codex Shredder — Artifact {1} (EDHREC rank 2400):
//
//	"{T}: Target player mills a card. (They put the top card of their
//	 library into their graveyard.)
//	 {5}, {T}, Sacrifice this artifact: Return target card from your
//	 graveyard to your hand."
//
// The one-mana mill rock that turns into a Regrowth. Two CR 602
// activated abilities sharing the tap: the mill targets any player,
// the caster included; the return is a graveyard target of any card
// type, the Shredder itself excluded by timing — it is sacrificed at
// announce, after the target was picked (CR 601.2c before 601.2h),
// so it cannot return itself, exactly as in paper.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef7b11ab-24e7-4e7c-91a7-920bade6e60b",
		Name:         "Codex Shredder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Target player mills a card.",
				Cost:    TapCost(),
				Targets: TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return MillCards{Player: t.ID, N: 1}.Apply(ctx)
						}
					}
					return nil
				},
			},
			{
				Label:   "{5}, {T}, Sacrifice this artifact: Return target card from your graveyard to your hand.",
				Cost:    Plus(ManaCost("{5}"), TapCost(), SacrificeThis()),
				Targets: TargetCardInGraveyard("target card from your graveyard", YouOwn()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
						}
					}
					return nil
				},
			},
		},
	})
}
