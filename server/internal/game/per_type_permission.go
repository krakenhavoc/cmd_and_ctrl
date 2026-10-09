package game

import (
	"slices"
	"sort"

	"github.com/google/uuid"
)

// per_type_permission.go — #2167, ADR 0066's 2026-10-09 amendment: a
// cast permission spent once per CARD TYPE rather than once per cast.
//
//	Muldrotha, the Gravetide: "During each of your turns, you may play
//	a land and cast a permanent spell of each permanent type from your
//	graveyard. (If a card has multiple permanent types, choose one as
//	you play it.)"
//
//	Aminatou's Augury: "Until end of turn, for each nonland card type,
//	you may cast a spell of that type from among the exiled cards
//	without paying its mana cost."
//
// CastsLeft (#1729) is one flat count, and nothing recorded which type a
// cast through a permission used, so "one of each" could not be said.
// This file says it in three parts:
//
//  1. THE BUDGET is data on the permission: CastPermission.PerType, the
//     card types it opens one use of each, and PerTypeUsed, the ones
//     spent. CoversCard reads both, so a permission whose every type
//     the card could use is spent does not open the card — and the cast
//     path, the bot enumerator and the view, which all ask CoversCard,
//     agree without a second check.
//
//  2. THE CHOICE is the caster's, made as the card is played (the
//     reminder text; Augury's ruling of 2018-07-13 says the same). It
//     is an announce-time parameter, CastSpellParams.PermissionType,
//     judged against the card AS IT IS PLAYED OR CAST — the face being
//     cast, face down if it is cast face down — because Muldrotha's
//     ruling says the type is read off the spell, not the card in the
//     graveyard. A cast with one possible type needs no answer.
//
//  3. WHERE THE SPENT TYPES LIVE depends on where the permission lives.
//     A stored permission (Augury's) carries them itself. A derived one
//     (Muldrotha's) is rebuilt from the battlefield on every query and
//     cannot remember anything, so they are written to
//     TurnTally.PermissionTypes under (holder, granting object, type)
//     and copied back in by the derivation. The turn tally empties as
//     each turn begins, which is "during each of your turns"; the
//     object in the key is CR 400.7, so a new Muldrotha grants a fresh
//     set (ruling 2020-11-10); the holder in the key is control, so a
//     Muldrotha that changes hands opens its new controller's own set.

// The card types a per-type budget may name (CR 205.2a), lowercase as
// Card.HasCardType reads them. Only the ones a cast or a play can spend
// are listed; a budget naming anything else names nothing.
const (
	PermissionTypeArtifact     = "artifact"
	PermissionTypeBattle       = "battle"
	PermissionTypeCreature     = "creature"
	PermissionTypeEnchantment  = "enchantment"
	PermissionTypeInstant      = "instant"
	PermissionTypeKindred      = "kindred"
	PermissionTypeLand         = "land"
	PermissionTypePlaneswalker = "planeswalker"
	PermissionTypeSorcery      = "sorcery"
)

// PermanentPermissionTypes are CR 110.4's six permanent types, in CR
// 205.2a's order — Muldrotha's budget.
var PermanentPermissionTypes = []string{
	PermissionTypeArtifact, PermissionTypeBattle, PermissionTypeCreature,
	PermissionTypeEnchantment, PermissionTypeLand, PermissionTypePlaneswalker,
}

// NonlandPermissionTypes are the card types a spell can have, in CR
// 205.2a's order — Aminatou's Augury's "each nonland card type".
// Conspiracy, dungeon, phenomenon, plane, scheme and vanguard are card
// types too, and none of them is ever cast.
var NonlandPermissionTypes = []string{
	PermissionTypeArtifact, PermissionTypeBattle, PermissionTypeCreature,
	PermissionTypeEnchantment, PermissionTypeInstant, PermissionTypeKindred,
	PermissionTypePlaneswalker, PermissionTypeSorcery,
}

// PermissionTypeArticle spells a type with its article for a label:
// "an artifact", "a creature".
func PermissionTypeArticle(t string) string {
	switch t {
	case PermissionTypeArtifact, PermissionTypeEnchantment, PermissionTypeInstant:
		return "an " + t
	}
	return "a " + t
}

// PermissionTypeChoices are the types a play or cast of `c` could spend
// under this permission's budget, in the budget's order: the types the
// card has that are in the budget and not yet spent. `c` is the card AS
// IT IS PLAYED OR CAST — the caller has set the face and, for a face-down
// cast, stamped it.
//
// A LAND is played, never cast (CR 305.1), so the one type it can spend
// is "land", whatever else it is: an artifact land played under
// Muldrotha uses the land play. A spell is never a land, so a spell
// cannot spend "land".
//
// Nil for a permission with no budget, and for a card the budget cannot
// open at all.
func (p *CastPermission) PermissionTypeChoices(c Card) []string {
	if p == nil || len(p.PerType) == 0 {
		return nil
	}
	land := c.IsLand()
	var out []string
	for _, t := range p.PerType {
		if slices.Contains(p.PerTypeUsed, t) || slices.Contains(out, t) {
			continue
		}
		switch {
		case land:
			if t == PermissionTypeLand {
				out = append(out, t)
			}
		case t != PermissionTypeLand && c.HasCardType(t):
			out = append(out, t)
		}
	}
	return out
}

