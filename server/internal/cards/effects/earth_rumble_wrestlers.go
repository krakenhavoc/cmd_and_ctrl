package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earth Rumble Wrestlers — Creature — Human Warrior Performer {3}{R/G},
// 3/4 (Avatar: The Last Airbender, #4459):
//
//	"Reach
//	 This creature gets +1/+0 and has trample as long as you control a
//	 land creature or a land entered the battlefield under your control
//	 this turn."
//
// One condition, two statics: a layer 7c +1/+0 and a layer 6 trample
// grant, both gated on earthRumbleWrestlersCondition and read on every
// layer pass. "A land creature" is a permanent you control that is both
// a land and a creature through the layers (an earthbent land, Dryad
// Arbor); "a land entered under your control this turn" is the turn's
// entry tally (PlayerTurnTally.LandsEntered), which a land that has since
// left still counts.
//
// No simplification.
func init() {
	when := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfCreature(target, g, source) && earthRumbleWrestlersCondition(g, source.Controller)
	}
	trample := KeywordGrant(when, "trample")
	trample.Label = "Has trample as long as you control a land creature or a land entered the battlefield under your control this turn."
	Register(Spec{
		OracleID:        "1de17747-27e8-40bb-b6f7-6d5402d4705c",
		Name:            "Earth Rumble Wrestlers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static: []game.StaticAbility{
			{
				Label:     "Gets +1/+0 as long as you control a land creature or a land entered the battlefield under your control this turn.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: when,
				Apply:     pumpBy(1, 0),
			},
			trample,
		},
	})
}

// earthRumbleWrestlersCondition is "you control a land creature or a land
// entered the battlefield under your control this turn".
func earthRumbleWrestlersCondition(g *game.Game, you uuid.UUID) bool {
	if g.TurnTallyFor(you).LandsEntered > 0 {
		return true
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == you && c.IsLand() && c.IsCreature() {
			return true
		}
	}
	return false
}
