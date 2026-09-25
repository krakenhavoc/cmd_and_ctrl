package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spiteful Banditry — Enchantment {X}{R}{R} (EDHREC rank 2274):
//
//	"When this enchantment enters, it deals X damage to each creature.
//	 Whenever one or more creatures your opponents control die, you
//	 create a Treasure token. This ability triggers only once each
//	 turn."
//
// A red sweeper that pays you back every turn. Two abilities:
//
//   - "When this enchantment enters, it deals X damage to each
//     creature" is a printed ENTERS TRIGGER, not an entry
//     replacement — before #1312 an enters trigger's Build had
//     nowhere to read the announced X from (the resolving spell's
//     StackItem is gone by the time the trigger fires), so this used
//     to move the read into OnResolve, a beat before the Banditry
//     left the stack, with a declared caveat that the damage arrived
//     without a trigger to respond to. #1357: Build now reads
//     source.CastX() (CastProvenance.X, CR 107.3m) and closes over
//     the plain int rather than the card, so the trigger is a real
//     CR 603 object on the stack — it can be countered, or the
//     Banditry can be removed in response to it (the trigger still
//     resolves on its own last-known-X, CR 603.10), which the
//     resolve-time shortcut could not model. The X damage still hits
//     EVERY creature, the controller's included, as printed, and is
//     dealt by the Banditry so it is red noncombat damage. Creatures
//     that took lethal die at the state-based sweep that follows —
//     by which time the Banditry is on the battlefield, so its own
//     Treasure trigger sees the deaths and the wipe makes a Treasure,
//     exactly as it does in paper.
//   - "One or more creatures your opponents control die" is one
//     trigger per batch of deaths and at most one per turn. The
//     engine emits one EventLTB per creature, and the harvester
//     emits EventTrigger the moment it queues an item, so
//     b11TriggeredThisTurn — the per-turn tally's count for a
//     trigger with this label — answers both clauses at once: the
//     second death in a wipe finds the first death's trigger already
//     recorded, and so does a death later in the same turn.
func init() {
	Register(Spec{
		OracleID:     "fde2653f-5270-4b8f-9642-0835dbb076c2",
		Name:         "Spiteful Banditry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventETB},
				AppliesTo: Self,
				Key:       "Spiteful Banditry — deal X damage to each creature",
				// #1312/#1357: read once, here (ADR 0041 P9's fill-in
				// Build), and stamp the plain int on Params.Amount
				// rather than closing over the card, which the
				// Effect must not capture.
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Spiteful Banditry — deal X damage to each creature")
					item.Params.Amount = source.CastX()
					return item
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return damageEachMatching(NewContext(g, item), Creature(), item.Params.Amount)
				},
			},
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b18OpponentsCreatureDied(ev, source, g) &&
					!b11TriggeredThisTurn(g, source.InstanceID, b21SpitefulBanditryLabel)
			}, b21SpitefulBanditryLabel, Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}

// b21SpitefulBanditryLabel is the Treasure trigger's stack label —
// named because the once-per-turn tally matches on it.
const b21SpitefulBanditryLabel = "Spiteful Banditry — create a Treasure"
