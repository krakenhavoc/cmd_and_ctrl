package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shielded by Faith — Enchantment — Aura {1}{W}{W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 Enchanted creature has indestructible.
//	 Whenever a creature enters, you may attach this Aura to that creature."
//
// The indestructible grant is Timely Ward's (GrantToAttached). The trigger
// watches every creature entering under any player's control (CR 603.2),
// and "you may" is the CR 603.5 prompt on resolution. The move is
// AttachForEffect, which is CR 701.3b's whole rule for an attach that
// cannot happen: an Aura that can't legally enchant that creature (its
// enchant clause, or protection from the Aura, CR 702.16c) stays where it is
// (CR 303.4i), and a creature that has left the battlefield by then is not
// there to be enchanted. An Aura that left and came back is a new object
// that this ability knows nothing about (CR 400.7, #1432), so it moves
// nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79085dea-8daa-41c0-a975-61e3e6da2e26",
		Name:         "Shielded by Faith",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			GrantToAttached("indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, AnotherCreatureEntered,
				"Shielded by Faith — attach this Aura to that creature",
				shieldedByFaithAttach), "Shielded by Faith — attach this Aura to the creature that entered?"),
		},
	})
}

// shieldedByFaithAttach attaches the Aura to the creature whose entering
// triggered it.
func shieldedByFaithAttach(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || sourceIsNewObject(g, item) {
		return nil
	}
	return g.AttachForEffect(item.SourceCardID, game.TargetRef{Kind: game.TargetCard, ID: item.Trigger.Event.CardID})
}
