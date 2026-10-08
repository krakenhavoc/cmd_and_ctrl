package effects

import (
	"fmt"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_alternative_cost.go — ADR 0118 §3, #2163: the statics that
// offer their controller one more alternative cost (CR 118.9) for each
// spell they cast, declared in Spec.GrantedAlternativeCosts. The engine
// half is game/granted_alternative_cost.go.
//
// Mechanic-named and append-only: a new shape (a spell filter, a
// condition — Hunting Velociraptor's granted prowl) is a new
// constructor here, never an edit to one a card already uses.
//
// The keys are ON-DISK identities. A claim lands on StackItem.AltCost
// and is captured with the stack, so like a token slug a key is never
// renamed or reused. They are namespaced "granted-…" so that a card's
// own offer ("bringer-wubrg", a printed free cast) is never shadowed by
// the per-key dedupe of the offer list.

// grantedAltCostKeyPrefix is the namespace every granted key lives in.
const grantedAltCostKeyPrefix = "granted-"

// Granted alternative-cost keys. Never renamed, never reused.
const (
	// GrantedAltCostWUBRG is "You may pay {W}{U}{B}{R}{G} rather than
	// pay the mana cost for spells you cast."
	GrantedAltCostWUBRG = "granted-wubrg"
	// GrantedAltCostFree is "You may cast spells from your hand without
	// paying their mana costs."
	GrantedAltCostFree = "granted-free"
	// GrantedAltCostEightEnergyPermanents is "You may pay eight {E}
	// rather than pay the mana cost for permanent spells you cast."
	GrantedAltCostEightEnergyPermanents = "granted-eight-energy-permanents"
	// GrantedAltCostEnergySmallCreatures is "You may cast creature
	// spells with mana value 3 or less by paying {E} rather than paying
	// their mana costs. If you cast a spell this way, you may cast it as
	// though it had flash."
	GrantedAltCostEnergySmallCreatures = "granted-energy-small-creatures"
)

// PayWUBRGForSpellsYouCast is Fist of Suns', Jodah, Archmage Eternal's
// and Leyline of Mutation's static: "You may pay {W}{U}{B}{R}{G} rather
// than pay the mana cost for spells you cast." Every zone a spell is
// cast from, wherever the printed mana cost could be paid (CR 118.9a).
func PayWUBRGForSpellsYouCast() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:      GrantedAltCostWUBRG,
			Label:    "Pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost",
			ManaCost: "{W}{U}{B}{R}{G}",
		},
	}
}

// CastFromHandWithoutPayingManaCost is Omniscience's static: "You may
// cast spells from your hand without paying their mana costs." An
// offer with no mana cost is free (CR 118.9's second phrasing), X is 0
// (CR 107.3b), and the spell keeps its own timing (CR 117.1a, 307.1).
func CastFromHandWithoutPayingManaCost() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostFree,
			Label: "Cast it without paying its mana cost",
		},
		Zones: []game.ZoneKind{game.ZoneHand},
	}
}

// PayEightEnergyForPermanentSpellsYouCast is Nissa, Worldsoul Speaker's
// static: "You may pay eight {E} rather than pay the mana cost for
// permanent spells you cast." (ADR 0129 §5). Every zone a permanent
// spell is cast from, wherever its printed mana cost could be paid (CR
// 118.9a). The energy is a cost (CR 107.14): a caster with fewer than
// eight is not offered it.
func PayEightEnergyForPermanentSpellsYouCast() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:    GrantedAltCostEightEnergyPermanents,
			Label:  "Pay eight {E} rather than pay this spell's mana cost",
			Energy: 8,
		},
		Spells: game.PermissionFilter{NonLandPermanentOnly: true},
	}
}

// PayEnergyForSmallCreatureSpellsWithFlash is Primal Prayers' static:
// "You may cast creature spells with mana value 3 or less by paying {E}
// rather than paying their mana costs. If you cast a spell this way, you
// may cast it as though it had flash." (ADR 0129 §5). The flash belongs
// to the claim (CR 601.3c), so the same creature cast for its printed
// cost keeps its own timing.
func PayEnergyForSmallCreatureSpellsWithFlash() game.GrantedAlternativeCost {
	three := 3
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:           GrantedAltCostEnergySmallCreatures,
			Label:         "Pay {E} rather than pay this spell's mana cost, as though it had flash",
			Energy:        1,
			AsThoughFlash: true,
		},
		Spells:       game.PermissionFilter{CreatureOnly: true},
		MaxManaValue: &three,
	}
}

// checkGrantedAlternativeCosts refuses a declaration the engine would
// read wrongly, at boot: a key outside the namespace (it could shadow a
// card's own offer, or be shadowed by one), a key declared twice, and a
// component the seam does not carry. A granted offer is a price and a
// label, plus energy (ADR 0129 §5) and "as though it had flash" (CR
// 601.3c). Life, a card to pay with, a target rewrite or a spell rider
// has never been checked against the seam, so none may be declared
// until a card needs one and the seam is grown for it.
func checkGrantedAlternativeCosts(name string, in []game.GrantedAlternativeCost) {
	seen := make(map[string]bool, len(in))
	for i, gr := range in {
		o := gr.Offer
		if !strings.HasPrefix(o.Key, grantedAltCostKeyPrefix) || len(o.Key) == len(grantedAltCostKeyPrefix) {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %d has key %q — it must be %q plus a name", name, i, o.Key, grantedAltCostKeyPrefix))
		}
		if seen[o.Key] {
			panic(fmt.Sprintf("effects.Register: %q declares granted alternative cost %q twice", name, o.Key))
		}
		seen[o.Key] = true
		if o.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q has no Label — the cost picker shows it", name, o.Key))
		}
		if _, err := game.ParseCost(o.ManaCost); err != nil {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q has an unparseable ManaCost %q: %v", name, o.Key, o.ManaCost, err))
		}
		if o.Energy < 0 {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q has a negative Energy", name, o.Key))
		}
		if gr.MaxManaValue != nil && *gr.MaxManaValue < 0 {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q has a negative MaxManaValue", name, o.Key))
		}
		if o.Life != 0 || o.PaysCards() || o.Condition != nil || o.Targets != nil || o.ClearsTargets ||
			o.SacrificeOnEntry || o.FromZone != "" || o.ExileOnLeavingStack || o.WarpExile ||
			o.EntersWithCounterName != "" || o.FaceDown != nil || o.RequiresGrant || o.CastsFace != 0 || o.Granted {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q declares more than a price and a label — narrow its zones with Zones, not FromZone", name, o.Key))
		}
	}
}
