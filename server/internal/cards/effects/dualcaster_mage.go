package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dualcaster Mage — Creature — Human Wizard {1}{R}{R}, 2/2 (EDHREC
// rank 668):
//
//	"Flash
//	 When this creature enters, copy target instant or sorcery
//	 spell. You may choose new targets for the copy."
//
// Reverberate on a body — and half of the Twinflame / Heat Shimmer
// combo. Flash rides PrintedKeywords, which is what lets the Mage be
// cast with the spell it wants to copy still on the stack. The ETB
// is a TARGETED trigger whose clause reaches the stack zone (the
// same instantOrSorcerySpell clause Reverberate declares): the
// engine computes the legal set when the Mage enters, removes the
// trigger with no prompt if no instant or sorcery spell is on the
// stack (CR 603.3d), and otherwise asks the controller to pick. The
// Effect is the CopySpell primitive with the CR 706.10c re-target
// prompt enabled; the copy is controlled by the Mage's controller
// (CR 706.10a), which is the whole point of flashing it in against
// an opponent's spell.
//
// A spell that leaves the stack between the pick and the
// resolution — countered in response — is the CR 608.2b outcome,
// and CopySpell already treats it as "nothing to copy".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8eb7c0a5-6190-40de-b473-2d1daa3bbe28",
		Name:            "Dualcaster Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: instantOrSorcerySpell("target instant or sorcery spell"),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Dualcaster Mage — copy target instant or sorcery spell",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
							return nil
						}
						return CopySpell{
							StackID:          item.Targets[0].ID,
							Controller:       item.Controller,
							ChooseNewTargets: true,
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
