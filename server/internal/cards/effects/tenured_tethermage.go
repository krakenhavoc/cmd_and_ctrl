package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tenured Tethermage — Creature — Human Artificer {1}{R}{G}, 1/1:
//
//	"When this creature enters, you may sacrifice a land. If you do,
//	 create two tapped Heartwood tokens. (They're red and green
//	 artifacts with "{T}: Add {R} or {G}.")
//	 Tap two untapped artifacts you control: Put two +1/+1 counters on
//	 this creature."
//
// The sacrifice is the trigger's own "you may" (CR 603.5) followed by a
// land pick; the tokens come only once a land was really sacrificed.
// The tap cost is not the {T} symbol (CR 302.6), so artifacts that
// arrived this turn, the freshly made Heartwoods among them once they
// untap, pay it freely; the Tethermage itself is no artifact and cannot.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "571cb06d-55b5-46e4-9678-1f74aca73068",
		Name:         "Tenured Tethermage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(WhenThisEnters("Tenured Tethermage — sacrifice a land to create two tapped Heartwood tokens", tenuredTethermageETB),
				"Tenured Tethermage — sacrifice a land to create two tapped Heartwood tokens?"),
		},
		Activated: []ActivatedAbility{{
			Label:   "Tap two untapped artifacts you control: Put two +1/+1 counters on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent("two untapped artifacts you control", Artifact()),
				Label:  "two untapped artifacts you control",
			}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			},
		}},
	})
}

func tenuredTethermageETB(g *game.Game, item *game.StackItem) error {
	// Keyed (#4290): the controller rides the run as its carry, so a
	// table waiting on the sacrifice is a restore point.
	return g.PlayerSacrificesThenKeyedForEffect(
		item.SourceCardID, item.Controller,
		sacrificeSpec("a land", Land()),
		"Tenured Tethermage — sacrifice a land",
		1,
		tenuredTethermageTokens, []uuid.UUID{item.Controller})
}

// tenuredTethermageTokens is "if you do, create two tapped Heartwood
// tokens", once the land has really been sacrificed.
var tenuredTethermageTokens = game.RegisterPromptRunThen("prompt-run/tenured-tethermage-tokens", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 || !r.Sacrifices().Sacrificed(r.Carry[0]) {
		return nil
	}
	return CreateTokenAdvanced{Spec: Token(HeartwoodToken()).EntersTapped(), N: 2}.Apply(keyedContext(g, r.Carry[0], r.Source))
})
