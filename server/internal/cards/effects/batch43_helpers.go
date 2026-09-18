package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch43_helpers.go — the shared bodies behind the card-coverage
// roadmap's batch 43 (#450, `edhrec_rank` 4456–4555). Own file per
// the #231 convention; every package-level name carries the b43
// prefix because other batches land beside this one.
//
// What is NOT here, because the package already had it: "this
// permanent enters" is b06SelfETB, "another creature you control
// entered" is b13AnotherCreatureYouControlEntered, "a creature you
// control attacks" is attackDeclaredByYou with the "one or more"
// dedup OncePerBatch, "this creature attacks" is attackDeclared, "a
// creature you control dealt combat damage to a player" is
// combatDamageToPlayerBy, "at the beginning of your upkeep" is
// AtYourUpkeep, "at the beginning of combat on your turn" is
// AtBeginningOfYourCombat, the tutor-to-hand body is b06TutorToHand,
// "destroy the chosen permanent" is destroyChosenPermanent, the
// two-colour tap is dualManaAbility, enters-tapped is
// SelfEntersTapped, "you control your commander" is
// b18ControlsYourCommander, the tribe count is
// b22OtherCreaturesOfSubtypeControlled, the creature count is
// b04CreaturesControlled, and the Panorama cycle table is
// b43PanoramaFetch in naya_panorama.go (a cycle table belongs with
// its cycle, not in this file).

// --- counting ------------------------------------------------------

// b43CreaturesOfSubtypeControlled counts every creature `controller`
// controls with the given subtype, the source INCLUDED — Shaman of
// the Pack's "the number of Elves you control", which counts the
// Shaman itself because the Shaman is an Elf.
//
// Post-layer subtype, read through the battlefield snapshot, so a
// changeling counts and a creature that is only an Elf through a
// type-changing effect counts too.
func b43CreaturesOfSubtypeControlled(g *game.Game, controller uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

// b43TotalPowerControlled sums the current power of every creature
// `controller` controls — Finneas's "if creatures you control have
// total power 10 or greater" and Gimli's Reckless Might's formidable
// check.
//
// CurrentPower, not printed power: counters, anthems and pumps all
// count, which is what "have total power" means (CR 208.3). Negative
// power is clamped at zero by CurrentPower, so a creature shrunk
// below zero contributes nothing rather than subtracting.
func b43TotalPowerControlled(g *game.Game, controller uuid.UUID) int {
	total := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() {
			total += c.CurrentPower()
		}
	}
	return total
}

// --- trigger conditions --------------------------------------------

// b43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer is Seafloor
// Oracle's "whenever a Merfolk you control deals combat damage to a
// player": combatDamageToPlayerBy narrowed to one creature subtype,
// read off the damage SOURCE.
//
// The source is looked up live rather than through LKI, so a creature
// that dies in first-strike damage and deals its damage in the
// regular step is still the creature it was. A source that has
// already left the battlefield by the time the event is harvested
// fails the lookup and the trigger does not fire — weaker than
// printed, never stronger.
func b43CreatureOfSubtypeYouControlDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game, subtype string) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.HasSubtype(subtype)
}

