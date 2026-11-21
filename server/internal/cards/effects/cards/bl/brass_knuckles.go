package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brass Knuckles — Artifact — Equipment {4} (Streets of New Capenna
// Commander, #5491):
//
//	"When you cast this spell, copy it. (The copy becomes a token.)
//	 Equipped creature has double strike as long as two or more Equipment
//	 are attached to it.
//	 Equip {1}"
//
// Combadge's cast trigger: the copy names the spell without targeting it,
// so one countered in response is still copied from last-known information
// (CR 608.2h), and CR 707.10f makes the copy a token as it resolves. The
// double strike is a layer-6 grant to the host while the Equipment attached
// to it, counting itself, number two or more (CR 301.5); the condition is
// re-read on every pass, so moving one away ends it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a8630a8-804e-4789-b9a3-96d9ede5e188",
		Name:         "Brass Knuckles",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttachedWhile(func(host *game.Card, g *game.Game, _ *game.Card) bool {
				return effects.S82SncEquipmentAttachedTo(g, host.InstanceID) >= 2
			}, "double strike"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
		Triggered: []game.TriggeredAbility{
			effects.WhenYouCastThisSpell("Brass Knuckles — copy this spell",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CopySpell{StackID: item.SourceCardID, Controller: item.Controller, FromLastKnown: true}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
