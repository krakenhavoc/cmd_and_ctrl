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
//
// #1703: and the creatures tapped for teamwork — a Llanowar Elves
// named to HULK SMASH!'s teamwork cannot also tap for its {R} — and the
// creature named to a blight, which a sacrifice-for-mana plan (an
// Eldrazi Spawn) would otherwise eat before the counters could land.
//
// #1727: and the cards and permanents named to the ALTERNATIVE cost's
// card component (AltCostIDs). A sacrifice alternative cost has no mana
// of its own, but a cost increase gives it some — Dread Return
// flashed back under Thalia owes {1} — and an Eldrazi Spawn named to
// "sacrifice three creatures" must not be cracked for that {1} first.
// The same was already true, and unexercised, of a Spirit Guide named
// to a pitch cost.
//
// ADR 0135 §4 (owner decision 3): a permanent named to a SACRIFICE cost
// on the cast — the additional cost's SacrificeIDs, and AltCostIDs when
// `alt` is a sacrifice offer (emerge, Dread Return's flashback,
// Fireblast) — is in the set as SACRIFICE-NAMED rather than excluded:
// the planner may use its mana abilities that do not sacrifice it, and
// none that does. CR 601.2g has mana abilities activated before CR
// 601.2h pays the costs, and the emerge rulings say so outright ("The
// creature chosen to be sacrificed is still on the battlefield … as you
// activate mana abilities to cast the emerge spell"), so a Llanowar
// Elves named to Village Rites may tap for the {B}'s neighbour and then
// be sacrificed. An Eldrazi Spawn named to emerge may not be cracked
// first: that would leave the sacrifice nothing to pay with. Every other
// component stays excluded outright. `alt` is the offer the cast claims,
// nil for none.
func CastAutoTapExclusions(params CastSpellParams, alt *AlternativeCost) map[uuid.UUID]bool {
	full := [][]uuid.UUID{params.LockedSources, params.TapIDs, params.DiscardIDs,
		params.TeamworkIDs, params.BlightIDs}
	sacrificed := [][]uuid.UUID{params.SacrificeIDs}
	if alt != nil && alt.Sacrifice != nil {
		sacrificed = append(sacrificed, params.AltCostIDs)
	} else {
		full = append(full, params.AltCostIDs)
	}
	return WithSacrificeNamedExclusions(unionIDs(full...), sacrificed...)
}

// castAutoTapExclusionsLocked is CastAutoTapExclusions for an
// announcement whose offer is still a key on `params`: the offer is
// resolved the way the pricer resolves it. Caller must hold g.mu.
func (g *Game) castAutoTapExclusionsLocked(playerID uuid.UUID, card Card, params CastSpellParams) map[uuid.UUID]bool {
	return CastAutoTapExclusions(params, g.claimedAltCostLocked(playerID, card, params))
}

// CastAutoTapExclusionsFor is castAutoTapExclusionsLocked under the read
// lock, for the auto-tap preview, so the preview plans exactly what
// CastSpell's auto-tap will.
func (g *Game) CastAutoTapExclusionsFor(playerID uuid.UUID, card Card, params CastSpellParams) map[uuid.UUID]bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.castAutoTapExclusionsLocked(playerID, card, params)
}

// An exclusion set maps a permanent or card to TRUE when the planner may
// not reach for it at all, and to FALSE when it is SACRIFICE-NAMED (ADR
// 0135 §4): the planner may still use its mana abilities that do not
// sacrifice it (autoTapMayUse). Absent is free. Every reader that copies
// a set keeps the value (WithAutoTapExclusions, MergeAutoTapExclusions),
// and a permanent that is both is excluded outright.

// WithSacrificeNamedExclusions returns `base` plus each of `ids` as
// sacrifice-named, copying rather than writing into `base`. An ID
// `base` already excludes outright stays excluded.
func WithSacrificeNamedExclusions(base map[uuid.UUID]bool, ids ...[]uuid.UUID) map[uuid.UUID]bool {
	n := 0
	for _, l := range ids {
		n += len(l)
	}
	if n == 0 {
		return base
	}
	out := make(map[uuid.UUID]bool, len(base)+n)
	for id, full := range base {
		out[id] = full
	}
	for _, l := range ids {
		for _, id := range l {
			if _, ok := out[id]; !ok {
				out[id] = false
			}
		}
	}
	return out
}

// MergeAutoTapExclusions writes `src` into `dst` (which must be
// non-nil), keeping each entry's meaning: outright wins over
// sacrifice-named.
func MergeAutoTapExclusions(dst, src map[uuid.UUID]bool) {
	for id, full := range src {
		if cur, ok := dst[id]; !ok || (full && !cur) {
			dst[id] = full
		}
	}
}

