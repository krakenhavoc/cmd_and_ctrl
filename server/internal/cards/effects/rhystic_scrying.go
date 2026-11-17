package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rhystic Scrying — Sorcery {2}{U}{U} (Prophecy, #5617):
//
//	"Draw three cards. Then if any player pays {2}, discard three
//	 cards."
//
// ADR 0150 (#1903): the three cards are drawn first, then the {2} is
// offered to every player in APNAP order from the active player, the
// caster included, until one pays (AnyPlayerPays, Rhystic Tutor's
// seam). The offer holds the step, so nothing happens between the draw
// and the answer. If a player pays, the caster discards three cards of
// their choice from the hand they hold then, the three just drawn among
// them (CR 701.9a); with fewer in hand they discard what they have.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a07971a-bfc3-4d36-ad8a-b9d395cbd6e9",
		Name:         "Rhystic Scrying",
		Purpose:      game.Purpose{Draws: 3},
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			DrawCards{N: 3},
			AnyPlayerPays{
				Cost:     "{2}",
				Question: "Rhystic Scrying — pay {2} to make its caster discard three cards?",
				Then:     rhysticScryingDiscard,
			},
		),
	})
}

// rhysticScryingDiscard is "if any player pays {2}, discard three
// cards", keyed (#4264): the caster is the asking item's controller.
var rhysticScryingDiscard = AnyPlayerConsequences("any-player/rhystic-scrying-discard",
	func(ctx *Context, _ game.AnyPlayerAnswer) error {
		ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   ctx.Controller(),
			Source:   ctx.Source(),
			N:        3,
			Question: "Rhystic Scrying — discard three cards",
		})
		return nil
	},
	nil, nil)
