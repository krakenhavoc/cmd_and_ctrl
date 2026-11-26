package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sharpened Pitchfork — {2} Artifact — Equipment (Shadows over Innistrad via
// Duel Decks: Blessed vs. Cursed, S82 OLD b294):
//
//	"Equipped creature has first strike.
//	 As long as equipped creature is a Human, it gets +1/+1.
//	 Equip {1}"
//
// Hero's Heirloom's conditional grant, read off the host every layer pass;
// the +1/+1 is the same condition on a layer 7c modification. "Human" is a
// word of the Pitchfork's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "409b44fc-fae2-4804-83a7-78cf95038f58",
		Name:         "Sharpened Pitchfork",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.GrantToAttached("first strike"),
			{
				Label:    "As long as equipped creature is a Human, it gets +1/+1.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && target.HasSubtype(source.TextSubtype("Human"))
				},
				Apply: effects.PumpBy(1, 1),
			},
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
