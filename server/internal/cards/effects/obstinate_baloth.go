package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obstinate Baloth — Creature — Beast {2}{G}{G}, 4/4:
//
//	"When this creature enters, you gain 4 life.
//	 If a spell or ability an opponent controls causes you to discard
//	 this card, put it onto the battlefield instead of putting it into
//	 your graveyard."
//
// ADR 0153, the Baloth shape every card in the family shares: a CR 614
// replacement the card carries while it is in a HAND (found by the
// engine's self-replacement pass, the way madness is), written as
// DiscardBecomes{ByOpponent, ThisCard, Dst: ZoneBattlefield}.
//
// The test is the stack item's controller (CR 113.8), never the source
// card's: Mind Rot, Painful Quandary's trigger and a stolen permanent's
// ability are an opponent's, and your own Careful Study, a discard paid
// as a cost and the cleanup step's discard are not, so those leave the
// Baloth in the graveyard. Painful Quandary asks you to discard while
// its trigger resolves, so it is an effect discard and the Baloth lands
// (owner decision 4). The card was discarded either way: "whenever you
// discard" payoffs still see it, with the zone it landed in.
//
// The card enters through the ordinary entry pipeline, under its
// owner's control, so the "you gain 4 life" trigger fires from the
// battlefield and an enters-tapped effect still applies (CR 614.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:          "ee2ab1ab-be1e-4e56-99f2-7784f7350b41",
		Name:              "Obstinate Baloth",
		Completeness:      CompletenessFull,
		DiscardByOpponent: game.DiscardByOpponentBattlefield,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Obstinate Baloth — you gain 4 life", Do(GainLife{Amount: 4})),
		},
		Replacements: []game.ReplacementEffect{
			OntoTheBattlefieldWhenAnOpponentMakesYouDiscardThis("Obstinate Baloth", nil),
		},
	})
}
