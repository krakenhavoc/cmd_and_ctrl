package game

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// granted_abilities.go — CR 113.10 and CR 613.1f: an effect that
// gives ANOTHER permanent an ability. ADR 0093 is the decision; this
// file is the seam it describes (the ADR's PR 1), and it is the whole
// of it outside three one-line hooks in the layer pass.
//
// # The shape
//
// A grant is a catalog BUNDLE (effects.AbilityGrant, the #665 bundle a
// copy's "except it has …" clause already names), registered under
// GrantKey. A layer-6 static that grants it declares the bundle's key
// in StaticAbility.GrantAbilities, and the engine appends one
// GrantedAbility — the key and the granting object — to the
// recipient's Characteristic in that static's timestamp slot. Nothing
// else is written anywhere: no closure on the recipient, no second
// field on Card, nothing persisted that a restore could not re-derive.
//
// # How a reader finds it
//
// There is still ONE lookup seam (ADR 0093 Decision 2). CatalogAbilityKey
// — the "what does this permanent DO" accessor every ability reader
// already goes through — composes the object's own key with its
// layered grants:
//
//	own                         (unchanged: no grants, the fast path)
//	own|grant:<a>|grant:<b>     (layered grants)
//	|grant:<a>                  (own abilities removed, face down, or
//	                             an uncatalogued card: the grant stays)
//
// and catalogDef answers the composite by merging each bundle into the
// card's definition, as it has since #665. So the trigger harvest, the
// cost-modifier gather and the rest see a granted ability with no
// change of their own. The two readers that DO change are the activated
// and mana ability accessors, because their rows carry an index the
// wire sends back and a granted row must be told apart from an own row
// — see ActivatedAbilitiesWithOrigins and ManaAbilitiesWithOrigins.
//
// # What "this creature" means
//
// The recipient. Every reader walks the RECIPIENT's key, so a granted
// ability is handed the host as its source everywhere: {T} taps the
// host, the host's controller activates it (CR 602.2), and CR 302.6's
// sickness check reads the host (ADR 0093 Decision 4).

// GrantedAbility is one layer-6 grant on an object: the bundle and the
// object that granted it.
type GrantedAbility struct {
	// Key is the bundle's catalog key, always in GrantKey form
	// ("grant:<name>").
	Key string
	// Source is the granting object. It is for the wire label and for
	// nothing else — the rules never ask who granted an ability, only
	// whether the object has it (CR 113.10).
	Source uuid.UUID
}

// GrantAbility appends one granted bundle to a characteristic. It is
// the ONE writer of Characteristic.GrantedAbilities, called by the
// layer engine for a static that declares StaticAbility.GrantAbilities.
// An empty key is ignored rather than recorded: a nameless grant would
// compose into a key that looks up nothing.
//
// Identical grants are NOT merged (ADR 0093 Decision 5, CR 113.2c): two
// Cryptolith Rites give a creature two instances of the ability.
func (c *Characteristic) GrantAbility(key string, source uuid.UUID) {
	key = GrantKey(key)
	if key == "" {
		return
	}
	c.GrantedAbilities = append(c.GrantedAbilities, GrantedAbility{Key: key, Source: source})
}

// layeredGrants returns the object's current layer-6 grants, or nil for
// every object that has none — which is every card off the battlefield
// (no effective characteristic) and nearly every card on it.
func layeredGrants(c Card) []GrantedAbility {
	if c.effective == nil {
		return nil
	}
	return c.effective.GrantedAbilities
}

// ownAbilityKey is the object's OWN abilities' catalog key: its printed
// and copy-granted abilities (CatalogKey), or "" when a CR 613.1f
// removal applies to it. Since ADR 0093 that is what AbilitiesRemoved
// means — "its own abilities are gone" — and a layered grant is not
// one of them.
func ownAbilityKey(c Card) string {
	if c.HasLostAllAbilities() {
		return ""
	}
	return CatalogKey(c)
}

// composeAbilityKey appends layered grant keys to an own key. The own
// key may be empty: an object with no text of its own still has the
// abilities other effects gave it (ADR 0093 Decision 2 §1), and the
// composite "|grant:<a>" is what says so.
func composeAbilityKey(own string, grants []GrantedAbility) string {
	if len(grants) == 0 {
		return own
	}
	var b strings.Builder
	b.Grow(len(own) + len(grants)*(len(grantKeySeparator)+len(GrantKeyPrefix)+24))
	b.WriteString(own)
	for _, g := range grants {
		if g.Key == "" {
			continue
		}
		b.WriteString(grantKeySeparator)
		b.WriteString(g.Key)
	}
	return b.String()
}

