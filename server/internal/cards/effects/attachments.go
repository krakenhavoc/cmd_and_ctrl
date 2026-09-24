package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachments.go is the shared catalog surface for S24's Equipment
// and Auras (ADR 0036). Everything an equipment or aura card file
// needs that is not the card itself lives here:
//
//	EquipAbility("{2}")           the CR 702.6 activated ability
//	AttachedToSource              "equipped creature" / "enchanted creature"
//	PumpAttached(2, 2)            a layer 7c static on the host
//	GrantToAttached("haste")      a layer 6 keyword grant on the host
//	EnchantCreature(...)          the aura's enchant clause
//
// The deliberate absence here is a `Spec.Equip` or `Spec.Enchant`
// field. Equip is an ordinary `Spec.Activated` entry, because it IS
// an ordinary activated ability — a mana cost, a target clause, a
// sorcery-speed gate — and every one of those already exists and is
// already enumerated by the bot, rendered by the right-click menu
// and validated at announce. An aura's enchant clause is an ordinary
// `Spec.Targets`. Neither needed a new engine verb.
//
// That is still true after #1208, which gave the ABILITY a name for
// its keyword (`ActivatedAbility.Equip`). It is not a kind and it
// changes nothing about how an equip is validated, paid or
// resolved; it exists so that a card which speaks about equip
// abilities — Leonin Shikari — can pick them out of a permanent's
// list. Set by the two constructors below and by no card file.

// EquipAbility builds the CR 702.6 equip ability: "{cost}: Attach to
// target creature you control. Equip only as a sorcery."
//
// Three clauses, all of them load-bearing and none of them new:
//
//   - SorcerySpeed is CR 702.6a's "any time you could cast a
//     sorcery", gated by the engine's existing main-phase /
//     empty-stack / active-player check.
//   - The target clause is "creature you control", validated at
//     announce (CR 601.2c) and re-checked at resolution (CR 608.2b),
//     so an equip whose target dies in response fizzles with no card
//     code. Note the restriction is checked only WHILE ACTIVATING:
//     once attached, control of the creature may change and the
//     Equipment stays put (see attachmentLegalLocked in the game
//     package for the other half of that rule).
//   - Re-activating moves the Equipment (CR 701.3a), which the
//     attach primitive gets right by overwriting.
func EquipAbility(cost string) ActivatedAbility {
	return ActivatedAbility{
		Label:        "Equip " + cost,
		Cost:         ManaCost(cost),
		Targets:      TargetCreature("target creature you control", YouControl()),
		SorcerySpeed: true,
		// #1208: the keyword, named. Equip is still an ordinary
		// activated ability — nothing above this line changed — but
		// Leonin Shikari's "you may activate EQUIP abilities any
		// time you could cast an instant" has to be able to pick
		// them out, and matching on the label's spelling would be a
		// rule written in string literals.
		Equip:  true,
		Effect: AttachSourceToTarget,
	}
}

// EquipOnlyAbility is EquipAbility with a NARROWED target clause and
// the card's own wording — "Equip Halfling {1}" (Bilbo's Ring),
// "Equip legendary creature {3}" (Blackblade Reforged, Excalibur,
// Sword of Eden).
//
// It exists because those three were written out by hand, and a
// hand-written equip is one that can forget a field. It forgot one
// the day #1208 added `Equip`: Leonin Shikari's "you may activate
// equip abilities any time you could cast an instant" would have
// opened Blackblade's {7} and not its {3}, which is a bug nobody
// would find by reading either card. TestEveryEquipAbilityIsMarked
// is the other half of that guarantee.
func EquipOnlyAbility(label, cost string, targets *game.TargetSpec) ActivatedAbility {
	ab := EquipAbility(cost)
	ab.Label = label
	ab.Targets = targets
	return ab
}

// AttachSourceToTarget is the equip ability's resolution: attach the
// source permanent to the chosen creature.
//
// Reads the source off the StackItem by ID rather than capturing a
// *Card, which is the standing rule for anything that resolves off
// the stack — the battlefield slice may have been reallocated since
// the ability was announced.
//
// Two ways it does nothing, and neither is an error:
//
//   - A TARGET that became illegal in response is skipped. CR 608.2b's
//     re-check says the ability does as much as it can, and for equip
//     with one target that is nothing at all.
//   - The SOURCE is gone — the Equipment was sacrificed, destroyed or
//     bounced while the equip sat on the stack, or bounced and
//     replayed so that the permanent with that ID is a new object
//     (CR 400.7). game.AttachSourceForEffect owns that half for every
//     "attach this permanent" ability, so there is no per-card check
//     here and none in any card file. #812.
func AttachSourceToTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return g.AttachSourceForEffect(item, t)
	}
	return nil
}

