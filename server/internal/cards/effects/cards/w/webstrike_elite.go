package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Webstrike Elite — Creature — Insect Archer {G}{G}, 3/3 (Aetherdrift,
// #4464):
//
//	"Reach
//	 Cycling {X}{G}{G} ({X}{G}{G}, Discard this card: Draw a card.)
//	 When you cycle this card, destroy up to one target artifact or
//	 enchantment with mana value X."
//
// Shark Typhoon's shape: the cycling ability draws the card, and a
// separate trigger fired from the graveyard (CR 702.29c) destroys the
// permanent, so countering either leaves the other. EventCycle carries the
// X the cycling cost was announced with, and the clause's "mana value X" is
// that Amount, fixed when the trigger fires and re-checked as it resolves
// (game.TargetSpec.ManaValueEqualsEventAmount). The {X} in a permanent's
// own mana cost is 0 there (CR 202.3e), so X=0 reaches an {X} artifact.
//
// No simplification.
func init() {
	cycling := effects.Cycling("{G}{G}")
	cycling.Label = "Cycling {X}{G}{G} ({X}{G}{G}, Discard this card: Draw a card.)"
	cycling.Cost = effects.Plus(effects.ManaCost("{X}{G}{G}"), effects.DiscardThis())
	effects.Register(effects.Spec{
		OracleID:        "48a1d55e-0b71-4724-ab2a-6d2c87e86b5b",
		Name:            "Webstrike Elite",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []effects.ActivatedAbility{cycling},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InGraveyard(effects.On(game.EventCycle, effects.Self, "Webstrike Elite — cycled: destroy up to one target artifact or enchantment with mana value X",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
					})),
				effects.ManaValueEqualsTheEventAmount(
					effects.TargetPermanent("up to one target artifact or enchantment with mana value X",
						effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 1))),
		},
	})
}