// AbilityKeyFromLKI is CatalogAbilityKey for a harvest that holds a
// last-known-information snapshot instead of a live permanent: the LTB
// harvest, whose permanent has already left the battlefield and lost its
// layer cache. CR 603.10a judges a leaves-the-battlefield trigger on
// what the permanent looked like while it was still there, so the own
// half is read against the snapshot's AbilitiesRemoved and the granted
// half is the snapshot's GrantedAbilities.
//
// `identity` supplies what the snapshot does not carry — the oracle ID,
// token key, face and copy grants the catalog is keyed by. This and
// CatalogAbilityKey are the only two places the composition is written.
func AbilityKeyFromLKI(identity Card, lki Characteristic) string {
	own := ""
	if !lki.AbilitiesRemoved {
		own = CatalogKey(identity)
	}
	return composeAbilityKey(own, lki.GrantedAbilities)
}

// --- ability refs (ADR 0093 Decision 5) -------------------------------

// Ability refs name ONE row of an object's activated or mana ability
// list in a way that survives the list changing under it. The wire
// indexes are positional, and a grant appearing or disappearing between
// the view and the announcement can move a row; a ref sent back beside
// the index lets the engine refuse the move rather than fire whatever
// now sits at that index (#544).
//
//	own:<i>                   the object's own ability i, counted in its
//	                          FULL declared list (so a designation gate
//	                          opening does not renumber it)
//	land:<colour>             a CR 305.6 intrinsic land ability
//	grant:<bundle>:<i>:<n>    ability i of the nth instance of that
//	                          granted bundle on this object
const (
	abilityRefOwn   = "own:"
	abilityRefLand  = "land:"
	abilityRefGrant = "grant:"
)

// OwnAbilityRef is the ref of an object's own ability at index i of its
// full declared list.
func OwnAbilityRef(i int) string { return abilityRefOwn + strconv.Itoa(i) }

// IntrinsicLandAbilityRef is the ref of a CR 305.6 intrinsic land
// ability, keyed by the colour it adds.
func IntrinsicLandAbilityRef(color string) string { return abilityRefLand + color }

// GrantedAbilityRef is the ref of ability i of the nth (0-based)
// instance of a granted bundle. `key` may be in either spelling.
func GrantedAbilityRef(key string, i, n int) string {
	return abilityRefGrant + strings.TrimPrefix(key, GrantKeyPrefix) + ":" + strconv.Itoa(i) + ":" + strconv.Itoa(n)
}

// AbilityOrigin says where one row of an ability list came from.
type AbilityOrigin struct {
	// Ref is the row's stable ref.
	Ref string
	// Grant is the bundle's catalog key for a granted row; "" for the
	// object's own and intrinsic abilities.
	Grant string
	// GrantedBy is the granting object for a granted row; uuid.Nil
	// otherwise.
	GrantedBy uuid.UUID
}

// Granted reports whether the row is a layer-6 grant.
func (o AbilityOrigin) Granted() bool { return o.Grant != "" }

// AbilityOrigins is a list of row origins, parallel to the ability list
// it describes. A NIL list is the fast path and means "every row is the
// object's own ability at that same index", which is what an object with
// no grant, no designation gate and no intrinsic land ability has.
type AbilityOrigins []AbilityOrigin

// Ref returns row i's ref.
func (o AbilityOrigins) Ref(i int) string {
	if o == nil {
		return OwnAbilityRef(i)
	}
	if i < 0 || i >= len(o) {
		return ""
	}
	return o[i].Ref
}

// At returns row i's origin; the zero-grant own origin on the fast path.
func (o AbilityOrigins) At(i int) AbilityOrigin {
	if o == nil || i < 0 || i >= len(o) {
		return AbilityOrigin{Ref: OwnAbilityRef(i)}
	}
	return o[i]
}

