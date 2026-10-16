package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Metrognome — Artifact {4}:
//
//	"When a spell or ability an opponent controls causes you to discard
//	 this card, create four 1/1 colorless Gnome artifact creature
//	 tokens.
//	 {4}, {T}: Create a 1/1 colorless Gnome artifact creature token."
//
// The activated ability is an ordinary tap-and-pay token maker. The
// first is the family's "this card" trigger (ADR 0153 §4), watching from
// wherever the card landed (WhereverDiscarded, #3859) on a discard an opponent's spell or ability caused (the
// stack item's controller, CR 113.8). The four Gnomes are the owner's,
// since "you" is the card's owner while it sits in the graveyard
// (CR 108.4). The row declares what it pays, four tokens, so the bot
// knows that an opponent's Mind Rot hands it four Gnomes (ADR 0153 §5).
//
// No simplification: the discard ability is watched from wherever the
// discard left the card (WhereverDiscarded, #3859), so madness or Rest in
// Peace sending it to exile instead of the graveyard does not stop it.
func init() {
	Register(Spec{
		OracleID:          "57fe5c94-ee93-45c6-bf7e-89d177d0fe70",
		Name:              "Metrognome",
		Completeness:      CompletenessFull,
		DiscardByOpponent: game.DiscardByOpponentTrigger,
		Triggered: []game.TriggeredAbility{
			WhereverDiscarded(TriggerWithPurpose(On(game.EventDiscardCard, DiscardedByAnOpponent(true),
				"Metrognome — discarded: create four 1/1 Gnome artifact creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 colorless Gnome artifact"), N: 4})),
				game.Purpose{OnOpponentDiscard: &game.DiscardPayoff{Any: true, Tokens: 4}})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Create a 1/1 colorless Gnome artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 colorless Gnome artifact"), N: 1}),
		}},
	})
}
