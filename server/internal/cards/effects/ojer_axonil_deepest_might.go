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
//   - "a red SOURCE YOU CONTROL" — the source is looked up wherever
//     it now is, since a burn spell is already in a graveyard by the
//     time its damage resolves, and its colour is read post-layers.
//     A source the engine cannot find is left alone, which errs
//     weaker and never stronger.
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
// ONE DECLARED SIMPLIFICATION: the dies trigger is not implemented.
// Nothing in the engine turns a permanent to its other face on the
// battlefield (the transform seam, #343), so the honest options were
// to return Ojer as itself — a recurring 4/4 God, strictly STRONGER
// than the land it is printed to come back as, which is the #259
// rule — or to leave the trigger out and say so. It is left out:
// Ojer dies like any other creature and stays in the graveyard, and
// the Temple of Power face is unreachable, so its mana ability and
// its transform-back ability are not registered either. The trigger
// and the back face land together when the transform verb does.
func init() {
	Register(Spec{
		OracleID:        "d3b7b541-6f05-46c1-8031-c848c4bd4635",
		Name:            "Ojer Axonil, Deepest Might",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"trample"},
		Caveats: []string{
			"When Ojer Axonil dies it stays in the graveyard — coming back tapped as Temple of Power isn't implemented yet, so the land half of the card can't be reached.",
		},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventDealDamage},
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
// The source is looked up wherever it has landed, so a spell already
// in its owner's graveyard still answers for its colour, and the
// colour is the effective one — a creature an effect has turned red
// counts.
func ojerRedSourceControlledBy(ev *game.ReplacementEvent, g *game.Game, controller uuid.UUID) bool {
	if ev.DamageSource == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.DamageSource)
	return ok && c.Controller == controller && c.HasColor("R")
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
