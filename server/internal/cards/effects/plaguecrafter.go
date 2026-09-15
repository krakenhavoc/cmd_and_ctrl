package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plaguecrafter — Creature — Human Shaman {2}{B}, 3/2 (EDHREC rank
// 636):
//
//	"When this creature enters, each player sacrifices a creature or
//	 planeswalker of their choice. Each player who can't discards a
//	 card."
//
// Fleshbag Marauder that also answers planeswalkers and never whiffs.
// The sacrifice half is the EachPlayerSacrifices primitive with the
// match widened to "a creature or planeswalker" — one prompt per
// player, each choosing from their own permanents, the Plaguecrafter
// itself a legal answer to its own trigger. The "who can't" half is
// decided at resolution from the board: a player with neither a
// creature nor a planeswalker gets the ordinary discard-a-card prompt
// instead (Mind Rot's), and a player with nothing in hand simply
// does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2e3ee458-fa3f-4452-ab95-5a7fb5a0483b",
		Name:         "Plaguecrafter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Plaguecrafter — each player sacrifices a creature or planeswalker", func(g *game.Game, item *game.StackItem) error {
				for _, p := range g.Seats {
					if p == nil || p.Eliminated || b05ControlsCreatureOrPlaneswalker(g, p.ID) {
						continue
					}
					if p.Hand.Size() > 0 {
						g.DiscardChoiceForEffect(p.ID, 1)
					}
				}
				return EachPlayerSacrifices{
					Match: Or(Creature(), Planeswalker()),
					Label: "a creature or planeswalker",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
