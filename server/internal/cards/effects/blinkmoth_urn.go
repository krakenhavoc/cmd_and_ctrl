package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blinkmoth Urn — Artifact {5} (Mirrodin, #5606):
//
//	"At the beginning of each player's first main phase, if this artifact
//	 is untapped, that player adds {C} for each artifact they control."
//
// Howling Mine's shape on Cheering Crowd's step: the trigger fires in
// every player's first main phase (the precombat main), and "that
// player" is the step event's actor, who gets the mana. "If this
// artifact is untapped" is an intervening if (CR 603.4), checked as the
// ability triggers and again as it resolves, so tapping the Urn in
// response stops the mana. The artifacts are counted as the ability
// resolves, the Urn among them if that player controls it; the mana
// empties with the phase (CR 106.4).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08497e8f-3521-48e4-a1e7-c192b485d891",
		Name:         "Blinkmoth Urn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginPrecombatMain, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return !source.Tapped
			}, "Blinkmoth Urn — that player adds {C} for each artifact they control",
				func(g *game.Game, item *game.StackItem) error {
					if c, ok := g.LookupCardForEffect(item.SourceCardID); !ok || c.Tapped {
						return nil
					}
					player := item.Trigger.Event.Actor
					n := b03ArtifactsControlled(g, player)
					if n <= 0 {
						return nil
					}
					return AddMana{Player: player, Amount: ManaN(n, "C")}.Apply(NewContext(g, item))
				}),
		},
	})
}
