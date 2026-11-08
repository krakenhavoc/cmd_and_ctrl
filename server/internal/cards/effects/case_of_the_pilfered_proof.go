package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Case of the Pilfered Proof — Enchantment — Case {1}{W} (Murders at
// Karlov Manor, #4469):
//
//	"Whenever a Detective you control enters or is turned face up, put a
//	 +1/+1 counter on it.
//	 To solve — You control three or more Detectives. (If unsolved, solve
//	 at the beginning of your end step.)
//	 Solved — If one or more tokens would be created under your control,
//	 those tokens plus a Clue token are created instead. (It's an artifact
//	 with "{2}, Sacrifice this token: Draw a card.")"
//
// The trigger is one ability with two conditions (OnAny, Gadget
// Technician's shape): a Detective entering under your control, or one
// you control being turned face up (CR 708.8). A disguised Detective
// enters face down with no creature types, so it counts only when it
// turns up. The counter goes on the permanent if it is still there. The
// solve condition counts the Detectives you control at the end step and
// again on resolution (CR 603.4). The solved line is Peregrin Took's
// token-creation replacement (CR 614.1a) with a Clue, existing only
// while the Case is solved (ActiveWhen: Solved()); it fires once per
// creation instruction, and a doubler that has not applied yet doubles
// the Clue too. "Detective" is a word of the Case's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68d3084b-0f4b-44a6-8954-6ebaf1db9269",
		Name:         "Case of the Pilfered Proof",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventTurnedFaceUp}, pilferedProofDetectiveArrived,
				"Case of the Pilfered Proof — put a +1/+1 counter on that Detective",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !onBattlefield(g, item.Trigger.Event.CardID) {
						return nil
					}
					return AddCounter{Target: item.Trigger.Event.CardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
			ToSolve("Case of the Pilfered Proof — to solve: you control three or more Detectives", youControlThreeOrMoreDetectives),
		},
		Replacements: []game.ReplacementEffect{{
			ActiveWhen: Solved(),
			Watches:    []game.EventKind{game.EventTokenCreated},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventCreateTokens && ev.TokenCount() > 0 && ev.TokenController == src.Controller
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.TokenGroups = append(ev.TokenGroups, game.TokenGroup{Template: ruleDefinedToken(ClueToken()), Count: 1})
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Case of the Pilfered Proof: plus a Clue",
		}},
	})
}

// pilferedProofDetectiveArrived is "a Detective you control enters or is
// turned face up".
func pilferedProofDetectiveArrived(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB && ev.Kind != game.EventTurnedFaceUp {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.HasSubtype(source.TextSubtype("Detective"))
}

// youControlThreeOrMoreDetectives is the Case's solve condition.
func youControlThreeOrMoreDetectives(g *game.Game, controller, source uuid.UUID) bool {
	detective := g.TextReaderOfForEffect(source).TextSubtype("Detective")
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype(detective) {
			n++
		}
	}
	return n >= 3
}
