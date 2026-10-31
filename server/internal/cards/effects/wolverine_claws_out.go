package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wolverine, Claws Out — {3}{G} Legendary Creature — Mutant Berserker Hero,
// 2/5 (Marvel Super Heroes Commander, #4455):
//
//	"You may have Wolverine assign his combat damage as though he weren't
//	 blocked.
//	 Whenever a Mutant you control attacks, double its power until end of
//	 turn."
//
// "As though he weren't blocked" is Characteristic.AssignsAsThoughUnblocked
// (CR 510.1c): a layer 6 static on Wolverine sets the flag on Wolverine
// alone, so a Wolverine that has lost his abilities (CR 613.1f) loses the
// option with them. The combat damage step reads it through the layered
// view and, for a blocked Wolverine, asks his controller the damage-
// assignment prompt even with a single blocker: divide his damage among
// his blockers as usual, or put all of it on the player, planeswalker or
// battle he is attacking (DamageAssignmentFrame.AsThoughUnblocked). The
// second answer owes no blocker lethal damage and needs no trample; it
// deals damage to what he attacks and none to the blocker, with lifelink,
// deathtouch and commander damage reading the same frame as any other
// assignment.
//
// The attack trigger is the per-attacker EventAttack one (Agents of
// S.H.I.E.L.D.'s shape): the Mutant is judged as it attacks, Wolverine is
// one of the Mutants he doubles, and the doubling is the CR 701.10 one
// (DoublePowerUntilEOT), its X read as the trigger resolves, so a pump
// later in the turn is not doubled again and a Mutant removed in response
// gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7100fe0c-1c2a-461a-83b3-684d44279081",
		Name:         "Wolverine, Claws Out",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "You may have Wolverine assign his combat damage as though he weren't blocked.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.AssignsAsThoughUnblocked = true
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, wolverineClawsOutMutantAttacks,
				"Wolverine, Claws Out — double that Mutant's power until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return DoublePowerUntilEOT(NewContext(g, item), item.Trigger.Event.CardID,
						"Wolverine, Claws Out — double power")
				}),
		},
	})
}

// wolverineClawsOutMutantAttacks is "a Mutant you control attacks": the
// attacker is a Mutant you control as it is declared. The subtype is a
// word of the source's text (CR 612.2), read through the text the way
// YouCast reads a spell's.
func wolverineClawsOutMutantAttacks(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() || c.Controller != source.Controller {
		return false
	}
	return game.ThroughTextOf(source, c).HasSubtypeWord("Mutant")
}
