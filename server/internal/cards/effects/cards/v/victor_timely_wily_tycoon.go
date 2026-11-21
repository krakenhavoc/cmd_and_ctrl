package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Victor Timely, Wily Tycoon — Legendary Creature — Human Artificer Villain
// {4}{U}, 3/4 (Marvel Super Heroes Commander, #4455):
//
//	"When Victor Timely enters, you may cast target artifact, instant, or
//	 sorcery card with mana value 4 or less from your graveyard without
//	 paying its mana cost. If that spell would be put into your graveyard,
//	 exile it instead."
//
// Torrential Gearhulk's free cast from the graveyard
// (freeCastTargetFromYourGraveyardExilingIt, extracted for the pair) over a
// wider target clause. The mana value is the card's in the graveyard
// (X is 0 there, CR 202.3e). The spell cast is the real card, so an artifact
// that resolves enters the battlefield and one that is countered is exiled.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Victor Timely — you may cast target artifact, instant, or sorcery card with mana value 4 or less from your graveyard without paying its mana cost",
		effects.FreeCastTargetFromYourGraveyardExilingIt("Victor Timely"))
	enters.Targets = effects.TargetCardInGraveyard("target artifact, instant, or sorcery card with mana value 4 or less from your graveyard",
		effects.Or(effects.Artifact(), effects.Instant(), effects.Sorcery()), effects.ManaValueLE(4), effects.YouOwn())
	effects.Register(effects.Spec{
		OracleID:     "2ad77eb7-9466-432f-87b4-e80e4b66e143",
		Name:         "Victor Timely, Wily Tycoon",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
