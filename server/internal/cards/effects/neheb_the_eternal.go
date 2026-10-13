package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Neheb, the Eternal — Legendary Creature — Zombie Minotaur Warrior
// {3}{R}{R}, 4/6:
//
//	"Afflict 3 (Whenever this creature becomes blocked, defending
//	 player loses 3 life.)
//	 At the beginning of each of your postcombat main phases, add {R}
//	 for each 1 life your opponents have lost this turn."
//
// Afflict is a triggered ability (CR 702.130a) and lives here as one,
// the posture Cyberman Patrol takes for the grant: EventBecomesBlocked
// is emitted once per blocked attacker at the block declaration's
// lock-in (CR 506.4), with the attacker in Target and the defending
// player in Actor, and the loss is life loss, not damage.
//
// The mana is Neheb's own triggered ability, not a mana ability: it
// uses the stack, so it can be responded to and countered, and the
// life lost is counted as it resolves (CR 608.2h), which includes
// the 3 from this turn's afflict because that trigger resolved during
// combat. The count is the total of every opponent's life lost this
// turn (the per-turn tally Rakdos, Lord of Riots reads), so an
// opponent who lost life and then gained it back still counts. The
// mana empties at the end of the phase (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0d3fd57-1420-4dd1-8410-9d659655287f",
		Name:         "Neheb, the Eternal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventBecomesBlocked},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventBecomesBlocked && ev.Target == source.InstanceID
				},
				Key:    "Neheb, the Eternal — afflict 3: defending player loses 3 life",
				Effect: afflictDefenderLoses(3),
			},
			AtYourPostcombatMain("Neheb, the Eternal — add {R} for each 1 life your opponents have lost this turn", func(g *game.Game, item *game.StackItem) error {
				n := opponentLifeLostThisTurn(g, item.Controller)
				if n <= 0 {
					return nil
				}
				return AddMana{Amount: ManaN(n, "R")}.Apply(NewContext(g, item))
			}),
		},
	})
}
