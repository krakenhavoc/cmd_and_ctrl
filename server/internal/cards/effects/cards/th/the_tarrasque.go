package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Tarrasque — Legendary Creature — Dinosaur {6}{G}{G}{G}, 10/10
// (Adventures in the Forgotten Realms):
//
//	"The Tarrasque has haste and ward {10} as long as it was cast.
//	 Whenever The Tarrasque attacks, it fights target creature defending
//	 player controls."
//
// "As long as it was cast" is the permanent's cast provenance (CR 400.7d):
// the record of its caster is written as the spell becomes the permanent, so
// a Tarrasque put onto the battlefield without being cast has neither. Haste
// is a layer 6 grant to itself and the ward is the granted-ward trigger over
// the same predicate. The fight is the attack trigger's own, on a creature
// the defending player controls chosen as it goes on the stack; a Tarrasque
// or target that has left deals and is dealt no damage (CR 701.14b).
//
// No simplification.
func init() {
	wasCast := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && tarrasqueWasCast(target)
	}
	haste := effects.KeywordGrant(wasCast, "haste")
	haste.Label = "The Tarrasque has haste as long as it was cast."
	fight := effects.WheneverThisAttacks("The Tarrasque — it fights target creature defending player controls", effects.B30SourceFightsFirstLegalTarget)
	fight.TargetsFrom = effects.TargetCreatureDefendingPlayerControls
	effects.Register(effects.Spec{
		OracleID:     "bfd66214-e576-4a7e-8559-4e4e51d56b98",
		Name:         "The Tarrasque",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{haste},
		Triggered: []game.TriggeredAbility{
			effects.WardGranted(effects.WardMana("{10}"), "The Tarrasque — ward {10} as long as it was cast", wasCast),
			fight,
		},
	})
}

// tarrasqueWasCast reports that the permanent entered as a resolving spell.
func tarrasqueWasCast(c *game.Card) bool { return c.Provenance.Caster != uuid.Nil }
