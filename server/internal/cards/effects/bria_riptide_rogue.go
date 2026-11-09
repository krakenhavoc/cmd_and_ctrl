package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bria, Riptide Rogue — Legendary Creature — Otter Rogue {2}{U}{R}, 3/3 (Bloomburrow):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1
//	 until end of turn.)
//	 Other creatures you control have prowess. (If a creature has multiple
//	 instances of prowess, each triggers separately.)
//	 Whenever you cast a noncreature spell, target creature you control can't
//	 be blocked this turn."
//
// Prowess is the engine's keyword trigger (game/prowess.go), derived from the
// effective ability list, so a creature that is granted it by the layer-6
// static triggers like a printed one, and one that has both a printed prowess
// and Bria's grant triggers twice (CR 702.108b). The cast trigger is the Artful
// Dodge restriction on the target, chosen as the trigger is put on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d627996b-cbcd-4162-bd13-9ab286de64d1",
		Name:            "Bria, Riptide Rogue",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Static:          []game.StaticAbility{KeywordGrant(OtherCreaturesYouControl, "prowess")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WheneverYouCast(Noncreature(), "Bria, Riptide Rogue — target creature you control can't be blocked this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return RestrictUntilEOT{
							Target:       id,
							Restrictions: game.CantBeBlocked,
							Label:        "Bria, Riptide Rogue — can't be blocked",
						}.Apply(ctx)
					}),
				TargetCreature("target creature you control", YouControl())),
				ForTargets(CantBeBlockedUntilEOT(0, 0, 0))),
		},
	})
}
