package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// disturb_batch_c_helpers.go — the pieces the harder disturb cards
// (ADR 0107 §4, #1855) share or need by name: Dorothea's two
// end-of-combat sacrifices, the "Spirit in addition to its other
// types" copy clause Mirrorhall Mimic's two faces print, Katilda's
// count, Brine Comber's "becomes the target of an Aura spell", and
// Covert Cutpurse's "was dealt damage this turn".

// dorotheaSacrificeBody is the delayed trigger "sacrifice it at end of
// combat" schedules (Dorothea, Vengeful Victim): its source, if that
// is still the same object on the battlefield. A registered key (ADR
// 0041 phase 3), so a table with it waiting is a restore point.
//
// Assigned in init: a var initialiser would be an initialisation cycle
// through the exit primitives, the shape delayed_bodies.go's own init
// avoids.
var dorotheaSacrificeBody game.BodyRef

func init() {
	dorotheaSacrificeBody = game.SimpleDelayedBody("dorothea/sacrifice-at-end-of-combat", sacrificeDelayedTriggerSource)
}

// sacrificeDelayedTriggerSource sacrifices the delayed trigger's source
// (CR 603.7d: the source of the ability that created it). Nothing
// happens if that permanent left the battlefield and came back as a
// new object meanwhile (CR 400.7), or if its controller is no longer
// the trigger's controller — a player can sacrifice only a permanent
// they control (CR 701.21a).
func sacrificeDelayedTriggerSource(g *game.Game, item *game.StackItem) error {
	if g.AbilitySourceGoneForEffect(item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || c.Controller != item.Controller {
		return nil
	}
	return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// sacrificeThisAtEndOfCombat is the trigger body of "sacrifice it at
// end of combat": the beginning of this combat's end of combat step
// (CR 511.2), through a delayed trigger whose source is this object.
func sacrificeThisAtEndOfCombat(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return ScheduleDelayedTrigger{
			At:    game.StepEndCombat,
			Label: label,
			Body:  dorotheaSacrificeBody,
		}.Apply(NewContext(g, item))
	}
}

// dorotheasRetributionSpirit is the ability Dorothea's Retribution
// grants: a 4/4 white Spirit with flying, tapped and attacking the
// player the enchanted creature was declared against (carried on
// item.Params.Player), sacrificed at end of combat.
func dorotheasRetributionSpirit(g *game.Game, item *game.StackItem) error {
	tmpl := TokenCard("4/4 white Spirit with flying")
	tmpl.Tapped = true
	cursor := b25LastEventSeq(g)
	if err := g.CreateTokensAttackingForEffect(item.Controller, tmpl, 1, item.Params.Player); err != nil {
		return err
	}
	tokens := b27TokensCreatedByAfter(g, item.Controller, cursor)
	if len(tokens) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:    game.StepEndCombat,
		Label: "Dorothea's Retribution — sacrifice the Spirit",
		Cards: tokens,
		Body:  sacrificeListedCardsBody,
	}.Apply(NewContext(g, item))
}

// withSubtypeInAddition is "except it's a <subtype> in addition to its
// other types" on a type line (CR 707.9a): the subtype is added, and
// every type, supertype and subtype the line had stays. A line that
// already has it comes back unchanged.
func withSubtypeInAddition(printed, subtype string) string {
	super, types, subs := game.ParseTypeLine(printed)
	for _, s := range subs {
		if strings.EqualFold(s, subtype) {
			return printed
		}
	}
	head := make([]string, 0, len(super)+len(types))
	head = append(head, super...)
	head = append(head, types...)
	return strings.Join(head, " ") + " — " + strings.Join(append(append([]string(nil), subs...), subtype), " ")
}

// spiritsAndEnchantmentsYouControl counts "permanents you control that
// are Spirits and/or enchantments" — Katilda's X. A permanent that is
// both counts once.
func spiritsAndEnchantmentsYouControl(g *game.Game, source *game.Card) int {
	return permanentsYouControl(g, source, func(c *game.Card) bool {
		return c.HasSubtype("Spirit") || c.IsEnchantment()
	})
}

// targetedByAnAuraSpell reports whether the event is `target` becoming
// the target of an Aura SPELL — not an ability, and not a spell of any
// other kind (Brine Comber). The spell is read off the stack, where it
// still is: the event fires as targets are chosen (CR 601.2c).
func targetedByAnAuraSpell(ev game.Event, g *game.Game, target uuid.UUID) bool {
	if ev.Kind != game.EventBecomesTarget || target == uuid.Nil || ev.CardID != target {
		return false
	}
	if !b40TargetedByASpell(ev, g) {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.Source)
	return ok && spell.IsAura()
}

// DealtDamageThisTurn passes for a permanent that was dealt damage
// this turn — Covert Cutpurse's "target creature you don't control
// that was dealt damage this turn".
//
// Read off the turn's events: a damage event with an amount names the
// damage actually dealt, after prevention. CR 400.7: a permanent that
// entered the battlefield after the damage is a new object that was
// never dealt it, so an entry wipes the record.
func DealtDamageThisTurn() CardPredicate {
	return func(g *game.Game, _ uuid.UUID, c game.Card) bool {
		dealt := false
		for _, ev := range g.EventsThisTurn() {
			switch {
			case ev.Kind == game.EventZoneMove && ev.CardID == c.InstanceID && ev.NewZone == game.ZoneBattlefield:
				dealt = false
			case ev.Kind == game.EventDealDamage && ev.Target == c.InstanceID && ev.Amount > 0:
				dealt = true
			}
		}
		return dealt
	}
}
