package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exclusion Ritual — Enchantment {4}{W}{W}:
//
//	"Imprint — When this enchantment enters, exile target nonland permanent.
//	 Players can't cast spells with the same name as the exiled card."
//
// The restriction reads the exiled-with link live (CR 607.2a), so it lifts
// the moment the Ritual leaves, and binds every player, its controller too.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e3fd5da-6724-4134-ab15-4616cb659f88",
		Name:         "Exclusion Ritual",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Exclusion Ritual — Imprint: exile target nonland permanent", effects.ImprintExileTheTarget),
				effects.TargetPermanent("target nonland permanent", effects.Nonland())),
		},
		CastRestrictions: []game.CastRestriction{effects.PlayersCantCastSpellsNamedLikeTheExiledCard()},
	})
}
