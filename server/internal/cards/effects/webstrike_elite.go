package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	cycling := Cycling("{G}{G}")
	cycling.Label = "Cycling {X}{G}{G} ({X}{G}{G}, Discard this card: Draw a card.)"
	cycling.Cost = Plus(ManaCost("{X}{G}{G}"), DiscardThis())
	Register(Spec{
		OracleID:        "48a1d55e-0b71-4724-ab2a-6d2c87e86b5b",
		Name:            "Webstrike Elite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated:       []ActivatedAbility{cycling},
		Triggered: []game.TriggeredAbility{
			Targeting(
				InGraveyard(On(game.EventCycle, Self, "Webstrike Elite — cycled: destroy up to one target artifact or enchantment with mana value X",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						t, ok := firstLegalTarget(ctx)
						if !ok {
							return nil
						}
						return DestroyTarget{Target: t.ID}.Apply(ctx)
					})),
				ManaValueEqualsTheEventAmount(
					TargetPermanent("up to one target artifact or enchantment with mana value X",
						Or(Artifact(), Enchantment())).WithCount(0, 1))),
		},
	})
}