// AttachedToSource is the AppliesTo predicate for every "equipped
// creature" and "enchanted creature" static. One card scoped by one
// relation — the same shape as selfOnly, which is the precedent for
// a static that applies to exactly one permanent.
//
// Evaluated per battlefield card per recompute, so it re-reads the
// current attachment every pass: moving a sword with a second equip
// moves its bonus in the same beat, with no invalidation bookkeeping
// beyond the EventAttach the primitive already emits.
func AttachedToSource(target *game.Card, _ *game.Game, source *game.Card) bool {
	return source.IsAttachedTo(target.InstanceID)
}

// PumpAttached is "Equipped creature gets +P/+T" / "Enchanted
// creature gets +P/+T" — a layer 7c modify, which is where every
// non-setting P/T change belongs (CR 613.4c).
func PumpAttached(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power += power
			c.Toughness += toughness
		},
	}
}

// GrantToAttached is "Equipped creature has X" / "Enchanted creature
// has X" for one or more keywords — a layer 6 ability grant.
//
// Keywords MUST be canonical lowercase tokens the engine actually
// honours (game.CanonicalKeyword is the table). A grant of something
// outside it would render as a badge the rules layer never backs,
// which is the one failure mode ADR 0038 exists to prevent.
//
// Deduped on apply, so an Equipment granting haste to a creature
// that already prints haste does not produce two badges.
func GrantToAttached(keywords ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, kw := range keywords {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
			}
		},
	}
}

// RemoveFromAttached is "Equipped creature ... loses X" — Colossus
// Hammer's "and loses flying", the mirror of GrantToAttached and the
// same layer 6.
//
// It strips whatever the layers have granted SO FAR, in timestamp
// order, which is the same shape (and the same declared limit) as
// b27LoseKeyword: a grant from a source with a later timestamp than
// this Equipment lands after this removal and survives it. That is
// CR 613.7 working correctly, not a bug — a Lightning Greaves
// equipped after the Hammer really does give the creature back its
// haste — but it does mean "loses flying" is not an absolute.
func RemoveFromAttached(keywords ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			kept := c.Abilities[:0]
			for _, a := range c.Abilities {
				drop := false
				for _, kw := range keywords {
					if equalFoldASCIIEffects(a, kw) {
						drop = true
						break
					}
				}
				if !drop {
					kept = append(kept, a)
				}
			}
			c.Abilities = kept
		},
	}
}

// --- the Auras that change what a permanent IS -------------------
//
// Darksteel Mutation, Kenrith's Transformation and Song of the
// Dryads all say the same sentence in three dialects: the enchanted
// permanent stops being what it was and becomes something small.
// Each clause of that sentence is a different CR 613 layer, and the
// four helpers below are one layer each — declared separately in the
// card file, in printed order, because a reader of the card file
// should be able to see which layer each phrase landed in.

// LoseAllAbilities is "loses all abilities" (CR 613.1f) — a layer 6
// ability REMOVAL, the mirror of GrantToAttached at the whole-card
// scale rather than the keyword scale.
//
// `keep` names keywords the SAME effect grants back in the same
// breath: Darksteel Mutation's "has indestructible, and it loses all
// OTHER abilities". They are not a timestamp exception. One
// continuous effect removes and grants at once, and CR 613.6's
// later-grant-survives rule is about a DIFFERENT effect with a later
// timestamp — which works here too, through the layer-6 sort, with
// no help from this helper: a Rancor attached after the Mutation
// gives the Insect trample, and one attached before it does not.
//
// The engine does the removing. This helper only declares
// RemovesAbilities and appends the keeps, because the removal has to
// be visible to the recompute (it is what stops the source's effects
// STARTING in any later layer, CR 613.6) and to
// game.CatalogAbilityKey (it is what stops the catalogued activated,
// triggered, mana, static and replacement abilities from answering).
// Clearing c.Abilities by hand would remove the keyword badges and
// leave the card fully functional underneath them, which was the
// exact shape of the gap this closes.
//
// NOT for "is a Forest land". That loss is CR 305.7's, it is part of
// the type change, and it belongs in layer 4 — see SetsBasicLandType.
func LoseAllAbilities(keep ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer:            game.Layer6Ability,
		RemovesAbilities: true,
		// Every catalogued user of this helper prints it as one
		// sentence with a layer-4 type change ("loses all abilities
		// and is a green Elk creature"), so CR 613.6 keeps this half
		// applying if something silences the AURA part-way through
		// the pass.
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, kw := range keep {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
			}
		},
	}
}

