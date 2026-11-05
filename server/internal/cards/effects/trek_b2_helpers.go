package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b2_helpers.go — shared shapes for Star Trek batch 2 (#4973).

// trekB2BeamMeUp is "Beam me up [cost] (You may cast this card from your
// graveyard for [cost] if you also return a creature you control to its
// owner's hand. Then exile this spell.)": flashback's graveyard cast and
// its exile on leaving the stack (Flashback), with the return of a
// creature you control as a second component of the same price, the
// ReturnToHand component Daze and web-slinging pay. The return is a
// cost, so it happens as the spell is cast and the creature stays in
// hand if the spell is countered; with no creature to return the cast
// is refused before anything is paid. The card file opens the graveyard
// in CastableZones, as for flashback.
func trekB2BeamMeUp(cost string) game.AlternativeCost {
	ac := Flashback(cost)
	ac.Key = "beam_me_up"
	ac.Label = "Beam me up " + cost
	ac.ReturnToHand = PermanentYouControl("a creature you control", Creature())
	ac.PayLabel = "a creature you control"
	return ac
}

// trekB2ControlsAPermanentWithACounter is "if you control a permanent
// with a counter on it": any kind of counter, read off the battlefield
// as the instruction is performed.
func trekB2ControlsAPermanentWithACounter(g *game.Game, player uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player {
			continue
		}
		for _, n := range c.Counters {
			if n > 0 {
				return true
			}
		}
	}
	return false
}

// trekB2DoublesYourLifeGain is Dr. Beverly Crusher's CR 603.2d doubler:
// "If you gaining life causes a triggered ability of a permanent you
// control to trigger, that ability triggers an additional time." The
// cause is an EventChangeLife that raised the doubler's controller's
// life; the doubled ability must be printed on (or granted to) a
// permanent that player controls, so a spell's, an emblem's (CR 114.4)
// and a delayed trigger's ability (which watches nothing, CR 603.7) are
// left alone.
func trekB2DoublesYourLifeGain(label string) game.TriggerDoubler {
	return game.TriggerDoubler{
		Label:    label,
		RowLabel: "Abilities that trigger when you gain life trigger an additional time.",
		Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || q.Ability == nil || len(q.Ability.Watches) == 0 {
				return false
			}
			ev := q.Event
			if ev.Kind != game.EventChangeLife || ev.Amount <= 0 || ev.Target != q.DoublerLKI.Controller {
				return false
			}
			return q.SourceLKI.Controller == q.DoublerLKI.Controller && onBattlefield(g, q.Source.InstanceID)
		},
	}
}

// trekB2CountersInsteadOfDestruction is Dr. McCoy's replacement: "If
// another creature you control with a +1/+1 counter on it would be
// destroyed, instead remove all +1/+1 counters from it and all damage
// already dealt to it is healed." A destruction (ReplacementEvent.
// Destruction, an effect's or a state-based action's) of another creature
// the source's controller controls that has at least one +1/+1 counter is
// cancelled; the counters come off and the damage is healed. The affected
// creature's controller orders it against other replacements (CR 616.1).
func trekB2CountersInsteadOfDestruction(label string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || !ev.Destruction || ev.OldZone != game.ZoneBattlefield || src == nil || ev.CardID == src.InstanceID {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && onBattlefield(g, ev.CardID) && c.IsCreature() && c.Controller == src.Controller && c.Counters[game.CounterPlusOne] > 0
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			c, ok := g.LookupCardForEffect(ev.CardID)
			if !ok {
				return nil
			}
			ev.Cancel()
			if n := c.Counters[game.CounterPlusOne]; n > 0 {
				if err := g.AddCounterForEffect(ev.CardID, game.CounterPlusOne, -n); err != nil {
					return err
				}
			}
			g.HealAllDamageForEffect(ev.CardID)
			return nil
		},
		Controller: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) uuid.UUID {
			return g.AffectedPermanentControllerForEffect(ev)
		},
		Label: label,
	}
}
