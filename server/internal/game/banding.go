package game

import (
	"errors"
	"slices"

	"github.com/google/uuid"
)

// banding.go — banding (CR 702.22, ADR 0224, #5723).
//
// Banding is a static ability that changes three things about combat,
// and this file is the one home of all three:
//
//   - ATTACKING IN A BAND. As the active player declares attackers they
//     may announce that one or more attacking creatures with banding and
//     up to one attacking creature without it are all in a band
//     (CR 508.1e, 702.22c). Every creature in a band attacks the same
//     player, planeswalker or battle (CR 702.22d). A band lasts for the
//     rest of combat even if a member loses banding (CR 702.22e), and a
//     member removed from combat leaves it (CR 702.22f). The band is
//     announced on declare_attackers: entries that share a `band` number
//     form one band (AttackDeclaration.Band), validated as a whole and
//     refused whole with a *BandRefusedError.
//   - BLOCKING A BAND. A creature that blocks one member of a band blocks
//     every member (CR 702.22h), and an effect that blocks one member
//     blocks the band (CR 702.22i). The block lock-in widens each
//     blocker's set to the band (expandBandBlocksLocked), so the rest of
//     the engine sees an ordinary creature blocking several attackers
//     (multi_block.go): every member deals its damage to the blocker, and
//     the blocker divides its damage among the members.
//   - WHO DIVIDES THE DAMAGE. An attacker blocked by a creature with
//     banding has its combat damage divided by the DEFENDING player among
//     the creatures blocking it (CR 702.22j), and a blocker blocking a
//     creature with banding has its damage divided by the ACTIVE player
//     among the creatures it blocks (CR 702.22k). Both read banding as the
//     damage is assigned, not as the band was formed, and both reuse the
//     CR 510.1c prompt with its chooser changed (DamageAssignmentFrame.
//     Banding). The defending player's division is among the blockers
//     only, so an attacker's trample puts nothing on the player then.
//
// "Bands with other [quality]" (CR 702.22b and the second half of
// 702.22c, j-k) is its special form, in bands_with_other.go (#6142): a
// band may also be [quality] creatures anchored by one with "bands with
// other [quality]", and such a pair among the blockers (or the blocked
// attackers) hands the damage division over as banding does.
//
// THE STATE. Game.attackBands, attacker -> band key, where the key is the
// instance ID of the band's first member in the declaration. Keys are
// compared, never resolved, so a band whose first member has left combat
// keeps its key. Every reader asks about members that are still on the
// battlefield and attacking (bandMatesLocked), which is CR 702.22f for
// the paths that take a creature out of combat without going through
// removeFromCombatLocked (a zone change mints a new instance ID).

// KeywordBanding is the canonical token for banding (CR 702.22a). The
// engine reads it at the attack declaration (CR 702.22c) and as combat
// damage is assigned (CR 702.22j-k).
const KeywordBanding = "banding"

// ErrIllegalBand is wrapped by every *BandRefusedError, so
// errors.Is(err, ErrIllegalBand) holds for any refused band.
var ErrIllegalBand = errors.New("game: that band can't be formed")

// BandRefusal is the stable token naming why a band was refused. It
// rides the `illegal_attack` frame's reason as "band".
type BandRefusal string

const (
	// BandRefusalTwoWithout: two or more creatures without banding in
	// one band (CR 702.22c allows up to one).
	BandRefusalTwoWithout BandRefusal = "two_without_banding"
	// BandRefusalSplitTarget: the band's creatures attack different
	// players, planeswalkers or battles (CR 702.22d).
	BandRefusalSplitTarget BandRefusal = "split_target"
	// BandRefusalSplitController: the band's creatures have different
	// controllers. Only an admin session driving several seats can send
	// that; a band is one attacking player's (CR 508.1e).
	BandRefusalSplitController BandRefusal = "split_controller"
	// BandRefusalNotHere: a band was sent where bands are not announced,
	// the confirmation half of a declaration another player chose.
	BandRefusalNotHere BandRefusal = "not_here"
)

// BandRefusedError is what declare_attackers returns for a band that
// breaks CR 702.22c-d. Nothing was declared, tapped or paid.
type BandRefusedError struct {
	Reason BandRefusal
	// Attacker and Other are the two creatures the refusal is about
	// (Other may be uuid.Nil), with their names for the sentence.
	Attacker     uuid.UUID
	AttackerName string
	Other        uuid.UUID
	OtherName    string
}

// Error is the debug form.
func (e *BandRefusedError) Error() string { return "game: " + e.Sentence() }

// Unwrap makes errors.Is(err, ErrIllegalBand) hold.
func (e *BandRefusedError) Unwrap() error { return ErrIllegalBand }