// SetAttachedTypes is "is an Insect artifact creature" / "is a green
// Elk creature" / "is a colorless Forest land" — a layer 4 type
// change that REPLACES rather than adds, which is what "is a"
// means and what every one of these Auras' reminder text spells out
// ("It loses all other card types and creature types").
//
// Supertypes are deliberately untouched. CR 205.4 supertypes are not
// card types and not creature types, so a legendary commander under
// a Darksteel Mutation is still legendary — which is the whole
// reason the card is a Commander staple and not a worse Pacifism.
// Nothing here is Mycosynth Lattice's additive shape; use
// game.Layer4Type directly for that.
//
// The subtype half goes through Characteristic.SetSubtypes, which is
// also what takes "is every creature type" away: a creature a
// Maskwood Nexus made every type earlier in layer 4 and a Kenrith's
// Transformation then set to Elk is an Elk (#670).
func SetAttachedTypes(types []string, subtypes []string) game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer4Type,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Types = append([]string(nil), types...)
			c.SetSubtypes(subtypes)
		},
	}
}

// SetAttachedColors is "is a GREEN Elk" / "is a COLORLESS Forest
// land" — a layer 5 colour change, and the first one in the catalog.
// Layer 5 has been in the engine's bucket order since S16 with
// nothing to put in it.
//
// No arguments means colourless, which is a real answer and not an
// empty one: Song of the Dryads turning an opposing commander
// colourless is how it dodges a colour-restricted removal spell.
func SetAttachedColors(colors ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer5Color,
		// CR 613.6: "is a colorless Forest land" is ONE continuous
		// effect that starts in layer 4, so the colour half keeps
		// applying even if something takes the Aura's abilities away
		// part-way through the pass. Same for the base-P/T half
		// below. See game.StaticAbility.ContinuesAfterRemoval.
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Colors = append([]string(nil), colors...)
		},
	}
}

// SetAttachedBasePT is "with base power and toughness 0/1" — layer
// 7b, which is where a SET belongs, as opposed to PumpAttached's 7c
// modify.
//
// The sub-layer is the whole difference and it is observable: 7b
// runs before 7c and before 7d, so a Darksteel Mutation'd creature
// really is 0/1 plus whatever anthems and +1/+1 counters are on it,
// rather than 0/1 flat. That is the printed behaviour, and it is
// what makes putting the Mutation on your own creature with counters
// on it a real (bad) decision rather than a no-op.
func SetAttachedBasePT(power, toughness int) game.StaticAbility {
	return game.StaticAbility{
		Layer:                 game.Layer7PT,
		SubLayer:              game.SubLayer7B_Set,
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power = power
			c.Toughness = toughness
		},
	}
}

// ControlAttachedBySource is "You control enchanted creature" — the
// CR 613.1b layer-2 continuous effect that Mind Control is.
//
// It is a continuous effect registered by the Aura's own static
// ability, NOT a one-shot mutation performed when the Aura resolves,
// and that distinction is the whole card:
//
//   - Destroy the Aura and control reverts by itself, because the
//     effect stops being in the active set on the next recompute.
//     No "remember who had it" bookkeeping, and nothing to get wrong
//     if the Aura leaves in an unusual way (bounced, exiled, its host
//     stops being a creature).
//   - Two control-changers on one creature sort by timestamp and the
//     later one wins (CR 613.7), for free, through the same sort
//     every other layer uses.
//   - The control change is visible to combat, targeting, activated
//     abilities and the wire, because the recompute materialises
//     layer 2's output back onto Card.Controller.
//
// The Aura's OWN controller is the new controller, not its owner: a
// Mind Control that has itself been stolen steals for whoever holds
// it now.
func ControlAttachedBySource() game.StaticAbility {
	return game.StaticAbility{
		Layer:     game.Layer2Control,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			c.Controller = source.Controller
			// #930: the Aura is what took it, and the control-change
			// event names the effect's source. Written by the same
			// assignment that decides the controller, so the winner of
			// CR 613.7's timestamp sort is the one the event reports.
			c.ControlSource = source.InstanceID
		},
	}
}

// EnchantCreature is an Aura's "Enchant creature" clause. It is an
// ordinary target spec, because at cast time that is exactly what it
// is (CR 303.4a) — and because the state-based action that keeps the
// Aura legal afterwards re-runs this same spec, one definition
// serves both.
func EnchantCreature(preds ...CardPredicate) *game.TargetSpec {
	return TargetCreature("enchant creature", preds...)
}

