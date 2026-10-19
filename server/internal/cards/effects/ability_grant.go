package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_grant.go — a BUNDLE of abilities one card gives another
// object. Two rules name one, and they are the same bundle:
//
//   - CR 707.9a, "except it has '<ability>'" (#665). Phantasmal Image
//     copies a creature "except … it has 'When this creature becomes the
//     target of a spell or ability, sacrifice it.'" The copy carries the
//     bundle's NAME in its copiable values (game.PrintedValues.
//     GrantAbility), so a Clone of a Phantasmal Image has the trigger
//     too.
//   - CR 113.10 and CR 613.1f, "creatures you control have '<ability>'"
//     (ADR 0093). Cryptolith Rite gives every creature you control
//     "{T}: Add one mana of any color." A layer-6 static names the
//     bundle in game.StaticAbility.GrantAbilities (build it with
//     GrantAbilities below) and the engine writes the name onto each
//     recipient's layered characteristic — which is NOT copied
//     (CR 707.2).
//
// Either way a grant is not a closure on anything. It is declared HERE,
// as static catalog data on the card that grants it, and Register files
// each bundle's own *game.CardDef in the same `defs` map cards use,
// under game.GrantKey(Key) — the shape #623 already uses for emblems,
// for the same reason: the object needs abilities the catalog can find,
// and it is not a card.
//
// The engine half is server/internal/game/copy_grants.go (the composite
// key and the merge) and server/internal/game/granted_abilities.go (the
// layer-6 grant, the readers and the refs).

// AbilityGrant is one bundle of abilities a card can give another
// object. A card declares its bundles in Spec.Grants and names one —
// from a copy's except clause, or from a layer-6 static:
//
//	Grants: []AbilityGrant{{
//	    Key:  "cryptolith-rite/any-color",
//	    Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color"}},
//	    Text: "{T}: Add one mana of any color.",
//	}},
//	Static: []game.StaticAbility{
//	    GrantAbilities(creaturesYouControl, "cryptolith-rite/any-color"),
//	},
//
// One bundle per printed quoted ability. Splitting a card's two granted
// abilities into two bundles costs nothing and makes the card read like
// the card.
type AbilityGrant struct {
	// Key is the bundle's name, unique across the whole catalog.
	// Namespace it with the granting card ("<card>/<what>") — it is
	// the string the recipient carries and the string a snapshot
	// carries, so it is as permanent as an oracle ID and a collision
	// would silently give one card another's ability.
	Key string

	// Triggered, Static, Activated and Mana are the ability kinds the
	// engine reads off a permanent, declared exactly as a card declares
	// its own.
	//
	// Static is any static for a COPY grant. A layer-6 grant may name a
	// bundle with a Static slot when every static in it is in layer 6 or
	// in layer 7b, 7c or 7e, and none removes, grants or borrows
	// abilities: "Equipped creature gets +3/+3" (ADR 0093 Decision 10,
	// amended 2026-10-10, #2562), "Other commanders you control get
	// +2/+2 and have lifelink" (amended 2026-10-15, #3784). A layer-6
	// static joins the layer-6 bucket the moment its grant applies, at
	// CR 613.7a's timestamp; a layer-7 one is gathered after layer 6
	// (game/granted_statics.go). TestEveryGrantKeyResolves and
	// GrantAbilitiesFor refuse any other (game.GrantedStaticProblem).
	//
	// Mana is for a LAYER-6 grant (ADR 0093) — no printed copy effect
	// grants one.
	Triggered []game.TriggeredAbility
	Static    []game.StaticAbility
	Activated []ActivatedAbility
	Mana      []ManaAbility

	// LegendRuleExempt is a granted "the legend rule doesn't apply to
	// permanents you control" (a copy of Sakashima of a Thousand Faces).
	LegendRuleExempt []game.LegendRuleExemption

	// AttackLimits is a granted CR 508.1c count limit — Tomik, Orzhov
	// Lawmage's "Planeswalkers you control have 'No more than one
	// creature can attack this planeswalker each combat.'" (#2821).
	// Not a layer effect, so Decision 10's refusal of a granted static
	// does not reach it: the attack check reads limits live under the
	// recipient's composite key, so "this" is the recipient and the
	// limit is gone the pass its grant is.
	AttackLimits []game.AttackLimit

	// DealtDamageCantBeRegenerated is a granted "Creatures dealt damage by
	// this creature this turn can't be regenerated this turn." — the
	// ability Bone Shaman gains (#1946). Not a layer effect either: the
	// regeneration gate reads it live under the recipient's composite
	// key, so "this creature" is the recipient and the rule is gone the
	// pass its grant is.
	DealtDamageCantBeRegenerated []game.DealtDamageCantBeRegenerated

	// GameEndGates is a granted "You can't lose the game and your
	// opponents can't win the game." — the ability Cloudsteel Kirin
	// gives the equipped creature (#3862, ADR 0093 amendment
	// 2026-10-15). Not a layer effect either: the loss and win checks
	// read gates live under the recipient's composite key
	// (game.CatalogGameEndGates), so "you" is the recipient's
	// controller and the gate is gone the pass its grant is — a grant
	// a newer removal took, and not one an older removal could not.
	GameEndGates []game.GameEndGate

	// TriggerDoublers is a granted CR 603.2d doubler — Dungeon Delver's
	// "Commander creatures you own have 'Room abilities of dungeons you
	// own trigger an additional time.'" (#4137, ADR 0018's trigger
	// doubling addendum, amended 2026-10-19). Not a layer effect: the
	// harvest's doubler scan reads doublers off each permanent's composite
	// key (game.CatalogTriggerDoublers via CatalogAbilityKey), so the
	// doubler's "you" is the RECIPIENT's controller (TriggerDoublingQuery.
	// Doubler is the recipient), a recipient that loses its abilities
	// loses it, and each recipient is one more instance: two commander
	// creatures are two additional triggers (CR 603.2d: doublers add).
	TriggerDoublers []game.TriggerDoubler

	// Zone is where the bundle's RECIPIENTS are: empty is the battlefield,
	// as for every bundle before #3233. A hand, graveyard or exile bundle
	// ("each creature card in your hand has ninjutsu {2}{U}{B}") is given
	// to cards by a static built with GrantAbilitiesInZone, and holds
	// activated abilities that function from that zone and nothing else —
	// a card off the battlefield has no layered characteristic for a
	// trigger, a static or a mana ability to live on (ADR 0093 amendment
	// 2026-10-13).
	Zone game.ZoneKind

	// Text is the bundle as the granting card prints it, quoted ability
	// and all: "{T}: Add one mana of any color." Required. The recipient
	// has no printing that says it has the ability, and a granted
	// trigger has no ability row on the wire at all, so this is the only
	// thing that tells a player what their permanent can now do (ADR
	// 0093 Decision 8; the reason ADR 0083 gives a token its Text).
	Text string
}

