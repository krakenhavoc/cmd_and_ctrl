package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrap Welder — Creature — Goblin Artificer {2}{R}, 3/3 (Kamigawa: Neon
// Dynasty):
//
//	"{T}, Sacrifice an artifact with mana value X: Return target artifact
//	 card with mana value less than X from your graveyard to the
//	 battlefield. It gains haste until end of turn."
//
// X is announced with the activation (CR 602.2b), as Sidisi's is: the
// sacrifice cost names an artifact with mana value X
// (neoSacrificeAnArtifactWithManaValueX, game.SacrificeManaValueX) and the
// target is an artifact card in your graveyard with mana value less than X
// (TargetSpec.WithManaValueLessThanX), judged at announce and again as the
// ability resolves (CR 608.2b). {X} in a cost counts as 0 and a token has mana
// value 0 (CR 202.3). The returned card gains haste until end of turn only
// if it came back; a card that left the graveyard in response gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ebce651-ef95-4a96-b720-a23757d641c8",
		Name:         "Scrap Welder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Sacrifice an artifact with mana value X: Return target artifact card with mana value less than X from your graveyard to the battlefield. It gains haste until end of turn.",
			Cost:  effects.Plus(effects.TapCost(), effects.NeoSacrificeAnArtifactWithManaValueX("an artifact you control with mana value X")),
			Targets: effects.TargetCardInGraveyard("target artifact card with mana value less than X from your graveyard",
				effects.YouOwn(), effects.Artifact()).WithManaValueLessThanX(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
						return err
					}
					if !effects.OnBattlefield(g, t.ID) {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"haste"},
						Label:    "Scrap Welder — it gains haste until end of turn",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
