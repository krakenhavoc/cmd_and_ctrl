package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiend Hunter — Creature — Human Cleric {1}{W}{W}, 1/3 (The Hobbit Eternal,
// HOC):
//
//	"When this creature enters, you may exile another target creature.
//	 When this creature leaves the battlefield, return the exiled card to
//	 the battlefield under its owner's control."
//
// The original two-trigger Fiend Hunter, not Oblivion Ring's "until" form
// (#4454): the two abilities are linked (CR 607.2a), so "the exiled card" is
// the one the entry trigger exiled, and only while it is still in exile.
// b27ExiledWith reads that record off the event log by the entry trigger's
// key, the way Gravegouger's leave trigger does, and a Hunter that comes
// back is a new object with an empty record (CR 400.7). The consequence the
// Oracle rulings state: if the Hunter leaves the battlefield before its
// entry trigger resolves, the leave trigger finds nothing, and the creature
// is then exiled for good.
//
// The "may" is answered as the trigger resolves (CR 603.5), after the
// target is chosen. "Another" excludes the Hunter itself. A token exiled
// this way ceases to exist and does not come back. The return is one
// simultaneous entry under the card's owner's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb9d557a-fc06-428c-8be6-7d28add33028",
		Name:         "Fiend Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:        []game.EventKind{game.EventETB},
				AppliesTo:      Self,
				OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Fiend Hunter — exile another target creature?"},
				Targets:        Another(TargetCreature("another target creature")),
				Key:            fiendHunterExileLabel,
				Effect:         exileFirstLegalCardTarget,
			},
			On(game.EventLTB, Self, "Fiend Hunter — return the exiled card to the battlefield under its owner's control",
				b41ReturnCardsExiledWithToTheBattlefield(fiendHunterExileLabel)),
		},
	})
}

// fiendHunterExileLabel is the entry trigger's key, which is what the leave
// trigger reads the linked exile by.
const fiendHunterExileLabel = "Fiend Hunter — exile another target creature"