// GrantAbilitiesInZone is GrantAbilities for the CARDS in a zone other
// than the battlefield (ADR 0093 amendment 2026-10-13, #3233): "Each
// creature card in your hand has ninjutsu {2}{U}{B}." is
//
//	GrantAbilitiesInZone(game.ZoneHand, creatureCardsYouOwn, "satoru-umezawa/ninjutsu")
//
// `appliesTo` is asked of each card in every seat's zone, with this
// static's source as `source`; a card the predicate accepts has the
// bundle's activated abilities for as long as it is in that zone and
// the source is on the battlefield with its abilities. The bundle names
// the same Zone (checked by TestEveryGrantKeyResolves).
func GrantAbilitiesInZone(zone game.ZoneKind, appliesTo func(card *game.Card, g *game.Game, source *game.Card) bool, keys ...string) game.StaticAbility {
	ab := GrantAbilities(appliesTo, keys...)
	ab.GrantZone = zone
	return ab
}

// CardsYouOwn is the common recipient predicate for a zone grant: the
// source's controller's cards (in whichever zone the static names) that
// `match` accepts. A card outside the battlefield is its owner's
// (CR 108.4), and the pass only ever offers a card from a hand, a
// graveyard or exile, so the owner test is the "your" of the printed
// text.
func CardsYouOwn(match func(card *game.Card) bool) func(card *game.Card, g *game.Game, source *game.Card) bool {
	return func(card *game.Card, _ *game.Game, source *game.Card) bool {
		return source != nil && card.Owner == source.Controller && match(card)
	}
}

// buildGrantDef projects one bundle into the shape the engine reads.
// The twin of buildEmblemDef, and like it a *game.CardDef that is not
// a card's.
func buildGrantDef(gr AbilityGrant) *game.CardDef {
	return &game.CardDef{
		Triggered:     gr.Triggered,
		Static:        gr.Static,
		Activated:     activatedShapes(gr.Activated),
		ManaAbilities: manaShapes(gr.Mana),
		GrantText:     gr.Text,

		LegendRuleExemptions: gr.LegendRuleExempt,
		AttackLimits:         gr.AttackLimits,

		DealtDamageCantBeRegenerated: gr.DealtDamageCantBeRegenerated,

		GameEndGates: gr.GameEndGates,

		TriggerDoublers: gr.TriggerDoublers,
	}
}

