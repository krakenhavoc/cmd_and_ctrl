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
	// Static is for a COPY grant only. A layer-6 grant that names a
	// bundle with a Static slot is refused (TestEveryGrantKeyResolves,
	// ADR 0093 Decision 10): the layer pass gathers every static before
	// layer 1, so a static that only exists after layer 6 would never
	// be gathered.
	//
	// Mana is for a LAYER-6 grant (ADR 0093) — no printed copy effect
	// grants one.
	Triggered []game.TriggeredAbility
	Static    []game.StaticAbility
	Activated []ActivatedAbility
	Mana      []ManaAbility

	// Text is the bundle as the granting card prints it, quoted ability
	// and all: "{T}: Add one mana of any color." Required. The recipient
	// has no printing that says it has the ability, and a granted
	// trigger has no ability row on the wire at all, so this is the only
	// thing that tells a player what their permanent can now do (ADR
	// 0093 Decision 8; the reason ADR 0083 gives a token its Text).
	Text string
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
// as every "any color" pipe does, and never narrows it (AGENTS.md §7).
func AnyColorManaGrant(key string) AbilityGrant {
	return TapForManaGrant(key, "{W|U|B|R|G}", "Add one mana of any color", "{T}: Add one mana of any color.")
}

// TribalAbilityGrant is GrantAbilities over a TribeFilter — the
// TribalKeywordGrant twin (ADR 0093 Decision 9). "All Slivers have …"
// (Gemhide Sliver) is TribeFilter{Tribes: []string{"Sliver"}}, which
// reaches every player's Slivers and the source itself; "Sliver
// creatures you control have …" (Manaweft Sliver) adds YoursOnly.
func TribalAbilityGrant(f TribeFilter, keys ...string) game.StaticAbility {
	return GrantAbilities(f.Matches, keys...)
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
		if len(gr.Triggered) == 0 && len(gr.Static) == 0 && len(gr.Activated) == 0 && len(gr.Mana) == 0 {
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
	}
	for i, s := range gr.Static {
		if s.ActiveWhen.IsGate() {
			panic(where("static", i) + " declares ActiveWhen — a designation is the host's; gate the grantor's static instead")
		}
		if len(s.GrantAbilities) > 0 {
			panic(where("static", i) + " grants abilities itself — a granted ability that grants another is not modelled")
		}
	}
	for i, a := range gr.Activated {
		if a.ActiveWhen.IsGate() {
			panic(where("activated ability", i) + " declares ActiveWhen — a designation is the host's; gate the grantor's static instead")
		}
		if offBattlefield(a.Zones) {
			panic(where("activated ability", i) + " functions outside the battlefield — a granted ability is a permanent's")
		}
	}
	for i, m := range gr.Mana {
		if offBattlefield(m.Zones) {
			panic(where("mana ability", i) + " functions outside the battlefield — a granted ability is a permanent's")
		}
	}
}

// checkStaticGrants refuses a StaticAbility.GrantAbilities outside the
// layer it belongs to. A grant is an ability-ADDING effect, CR 613.1f's
// layer 6; declared anywhere else it would sort against the wrong
// effects and a removal in layer 6 could never reach it.
func checkStaticGrants(card, what string, statics []game.StaticAbility) {
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
}
