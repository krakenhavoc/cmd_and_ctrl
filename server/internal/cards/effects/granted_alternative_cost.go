package effects

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
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
	// GrantedAltCostEmerge is "Each creature spell you cast has emerge.
	// The emerge cost is equal to its mana cost." (Herigast, Erupting
	// Nullkite; ADR 0135 PR 6).
	GrantedAltCostEmerge = "granted-emerge"
	// GrantedAltCostZeroZombieCreatures is "You may pay {0} rather than
	// pay the mana cost for Zombie creature spells you cast." (Rooftop
	// Storm).
	GrantedAltCostZeroZombieCreatures = "granted-zero-zombie-creatures"
	// GrantedAltCostZeroFromExileOnceEachTurn is "Once each turn, you
	// may pay {0} rather than pay the mana cost for a spell you cast
	// from exile." (Warped Space, #3174).
	GrantedAltCostZeroFromExileOnceEachTurn = "granted-zero-from-exile-once-each-turn"
	// GrantedAltCostLifeFromHandOnceYourTurn is "Once during each of
	// your turns, you may cast a spell from your hand by paying life
	// equal to its mana value rather than paying its mana cost."
	// (Access Maze, #3174).
	GrantedAltCostLifeFromHandOnceYourTurn = "granted-life-from-hand-once-your-turn"
)

// PaidEmerge reports whether `key` — a StackItem's or a permanent's
// AltCost — is an emerge cost: the card's own ("emerge") or one
// Herigast gave it ("granted-emerge"). "If this creature's emerge cost
// was paid" (Adipose Offspring) is true for either: a spell given emerge
// has that emerge cost (CR 702.119a), and the Herigast ruling lets a
// spell with two emerge costs be cast for either.
func PaidEmerge(key string) bool {
	return key == AltCostKeyEmerge || key == GrantedAltCostEmerge
}

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

// GrantedAltCostFreeUpToCreatureCount is "You may cast spells with mana
// value less than or equal to the number of creatures you control from
// your hand without paying their mana costs." (Omnipresence, #3503).
// Its own key, not Omniscience's "granted-free": a table with both has
// two free offers whose reach differs, and a claim is stored by key.
const GrantedAltCostFreeUpToCreatureCount = "granted-free-up-to-creature-count"

// CastFromHandWithoutPayingManaCostUpToCreatureCount is Omnipresence's
// static. Omniscience's free cast from hand with a cap that FOLLOWS THE
// BOARD (GrantedAlternativeCost.MaxManaValueFunc): the number of
// creatures its controller controls right now, counted as the offer is
// derived, so the cast gate, the move list and the view's picker all
// see the same count and none of them holds a stale one. The caster is
// the one the offer is derived for (CR 601.2), so "you" is the caster.
// The mana value is the spell's as the offer replaces its mana cost, an
// {X} counting as 0 (CR 107.3b).
func CastFromHandWithoutPayingManaCostUpToCreatureCount() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostFreeUpToCreatureCount,
			Label: "Cast it without paying its mana cost",
		},
		Zones:            []game.ZoneKind{game.ZoneHand},
		MaxManaValueFunc: creaturesControlledCount,
		MaxManaValueText: "the number of creatures you control",
	}
}

// creaturesControlledCount is the number of creatures `controller`
// controls, off effective characteristics, so a type-changing effect
// counts. Caller must hold g.mu.
func creaturesControlledCount(g *game.Game, controller uuid.UUID) int {
	if g == nil || g.Battlefield == nil || controller == uuid.Nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.IsCreature() {
			n++
		}
	}
	return n
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

// EmergeForCreatureSpellsYouCast is Herigast, Erupting Nullkite's
// static: "Each creature spell you cast has emerge. The emerge cost is
// equal to its mana cost." (ADR 0135 §4 and PR 6, CR 702.119a). Each
// creature spell its controller casts, from any zone where its printed
// mana cost could be paid (CR 118.9a), may instead be cast by
// sacrificing a creature and paying its own mana cost reduced by that
// creature's mana value: the same one-creature sacrifice offer, flagged
// ReducedBySacrificedManaValue, that Emerge builds, priced at the
// spell's mana cost as it is cast (GrantedAlternativeCost.
// PricedAtManaCost). A creature that prints its own emerge keeps it
// beside this one, and either may be paid. The grant is read as the
// cost is chosen, so Herigast itself may be the creature sacrificed.
func EmergeForCreatureSpellsYouCast() game.GrantedAlternativeCost {
	what := "a creature"
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:                          GrantedAltCostEmerge,
			Label:                        "Emerge",
			Sacrifice:                    sacrificeSpec(what, Creature()).WithCount(1, 1),
			ReducedBySacrificedManaValue: true,
			PayLabel:                     what,
		},
		Spells:           game.PermissionFilter{CreatureOnly: true},
		PricedAtManaCost: true,
	}
}