// perTypeOpens is CoversCard's budget half: true for a permission with
// no budget, and otherwise when at least one face this permission lets
// its holder play has a type left to spend. Asked of every face rather
// than the one the card is showing because CR 712.11c judges the face
// that will be cast: a creature whose Adventure is a sorcery still opens
// under Augury once its creature use is spent.
//
// Pure, like the rest of CoversCard: a derived permission arrives with
// its spent types already copied in (standingCastPermissionsLocked).
func (p *CastPermission) perTypeOpens(c Card) bool {
	if p == nil || len(p.PerType) == 0 {
		return true
	}
	for _, face := range CastableFacesUnder(c, p, p.Player) {
		played := c
		played.SetFace(face)
		if len(p.PermissionTypeChoices(played)) > 0 {
			return true
		}
		// ADR 0141, CR 702.103b: a face cast bestowed is an Aura
		// enchantment spell, so it may spend "enchantment" once its
		// creature use is gone (Muldrotha's ruling of 2020-11-10).
		if hasBestowOffer(played) {
			played.Bestowed = true
			if len(p.PermissionTypeChoices(played)) > 0 {
				return true
			}
		}
	}
	return false
}

// hasBestowOffer reports whether the card prints a bestow cost.
func hasBestowOffer(c Card) bool {
	for _, ac := range AlternativeCostsFor(CatalogKey(c)) {
		if ac.Bestow {
			return true
		}
	}
	return false
}

// permissionTypeTallyKey is TurnTally.PermissionTypes' key for one type
// a derived permission has spent: whose permission, which granting
// object (CR 400.7), which type.
func permissionTypeTallyKey(holder, source uuid.UUID, epoch int, typ string) string {
	return holder.String() + "/" + ObjectTallyKey(source, epoch, typ)
}

// fillDerivedPerTypeUsedLocked copies the types this turn's tally says a
// derived permission has spent onto the freshly stamped permission, so
// CoversCard reads it exactly as it reads a stored one. `source` is the
// granting object.
//
// Caller must hold g.mu.
func (g *Game) fillDerivedPerTypeUsedLocked(perm *CastPermission, holder uuid.UUID, source *Card) {
	if len(perm.PerType) == 0 || source == nil {
		return
	}
	perm.PerTypeUsed = nil
	for _, t := range perm.PerType {
		if g.TurnTally.PermissionTypes[permissionTypeTallyKey(holder, source.InstanceID, source.ObjectEpoch, t)] > 0 {
			perm.PerTypeUsed = append(perm.PerTypeUsed, t)
		}
	}
}

// settlePermissionTypeLocked is the announce-time half of the budget:
// which type this play or cast spends. `card` is the card as it is
// played or cast; `asked` is the caster's answer, empty when they gave
// none. A cast with exactly one possible type needs no answer. One with
// several needs one, because the reminder text makes it the player's
// choice and the engine picking would spend a type they meant to keep.
//
// Returns "" with no error for a permission with no budget.
func settlePermissionTypeLocked(grant *CastPermission, card Card, asked string) (string, error) {
	if grant == nil || len(grant.PerType) == 0 {
		if asked != "" {
			return "", ErrPermissionTypeNotOffered
		}
		return "", nil
	}
	choices := grant.PermissionTypeChoices(card)
	if len(choices) == 0 {
		return "", ErrPermissionTypeSpent
	}
	if asked == "" {
		if len(choices) == 1 {
			return choices[0], nil
		}
		return "", ErrPermissionTypeRequired
	}
	if !slices.Contains(choices, asked) {
		return "", ErrPermissionTypeNotOffered
	}
	return asked, nil
}

