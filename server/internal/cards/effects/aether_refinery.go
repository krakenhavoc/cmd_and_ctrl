package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aether Refinery — Artifact {4}{R}{R}:
//
//	"If you would get one or more {E} (energy counters), you get twice
//	 that many {E} instead.
//	 {T}: You get {E}, then you may pay one or more {E}. If you do,
//	 create an X/X black Aetherborn creature token, where X is the
//	 amount of {E} paid this way."
//
// ADR 0129 §3 and §6 (#1995). Getting energy runs ADR 0056's counter
// window with the player set (RepEventCounter, CounterPlayer), so the
// replacement doubles any energy its controller would get, from any
// source (CR 614.1a). Two Refineries double twice (CR 616.1). The tap
// ability's own {E} is doubled too. "One or more" is the pay_amount
// prompt with a floor of 1; the bot pays all of it, since every
// counter is another +1/+1 on the token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e9cde32-0568-44bc-a00a-9709f15c1350",
		Name:         "Aether Refinery",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCounter && ev.CounterPlayer == src.Controller &&
					ev.CounterName == game.CounterEnergy && ev.CounterDelta > 0
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta *= 2
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Aether Refinery: twice that many {E}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: You get {E}, then you may pay one or more {E}. If you do, create an X/X black Aetherborn creature token, where X is the amount of {E} paid this way.",
			Cost:    TapCost(),
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Energy: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GetEnergy{N: 1}).Apply(ctx); err != nil {
					return err
				}
				return PayEnergyAmount{
					Min:      1,
					Question: "Aether Refinery — pay one or more {E} for an X/X Aetherborn",
					Unit:     game.PayAmountPower,
					Goal:     AsMuchAsYouCan,
					Then: func(ctx *Context, paid int) error {
						if paid <= 0 {
							return nil
						}
						return CreateToken{
							Controller: ctx.Controller(),
							Template: game.Card{
								Name:      "Aetherborn",
								TypeLine:  "Token Creature — Aetherborn",
								Colors:    []string{"B"},
								Power:     paid,
								Toughness: paid,
							},
							N: 1,
						}.Apply(ctx)
					},
				}.Apply(ctx)
			},
		}},
	})
}
