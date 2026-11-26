package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grixis Sojourners — Creature — Zombie Ogre {1}{U}{B}{R}, 4/3 (Alara Reborn,
// #5586):
//
//	"When you cycle this card and when this creature dies, you may exile target
//	 card from a graveyard.
//	 Cycling {2}{B}"
//
// One ability with two conditions (S82OldB504CycledOrDies); the card, from any
// graveyard, is chosen as the trigger goes on the stack and judged again as it
// resolves (CR 608.2b); the exile is optional. The Sojourners itself is in a
// graveyard when the ability triggers, and so may be the target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "21f507fd-6a26-455c-879a-f838e0aa2a43",
		Name:         "Grixis Sojourners",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}{B}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.S82OldB504CycledOrDies("Grixis Sojourners — you may exile target card from a graveyard", effects.S82OldB504ExileTheFirstTarget),
					"Grixis Sojourners — exile the target card?"),
				effects.TargetCardInGraveyard("target card from a graveyard")),
		},
	})
}