// GrantAbilities is the layer-6 static that gives every object
// `appliesTo` matches the named ability bundles (CR 113.10, ADR 0093
// Decision 9) — "Creatures you control have '{T}: Add one mana of any
// color.'" is
//
//	GrantAbilities(creaturesYouControl, "cryptolith-rite/any-color")
//
// Each key names a bundle some card declares in Spec.Grants; either
// spelling of game.GrantKey works. The engine writes the grant, with
// this static's source as the grantor, in this static's timestamp slot
// of the layer-6 bucket — so a later ability removal takes it and an
// earlier one does not (CR 613.6), and a removal on the GRANTOR takes
// it from everything whatever the timestamps (CR 613.8a).
//
// `appliesTo` is the ordinary StaticAbility predicate; the same value
// can drive an anthem beside it.
func GrantAbilities(appliesTo func(target *game.Card, g *game.Game, source *game.Card) bool, keys ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer:          game.Layer6Ability,
		AppliesTo:      appliesTo,
		GrantAbilities: append([]string(nil), keys...),
	}
}

// TapForManaGrant is the commonest granted bundle, "{T}: Add <mana>."
// — Jaheira's "{T}: Add {G}.", and with AnyColorManaGrant below,
// Cryptolith Rite's and Chromatic Lantern's. `produced` is the
// ordinary ParseProducedMana grammar; `text` is the quoted ability as
// the granting card prints it.
func TapForManaGrant(key, produced, label, text string) AbilityGrant {
	return AbilityGrant{
		Key:  key,
		Mana: []ManaAbility{{Cost: ManaAbilityCost{Tap: true}, Produced: produced, Label: label}},
		Text: text,
	}
}

// AnyColorManaGrant is "{T}: Add one mana of any color." as a granted
// bundle. The picker lists the controller's commander identity first,
// as every "any color" pipe does, and never narrows it (docs/adding-cards.md).
func AnyColorManaGrant(key string) AbilityGrant {
	return TapForManaGrant(key, "{W|U|B|R|G}", "Add one mana of any color", "{T}: Add one mana of any color.")
}

// TribalAbilityGrant is GrantAbilities over a TribeFilter — the
// TribalKeywordGrant twin (ADR 0093 Decision 9). "All Slivers have …"
// (Gemhide Sliver) is TribeFilter{Tribes: []string{"Sliver"}}, which
// reaches every player's Slivers and the source itself; "Sliver
// creatures you control have …" (Manaweft Sliver) adds YoursOnly.
func TribalAbilityGrant(f TribeFilter, keys ...string) game.StaticAbility {
	ab := GrantAbilities(f.Matches, keys...)
	ab.Label = f.subject() + " have additional abilities."
	return ab
}

// GrantAbilitiesToAttached is GrantAbilities over the permanent this
// Equipment or Aura is attached to — the GrantToAttached twin (ADR 0093
// Decision 9). "Equipped creature has '{T}: Add one mana of any
// color.'" (Paradise Mantle) and "Enchanted land has '{T}: Create a 1/1
// green Squirrel creature token.'" (Squirrel Nest). The grant follows
// the attachment: it ends the moment the Equipment is moved or the
// Aura falls off, because AttachedToSource is re-read every layer pass.
func GrantAbilitiesToAttached(keys ...string) game.StaticAbility {
	return GrantAbilities(AttachedToSource, keys...)
}

// checkGrants is the registration guard. Each failure is a card file
// that is wrong in a way no test of the card would catch — a recipient
// naming a bundle nobody registered gets a silent no-ability grant — so
// it fails at boot.
func checkGrants(name string, grants []AbilityGrant) {
	for _, gr := range grants {
		if gr.Key == "" {
			panic("effects.Register: " + name + " declares an ability grant with no Key — the recipient has no name to carry")
		}
		if len(gr.Triggered) == 0 && len(gr.Static) == 0 && len(gr.Activated) == 0 && len(gr.Mana) == 0 && len(gr.LegendRuleExempt) == 0 && len(gr.AttackLimits) == 0 &&
			len(gr.DealtDamageCantBeRegenerated) == 0 && len(gr.GameEndGates) == 0 && len(gr.TriggerDoublers) == 0 {
			panic("effects.Register: " + name + " declares the ability grant " + gr.Key + " with no abilities — a grant gives an ability or nothing")
		}
		if gr.Text == "" {
			panic("effects.Register: " + name + " declares the ability grant " + gr.Key + " with no Text — the recipient has no printing that says what it can now do")
		}
		if _, taken := defs[game.GrantKey(gr.Key)]; taken {
			panic("effects.Register: " + name + " declares the ability grant " + gr.Key + ", which is already registered — grant keys are catalog-wide")
		}
		checkGrantAbilities(name, gr)
	}
}

