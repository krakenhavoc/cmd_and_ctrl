package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent of the Iron Throne — Legendary Enchantment — Background {2}{B}
// (EDHREC rank 2270):
//
//	"Commander creatures you own have "Whenever an artifact or creature
//	 you control is put into a graveyard from the battlefield, each
//	 opponent loses 1 life.""
//
// The aristocrats Background. Since ADR 0093 (PR 3) the ability is
// really GRANTED — a layer-6 grant of a trigger bundle to every
// commander creature the Background's controller OWNS — rather than
// the Background's own trigger gated on such a commander existing.
// Three things that the old posture got wrong are right now:
//
//   - a commander an opponent has stolen still has the ability (you
//     still own it), and the ability is the THIEF's: "you control"
//     reads the commander's controller, so the thief's artifacts and
//     creatures dying drain the thief's opponents;
//   - a commander that loses all its abilities (a later Darksteel
//     Mutation) loses this one;
//   - the commander's OWN death counts because the harvest reads its
//     last-known abilities (CR 603.10a), not because a gate special-
//     cases it.
//
// A commander whose death is replaced by the command zone (the
// engine's CR 903.9 prompt) never reaches a graveyard, so that death
// does not drain, as printed. "Choose a Background" is a
// deck-construction rule (CR 702.124), the deck importer's business.
//
// No simplification.
const agentOfTheIronThroneGrant = "agent-of-the-iron-throne/drain"

func init() {
	Register(Spec{
		OracleID:     "325032d5-c452-4454-8976-82f86fee5ab8",
		Name:         "Agent of the Iron Throne",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: agentOfTheIronThroneGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					_, ok := b21ArtifactOrCreatureYouControlDied(ev, source, g)
					return ok
				}, "Agent of the Iron Throne — each opponent loses 1 life", func(g *game.Game, item *game.StackItem) error {
					return eachOpponentLosesLife(g, item, 1)
				}),
			},
			Text: "Whenever an artifact or creature you control is put into a graveyard from the battlefield, each opponent loses 1 life.",
		}},
		Static: []game.StaticAbility{GrantAbilities(commanderCreatureYouOwn, agentOfTheIronThroneGrant)},
	})
}

// commanderCreatureYouOwn is the "commander creatures you own" AppliesTo
// a Background's grant hangs on: OWNED by the Background's controller,
// whoever controls it now.
func commanderCreatureYouOwn(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCommander && target.IsCreature() && target.Owner == source.Controller
}
