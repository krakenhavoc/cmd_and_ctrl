package heuristic

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mana_sink.go is #4559: a repeatable mana sink is a use of what the
// turn leaves over, and a graveyard exile is worth what it takes away.
//
// A mana sink is an activated ability of the bot's own permanent whose
// only cost is mana: no {T}, nothing sacrificed, discarded, exiled or
// paid in life, counters, loyalty or crew. Nothing but mana stops it
// being activated again, so each activation spends the mana a spell
// would use. Lion Sash's "{W}: Exile target card from a graveyard" is the
// card it was found on; Scavenging Ooze, firebreathing, a shade's pump
// and "{2}: draw a card" have the same shape. Equip and reconfigure are
// not sinks: they move an attachment, and equip_move.go prices them.
//
// Priced one window at a time, a sink competed with the turn's spells on
// equal terms, so anything it scored above a cheap creature won. The bot
// spent its white mana exiling graveyard cards one at a time and never
// cast the creature the mana was for. With Config.SinkManaLast a sink is
// taken only when no other move clears its bar (decideGeneral): every
// cast, land, held instant and other activation the window would make
// goes first, and the sink spends what they leave. A sink is also left
// out of the best single move the turn plan must beat (ADR 0136 §3), so
// a plan of casts is not turned down for it. Its own price and bar are
// unchanged: in the bot's main phase it still clears PassThreshold,
// elsewhere InstantThreshold or the combat and save rules, and in a
// leftover window (ADR 0126 §5) LeftoverThreshold, which is where an
// instant-speed sink spends mana the bot kept up through the round.
//
// The bot's own upkeep and draw step, stack empty, are not a spend window
// for a sink (waitmain.go's premise, the owner's gate on casts carried
// over to sinks under this flag): a land tapped there stays tapped
// through both main phases, and the same activation is on offer in the
// main phase, where it goes after the spells. An exile worth more than
// InstantThreshold would otherwise eat the turn's mana before the main
// phase began.
//
// With Config.PriceGraveyardExile, a pick the row declares it exiles
// from a graveyard (TargetPurpose.Exiles) is priced by what the exile
// takes away (graveyardExileValue) rather than by cardTargetValue, which
// reads a graveyard pick as recursion: the bot's own card as one it could
// use, and an opponent's at a flat OwnPermanentTarget whatever it was.

// manaSink reports whether m activates a repeatable mana sink of the
// bot's own (see above).
func (p *Policy) manaSink(st *state, m legal.Move) bool {
	if m.Kind != legal.KindActivate || !p.costsOnlyManaAndTaps(st, m) {
		return false
	}
	ap := decode[activateParams](m.Params)
	if len(ap.CrewIDs) > 0 || len(ap.WaterbendIDs) > 0 || len(ap.TapIDs) > 0 {
		return false
	}
	src := st.bf[ap.SourceCardID]
	if src == nil || src.Controller != st.me {
		return false
	}
	row := activatedRow(src, ap.AbilityIndex)
	if row == nil || row.TapCost || row.Equip || row.Reconfigure || row.Saddle || row.Exert ||
		row.CrewCost > 0 || row.EnergyCost > 0 || row.EnergyCostX {
		return false
	}
	if row.ManaCost == "" || (manaValue(row.ManaCost, 0) <= 0 && ap.XValue <= 0) {
		return false
	}
	// #4980: a run to a resolution that returns more mana than it costs
	// makes mana for the casts rather than spending what they leave
	// (resolution_mana.go).
	if _, ok := p.resolutionManaValue(st, m); ok {
		return false
	}
	return true
}

// graveyardExileValue prices a pick its entry declares it exiles from a
// graveyard (#4559):
//
//   - a card of the bot's own costs fuelValue, what a cost that exiled it
//     would charge: FuelIdle for a card it cannot cast from there (the
//     graveyard synergies the policy cannot see), more for one it can;
//   - an opponent's card denies them FuelIdle, or FuelFloor for a land,
//     the same unseen synergy from their side, scaled by leaderBoost. The
//     bot cannot see what that player can recast, so this is a floor;
//   - an exiled card of the type ExileGrows names puts a +1/+1 counter on
//     the source while it is on the battlefield, at what the evaluation
//     counts a point of power and toughness.
//
// So exiling an opponent's instant with Lion Sash buys almost nothing,
// and its land or creature buys the counter. False when the entry does
// not exile or the pick is not a graveyard card the view shows.
func (p *Policy) graveyardExileValue(st *state, t targetRef, e *protocol.TargetPurposeView, source *protocol.CardView) (float64, bool) {
	if !p.cfg.PriceGraveyardExile || e == nil || !e.Exiles || t.Kind != "card" {
		return 0, false
	}
	c := st.graveyard[t.ID]
	if c == nil {
		return 0, false
	}
	var v float64
	if c.Owner == st.me {
		v = -p.fuelValue(st, t.ID)
	} else {
		deny := p.cfg.FuelIdle
		if isLand(c) {
			deny = p.cfg.FuelFloor
		}
		v = deny * st.leaderBoost(p.cfg, c.Owner)
	}
	if source != nil && st.bf[source.InstanceID] != nil && exileGrowsOn(e.ExileGrows, c) {
		v += st.w.Power + st.w.Toughness
	}
	return v, true
}

// exileGrowsOn reports whether exiling c puts the counter ExileGrows
// promises on the source.
func exileGrowsOn(grows string, c *protocol.CardView) bool {
	switch grows {
	case "permanent":
		// The type line read as game.Card.IsPermanent reads it.
		return isPermanentSpell(c)
	case "creature":
		return isCreature(c)
	}
	return false
}
