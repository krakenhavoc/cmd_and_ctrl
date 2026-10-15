package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elena, Turk Recruit — Legendary Creature — Human Assassin {2}{W},
// 1/4 (#3656):
//
//	"When Elena enters, return target non-Assassin historic card from
//	 your graveyard to your hand. (Artifacts, legendaries, and Sagas
//	 are historic.)
//	 Whenever you cast a historic spell, put a +1/+1 counter on Elena."
//
// Archaeomancer's targeted ETB with the historic predicate (CR 700.6)
// minus the Assassin subtype, and Jhoira's historic cast trigger with a
// counter in place of the draw. Elena's own cast does not count: she is
// on the stack, not the battlefield, when it fires.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5b6835b-7ebe-4855-8b3e-bffb74fd50ee",
		Name:         "Elena, Turk Recruit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: b06SelfETB,
				Targets: TargetCardInGraveyard("target non-Assassin historic card in your graveyard",
					And(b14Historic(), Not(OfSubtype("Assassin"))), YouOwn()),
				Key:    "Elena, Turk Recruit — return a non-Assassin historic card to hand",
				Effect: returnTargetCardToHand,
			},
			On(game.EventCast, youCastAHistoricSpell,
				"Elena, Turk Recruit — put a +1/+1 counter on Elena",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
