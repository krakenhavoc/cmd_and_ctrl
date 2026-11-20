package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "e100f9fc-b6fb-42ab-b9b3-796bede6804d",
		Name:            "Kalemne, Disciple of Iroas",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(And(Creature(), ManaValueGE(5)),
				"Kalemne, Disciple of Iroas — you get an experience counter", youGetAnExperienceCounter),
		},
		Static: []game.StaticAbility{{
			Label:     "Kalemne gets +1/+1 for each experience counter you have.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := experienceCounters(g, source.Controller)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
