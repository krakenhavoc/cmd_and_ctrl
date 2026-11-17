package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Attended Socialite — Creature — Elf Druid {1}{G}, 2/1 (Streets of New
// Capenna, #5491):
//
//	"Alliance — Whenever another creature you control enters, this creature
//	 gets +1/+1 until end of turn."
//
// Alliance is "another creature you control enters". The pump is a layer-7c
// effect on the Socialite until end of turn; it stacks with each trigger,
// and does nothing if the Socialite has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5f7c6fa9-4208-4dc8-a75b-9586200d8e19",
		Name:         "Attended Socialite",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureEnteredUnderYourControl,
				"Attended Socialite — alliance: this creature gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1,
						Label: "Attended Socialite — +1/+1"}.Apply(NewContext(g, item))
				}),
		},
	})
}
