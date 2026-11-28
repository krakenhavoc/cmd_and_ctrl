package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roots of Life — Enchantment {1}{G}{G} (Mirage, S82 OLD batch 809, #5633):
//
//	"As this enchantment enters, choose Island or Swamp.
//	 Whenever a land of the chosen type an opponent controls becomes tapped,
//	 you gain 1 life."
//
// The choice is CR 614.12's, made as it enters (ChooseOptionAsEnters). Each
// word has its own "becomes tapped" trigger gated on it (WhenChosen), so
// exactly one is live once the choice is made and none before. The trigger
// is Spireside Infiltrator's tap-or-attack watcher aimed at an opponent's
// land of that type.
//
// No simplification.
func init() {
	trigger := func(land string) game.TriggeredAbility {
		return effects.WhenChosen(land, effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
			effects.S82OldB809OpponentsLandOfTypeBecameTapped(land),
			"Roots of Life — you gain 1 life ("+land+")",
			effects.Do(effects.GainLife{Amount: 1})))
	}
	effects.Register(effects.Spec{
		OracleID:     "29ec9b49-fbe5-49ac-8ad6-7159cb88bc66",
		Name:         "Roots of Life",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseOptionAsEnters("Roots of Life", "Island", "Swamp"),
		Triggered:    []game.TriggeredAbility{trigger("Island"), trigger("Swamp")},
	})
}