// b43ALandAnOpponentControlsEntered is Nightshade Harvester's
// "whenever a land an opponent controls enters": the entering
// permanent is a land and its controller is anyone other than the
// source's controller.
//
// Returns the land's controller alongside the verdict, because the
// trigger has to drain THAT player and not an arbitrary opponent.
func b43ALandAnOpponentControlsEntered(ev game.Event, source *game.Card, g *game.Game) (uuid.UUID, bool) {
	if ev.Kind != game.EventETB {
		return uuid.Nil, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsLand() || c.Controller == source.Controller {
		return uuid.Nil, false
	}
	if p := g.PlayerByIDForEffect(c.Controller); p == nil {
		return uuid.Nil, false
	}
	return c.Controller, true
}

// b43AnotherHumanEnteredThisTurn is Éowyn, Shieldmaiden's
// intervening-if: a Human other than the source entered the
// battlefield under the source's controller's control this turn.
//
// The tally counts permanents by the subtype they HAD as they
// entered, so a Human that has since died still counts and a
// changeling counts. "Another" is the reason for the second term:
// Éowyn is herself a Human, so when she is one of this turn's
// arrivals the tally has to reach two before the clause is
// satisfied.
func b43AnotherHumanEnteredThisTurn(g *game.Game, source *game.Card) bool {
	need := 1
	if b06EnteredThisTurn(g, source.InstanceID) {
		need = 2
	}
	return g.EnteredWithSubtypeThisTurn(source.Controller, "Human") >= need
}

// b43NontokenMerfolkYouControlBecameTapped is Deeproot Pilgrimage's
// condition — b32CreatureYouControlBecameTapped narrowed to nontoken
// Merfolk. Two event kinds feed it for the reason that helper gives:
// the engine taps an attacker without an EventTapCard, so an
// EventAttack whose creature is now tapped is "became tapped" and a
// vigilance attacker is not.
func b43NontokenMerfolkYouControlBecameTapped(ev game.Event, source *game.Card, g *game.Game) bool {
	if !b32CreatureYouControlBecameTapped(ev, source, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !IsToken(c) && c.HasSubtype("Merfolk")
}

// b43YouWereDealtDamage is Darien, King of Kjeldor's "whenever you're
// dealt damage": any damage event, combat or not, from any source,
// whose target is the source's controller. Returns the amount, which
// is the token count.
//
// "You're dealt damage" is not "you lose life": a Vizkopa drain or a
// paid life cost is life loss, not damage, and Darien sees neither
// (CR 119.3c). Damage prevented or replaced never happens, so a
// prevention shield really does stop the tokens as well as the
// damage — the event is emitted with the amount that was APPLIED.
func b43YouWereDealtDamage(ev game.Event, source *game.Card, g *game.Game) (int, bool) {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Target != source.Controller {
		return 0, false
	}
	if p := g.PlayerByIDForEffect(ev.Target); p == nil {
		return 0, false
	}
	return ev.Amount, true
}

// --- conditional attachment statics --------------------------------

// b43AttachedHasColor is Shield of the Oversoul's "as long as
// enchanted creature is green / is white": AttachedToSource narrowed
// by the target's POST-LAYER colour.
//
// EffectiveColors reads the characteristic the recompute pass is
// building, and layer 5 (colour) runs before layer 6 (abilities) and
// layer 7 (P/T), so by the time either of the statics below asks, a
// colour-changing effect has already been applied. That ordering is
// what lets the two halves be plain conditional statics rather than a
// CR 613.8 dependency.
func b43AttachedHasColor(color string) func(target *game.Card, g *game.Game, source *game.Card) bool {
	return func(target *game.Card, g *game.Game, source *game.Card) bool {
		if !AttachedToSource(target, g, source) {
			return false
		}
		for _, c := range target.EffectiveColors() {
			if c == color {
				return true
			}
		}
		return false
	}
}

// b43PumpAttachedWhen is PumpAttached under a condition — layer 7c,
// so it composes additively with anthems and with the card's other
// conditional pump.
func b43PumpAttachedWhen(when func(target *game.Card, g *game.Game, source *game.Card) bool, power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: when,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// b43GrantToAttachedWhen is GrantToAttached under a condition — layer
// 6. Keywords must be canonical lowercase tokens the engine honours;
// both of Shield of the Oversoul's are.
func b43GrantToAttachedWhen(when func(target *game.Card, g *game.Game, source *game.Card) bool, keywords ...string) game.StaticAbility {
	granted := append([]string(nil), keywords...)
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: when,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, kw := range granted {
				if !keywordSliceContains(c.Abilities, kw) {
					c.Abilities = append(c.Abilities, kw)
				}
			}
		},
	}
}

// --- attachments ---------------------------------------------------

// b43AttachedCreatureDealtCombatDamage is Lost Jitte's trigger
// condition: "whenever equipped creature deals combat damage" — to
// ANYTHING.
//
// The difference from attachedCreatureDealtCombatDamageToPlayer (the
// Swords' condition) is the whole point of the separate helper: a
// blocked attacker that kills its blocker banks a Jitte counter, and
// a Sword of Fire and Ice on the same creature does nothing. Both
// halves of that are printed.
func b43AttachedCreatureDealtCombatDamage(ev game.Event, source *game.Card, _ *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	return source.IsAttachedTo(ev.Source)
}

// b43AttachedCreature is "equipped creature" read off the Equipment —
// the instance it is attached to right now, or false when it is
// attached to nothing.
//
// Read live rather than cached: an Equipment that came unattached
// between the activation and the resolution equips nothing, and the
// bullet that would have grown its host does nothing at all, which is
// what the printed card does.
func b43AttachedCreature(g *game.Game, equipment uuid.UUID) (uuid.UUID, bool) {
	c, ok := g.LookupCardForEffect(equipment)
	if !ok || c.AttachedTo.Kind != game.TargetCard || c.AttachedTo.ID == uuid.Nil {
		return uuid.Nil, false
	}
	host, ok := g.LookupCardForEffect(c.AttachedTo.ID)
	if !ok || !host.IsCreature() {
		return uuid.Nil, false
	}
	return host.InstanceID, true
}

// --- reads ---------------------------------------------------------

// b43PowerNowOrLastKnown is "equal to its power" read at resolution by
// a trigger whose source may already be gone — Caldera Pyremaw.
//
// On the battlefield it is CurrentPower, the post-layer answer, so
// anthems, pumps and counters all count. Off the battlefield it is
// CR 608.2h last known information: the printed power plus the
// +1/+1 counters it had as it left, minus the -1/-1 counters, which
// is b17LastKnownPowerOffBattlefield. Killing the source in response
// is therefore not a way to fog the effect, which is what the rule
// says and what a player expects.
func b43PowerNowOrLastKnown(g *game.Game, cardID uuid.UUID) int {
	if z := g.FindCardZoneForEffect(cardID); z != nil && z.Kind == game.ZoneBattlefield {
		if c, ok := g.LookupCardForEffect(cardID); ok {
			return c.CurrentPower()
		}
	}
	return b17LastKnownPowerOffBattlefield(g, cardID)
}

// --- effect bodies -------------------------------------------------

// b43TargetPlayerLoses is "target opponent loses N life" for a trigger
// whose target clause is a player, with N computed at resolution.
// A target that is no longer legal is skipped.
func b43TargetPlayerLoses(g *game.Game, item *game.StackItem, n int) error {
	if n <= 0 {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, t.ID, -n)
		}
	}
	return nil
}
