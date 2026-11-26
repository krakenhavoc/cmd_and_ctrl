package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Séance — Enchantment {2}{W}{W}:
//
//	"At the beginning of each upkeep, you may exile target creature card from
//	 your graveyard. If you do, create a token that's a copy of that card,
//	 except it's a Spirit in addition to its other types. Exile it at the
//	 beginning of the next end step."
//
// The upkeep trigger targets a creature card in your graveyard and is asked
// "you may" as it resolves. The copy is made from the exiled card (CR 707.9b
// adds Spirit to the copiable values) and its exile is a delayed trigger on
// that token alone.
func init() {
	const label = "Séance — you may exile target creature card from your graveyard; if you do, create a token that's a copy of it, a Spirit in addition to its other types, and exile it at the beginning of the next end step"
	effects.Register(effects.Spec{
		OracleID:     "a375b935-c02b-44bc-a942-37c5de58b1e4",
		Name:         "Séance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(
				effects.Targeting(
					effects.AtEachUpkeep(label, effects.OldB443ExileTargetThenSeanceCopy),
					effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())),
				"Séance — exile the creature card and create a Spirit copy of it?"),
		},
	})
}