// staleAbilityRef reports whether an announcement's ref no longer names
// the row at its index. An absent ref is accepted — one release of
// client skew, ADR 0093 Decision 5 — and so is a ref that matches.
func staleAbilityRef(ref string, index, n int, origins AbilityOrigins) bool {
	if ref == "" {
		return false
	}
	if index < 0 || index >= n {
		return true
	}
	return origins.Ref(index) != ref
}

// ActivatedAbilityOrigins is ActivatedAbilitiesWithOrigins' origin list
// alone.
func ActivatedAbilityOrigins(c Card) AbilityOrigins {
	_, o := activatedAbilityRows(c, true)
	return o
}

// ActivatedAbilitiesWithOrigins returns the object's activated abilities
// — its own, then every layered grant's in layer-6 order — with the
// origin of each row (ADR 0093 Decision 5). Granted rows are always
// LAST, so a grant appearing never renumbers a row that was already
// there.
func ActivatedAbilitiesWithOrigins(c Card) ([]ActivatedAbilityShape, AbilityOrigins) {
	return activatedAbilityRows(c, true)
}

// activatedAbilityRows is the one body behind ActivatedAbilitiesForCard
// and ActivatedAbilitiesWithOrigins. With wantOrigins false it builds no
// origin list and, for an object with no grant, allocates nothing.
func activatedAbilityRows(c Card, wantOrigins bool) ([]ActivatedAbilityShape, AbilityOrigins) {
	var own []ActivatedAbilityShape
	var ownIdx []int
	// S24 layer 6: a permanent an ability-removing effect applies to
	// has none of its OWN activated abilities, and that has to be
	// checked BEFORE the instance-carried list as well as before the
	// catalog. A Clue token's "{2}, Sacrifice this: Draw a card" is an
	// activated ability like any other, and Darksteel Mutation takes it
	// away exactly as it takes away Sol Ring's. What it does NOT take
	// is a grant with a later timestamp (CR 613.6) — the layer pass has
	// already decided which of those survived, and they are below.
	if !c.HasLostAllAbilities() {
		switch {
		// S21 sub-PR 4: intrinsic abilities win — a token has no oracle
		// ID for the catalog to key on, and Food / Clue / Blood ARE
		// their activated ability. Not designation-gated: a token
		// carries its own abilities and has no catalog entry to print a
		// designation on.
		case len(c.ActivatedAbilities) > 0:
			own = c.ActivatedAbilities
		case CatalogActivatedAbilities != nil:
			// #521: an object with no entry has the empty key, whether
			// because it is uncatalogued or because CR 708.2a has
			// silenced it.
			if key := CatalogKey(c); key != "" {
				// ADR 0071: an ability gated on a designation the
				// permanent does not have is not on the permanent.
				own, ownIdx = activeOnlyIndexed(c, CatalogActivatedAbilities(key), func(a ActivatedAbilityShape) Designation {
					return a.ActiveWhen
				}, wantOrigins)
			}
		}
	}
	grants := layeredGrants(c)
	if len(grants) == 0 || CatalogActivatedAbilities == nil {
		if !wantOrigins || ownIdx == nil {
			return own, nil
		}
		return own, ownOrigins(len(own), ownIdx)
	}
	var out []ActivatedAbilityShape
	var origins AbilityOrigins
	for gi, gr := range grants {
		abs := CatalogActivatedAbilities(gr.Key)
		if len(abs) == 0 {
			continue
		}
		if out == nil {
			out = make([]ActivatedAbilityShape, 0, len(own)+len(abs))
			out = append(out, own...)
			if wantOrigins {
				origins = ownOrigins(len(own), ownIdx)
			}
		}
		n := grantOccurrence(grants, gi)
		for i, a := range abs {
			out = append(out, a)
			if wantOrigins {
				origins = append(origins, AbilityOrigin{
					Ref:       GrantedAbilityRef(gr.Key, i, n),
					Grant:     gr.Key,
					GrantedBy: gr.Source,
				})
			}
		}
	}
	if out == nil {
		// Every grant on this object is a bundle with no activated
		// ability (a granted trigger, a granted mana ability).
		if !wantOrigins || ownIdx == nil {
			return own, nil
		}
		return own, ownOrigins(len(own), ownIdx)
	}
	return out, origins
}

// ManaAbilityOrigins is ManaAbilitiesWithOrigins' origin list alone.
func ManaAbilityOrigins(c Card) AbilityOrigins {
	_, o := manaAbilityRows(c, true)
	return o
}

