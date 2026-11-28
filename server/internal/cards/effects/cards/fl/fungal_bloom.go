package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fungal Bloom — Enchantment {G}{G} (Fallen Empires, S82 OLD b853, #5637):
//
//	"{G}{G}: Put a spore counter on target Fungus."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa76ee16-ff48-418b-819a-69f1139b36f3",
		Name:         "Fungal Bloom",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{G}: Put a spore counter on target Fungus.",
			Cost:    effects.ManaCost("{G}{G}"),
			Targets: effects.TargetPermanent("target Fungus", effects.HasSubtype("Fungus")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (effects.AddCounter{Target: t.ID, Kind: effects.S82OldB853SporeCounter, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
