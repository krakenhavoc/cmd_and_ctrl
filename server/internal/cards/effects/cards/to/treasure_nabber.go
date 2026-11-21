package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Treasure Nabber — Creature — Goblin Rogue {2}{R}, 3/2 (Commander 2018,
// #5526):
//
//	"Whenever an opponent taps an artifact for mana, gain control of that
//	 artifact until the end of your next turn."
//
// Manabarbs' watch (EventManaAbilityActivated, emitted for a manual
// activation and the auto-tapper alike) narrowed to an artifact that is now
// tapped and an activating player who is an opponent: "taps … for mana" is
// a mana ability with a tap in its cost, and a Treasure sacrificed for its
// mana is gone before the trigger can take it. The trigger is not a mana
// ability, so it uses the stack (CR 605.5a).
//
// "That artifact" is pinned to the object that was tapped (its object
// epoch is read as the trigger is put on the stack), so an artifact that
// left the battlefield or came back as a new object in response is not
// taken (CR 400.7). The control lasts until the end of your next turn,
// Evil's Thrall's duration: from an opponent's turn that is your coming
// turn, and from your own turn the one after it.
//
// No simplification.
func init() {
	const label = "Treasure Nabber — gain control of that artifact until the end of your next turn"
	effects.Register(effects.Spec{
		OracleID:     "20a6d576-a5b2-43b4-8be9-75be001e567d",
		Name:         "Treasure Nabber",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventManaAbilityActivated},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventManaAbilityActivated || ev.Source == uuid.Nil || ev.Actor == uuid.Nil ||
					source == nil || ev.Actor == source.Controller {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Source)
				return ok && c.IsArtifact() && c.Tapped && effects.OnBattlefield(g, c.InstanceID)
			},
			Key:   label,
			Build: effects.BuildPinnedToEventSource(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ref := item.Params.Object
				c, ok := g.LookupCardForEffect(ref.ID)
				if !ok || ref.ID == uuid.Nil || !effects.OnBattlefield(g, c.InstanceID) || c.ObjectEpoch != ref.Epoch {
					return nil
				}
				return effects.GainControl{
					Target:   c.InstanceID,
					Duration: g.UntilEndOfYourNextTurnDuration(item.Controller),
					Label:    label,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