// PayZeroForZombieCreatureSpellsYouCast is Rooftop Storm's static: "You
// may pay {0} rather than pay the mana cost for Zombie creature spells
// you cast." An offer with no mana cost is free, so X is 0 (CR
// 107.3b); every zone a Zombie creature spell is cast from, wherever its
// printed mana cost could be paid (CR 118.9a). Additional costs (a
// kicker, an emerge sacrifice) are still paid on top (CR 118.9d).
func PayZeroForZombieCreatureSpellsYouCast() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostZeroZombieCreatures,
			Label: "Pay {0} rather than pay this spell's mana cost",
		},
		Spells: game.PermissionFilter{CreatureOnly: true, CreatureType: "Zombie"},
	}
}

// PayZeroForASpellFromExileOnceEachTurn is Warped Space's static
// (#3174): "Once each turn, you may pay {0} rather than pay the mana
// cost for a spell you cast from exile." Free, so X is 0 (CR 107.3b);
// only where the printed mana cost could be paid from exile (an
// impulse-exiled card, CR 118.9a), so a foretold or plotted card, which
// is already alternatively costed, gets none. The limit is per Warped
// Space and spent as the spell becomes cast. Put it in a Door so the
// door gates it.
func PayZeroForASpellFromExileOnceEachTurn() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostZeroFromExileOnceEachTurn,
			Label: "Pay {0} rather than pay this spell's mana cost",
		},
		Zones:        []game.ZoneKind{game.ZoneExile},
		OnceEachTurn: true,
	}
}

// PayLifeForASpellFromHandOnceEachOfYourTurns is Access Maze's static
// (#3174): "Once during each of your turns, you may cast a spell from
// your hand by paying life equal to its mana value rather than paying
// its mana cost." Bolas's Citadel's price on a granted offer: a cost
// (CR 119.4), so a caster without the life is not offered it. A spell
// with {X} is cast with X = 0 (CR 107.3b), so it costs the rest of its
// mana value, and a spell with no mana cost (CR 118.6) costs 0 life.
func PayLifeForASpellFromHandOnceEachOfYourTurns() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostLifeFromHandOnceYourTurn,
			Label: "Pay life equal to its mana value rather than pay its mana cost",
		},
		Zones:                []game.ZoneKind{game.ZoneHand},
		OnceEachTurn:         true,
		YourTurnOnly:         true,
		LifeEqualToManaValue: true,
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
		if gr.MaxManaValueFunc != nil && gr.MaxManaValueText == "" {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q has a MaxManaValueFunc and no MaxManaValueText — the ability row words it", name, o.Key))
		}
		if gr.LifeEqualToManaValue && (o.ManaCost != "" || o.Energy != 0 || gr.PricedAtManaCost) {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q pays life equal to mana value and also declares another price", name, o.Key))
		}
		if gr.PricedAtManaCost && o.ManaCost != "" {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q is priced at the spell's mana cost and also declares ManaCost %q", name, o.Key, o.ManaCost))
		}
		// ADR 0135 PR 6: the one card component a granted offer may
		// carry is emerge's, a sacrifice of exactly one permanent that
		// reduces the price (Herigast). checkEmerge holds its shape.
		emerge := o.ReducedBySacrificedManaValue && o.Sacrifice != nil &&
			o.ExileFromHand == nil && o.DiscardFromHand == nil && o.ReturnToHand == nil &&
			o.ExileFromGraveyard == nil && o.TapOthers == nil
		if emerge {
			checkEmerge(name, o)
		}
		if o.Life != 0 || (o.PaysCards() && !emerge) || o.Condition != nil || o.Targets != nil || o.ClearsTargets ||
			o.SacrificeOnEntry || o.FromZone != "" || o.ExileOnLeavingStack || o.WarpExile || o.Dash || o.Blitz || o.Sneak ||
			o.EntersWithCounterName != "" || o.FaceDown != nil || o.RequiresGrant || o.CastsFace != 0 || o.Granted || o.XCeiling != 0 {
			panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q declares more than a price and a label — narrow its zones with Zones, not FromZone", name, o.Key))
		}
	}
}
