package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Djinn Illuminatus — Creature — Djinn {5}{U/R}{U/R}, 3/5 (Commander 2020, S82
// OLD, #5513):
//
//	"Flying
//	 Each instant and sorcery spell you cast has replicate. The replicate cost
//	 is equal to its mana cost."
//
// Hatchery Sliver's grant over instants and sorceries: a granted optional
// cost on each of your instant and sorcery spells, priced at the spell's own
// mana cost, with the trigger that copies the spell once per payment (CR
// 702.56a). Only the Djinn's controller's spells have it.
//
// No simplification.
func init() {
	spellCost := effects.ReplicateMana("", 20)
	spellCost.ManaCostIsSpellCost = true
	spellCost.Label = "Replicate"
	isInstantOrSorcery := func(_ *game.Game, _ uuid.UUID, spell game.Card) bool {
		return spell.IsInstant() || spell.IsSorcery()
	}
	effects.Register(effects.Spec{
		OracleID:        "d7dcf1be-2779-419d-8c90-03800f9e7ea2",
		Name:            "Djinn Illuminatus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		GrantedOptionalCosts: []game.GrantedOptionalCost{effects.SpellsYouCastHave(
			"Each instant and sorcery spell you cast has replicate. The replicate cost is equal to its mana cost.", isInstantOrSorcery, spellCost)},
		Triggered: []game.TriggeredAbility{
			effects.S82SncReplicateTrigger("Djinn Illuminatus — replicate: copy the instant or sorcery spell", isInstantOrSorcery),
		},
	})
}
