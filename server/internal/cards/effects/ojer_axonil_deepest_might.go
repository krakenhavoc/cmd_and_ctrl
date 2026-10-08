package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ojer Axonil, Deepest Might — Legendary Creature — God {2}{R}{R},
// 4/4, the front face of a transforming card whose back is Temple of
// Power:
//
//	"Trample
//	 If a red source you control would deal an amount of noncombat
//	 damage less than Ojer Axonil's power to an opponent, that source
//	 deals damage equal to Ojer Axonil's power instead.
//	 When Ojer Axonil dies, return it to the battlefield tapped and
//	 transformed under its owner's control."
//
// A four-mana 4/4 trample that turns every ping in a red deck into a
// four-point hit: a Shock at the table's face becomes a Lightning
// Helix-sized swing, and a Vivi Ornitier or a Firebrand Archer
// trigger stops being chip damage. In a deck full of one-damage
// triggers it is the single largest multiplier on the board.
//
// The damage clause is a CR 614 replacement, and all three of its
// conditions are load-bearing:
//
//   - "a red SOURCE YOU CONTROL" — the source is the damage event's
//     snapshot of it: a burn spell as it stands on the stack, and a
//     creature that has already left as it last existed on the
//     battlefield (#1417, CR 608.2h), its colour and controller read
//     post-layers. A source the engine cannot read is left alone,
//     which errs weaker and never stronger.
//   - "NONCOMBAT damage" — the whole reason the card is not simply a
//     Gratuitous Violence. A 1/1 attacker still deals 1 in combat;
//     only damage from spells, abilities and triggers is raised.
//   - "to an OPPONENT" — a player, and not a permanent. Damage at
//     your own face, at any creature, and at a planeswalker is
//     untouched.
//
// It raises damage to Ojer's power and never lowers it, so a source
// already dealing more than Ojer's power is unaffected, and two
// copies of the effect reach the same number in either CR 616 order.
// Power is read at the moment the damage would be dealt, post-layers
// and with counters folded in, so an anthem or a +1/+1 counter makes
// every ping bigger at once.
//
// The dies trigger is "return it to the battlefield tapped and
// transformed under its owner's control" (CR 712.14a), built on
// ReturnFromGraveyard{Transformed: true} (#1900, ADR 0079 amendment
// of 2026-10-08). The card comes back as Temple of Power, tapped, a
// new object that never transformed.
//
// ONE DECLARED SIMPLIFICATION: Temple of Power's second ability,
// "{2}{R}, {T}: Transform this land. Activate only if red sources you
// controlled dealt 4 or more noncombat damage this turn and only as a
// sorcery", is not registered. The turn tally keeps no per-source
// amount of noncombat damage by colour (DamageDealtRecord is a
// once-per-creature flag), and an activation with its condition
// dropped would be STRONGER than printed (#259). The land taps for
// {R} and stays a land, which is weaker; the condition lands with
// that tally.
func init() {
	Register(Spec{
		OracleID:        ojerAxonilOracleID,
		Name:            "Ojer Axonil, Deepest Might",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"trample"},
		Caveats: []string{
			"Temple of Power's ability to transform back into Ojer Axonil isn't implemented — once the God dies and returns as the land, it stays a land.",
		},
		Triggered: []game.TriggeredAbility{ojerDiesReturnTransformed("Ojer Axonil", nil)},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
			// It only ever raises the amount ("deals damage equal to
			// Ojer Axonil's power instead"), so it is not a CR 615
			// prevention effect and "can't be prevented" leaves it be.
			Prevention: false,
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				if ev.Kind != game.RepEventDamage || ev.IsCombatDamage || src == nil {
					return false
				}
				if ev.DamageAmount <= 0 || ev.DamageAmount >= src.CurrentPower() {
					return false
				}
				return ojerRedSourceControlledBy(ev, g, src.Controller) &&
					ojerDamageHitsAnOpponentOf(ev, g, src.Controller)
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) error {
				ev.DamageAmount = src.CurrentPower()
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Ojer Axonil, Deepest Might: damage equal to its power",
		}},
	})
}

// ojerRedSourceControlledBy is "a red source you control". Narrower
// than Mechanized Warfare's red-OR-ARTIFACT reader on purpose: a
// colourless artifact pinging an opponent is not raised by Ojer.
//
// The source is the damage event's snapshot of it
// (damageSourceCharacteristics): a burn spell as it stands on the
// stack, and a creature that has left the battlefield as it last
// existed there (#1417, CR 608.2h), so a creature an effect had turned
// red still counts after it dies, and the colour is the effective one.
func ojerRedSourceControlledBy(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID) bool {
	return damageSourceIsRedControlledBy(ev, g, controller, false)
}

// ojerDamageHitsAnOpponentOf is "to an opponent" — a seated,
// non-eliminated PLAYER other than `controller`. Damage aimed at a
// permanent is not raised, whoever controls it, which is the
// difference between this clause and Mechanized Warfare's wider "an
// opponent or a permanent an opponent controls".
func ojerDamageHitsAnOpponentOf(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID) bool {
	if ev.DamageTarget == uuid.Nil || ev.DamageTarget == controller {
		return false
	}
	p := g.PlayerByIDForEffect(ev.DamageTarget)
	return p != nil && !p.Eliminated
}

// Temple of Power — Land, the back face of Ojer Axonil, Deepest Might:
//
//	"(Transforms from Ojer Axonil, Deepest Might.)
//	 {T}: Add {R}.
//	 {2}{R}, {T}: Transform this land. Activate only if red sources you
//	 controlled dealt 4 or more noncombat damage this turn and only as
//	 a sorcery."
//
// Only the mana ability is registered; see the declared simplification
// on the front face.
func init() {
	Register(Spec{
		OracleID:     ojerAxonilOracleID + "#1",
		Name:         "Temple of Power",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The {2}{R}, {T} ability to transform back into Ojer Axonil isn't implemented — the land only taps for {R}.",
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "{T}: Add {R}",
		}},
	})
}

const ojerAxonilOracleID = "d3b7b541-6f05-46c1-8031-c848c4bd4635"
