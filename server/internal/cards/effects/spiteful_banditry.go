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
//   - The X damage hits EVERY creature, the controller's included, as
//     printed, and is dealt by the Banditry so it is red noncombat
//     damage. Creatures that took lethal die at the state-based
//     sweep that follows — by which time the Banditry is on the
//     battlefield, so its own Treasure trigger sees the deaths and
//     the wipe makes a Treasure, exactly as it does in paper.
//   - "One or more creatures your opponents control die" is one
//     trigger per batch of deaths and at most one per turn. The
//     engine emits one EventLTB per creature, and the harvester
//     emits EventTrigger the moment it queues an item, so
//     b11TriggeredThisTurn — a walk back to the turn's upkeep for a
//     trigger with this label — answers both clauses at once: the
//     second death in a wipe finds the first death's trigger already
//     recorded, and so does a death later in the same turn.
//
// SANDBOX SIMPLIFICATION, declared — the Springleaf Parade posture:
// the X damage is dealt as the spell RESOLVES, a beat before the
// Banditry moves from the stack to the battlefield, rather than by
// an enters trigger on the stack. An ETB trigger cannot see the X
// announced for the spell (the stack item is gone by the time the
// trigger fires), and resolution is the last moment X is readable.
// So the damage comes without a trigger to respond to; nothing is
// stronger for it, because the spell itself could be countered, and
// the deaths still happen with the Banditry in play.
func init() {
	Register(Spec{
		OracleID:     "fde2653f-5270-4b8f-9642-0835dbb076c2",
		Name:         "Spiteful Banditry",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The X damage is dealt as the spell resolves rather than by an enters trigger you can respond to."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, Creature(), ctx.X())
		},
		Triggered: []game.TriggeredAbility{
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
