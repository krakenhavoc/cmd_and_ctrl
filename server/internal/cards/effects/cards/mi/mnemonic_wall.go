package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mnemonic Wall — Creature — Wall {4}{U}, 0/4 (Rise of the Eldrazi, #5581):
//
//	"Defender
//	 When this creature enters, you may return target instant or sorcery
//	 card from your graveyard to your hand."
//
// Gravedigger's optional enters trigger with an instant-or-sorcery clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f2933616-211b-4d3f-8064-2be61c26079a",
		Name:            "Mnemonic Wall",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Mnemonic Wall — return target instant or sorcery card from your graveyard to your hand", effects.ReturnTargetCardToHand),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery()))),
				"Mnemonic Wall — return it to your hand?"),
		},
	})
}