// Sentence is the player-facing explanation.
func (e *BandRefusedError) Sentence() string {
	a, o := bandMemberName(e.AttackerName), bandMemberName(e.OtherName)
	switch e.Reason {
	case BandRefusalTwoWithout:
		return "A band can have only one creature without banding, and " + a + " and " + o + " both lack it."
	case BandRefusalSplitTarget:
		return "Creatures in a band must attack the same player, planeswalker or battle, and " + a + " and " + o + " don't."
	case BandRefusalSplitController:
		return "Creatures in a band must all be controlled by the attacking player."
	case BandRefusalNotHere:
		return "Bands are announced by the player declaring the attackers, not as the attacks are confirmed."
	}
	return "That band can't be formed."
}

func bandMemberName(name string) string {
	if name == "" {
		return "an unknown creature"
	}
	return name
}

// bandGroups is the declaration's bands, in first-seen order: each group
// is the entries sharing one positive Band number. A group of one entry
// is not a band (it changes nothing, CR 702.22h-k), and is left out.
func bandGroups(decls []AttackDeclaration) [][]AttackDeclaration {
	var order []int
	byBand := map[int][]AttackDeclaration{}
	for _, d := range decls {
		if d.Band <= 0 {
			continue
		}
		if _, seen := byBand[d.Band]; !seen {
			order = append(order, d.Band)
		}
		byBand[d.Band] = append(byBand[d.Band], d)
	}
	var out [][]AttackDeclaration
	for _, n := range order {
		if len(byBand[n]) >= 2 {
			out = append(out, byBand[n])
		}
	}
	return out
}

// bandRefusalLocked judges the bands of a declaration against
// CR 702.22c-d: within each band, one target, one controller, and at
// most one creature without banding. `decls` is the entries the verb
// will declare (the ineligible ones already skipped). Layers must be
// fresh: banding is read off the effective characteristic, so a
// creature Helm of Chatzuk gave banding this turn counts.
//
// Caller must hold g.mu.
func (g *Game) bandRefusalLocked(decls []AttackDeclaration) *BandRefusedError {
	for _, band := range bandGroups(decls) {
		first := findBattlefieldCard(g, band[0].Attacker)
		if first == nil {
			continue
		}
		var without *Card
		// #6142, CR 702.22c's second sentence: [quality] creatures
		// anchored by one with "bands with other [quality]" are a band
		// whatever banding they have, so none counts as "without".
		bwo := g.bandsWithOtherBandLocked(bandAttackers(band))
		for _, d := range band {
			c := findBattlefieldCard(g, d.Attacker)
			if c == nil {
				continue
			}
			if d.Target != band[0].Target {
				return &BandRefusedError{Reason: BandRefusalSplitTarget,
					Attacker: first.InstanceID, AttackerName: first.Name, Other: c.InstanceID, OtherName: c.Name}
			}
			if c.Controller != first.Controller {
				return &BandRefusedError{Reason: BandRefusalSplitController,
					Attacker: first.InstanceID, AttackerName: first.Name, Other: c.InstanceID, OtherName: c.Name}
			}
			if bwo || HasKeyword(c, KeywordBanding) {
				continue
			}
			if without != nil {
				return &BandRefusedError{Reason: BandRefusalTwoWithout,
					Attacker: without.InstanceID, AttackerName: without.Name, Other: c.InstanceID, OtherName: c.Name}
			}
			without = c
		}
	}
	return nil
}

// bandAttackers is a band's attacking creatures.
func bandAttackers(band []AttackDeclaration) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(band))
	for _, d := range band {
		out = append(out, d.Attacker)
	}
	return out
}

// stageBandsLocked records the declaration's bands (CR 508.1e). Called
// after bandRefusalLocked has accepted them and the attacks are staged.
//
// Caller must hold g.mu in write mode.
func (g *Game) stageBandsLocked(decls []AttackDeclaration) {
	for _, band := range bandGroups(decls) {
		key := band[0].Attacker
		for _, d := range band {
			if g.attackBands == nil {
				g.attackBands = map[uuid.UUID]uuid.UUID{}
			}
			g.attackBands[d.Attacker] = key
		}
	}
}

// leaveBandLocked takes `id` out of its band (CR 702.22f): it was
// removed from combat, or its attack was re-pointed at something else
// before the declaration ended (CR 702.22d).
//
// Caller must hold g.mu in write mode.
func (g *Game) leaveBandLocked(id uuid.UUID) {
	if _, ok := g.attackBands[id]; !ok {
		return
	}
	delete(g.attackBands, id)
	if len(g.attackBands) == 0 {
		g.attackBands = nil
	}
}

