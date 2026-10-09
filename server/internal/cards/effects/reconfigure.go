package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reconfigure.go — the card side of CR 702.151, reconfigure (#2639).
// The rules side is game/reconfigure.go.
//
//	Activated: Reconfigure("{2}"),                       // Lizard Blades
//	Activated: append([]ActivatedAbility{lionSashExile}, Reconfigure("{2}")...),
//	Activated: ReconfigureOneOf("Reconfigure—Pay {2} or {E}{E}{E}",
//	    ReconfigureOption{Pay: "{2}", Cost: ManaCost("{2}")},
//	    ReconfigureOption{Pay: "{E}{E}{E}", Cost: PayEnergy(3)}),  // Razorfield Ripper
//
// "Reconfigure [cost]" is two activated abilities (CR 702.151a), so it
// is two rows: attach to another target creature you control, and
// unattach, which may be activated only while this permanent is
// attached to a creature. Both are sorcery speed, both are marked
// Reconfigure, and neither is marked Equip. A cost printed with an
// alternative ("Pay {2} or {E}{E}{E}") is one pair of rows per way to
// pay, the shape the issue asked for and the one Inventor's Axe's
// energy equip already uses for its single cost.
//
// The labels are the printed keyword line, so the oracle check finds
// them, followed by a parenthesis saying which half (and, for an
// alternative, which payment) the row is. The parenthesis is what keeps
// the labels distinct: a restore point and the activation tally both
// name a row by its label.
//
// What the engine does for you: an attached reconfigure Equipment is
// not a creature (CR 702.151b), and an Equipment that is a creature
// stays attached only if it has reconfigure (CR 301.5c). Card files
// write their "Equipped creature …" statics exactly as any Equipment
// does (PumpAttached, GrantToAttached).

// ReconfigureOption is one way to pay a reconfigure cost printed with
// alternatives: Pay is the printed cost ("{2}", "{E}{E}{E}") and Cost
// is the component that pays it.
type ReconfigureOption struct {
	Pay  string
	Cost game.AbilityCost
}

// Reconfigure is "Reconfigure {cost}" with a mana cost: the attach
// and the unattach rows.
func Reconfigure(cost string) []ActivatedAbility {
	printed := "Reconfigure " + cost
	return reconfigurePair(printed+" (attach)", printed+" (unattach)", ManaCost(cost))
}

// ReconfigureOneOf is a reconfigure keyword printed with a choice of
// costs, "Reconfigure—Pay {2} or {E}{E}{E}": `printed` is that line, and
// each option adds an attach row and an unattach row.
func ReconfigureOneOf(printed string, options ...ReconfigureOption) []ActivatedAbility {
	out := make([]ActivatedAbility, 0, 2*len(options))
	for _, o := range options {
		out = append(out, reconfigurePair(
			printed+" (pay "+o.Pay+", attach)",
			printed+" (pay "+o.Pay+", unattach)",
			o.Cost)...)
	}
	return out
}

// reconfigurePair builds CR 702.151a's two abilities at one cost.
func reconfigurePair(attachLabel, unattachLabel string, cost game.AbilityCost) []ActivatedAbility {
	return []ActivatedAbility{
		{
			// "Attach this permanent to another target creature you
			// control. Activate only as a sorcery." "Another" is the
			// object itself (Another), which is the Equipment while it
			// is an unattached creature.
			Label:        attachLabel,
			Cost:         cost,
			Targets:      Another(TargetCreature("another target creature you control", YouControl())),
			SorcerySpeed: true,
			Reconfigure:  true,
			Effect:       AttachSourceToTarget,
		},
		{
			// "Unattach this permanent. Activate only if this permanent
			// is attached to a creature and only as a sorcery."
			Label:        unattachLabel,
			Cost:         cost,
			SorcerySpeed: true,
			Reconfigure:  true,
			Condition:    SourceAttachedToACreature(),
			Effect:       UnattachSource,
		},
	}
}

// SourceAttachedToACreature — "Activate only if this permanent is
// attached to a creature" (CR 702.151a). The host must be on the
// battlefield and a creature now.
func SourceAttachedToACreature() ActivationCondition {
	return func(g *game.Game, _, source uuid.UUID) bool {
		return g.AttachedToACreatureForEffect(source)
	}
}

// UnattachSource is the unattach ability's resolution: "Unattach this
// permanent." The source has to still be the same object
// (game.UnattachSourceForEffect).
func UnattachSource(g *game.Game, item *game.StackItem) error {
	return g.UnattachSourceForEffect(item)
}

// ThisOrEquippedCreature reports whether `id` is the source itself or
// the creature it is attached to — the subject of "Whenever this
// creature or equipped creature …", which every reconfigure card with
// a trigger prints. An attached reconfigure Equipment is not a
// creature, so at most one of the two can be attacking or dealing
// combat damage at a time.
func ThisOrEquippedCreature(id uuid.UUID, source *game.Card) bool {
	return id != uuid.Nil && (id == source.InstanceID || source.IsAttachedTo(id))
}

// ThisOrEquippedCreatureAttacks is "Whenever this creature or equipped
// creature attacks". The attacker is the event's CardID, which the
// effect reads off item.Trigger.Event.CardID.
func ThisOrEquippedCreatureAttacks(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventAttack && ThisOrEquippedCreature(ev.CardID, source)
}

// ThisOrEquippedCreatureDealsCombatDamageToAPlayer is "Whenever this
// creature or equipped creature deals combat damage to a player". The
// player is the event's Target.
func ThisOrEquippedCreatureDealsCombatDamageToAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	return ThisOrEquippedCreature(ev.Source, source) && g.PlayerByIDForEffect(ev.Target) != nil
}

// ThisOrEquippedCreatureBecomesBlocked is "Whenever this creature or
// equipped creature becomes blocked": once per blocked attacker
// (game.EventBecomesBlocked), the attacker being the event's CardID and
// the defending player its Actor.
func ThisOrEquippedCreatureBecomesBlocked(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventBecomesBlocked && ThisOrEquippedCreature(ev.CardID, source)
}

// pumpTriggeringCreatureUntilEOT is "it gets +P/+T until end of turn"
// for the creature a "this creature or equipped creature" trigger was
// about: the triggering permanent, if it is still that object on the
// battlefield (CR 400.7). One that has left gets nothing.
func pumpTriggeringCreatureUntilEOT(ctx *Context, power, toughness int, label string) error {
	info, ok := ctx.TriggeringPermanent()
	if !ok || info.Left {
		return nil
	}
	return BoostUntilEOT{
		Target:    ctx.Trigger().Event.CardID,
		Power:     power,
		Toughness: toughness,
		Label:     label,
	}.Apply(ctx)
}

// equippedCreatureIs is one layer's part of a change to the equipped
// creature that is ONE continuous effect with a part in a later layer
// (Blade of the Oni's "has base power and toughness 5/5, has menace,
// and is a black Demon"): CR 613.6 keeps it applying in the later
// layers once it has started, whatever removes the source's abilities
// part-way through the pass.
func equippedCreatureIs(layer game.Layer, apply func(c *game.Characteristic)) game.StaticAbility {
	return game.StaticAbility{
		Layer:                 layer,
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			apply(c)
		},
	}
}
