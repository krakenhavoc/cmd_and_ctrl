package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cyberman Patrol — Artifact Creature — Cyberman {2}, 2/2 (EDHREC
// rank 2786):
//
//	"Artifact creatures you control have afflict 3. (Whenever a
//	 creature with afflict 3 becomes blocked, defending player loses
//	 3 life.)"
//
// The artifact deck's "block or lose 3". Afflict is a triggered
// ability (CR 702.130a), and the printed card GRANTS it; a static
// cannot grant a trigger (the layer engine rewrites characteristics,
// and the harvester reads triggers off the catalog by oracle ID), so
// the ability lives on the Patrol — the Agent of the Iron Throne
// posture — and fires for every artifact creature its controller
// controls, the Patrol included since it is one. "Becomes blocked" is
// EventBecomesBlocked, emitted once per blocked ATTACKER when the
// block declaration is locked in (CR 506.4, #830), so a double block
// is one trigger and a blocker re-pointed away before the lock-in
// never made the Patrol blocked at all. The defending player is the
// blockers' controller, and the loss is life loss, not damage.
//
// Two consequences of that placement are both as printed: a Patrol
// that leaves between the block and the trigger resolving still
// makes the player lose 3 (the ability had already triggered), and
// two Patrols give afflict 3 twice, since two instances of afflict
// trigger separately (CR 702.130b).
//
// The #388 timing gap is closed on both halves and the card prints
// as printed. The lock-in (#830) drains the trigger onto the stack
// INSIDE the declare blockers step, so ordinary priority play loses
// the 3 life before combat damage; and since #914 the sandbox's
// skip-ahead button does too — advance_step passes priority until the
// step ends (CR 117.4) rather than walking the cursor past a trigger
// the step still owes. The caveat this card carried for both of them
// is gone.
func init() {
	Register(Spec{
		OracleID:     "1e51fab7-3ca5-4fbb-a1e9-b39c842895e8",
		Name:         "Cyberman Patrol",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesBlocked},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b26ArtifactCreatureYouControlBecameBlocked(ev, source, g)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				defender := ev.Actor
				return game.NewTriggeredItem(source, "Cyberman Patrol — afflict 3: defending player loses 3 life",
					func(g *game.Game, item *game.StackItem) error {
						if g.PlayerByIDForEffect(defender) == nil {
							return nil
						}
						return g.ChangePlayerLifeForEffect(item.SourceCardID, defender, -3)
					})
			},
		}},
	})
}