// autoTapMayUse is the planner's question about one candidate: may it
// plan `ab` on the permanent or card `id`? Not when the set excludes it
// outright; when it is sacrifice-named, only an ability that does not
// sacrifice it (ADR 0135 §4).
func autoTapMayUse(excluded map[uuid.UUID]bool, id uuid.UUID, ab *ManaAbilityShape) bool {
	full, named := excluded[id]
	if !named {
		return true
	}
	if full {
		return false
	}
	return ab == nil || !ab.SacrificeCost
}

// AbilityAutoTapExclusions is the same set for a CR 602 activation's
// mana component: the source when the cost taps, sacrifices, exiles
// or returns it (#2028), and
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
	if cost.Tap || cost.SacrificeSelf || cost.ExileSelf || cost.ReturnSelf {
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
//
// ADR 0109 §7: and the hand cards put on top of the library or drawn
// for a random discard. A Spirit Guide named to Penance's cost, or
// drawn for Pyromancy's, is the payment and can't also be exiled for
// mana (CR 118.3), exactly as one named to a discard can't.
//
// #1600: and the permanents named to an exile-a-permanent cost, for the
// sacrifice's reason (#1242): an Eldrazi Spawn named to The Soul
// Stone's "Exile a creature you control" is also a sacrifice-for-mana
// source, and a plan that cracked it for the {6}{B} would leave the
// exile with nothing to pay. A mana creature that merely TAPS for the
// mana and is then exiled would be a legal paper line (CR 601.2g before
// 601.2h); refusing it is the weaker-than-printed direction, and the
// player can float the mana by hand first.
func ActivationAutoTapExclusions(sourceID uuid.UUID, cost AbilityCost, params ActivateAbilityParams) map[uuid.UUID]bool {
	return WithAutoTapExclusions(
		AbilityAutoTapExclusions(sourceID, cost, params.TapIDs, params.SacrificeIDs, params.DiscardIDs, params.ExileIDs),
		params.WaterbendIDs, params.TopIDs, params.randomDiscards, params.ExilePermanentIDs)
}

// ManaActivationAutoTapExclusions is the set a MANA ability's own
// auto-tap (#2215, ManaAbilityParams.AutoTap) may not spend on the
// ability's mana component: the SOURCE, always, and every permanent and
// card another component of the cost names (its sacrifice, tap-another,
// discard, exile and exile-a-permanent picks).
//
// The source is excluded whatever the cost prints. With a {T} it is
// CR 602.2b's reason, as on a CR 602 activation: the tap and the mana
// are components of one cost. Without one, the only way the source
// could pay is through another of its own mana abilities, and a mana
// ability funding itself through its own permanent is not a payment a
// player would want planned behind their back. ONE function for the
// activation and the legal-move enumerator, so the move list never
// offers an activation the engine then cannot fund.
func ManaActivationAutoTapExclusions(sourceID uuid.UUID, params ManaAbilityParams) map[uuid.UUID]bool {
	return unionIDs([]uuid.UUID{sourceID}, params.SacrificeIDs, params.TapIDs, params.DiscardIDs,
		params.ExileIDs, params.ExilePermanentIDs)
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
	// ADR 0135 §4: a copy keeps a sacrifice-named entry's meaning, and an
	// ID named here is excluded outright whatever it was.
	for id, full := range base {
		out[id] = full
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

// excludeNonQualifyingSourcesLocked is `excluded` plus every source of
// `controller`'s that could not make mana a source-restricted payment
// accepts (#2556): "Spend only mana produced by basic lands to cast
// this spell" (Imperiosaur) is a rule about the SOURCE, and the planner
// plans from permanents, so a source whose kinds miss `only` is simply
// not reachable — the same exclusion list a convoked creature or a
// sacrificed Treasure gets, which every gatherer already honours.
//
// The kinds are read the way a mint site reads them (manaSourceKindsOf),
// so a permanent the planner skips here is exactly one whose mana the
// pool solver would then refuse. `only` zero returns `excluded`
// unchanged (nil stays nil): the walk is paid by the few casts that
// print the clause. Never mutates its argument. Caller must hold g.mu.
func (g *Game) excludeNonQualifyingSourcesLocked(controller uuid.UUID, excluded map[uuid.UUID]bool, only ManaSourceKinds) map[uuid.UUID]bool {
	if only == 0 {
		return excluded
	}
	out := make(map[uuid.UUID]bool, len(excluded))
	for id, full := range excluded {
		out[id] = full
	}
	skip := func(c Card) {
		if !manaSourceKindsOf(c).HasAny(only) {
			out[c.InstanceID] = true
		}
	}
	if g.Battlefield != nil {
		for _, c := range g.Battlefield.Cards {
			if c.Controller == controller {
				skip(c)
			}
		}
	}
	if p := g.playerByIDLocked(controller); p != nil {
		for _, kind := range supportedManaAbilityZones {
			if pile := playerManaZone(p, kind); pile != nil {
				for _, c := range pile.Cards {
					skip(c)
				}
			}
		}
	}
	return out
}
