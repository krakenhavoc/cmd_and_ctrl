package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Master Pakku — Legendary Creature — Human Advisor Ally {1}{U}, 1/3:
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 Whenever Master Pakku becomes tapped, target player mills X cards,
//	 where X is the number of Lesson cards in your graveyard."
//
// "Becomes tapped" is any tap, an attack included (ThisBecameTapped over
// the tap and attack events, as Hemmer). X is counted as the ability
// resolves (CR 608.2h), from the graveyard of Pakku's controller, not
// the target's.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "b1727c07-c829-4d64-a1c5-8b53c743888a",
		Name:            "Master Pakku",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
					"Master Pakku — target player mills X cards, where X is the number of Lesson cards in your graveyard",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetPlayer {
								continue
							}
							return MillCards{Player: t.ID, N: avatarB5LessonCardsInYourGraveyard(g, item.Controller)}.Apply(ctx)
						}
						return nil
					}),
				TargetPlayer("target player")),
		},
	})
}