// ManaAbilitiesWithOrigins returns the object's mana abilities — its
// own declared abilities, then its CR 305.6 intrinsic land abilities,
// then every layered grant's in layer-6 order — with the origin of each
// row (ADR 0093 Decision 5).
func ManaAbilitiesWithOrigins(c Card) ([]ManaAbilityShape, AbilityOrigins) {
	return manaAbilityRows(c, true)
}

// manaAbilityRows is the one body behind ManaAbilitiesForCard and
// ManaAbilitiesWithOrigins.
func manaAbilityRows(c Card, wantOrigins bool) ([]ManaAbilityShape, AbilityOrigins) {
	var declared []ManaAbilityShape
	switch {
	// S24 layer 6: an ability-removing effect takes the DECLARED half
	// and leaves the INTRINSIC half, and the split is the whole of
	// CR 305.7. "Enchanted permanent is a colorless Forest land"
	// removes the abilities the permanent's rules text generated — Sol
	// Ring's "{T}: Add {C}{C}", a Signet's filter — and grants the mana
	// ability that comes with the new land type. That second half is
	// not printed on the card and is not in the catalog: it is derived
	// from the effective subtypes, below, by the same effect that did
	// the removing, so it survives on the other side of this switch
	// rather than being re-granted. And since ADR 0093 a THIRD half
	// survives too — a layer-6 grant, which is not the object's own.
	case c.HasLostAllAbilities():
		declared = nil
	// CR 708.2a: a face-down permanent has no text, so nothing it
	// carries declares a mana ability — CatalogKey is already silent,
	// and this closes the other door, a TOKEN's card-carried slice (a
	// Treasure turned face down by Cyber Conversion is not a Treasure).
	// A listed Forest still taps for {G}: that is the intrinsic half
	// below, read off the listed subtype (#1270).
	case c.FaceDownIsPermanent():
		declared = nil
	// S21 sub-PR 1: instance abilities win — a token has no oracle ID
	// for the catalog to key on.
	case len(c.ManaAbilities) > 0:
		declared = c.ManaAbilities
	// The OWN key, not CatalogAbilityKey: that one composes the
	// layered grants in too, and they are the third half below. And
	// CatalogKey, not c.OracleID: an MDFC back face keys on
	// "<oracle_id>#N" (#357).
	case CatalogManaAbilities != nil:
		declared = CatalogManaAbilities(CatalogKey(c))
	}
	intrinsic := intrinsicLandManaAbilities(c)
	grants := layeredGrants(c)
	var granted []ManaAbilityShape
	var grantedOrigins AbilityOrigins
	if len(grants) > 0 && CatalogManaAbilities != nil {
		for gi, gr := range grants {
			abs := CatalogManaAbilities(gr.Key)
			if len(abs) == 0 {
				continue
			}
			n := grantOccurrence(grants, gi)
			for i, a := range abs {
				granted = append(granted, a)
				if wantOrigins {
					grantedOrigins = append(grantedOrigins, AbilityOrigin{
						Ref:       GrantedAbilityRef(gr.Key, i, n),
						Grant:     gr.Key,
						GrantedBy: gr.Source,
					})
				}
			}
		}
	}
	// The fast path, and the pre-ADR-0093 answer exactly: no intrinsic
	// half and no granted half.
	if len(intrinsic) == 0 && len(granted) == 0 {
		return declared, nil
	}
	// Still no grant and no origins wanted: hand back the half that is
	// there without allocating, as the accessor always has.
	if len(granted) == 0 && !wantOrigins {
		if len(declared) == 0 {
			return intrinsic, nil
		}
	}
	// A declared ability and an intrinsic one can name the same
	// colour — every catalog dual land ("{T}: Add {B} or {G}") is
	// printed with the land types that would have produced the same
	// mana, and doubling it up would put two ways to make {B} in the
	// client's ability row. Keep the declared shape (it may carry a
	// rider or a pipe) and add only colours it cannot make. The
	// dedupe is against the OWN half only: a granted any-colour
	// ability is a different ability, and a Forest under Chromatic
	// Lantern has both (ADR 0093 Decision 5).
	covered := producibleColors(declared)
	out := make([]ManaAbilityShape, 0, len(declared)+len(intrinsic)+len(granted))
	var origins AbilityOrigins
	if wantOrigins {
		origins = make(AbilityOrigins, 0, cap(out))
	}
	for i, ab := range declared {
		out = append(out, ab)
		if wantOrigins {
			origins = append(origins, AbilityOrigin{Ref: OwnAbilityRef(i)})
		}
	}
	for _, ab := range intrinsic {
		color := landTypeColorOf(ab)
		if len(declared) > 0 && covered[color] {
			continue
		}
		out = append(out, ab)
		if wantOrigins {
			origins = append(origins, AbilityOrigin{Ref: IntrinsicLandAbilityRef(color)})
		}
	}
	out = append(out, granted...)
	if wantOrigins {
		origins = append(origins, grantedOrigins...)
	}
	return out, origins
}

