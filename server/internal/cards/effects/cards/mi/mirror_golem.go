package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirror Golem — Artifact Creature — Golem {6}, 3/4 (Mirrodin, #5606):
//
//	"Imprint — When this creature enters, you may exile target card from a
//	 graveyard.
//	 This creature has protection from each of the exiled card's card types.
//	 (Artifact, battle, creature, enchantment, instant, kindred, land,
//	 planeswalker, and sorcery are card types.)"
//
// Mourner's Shield's optional imprint (the exile is linked to this object,
// CR 607.2a) and Phyrexian Ingester's static read of the exiled card, as it
// sits in exile: one protection ability per card type on its type line
// (CR 702.16). An object that left and came back remembers nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3460e86f-c182-4ca9-bd90-7949b0759c98",
		Name:         "Mirror Golem",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB606ProtectionFromExiledCardTypes()},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Mirror Golem — you may exile target card from a graveyard", effects.ExileLinkedTargetsEffect),
				effects.TargetCardInGraveyard("target card from a graveyard")),
				"Mirror Golem — exile target card from a graveyard?"),
		},
	})
}