// checkGrantAbilities is ADR 0093 Decision 4's refusal: a bundle's
// abilities are the RECIPIENT's, so two things a card's own ability may
// declare mean nothing on one.
//
//   - ActiveWhen. A designation is the host's. Gate the GRANT — the
//     grantor's own static takes ActiveWhen — not the granted ability.
//   - Zones other than the battlefield. The layer pass reaches only
//     permanents, so a bundle ability that functions from a hand or a
//     graveyard could never be anywhere it functions.
func checkGrantAbilities(name string, gr AbilityGrant) {
	where := func(kind string, i int) string {
		return fmt.Sprintf("effects.Register: %q ability grant %s %s %d", name, gr.Key, kind, i)
	}
	if gr.Zone != "" && gr.Zone != game.ZoneBattlefield {
		checkZoneGrantBundle(name, gr, where)
		return
	}
	offBattlefield := func(zones []game.ZoneKind) bool {
		for _, z := range zones {
			if z != game.ZoneBattlefield {
				return true
			}
		}
		return false
	}
	for i, t := range gr.Triggered {
		if t.ActiveWhen.IsGate() {
			panic(where("trigger", i) + " declares ActiveWhen — a designation is the host's; gate the grantor's static instead")
		}
		if offBattlefield(t.Zones) {
			panic(where("trigger", i) + " watches from outside the battlefield — a granted ability is a permanent's")
		}
		checkTriggeredPurpose(name, fmt.Sprintf("ability grant %s trigger %d", gr.Key, i), t)
	}
	for i, s := range gr.Static {
		if s.ActiveWhen.IsGate() {
			panic(where("static", i) + " declares ActiveWhen — a designation is the host's; gate the grantor's static instead")
		}
		if len(s.GrantAbilities) > 0 {
			panic(where("static", i) + " grants abilities itself — a granted ability that grants another is not modelled")
		}
		// ADR 0155 Decision 1: a bundle that lends abilities would put a
		// borrow inside a borrow.
		if s.Borrows != nil {
			panic(where("static", i) + " borrows abilities — a granted ability that lends others is not modelled")
		}
	}
	for i, d := range gr.TriggerDoublers {
		if d.ActiveWhen.IsGate() {
			panic(where("trigger doubler", i) + " declares ActiveWhen — a designation is the host's; gate the grantor's static instead")
		}
		if d.Applies == nil {
			panic(where("trigger doubler", i) + " has no Applies — it would never double anything")
		}
	}
	for i, a := range gr.Activated {
		if a.ActiveWhen.IsGate() {
			panic(where("activated ability", i) + " declares ActiveWhen — a designation is the host's; gate the grantor's static instead")
		}
		if offBattlefield(a.Zones) {
			panic(where("activated ability", i) + " functions outside the battlefield — a granted ability is a permanent's")
		}
		checkAnyPlayerAbility(name, fmt.Sprintf("ability grant %s activated ability %d", gr.Key, i), a)
		checkActivatedPurpose(name, fmt.Sprintf("ability grant %s activated ability %d", gr.Key, i), a, false)
		checkExileGrantor(name, fmt.Sprintf("ability grant %s activated ability %d", gr.Key, i), a, true, gr.Zone)
	}
	for i, m := range gr.Mana {
		if offBattlefield(m.Zones) {
			panic(where("mana ability", i) + " functions outside the battlefield — a granted ability is a permanent's")
		}
	}
	checkManaAbilitiesAnswers(name, "ability grant "+gr.Key, gr.Mana)
}