// grantOccurrence is how many earlier entries in `grants` name the same
// bundle as grants[i] — the <n> of a granted ref.
func grantOccurrence(grants []GrantedAbility, i int) int {
	n := 0
	for j := 0; j < i; j++ {
		if grants[j].Key == grants[i].Key {
			n++
		}
	}
	return n
}

// ownOrigins builds the origin list for an own-only ability list,
// mapping each row back to its index in the full declared list when a
// designation gate filtered it (ownIdx), or 1:1 when nothing did.
func ownOrigins(n int, ownIdx []int) AbilityOrigins {
	out := make(AbilityOrigins, n)
	for i := range out {
		k := i
		if ownIdx != nil {
			k = ownIdx[i]
		}
		out[i] = AbilityOrigin{Ref: OwnAbilityRef(k)}
	}
	return out
}

// activeOnlyIndexed is activeOnly that can also report, for a filtered
// list, each survivor's index in the input. The index list is nil when
// nothing was filtered or when it was not asked for.
func activeOnlyIndexed[T any](c Card, all []T, gate func(T) Designation, wantIdx bool) ([]T, []int) {
	gated := false
	for i := range all {
		if gate(all[i]).IsGate() {
			gated = true
			break
		}
	}
	if !gated {
		return all, nil
	}
	out := make([]T, 0, len(all))
	var idx []int
	for i := range all {
		if gate(all[i]).Active(c) {
			out = append(out, all[i])
			if wantIdx {
				idx = append(idx, i)
			}
		}
	}
	if len(out) == 0 {
		return nil, nil
	}
	return out, idx
}

// --- the wire's text ---------------------------------------------------

// GrantedAbilityInfo is one ability another effect gave an object, as
// the view projects it (ADR 0093 Decision 8): the bundle's printed
// text and who granted it. Granted TRIGGERS have no ability row, so
// this is the only place a player sees them.
type GrantedAbilityInfo struct {
	// Key is the bundle's catalog key.
	Key string
	// Text is the bundle's printed text (effects.AbilityGrant.Text).
	Text string
	// Source is the granting object; uuid.Nil for a copy grant
	// (CR 707.9a), which is part of the object's own copiable values
	// and has no granting object on the board.
	Source uuid.UUID
}

// GrantedAbilitiesOf lists the granted abilities an object has right
// now: its copy grants first (they are part of its own abilities, so a
// removal takes them too), then its layered grants in layer-6 order.
// A bundle with no registered text is skipped — there is nothing to
// show — and so is every grant on an object whose own abilities are
// gone, for the copy half.
func GrantedAbilitiesOf(c Card) []GrantedAbilityInfo {
	grants := layeredGrants(c)
	copyGrants := c.GrantedAbilities
	if c.HasLostAllAbilities() || c.FaceDownIsPermanent() {
		copyGrants = nil
	}
	if len(grants) == 0 && len(copyGrants) == 0 {
		return nil
	}
	var out []GrantedAbilityInfo
	for _, k := range copyGrants {
		key := GrantKey(k)
		if text := GrantTextFor(key); text != "" {
			out = append(out, GrantedAbilityInfo{Key: key, Text: text})
		}
	}
	for _, gr := range grants {
		if text := GrantTextFor(gr.Key); text != "" {
			out = append(out, GrantedAbilityInfo{Key: gr.Key, Text: text, Source: gr.Source})
		}
	}
	return out
}

// GrantTextFor is a bundle's printed text, or "" when the key names no
// registered bundle.
func GrantTextFor(key string) string {
	if d := catalogDef(GrantKey(key)); d != nil {
		return d.GrantText
	}
	return ""
}