// spendPermissionTypeLocked spends `typ` of the budget the play or cast
// of `card` just used. `card` is the card as it sat in its source zone,
// epoch and all, because a stored permission names that object.
//
// A derived permission writes the turn tally against the object that
// grants it, read now: the granting permanent is on the battlefield,
// which is why the permission exists at all. A stored one is found as
// consumeLimitedGrantLocked finds a limited one — same source, label and
// claim, and it still covers the card — and gets a fresh slice, because
// the backing array is shared with the undo snapshots (the reason
// sweepCastPermissionsLocked gives). One whose whole budget is spent is
// dropped: it opens nothing any more.
//
// Caller must hold g.mu (write).
func (g *Game) spendPermissionTypeLocked(holder uuid.UUID, card Card, grant *CastPermission, typ string) {
	if grant == nil || typ == "" || len(grant.PerType) == 0 {
		return
	}
	if grant.Scope == ScopeStanding {
		key := permissionTypeTallyKey(holder, grant.Source, g.objectEpochLocked(grant.Source), typ)
		if g.TurnTally.PermissionTypes == nil {
			g.TurnTally.PermissionTypes = map[string]int{}
		}
		g.TurnTally.PermissionTypes[key]++
		return
	}
	p := g.playerByIDLocked(holder)
	if p == nil || len(p.CastPermissions) == 0 {
		return
	}
	kept := make([]CastPermission, 0, len(p.CastPermissions))
	spent := false
	for _, perm := range p.CastPermissions {
		if !spent && len(perm.PerType) > 0 && perm.Source == grant.Source && perm.Label == grant.Label &&
			perm.AltCostKey == grant.AltCostKey && perm.CoversCard(card, perm.Zone) {
			spent = true
			perm.PerTypeUsed = append(append([]string(nil), perm.PerTypeUsed...), typ)
			if len(perm.PerTypeUsed) >= len(perm.PerType) {
				continue
			}
		}
		kept = append(kept, perm)
	}
	if !spent {
		return
	}
	if len(kept) == 0 {
		kept = nil
	}
	p.CastPermissions = kept
}

// PermissionTypeOptionsLocked is the bot enumerator's and the view's
// question: the types a play or cast of `card` out of `zone` under
// `offer` may announce, ranked by RankPermissionTypesLocked. `card` has
// its face set and nothing else, as CastSpell's copy has when it reads
// the grant; the offer's own changes to what the spell is (a face-down
// cast, a disturbed back face) are applied here, as CastSpell applies
// them before it settles the type.
//
// Three answers:
//
//   - ([""], true): no budget applies — the permission keeps none, or
//     the card's own text opens the zone and the cast spends nothing.
//     Announce no type.
//   - (types, true): announce one of these. One entry needs no picker.
//   - (nil, false): a budget applies and this play or cast has no type
//     left to spend. CastSpell refuses it, so offer nothing.
//
// Caller must hold g.mu.
func (g *Game) PermissionTypeOptionsLocked(holder uuid.UUID, card Card, zone ZoneKind, grant *CastPermission, offer *AlternativeCost) ([]string, bool) {
	if grant == nil || len(grant.PerType) == 0 || !g.castUsesGrantLocked(card, zone, offer) {
		return []string{""}, true
	}
	played := card
	if offer != nil && offer.Bestow {
		played.Bestowed = true
	}
	if offer != nil && offer.FaceDown != nil {
		played.SetFaceDown(offer.FaceDown.Kind)
	}
	played = offer.CastFaceOf(played)
	out := g.RankPermissionTypesLocked(holder, grant, played)
	return out, len(out) > 0
}

// RankPermissionTypesLocked orders the types a play or cast of `card`
// could spend so the one the REST of the permission's cards need least
// comes first: an artifact creature cast under Muldrotha with another
// creature in the graveyard and no other artifact should be the
// artifact. It is advice, not a rule — the bot takes the first, and the
// client opens its picker on it — so ties keep the budget's order.
//
// "The rest" is every other card the permission still opens: a standing
// graveyard permission's other cards in its holder's graveyard, a stored
// one's other named objects. Each counts once for each type it could
// spend.
//
// Caller must hold g.mu.
func (g *Game) RankPermissionTypesLocked(holder uuid.UUID, grant *CastPermission, card Card) []string {
	choices := grant.PermissionTypeChoices(card)
	if len(choices) < 2 {
		return choices
	}
	need := make(map[string]int, len(choices))
	for _, other := range g.perTypeOthersLocked(holder, grant, card.InstanceID) {
		for _, face := range CastableFacesUnder(other, grant, holder) {
			played := other
			played.SetFace(face)
			for _, t := range grant.PermissionTypeChoices(played) {
				need[t]++
			}
		}
	}
	out := append([]string(nil), choices...)
	sort.SliceStable(out, func(i, j int) bool { return need[out[i]] < need[out[j]] })
	return out
}

// perTypeOthersLocked lists the cards besides `except` the permission
// still opens, for RankPermissionTypesLocked.
//
// Caller must hold g.mu.
func (g *Game) perTypeOthersLocked(holder uuid.UUID, grant *CastPermission, except uuid.UUID) []Card {
	var pile []Card
	if grant.Scope == ScopeStanding {
		if z := g.permissionZoneLocked(grant.PileOwnerFor(holder), grant.Zone); z != nil {
			pile = z.Cards
		}
	} else {
		for _, ref := range grant.Cards {
			if z := g.findCardZoneLocked(ref.ID); z != nil && z.Kind == grant.Zone {
				for _, c := range z.Cards {
					if c.InstanceID == ref.ID {
						pile = append(pile, c)
					}
				}
			}
		}
	}
	var out []Card
	for _, c := range pile {
		if c.InstanceID != except && grant.CoversCard(c, grant.Zone) {
			out = append(out, c)
		}
	}
	return out
}
