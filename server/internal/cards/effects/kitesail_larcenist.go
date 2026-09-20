package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kitesail Larcenist — Creature — Human Pirate {2}{U}, 2/3:
//
//	"Flying, ward {1}
//	 When this creature enters, for each player, choose up to one
//	 other target artifact or creature that player controls. For as
//	 long as this creature remains on the battlefield, the chosen
//	 permanents become Treasure artifacts with '{T}, Sacrifice this
//	 artifact: Add one mana of any color' and lose all other
//	 abilities."
//
// The two keywords ship; the enters trigger does not ship at all.
// A 2/3 flier with ward {1} for three is a real card, and half a
// Treasure-ifying trigger would not be — see below on why there is no
// honest partial here.
//
// # The three questions the trigger asks, answered separately
//
// (a) THE TARGET CLAUSE. "For each player, choose up to one other
// target artifact or creature that player controls" is one optional
// target SLOT PER PLAYER, and the number of slots is not known when
// the card is registered. game.TargetSpec grew multi-clause statements
// in #764 (TargetSpec.Rest, built with Clauses(first, then…)), but the
// list is FLAT and FIXED at catalog-build time: there is no clause
// count derived from the seat count, and nothing scopes a clause's
// candidate set to "the player this slot belongs to". Hard-coding four
// clauses would be wrong at any other table size and wrong again the
// moment a player is eliminated. This is the smaller of the two gaps
// and could be a per-player clause generator; it is recorded on the
// seam list rather than invented here.
//
// (b) THE TYPE CHANGE AND THE ABILITY LOSS. Both of these DO have a
// shape. Layer 4 is authoritative since S24 (ADR 0039) and layer 6's
// removal reaches the Catalog* hooks (ADR 0046), so "become Treasure
// artifacts" is a layer-4 static and "lose all other abilities" is
// LoseAllAbilities() — a DECLARATION (RemovesAbilities: true), never
// something an Apply closure does by hand.
//
// (c) THE ABILITY IT GRANTS. This is the blocker. The chosen
// permanents gain "{T}, Sacrifice this artifact: Add one mana of any
// color" — a MANA ability given to another permanent by a continuous
// effect. Mana abilities are read out of the catalog by the
// permanent's OWN oracle ID (game.CatalogManaAbilities via
// CatalogAbilityKey); Characteristic carries granted KEYWORDS as
// strings and has nowhere to put an ability with a cost and a
// production. That is the "Ability granted to another permanent by a
// static" seam (#754), which twelve other cards already wait on —
// Chromatic Lantern, Cryptolith Rite, Gemhide Sliver and friends.
//
// # Why there is no partial
//
// Shipping (b) without (c) would turn an opponent's Sol Ring into a
// Treasure that cannot be tapped for anything — strictly better
// removal than the card prints. Shipping (c) without (b) is not a
// thing. And the duration, which looked like the third problem, is
// not one: "for as long as this creature remains on the battlefield"
// is CR 611.2b and has had a shape since #755 / ADR 0063
// (DurationWhileSourceRemains + StaticForDuration + SnapshotAffected,
// the same machinery Sower of Temptation uses). So the trigger is one
// engine seam away, not three, and the file is written so that when
// #754 lands the whole ability can be added in one place.
func init() {
	Register(Spec{
		OracleID:     "2452be47-cc23-47f7-a3a1-fec900bb0119",
		Name:         "Kitesail Larcenist",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The ability that turns other permanents into Treasures when this creature enters isn't implemented — nothing happens when it arrives.",
		},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Kitesail Larcenist — ward {1}"),
		},
	})
}