// checkZoneGrantBundle is checkGrantAbilities for a bundle given to
// CARDS in a hand, graveyard or exile (AbilityGrant.Zone). The recipient
// has no layered characteristic, so the only kind of ability that can be
// anywhere it functions is an activated one that functions from that
// zone; everything else is refused at boot rather than shipped as text
// that never does anything.
func checkZoneGrantBundle(name string, gr AbilityGrant, where func(kind string, i int) string) {
	if why := game.GrantZoneUnsupported(gr.Zone); why != "" {
		panic(fmt.Sprintf("effects.Register: %q ability grant %s is given to cards in %s — %s", name, gr.Key, gr.Zone, why))
	}
	if len(gr.Triggered) > 0 || len(gr.Static) > 0 || len(gr.Mana) > 0 || len(gr.LegendRuleExempt) > 0 || len(gr.AttackLimits) > 0 ||
		len(gr.DealtDamageCantBeRegenerated) > 0 || len(gr.GameEndGates) > 0 || len(gr.TriggerDoublers) > 0 {
		panic(fmt.Sprintf("effects.Register: %q ability grant %s is given to cards in %s but declares more than activated abilities — only an activated ability is modelled on a card off the battlefield", name, gr.Key, gr.Zone))
	}
	for i, a := range gr.Activated {
		if a.ActiveWhen.IsGate() {
			panic(where("activated ability", i) + " declares ActiveWhen — a designation is the host's; gate the grantor's static instead")
		}
		functions := false
		for _, z := range a.Zones {
			if z == gr.Zone {
				functions = true
			}
		}
		if !functions {
			panic(where("activated ability", i) + " does not function from " + string(gr.Zone) + " — a granted ability must name the zone its recipients are in (ActivatedAbility.Zones)")
		}
		if why := game.AbilityNeedsPermanentSource(a.Cost); why != "" {
			panic(where("activated ability", i) + " functions from " + string(gr.Zone) + " but declares " + why + " — that component needs a permanent on the battlefield")
		}
		checkAnyPlayerAbility(name, fmt.Sprintf("ability grant %s activated ability %d", gr.Key, i), a)
		checkActivatedPurpose(name, fmt.Sprintf("ability grant %s activated ability %d", gr.Key, i), a, false)
		checkExileGrantor(name, fmt.Sprintf("ability grant %s activated ability %d", gr.Key, i), a, true, gr.Zone)
	}
}

// checkStaticGrants refuses a StaticAbility.GrantAbilities outside the
// layer it belongs to. A grant is an ability-ADDING effect, CR 613.1f's
// layer 6; declared anywhere else it would sort against the wrong
// effects and a removal in layer 6 could never reach it.
func checkStaticGrants(card, what string, statics []game.StaticAbility) {
	// ADR 0155: the borrow, the other layer-6 ability-adding declaration,
	// is held to its layer at every place a grant is.
	checkStaticBorrows(card, what, statics)
	for i, s := range statics {
		if len(s.GrantAbilities) == 0 {
			continue
		}
		if s.Layer != game.Layer6Ability {
			panic(fmt.Sprintf("effects.Register: %q %s %d grants abilities in layer %d — an ability grant is a layer-6 effect (CR 613.1f)",
				card, what, i, s.Layer))
		}
		for _, k := range s.GrantAbilities {
			if k == "" {
				panic(fmt.Sprintf("effects.Register: %q %s %d grants an ability bundle with no key", card, what, i))
			}
		}
	}
	// ADR 0093 amendment 2026-10-13: a grant to cards off the battlefield.
	for i, s := range statics {
		if s.GrantZone == "" {
			continue
		}
		where := fmt.Sprintf("effects.Register: %q %s %d", card, what, i)
		switch {
		case what != "static":
			panic(where + " grants to cards in a zone — only a battlefield permanent's static is gathered for it")
		case len(s.GrantAbilities) == 0:
			panic(where + " declares GrantZone but grants no ability bundle")
		case s.Layer != game.Layer6Ability:
			panic(fmt.Sprintf("%s grants to cards in a zone in layer %d — an ability grant is a layer-6 effect (CR 613.1f)", where, s.Layer))
		case s.AppliesTo == nil:
			panic(where + " grants to cards in a zone but has no AppliesTo")
		case s.Apply != nil || s.RemovesAbilities || s.Borrows != nil || s.AffectsSpells || len(s.Zones) > 0:
			panic(where + " grants to cards in a zone beside an Apply, a removal, a borrow, a spell clause or a declared source zone — declare those as a static of their own")
		}
		if why := game.GrantZoneUnsupported(s.GrantZone); why != "" {
			panic(fmt.Sprintf("%s grants to cards in %s — %s", where, s.GrantZone, why))
		}
	}
}