// EnchantPermanent is the widest enchant clause: "Enchant permanent"
// (Faith's Fetters, Song of the Dryads). Distinct from
// EnchantCreature because the CR 704.5m legality re-check runs this
// very spec every turn — an Aura declared as "enchant creature" falls
// off a host that stops being one, and an Aura declared this way does
// not.
//
// That asymmetry is the whole of Song of the Dryads. Its layer-4
// change stops the host being a creature, so every Equipment on it
// unattaches (CR 704.5n) and every "enchant creature" Aura on it goes
// to the graveyard (CR 704.5m) — and the Song stays, because a land
// is still a permanent.
func EnchantPermanent(preds ...CardPredicate) *game.TargetSpec {
	return TargetPermanent("enchant permanent", preds...)
}

// EnchantLand is "Enchant land" (Wild Growth, Overgrowth, Fertile
// Ground) and, with a predicate, the narrower printings — "Enchant
// Forest" is EnchantLand(Subtype("Forest")) on Utopia Sprawl.
//
// Like every other enchant clause this is the spec the CR 704.5m
// legality re-check reruns, so an enchanted land that stops being one
// takes the Aura to the graveyard with it — which is the whole reason
// it is not just EnchantPermanent(Land()) at each call site.
func EnchantLand(preds ...CardPredicate) *game.TargetSpec {
	return TargetPermanent("enchant land", append([]CardPredicate{Land()}, preds...)...)
}

// EnchantPlayer is a Curse's "Enchant player" clause. The reason
// Card.AttachedTo is a TargetRef and not a card ID.
func EnchantPlayer(preds ...PlayerPredicate) *game.TargetSpec {
	return TargetPlayer("enchant player", preds...)
}

// equippedCreatureDied reports whether ev is the death of the
// creature this Equipment is attached to — Skullclamp's trigger
// condition.
//
// The ordering this depends on was checked end to end and is not
// obvious: the trigger harvester runs synchronously inside EmitEvent
// under the write lock, and the CR 704.5n unattach is a state-based
// action that has not run yet. So at the instant EventLTB fires, the
// Equipment is still on the battlefield still pointing at the card
// that just died, and a direct read of AttachedTo is correct. Had it
// gone the other way this would have needed the LKI snapshot to grow
// an attachment field, which it has not got.
func equippedCreatureDied(ev game.Event, source *game.Card) bool {
	return ev.Kind == game.EventLTB &&
		ev.NewZone == game.ZoneGraveyard &&
		source.IsAttachedTo(ev.CardID)
}

// attachedCreatureDealtCombatDamageToPlayer is the Swords' trigger
// condition: "Whenever equipped creature deals combat damage to a
// player". The same EventDealDamage + ev.Combat shape Bident of
// Thassa uses, with the controller check swapped for the attachment.
func attachedCreatureDealtCombatDamageToPlayer(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 {
		return false
	}
	if !source.IsAttachedTo(ev.Source) {
		return false
	}
	return g.PlayerByIDForEffect(ev.Target) != nil
}

// untapAllLandsControlledBy is Sword of Feast and Famine's back half
// ("untap all lands you control"), which is a plain loop rather than
// a primitive because nothing else in the catalog untaps a set.
//
// Reads the battlefield through the effect API under the resolution
// write lock; collects IDs first so the untap does not mutate the
// slice being walked.
func untapAllLandsControlledBy(g *game.Game, controller uuid.UUID, ctx *Context) error {
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsLand() && c.Tapped {
			ids = append(ids, c.InstanceID)
		}
	}
	for _, id := range ids {
		if err := (UntapTarget{Target: id}.Apply(ctx)); err != nil {
			return err
		}
	}
	return nil
}

