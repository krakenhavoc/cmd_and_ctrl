package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Serenity — Creature — Angel {4}{W}{W}{W}, 5/6 (EDHREC rank
// 3758):
//
//	"Flying
//	 When this creature enters, you may exile up to three other target
//	 creatures from the battlefield and/or creature cards from
//	 graveyards.
//	 When this creature leaves the battlefield, return the exiled
//	 cards to their owners' hands."
//
// Seven mana for a 5/6 flier that takes three creatures with it, and
// gives them back — to their owners' HANDS, not the battlefield — if
// she ever leaves. Flying rides PrintedKeywords. The entry trigger is
// optional and targeted at up to three creatures; the leave trigger
// fires on ANY exit (dying, exile, bounce, a flicker) and returns
// every card her entry trigger exiled that is still in exile, read
// back off the event log (b27ExiledWith — the same record Duplicant
// keeps), so a card something else has since moved on is not pulled
// out of wherever it went.
//
// Sandbox simplifications, declared, both weaker than printed:
//
//   - The target clause covers creatures on the BATTLEFIELD only.
//     The picker is single-zone — a clause spanning the battlefield
//     and the graveyards has no client mode — so the "and/or creature
//     cards from graveyards" half is not offered. She still exiles
//     three opposing creatures; she cannot exile a graveyard threat,
//     and cannot be used to return your own dead creatures to hand.
//   - "Another" is enforced by NAME rather than by instance, because
//     the clause is declared at init() before any Angel exists. In a
//     singleton format that is the same creature; a token copy of
//     her would also be excluded. The Effect declines its own source
//     as well.
func init() {
	Register(Spec{
		OracleID:     "f4ce6078-8c7b-4f68-b324-13130c63a983",
		Name:         "Angel of Serenity",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The trigger can only exile creatures on the battlefield — creature cards in graveyards can't be chosen.",
			"The trigger can't target another creature named Angel of Serenity, such as a token copy of her.",
		},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:        []game.EventKind{game.EventETB},
				AppliesTo:      b06SelfETB,
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Angel of Serenity — exile up to three other target creatures?"},
				Targets:        TargetCreature("up to three other target creatures", b03NotNamed("Angel of Serenity")).WithCount(0, 3),
				Key:            b36AngelOfSerenityExileLabel,
				Effect:         b36ExileChosenCreatures,
			},
			On(game.EventLTB, Self, "Angel of Serenity — return the exiled cards to their owners' hands", b36ReturnCardsExiledWithToOwnersHands),
		},
	})
}