// bandMatesLocked is every OTHER creature in `id`'s band that is still
// attacking, in battlefield order: nil when `id` is not attacking, is in
// no band, or is the band's last member in combat.
//
// Caller must hold g.mu.
func (g *Game) bandMatesLocked(id uuid.UUID) []uuid.UUID {
	key, ok := g.attackBands[id]
	if !ok || g.Battlefield == nil {
		return nil
	}
	if c := findBattlefieldCard(g, id); c == nil || c.AttackingTarget == uuid.Nil {
		return nil
	}
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID == id || c.AttackingTarget == uuid.Nil {
			continue
		}
		if k, in := g.attackBands[c.InstanceID]; in && k == key {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// BandMatesForEffect is bandMatesLocked for readers outside the package.
//
// Caller must hold g.mu.
func (g *Game) BandMatesForEffect(id uuid.UUID) []uuid.UUID {
	return g.bandMatesLocked(id)
}

// AttackBandNumbersLocked numbers this combat's bands for the wire: each
// attacker in a band with at least one other member still attacking ->
// its band's number, 1 for the band whose first member comes first on
// the battlefield, 2 for the next, and so on. A number, not the key, so
// the view says which attackers are together without naming an
// instance the viewer may not be shown. Empty when nothing is banded.
//
// Caller must hold g.mu.
func (g *Game) AttackBandNumbersLocked() map[uuid.UUID]int {
	if len(g.attackBands) == 0 || g.Battlefield == nil {
		return nil
	}
	numbers := map[uuid.UUID]int{}
	var out map[uuid.UUID]int
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget == uuid.Nil {
			continue
		}
		key, ok := g.attackBands[c.InstanceID]
		if !ok || len(g.bandMatesLocked(c.InstanceID)) == 0 {
			continue
		}
		n, seen := numbers[key]
		if !seen {
			n = len(numbers) + 1
			numbers[key] = n
		}
		if out == nil {
			out = map[uuid.UUID]int{}
		}
		out[c.InstanceID] = n
	}
	return out
}

// withBandMatesLocked is `attackers` with every live band-mate of each
// added after it, without repeats, in the order given then battlefield
// order. CR 702.22h-i: whatever blocks one member of a band blocks all
// of them.
//
// Caller must hold g.mu.
func (g *Game) withBandMatesLocked(attackers []uuid.UUID) []uuid.UUID {
	if len(g.attackBands) == 0 {
		return attackers
	}
	out := append([]uuid.UUID(nil), attackers...)
	for _, a := range attackers {
		for _, m := range g.bandMatesLocked(a) {
			if !slices.Contains(out, m) {
				out = append(out, m)
			}
		}
	}
	return out
}

// expandBandBlocksLocked is CR 702.22h at the block lock-in: each blocker
// whose declaration is being announced, and that blocks a member of a
// band, also blocks every other member still attacking. The declared
// pairs come first and the band-mates after them, so the EventBlock the
// lock-in emits for a band-mate carries a place past 1 (EventBlock.Amount)
// and "whenever this creature blocks" still triggers once. The band-mates
// do become blocked by the blocker, so "becomes blocked by a creature"
// triggers for each of them.
//
// The block's legality was judged on the declared pairs, which is the
// rule: the rulebook's example blocks a flier in a band with a creature
// that has swampwalk, and the swampwalker becomes blocked too. A
// blocker's capacity (multi_block.go) is not consulted either; the
// band-mates are blocked by the band, not declared.
//
// Caller must hold g.mu in write mode.
func (g *Game) expandBandBlocksLocked() {
	if len(g.attackBands) == 0 || g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.BlockingTarget == uuid.Nil || !g.blockAnnounceableLocked(c) {
			continue
		}
		atks := c.BlockedAttackers()
		if out := g.withBandMatesLocked(atks); len(out) != len(atks) {
			c.setBlockingSet(out)
		}
	}
}

// anyHasBandingLocked reports whether the damage division is handed over
// for `ids` right now: one of them is a creature on the battlefield with
// banding, or they hold both a [quality] creature with "bands with other
// [quality]" and another [quality] creature (#6142). CR 702.22j asks it
// of an attacker's blockers, CR 702.22k of a blocker's attackers.
//
// Caller must hold g.mu.
func (g *Game) anyHasBandingLocked(ids []uuid.UUID) bool {
	for _, id := range ids {
		if HasKeyword(findBattlefieldCard(g, id), KeywordBanding) {
			return true
		}
	}
	return len(ids) >= 2 && g.bandsWithOtherPairLocked(ids)
}

// bandingDefenderLocked is who divides an attacker's combat damage under
// CR 702.22j: the defending player of its attack, else (an attack whose
// defender can no longer be resolved) the controller of a blocker with
// banding, who is that player.
//
// Caller must hold g.mu.
func (g *Game) bandingDefenderLocked(atk *Card, blockers []uuid.UUID) uuid.UUID {
	if d := g.defendingPlayerForAttackerLocked(atk); d != uuid.Nil {
		return d
	}
	for _, id := range blockers {
		if b := findBattlefieldCard(g, id); HasKeyword(b, KeywordBanding) || (b != nil && hasAnyBandsWithOther(b)) {
			return b.Controller
		}
	}
	return uuid.Nil
}