// SetsBasicLandType is CR 305.7 — "an effect that sets a land's
// subtype to one or more of the basic land types" — as one layer-4
// static. Song of the Dryads' "enchanted permanent is a colorless
// Forest land" is the catalog's first; Magus of the Moon's "nonbasic
// lands are Mountains" is the second, and it is not an attachment,
// which is why `applies` is an argument rather than AttachedToSource.
//
// CR 305.7 is three clauses and this is all three:
//
//   - the permanent's old land types go and the named basic land
//     types replace them — the type SET, so game.SetSubtypes;
//   - it loses the abilities generated from its rules text —
//     RemovesAbilities, and crucially in LAYER 4, because that is
//     where the type change is;
//   - it gains the basic land type's intrinsic mana ability. That
//     one is not here at all: game.ManaAbilitiesForCard derives it
//     from the EFFECTIVE subtypes (CR 305.6), so it arrives with the
//     type and survives the same effect's removal, which is what the
//     rule says happens.
//
// The layer is the whole point, and it is the difference between this
// helper and LoseAllAbilities. A removal in layer 4 cannot touch a
// grant that lands in layer 6, so an ability another effect gave the
// permanent survives whenever it was given — which is CR 305.7's last
// sentence ("this doesn't remove any abilities that were granted to
// the land by other effects") and Song's 2014-11-07 ruling ("it will
// still have any abilities it gained from other effects"). Modelling
// the loss as a layer-6 wipe took Boros Charm's indestructible off a
// Song'd Sol Ring and let a wrath destroy a Forest (#669).
//
// `types` is the whole card-type line the effect writes ("Land" for
// Song, which replaces artifact, creature and planeswalker; nil for
// Magus of the Moon, which changes only the subtypes of permanents
// that are lands already).
func SetsBasicLandType(applies func(target *game.Card, g *game.Game, source *game.Card) bool, types []string, subtypes []string) game.StaticAbility {
	return game.StaticAbility{
		Layer:            game.Layer4Type,
		RemovesAbilities: true,
		AppliesTo:        applies,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if len(types) > 0 {
				c.Types = append([]string(nil), types...)
			}
			c.SetSubtypes(subtypes)
		},
	}
}

// SetAttachedBasePTPer is SetAttachedBasePT with the value COUNTED
// from the game state on every recompute rather than printed —
// Aettir and Priwen's "base power and toughness X/X, where X is your
// life total".
//
// Same layer and sub-layer as its fixed sibling, and for the same
// reason: 7b SETS, so counters and anthems still land on top of the
// result in 7c and 7d. A card that set the base in 7c would fight
// the anthems instead of being modified by them, and a +1/+1 counter
// on the host would vanish.
//
// `value` runs inside a layer recompute where the caller may hold
// only the read lock, so it must read g.Battlefield / g.Seats
// directly rather than through a locking accessor — the same
// contract PumpAttachedPer's `count` carries.
//
// "Your" is the ATTACHMENT's controller (CR 109.5), which is what
// `source` is for: a stolen Aettir sizes its host off whoever holds
// the Equipment now, not off the creature's controller.
func SetAttachedBasePTPer(value func(g *game.Game, source *game.Card) (power, toughness int)) game.StaticAbility {
	return game.StaticAbility{
		Layer:                 game.Layer7PT,
		SubLayer:              game.SubLayer7B_Set,
		ContinuesAfterRemoval: true,
		AppliesTo:             AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			p, t := value(g, source)
			c.Power = p
			c.Toughness = t
		},
	}
}

// --- moving an EXISTING Equipment onto a creature (#1107) --------
//
// Brass Squire's activated ability and Magnetic Theft's spell say the
// same sentence: "Attach target Equipment [...] to target creature
// [...]". Neither is an equip ability — the Equipment being moved is
// not the source, so AttachSourceToTarget doesn't apply — and both are
// the ADR 0065 two-slot positional shape (Bite Down's shape, ported to
// an attachment): clause 0 names the Equipment, clause 1 the creature
// it lands on, read back positionally rather than by predicate, since
// nothing else distinguishes the two slots once both narrow to "a
// permanent".

// TwoSlotAttachTargets builds that clause pair. `equipmentPreds` /
// `creaturePreds` narrow each half exactly as printed — YouControl()
// on both for Brass Squire's "Equipment you control" / "creature you
// control", nothing for Magnetic Theft's unrestricted "target
// Equipment" / "target creature", which is why control of the
// Equipment never changes (AttachClauseTargets only ever writes the
// AttachedTo link, never Controller).
func TwoSlotAttachTargets(equipmentLabel string, equipmentPreds []CardPredicate, creatureLabel string, creaturePreds []CardPredicate) *game.TargetSpec {
	return Clauses(
		TargetPermanent(equipmentLabel, append([]CardPredicate{HasSubtype("Equipment")}, equipmentPreds...)...),
		TargetCreature(creatureLabel, creaturePreds...),
	)
}

// AttachClauseTargets is the resolution both cards share: attach
// clause 0's Equipment to clause 1's creature. Each clause is
// re-checked independently (CR 608.2b) — a target that left in
// response leaves nothing to attach, which AttachForEffect already
// treats as a quiet no-op (CR 701.3b) rather than an error, so this
// does too: there is no such thing as attaching nothing to a creature
// or an Equipment to nobody, so either miss is the whole ability
// doing as much as it can, which here is nothing.
func AttachClauseTargets(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	equip, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	host, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	return g.AttachForEffect(equip.ID, host)
}
