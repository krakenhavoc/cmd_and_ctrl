package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thought Prison — Artifact {5} (Mirrodin, #5606):
//
//	"Imprint — When this artifact enters, you may have target player reveal
//	 their hand. If you do, choose a nonland card from it and exile that card.
//	 Whenever a player casts a spell that shares a color or mana value with the
//	 exiled card, this artifact deals 2 damage to that player."
//
// Null Summoner's reveal-and-choose with a linked exile (ChooseFromRevealedHand
// Exile + Linked, CR 607.2a), and a watch on every cast that reads the exiled
// card: it shares a color when one of its colors is one of the spell's, and a
// mana value when the two are equal (a card with no mana cost has mana
// value 0). A colorless exiled card shares no color.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da1ea03f-747f-46c9-8ff5-7e85996df804",
		Name:         "Thought Prison",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Thought Prison — you may have target player reveal their hand, then exile a nonland card from it",
				effects.NullSummonerExileFromHand), effects.TargetPlayer("target player")),
				"Thought Prison — have target player reveal their hand?"),
			effects.On(game.EventCast, effects.OldB606ExiledCardSharesWithSpell,
				"Thought Prison — it deals 2 damage to that player", effects.OldB606PrisonDamagesTheCaster),
		},
	})
}
