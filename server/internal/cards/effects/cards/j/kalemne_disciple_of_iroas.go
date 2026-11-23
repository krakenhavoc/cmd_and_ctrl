package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kalemne, Disciple of Iroas — Legendary Creature — Giant Soldier
// {2}{R}{W}, 3/3 (Commander 2015, #5546):
//
//	"Double strike, vigilance
//	 Whenever you cast a creature spell with mana value 5 or greater, you
//	 get an experience counter.
//	 Kalemne gets +1/+1 for each experience counter you have."
//
// The cast trigger is Skybeast Tracker's (WheneverYouCast over ManaValueGE,
// which counts an announced X) narrowed to creature spells; the experience
// counter is a player counter (CR 122.1, Ezuri's youGetAnExperienceCounter).
// The pump is a layer 7c static over Kalemne alone that reads its
// controller's experience counters on every recompute; a player counter
// being placed invalidates the layers (EventPlayerCounterPlaced).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e100f9fc-b6fb-42ab-b9b3-796bede6804d",
		Name:            "Kalemne, Disciple of Iroas",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.And(effects.Creature(), effects.ManaValueGE(5)),
				"Kalemne, Disciple of Iroas — you get an experience counter", effects.YouGetAnExperienceCounter),
		},
		Static: []game.StaticAbility{{
			Label:     "Kalemne gets +1/+1 for each experience counter you have.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.ExperienceCounters(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
