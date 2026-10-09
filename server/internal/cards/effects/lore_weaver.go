package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lore Weaver — Creature — Human Wizard {3}{U}, 2/2:
//
//	"Partner with Ley Weaver (When this creature enters, target player
//	 may put Ley Weaver into their hand from their library, then
//	 shuffle.)
//	 {5}{U}{U}: Target player draws two cards."
//
// Not legendary, so it can't be a commander: of CR 702.124j's two
// abilities only the entry search does anything (PartnerWith, #2142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "040f6f11-7fbe-4635-b935-442e41f1e704",
		Name:         "Lore Weaver",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Lore Weaver", "Ley Weaver"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}{U}{U}: Target player draws two cards.",
			Cost:    ManaCost("{5}{U}{U}"),
			Targets: TargetPlayer("target player"),
			Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: 2}),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (DrawCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
