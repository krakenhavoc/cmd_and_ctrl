package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tough Cookie — Artifact Creature — Food Golem {1}{G}, 2/2 (Wilds of
// Eldraine, #4471):
//
//	"When this creature enters, create a Food token.
//	 {2}{G}: Until end of turn, target noncreature artifact you control
//	 becomes a 4/4 artifact creature.
//	 {2}, {T}, Sacrifice this creature: You gain 3 life."
//
// The Food is the catalog's token. The animation is Cyberdrive Awakener's
// (a layer-4 type addition and a layer-7b base 4/4, CR 613.1d, 613.4b) on
// Doc Ock's single target, a noncreature artifact you control as the
// ability resolves (CR 608.2b), until end of turn: the Food token this
// made can become a 4/4 that still has its Food ability. It keeps every
// type it had. The last row is Ice Cream Kitty's Food ability; as a
// creature's {T} ability it waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a1a5862-c027-4c15-99a7-1795d0da96da",
		Name:         "Tough Cookie",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Tough Cookie — create a Food token", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{G}: Until end of turn, target noncreature artifact you control becomes a 4/4 artifact creature.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    ManaCost("{2}{G}"),
				Targets: TargetPermanent("target noncreature artifact you control", Artifact(), Noncreature(), YouControl()),
				Effect:  toughCookieAnimate,
			},
			{
				Label:   "{2}, {T}, Sacrifice this creature: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Effect:  Do(GainLife{Amount: 3}),
			},
		},
	})
}

// toughCookieAnimate makes the target a 4/4 artifact creature until end
// of turn.
func toughCookieAnimate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return ScopedEffectFor{
		Target: id,
		Mods: []game.Mod{
			game.AddTypesMod("Artifact", "Creature"),
			game.SetBasePowerMod(4),
			game.SetBaseToughnessMod(4),
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Tough Cookie — a 4/4 artifact creature until end of turn",
	}.Apply(ctx)
}
