package game

import "github.com/google/uuid"

// autotap_exclusions.go — #1242: what an announcement has ALREADY
// SPENT, and therefore what the auto-tapper must not also spend on the
// mana half of the same payment.
//
// The rule is CR 118.3 read from the planner's side: one object pays
// one cost component once. It has held for the {T} of an ability since
// S15 (Treasure Vault's own "{T}: Add {C}" must not fund its
// "{X}{X}, {T}") and for convoke and Springleaf Drum's tapped creatures
// since S22 and #758. #1242 made it matter for two more components,
// because the planner may now CRACK a source that asks for no {T}:
//
//   - a SACRIFICE. An Eldrazi Spawn named to Village Rites' "sacrifice
//     a creature" is also a mana source, and auto-tap runs BEFORE the
//     additional cost is paid — so a plan could eat the Spawn and the
//     sacrifice would then find nothing to pay with, after the mana had
//     been made. The same was true of a Treasure named to Deadly
//     Dispute since #1215; sacrifice-only sources made it the common
//     case rather than a corner.
//   - a DISCARD. Since #1228 a Spirit Guide in hand is a mana source,
//     and one named to Thrill of Possibility's discard could be exiled
//     for mana first.
//
// Two helpers because a cast and an activation name their payments on
// different params structs; each is the ONE list its engine path and
// the legal-move enumerator both build from, so the enumerator never
// offers a move whose affordability leaned on a source the engine then
// refuses to plan (#544).

// CastAutoTapExclusions is the set the auto-tapper may not reach for
// while paying a cast's mana: the lock-tap reservations, the permanents
// tapped for convoke / waterbend, and the permanents and cards named to
// the additional cost's sacrifice and discard. Nil when empty.
func CastAutoTapExclusions(params CastSpellParams) map[uuid.UUID]bool {
	return unionIDs(params.LockedSources, params.TapIDs, params.SacrificeIDs, params.DiscardIDs)
}

// AbilityAutoTapExclusions is the same set for a CR 602 activation's
// mana component: the source when the cost taps, sacrifices or exiles
// it, and
// the permanents and cards named to its TapOthers, sacrifice, discard
// and exile-N-cards components. Nil when empty.
//
// #1297: the exile ids are here for the hand form. A Spirit Guide named
// to Holistic Wisdom's "Exile a card from your hand" is also a mana
// source (#1228), and the planner would otherwise exile it for mana
// first and leave the cost with nothing to pay. No card in a GRAVEYARD
// is a mana source today (supportedManaAbilityZones is the hand alone),
// so for the graveyard form this is the list being right in advance
// rather than a live exclusion.
//
// #1404: and the source when the cost EXILES it. From the battlefield
// that is the sacrifice-this reason exactly: a permanent that is also
// a mana source — a land with "{1}, Exile this land:", a mana rock
// with an exile-to-shuffle ability — could otherwise be cracked by the
// planner for the mana half (a Treasure-shaped sacrifice-for-mana
// source is plannable since #1242), and the exile would then find
// nothing to pay with, after the mana had been made. From the
// graveyard it is the list being right in advance, as above.
func AbilityAutoTapExclusions(sourceID uuid.UUID, cost AbilityCost, tapIDs, sacrificeIDs, discardIDs, exileIDs []uuid.UUID) map[uuid.UUID]bool {
	var self []uuid.UUID
	if cost.Tap || cost.SacrificeSelf || cost.ExileSelf {
		self = []uuid.UUID{sourceID}
	}
	return unionIDs(self, tapIDs, sacrificeIDs, discardIDs, exileIDs)
}

// ActivationAutoTapExclusions is the WHOLE set ActivateCatalogAbility's
// auto-tap may not spend on the mana half of an activation announced
// with `params`: AbilityAutoTapExclusions over the named components,
// plus the permanents tapped to waterbend (#1310 — a Birds of Paradise
// named to Katara's waterbend cannot also make the mana that pays the
// rest). Nil when empty.
//
// #1422: ONE function for the activation and the auto-tap preview's
// ?ability= branch, so the plan the preview shows is the plan the
// payment makes. The preview used to exclude only the lock-tap
// reservations and could plan the ability's own {T} source for mana.
func ActivationAutoTapExclusions(sourceID uuid.UUID, cost AbilityCost, params ActivateAbilityParams) map[uuid.UUID]bool {
	return WithAutoTapExclusions(
		AbilityAutoTapExclusions(sourceID, cost, params.TapIDs, params.SacrificeIDs, params.DiscardIDs, params.ExileIDs),
		params.WaterbendIDs)
}

// WithAutoTapExclusions returns `base` plus `ids`, copying rather than
// writing into `base` — the legal-move enumerator keeps one base set
// per ability and widens it per payment it offers.
func WithAutoTapExclusions(base map[uuid.UUID]bool, ids ...[]uuid.UUID) map[uuid.UUID]bool {
	n := 0
	for _, l := range ids {
		n += len(l)
	}
	if n == 0 {
		return base
	}
	out := make(map[uuid.UUID]bool, len(base)+n)
	for id := range base {
		out[id] = true
	}
	for _, l := range ids {
		for _, id := range l {
			out[id] = true
		}
	}
	return out
}

func unionIDs(lists ...[]uuid.UUID) map[uuid.UUID]bool {
	return WithAutoTapExclusions(nil, lists...)
}
