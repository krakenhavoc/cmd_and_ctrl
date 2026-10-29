package game

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/google/uuid"
)

// attack_split.go — "create N tokens that are tapped and attacking",
// each attacking what its controller chooses (CR 508.4), as data (#4294).
//
// CR 508.4: when a creature is put onto the battlefield attacking, its
// controller chooses which defending player or planeswalker or battle it
// is attacking. A group of tokens made together is split among the things
// they could attack: the controller is asked how many attack each in turn,
// the first thing being the one the creature whose trigger this is
// attacks, and the last takes the rest, so there is at most one question
// per attackable thing but the last, and none once every token is placed.
//
// The questions are number prompts (number_then.go). The state between
// them is an AttackSplit, plain data the prompt carries
// (PendingChoice.AttackSplit): the template, how many, the things to
// attack in asking order, the shares fixed so far and the continuation as
// a keyed tail, so a table waiting on any of the questions is a restore
// point. The tokens are made by one creation once the last share is
// known, so a token doubler doubles each share and a "whenever one or more
// tokens are created" trigger fires once.

// AttackSplit is where a tokens-attacking split is.
type AttackSplit struct {
	// Controller answers the questions and controls the tokens; Source is
	// the card that makes them.
	Controller, Source uuid.UUID
	// Template is the token as it is made; Count how many; Tapped whether
	// they enter tapped.
	Template Card
	Count    int
	Tapped   bool
	// Targets are the things the tokens could attack, in asking order, and
	// Shares how many attack each of Targets[:len(Shares)].
	Targets []uuid.UUID
	Shares  []int
	// Label names the card in the questions.
	Label string
	// Tail is what the card does once the tokens are made, told the tokens
	// that landed (RegisterTail). The zero tail does nothing.
	Tail KeyedTail

	// settled, when set, makes this the split of a creation whose CR 701.7b
	// window has already settled (myriad, CR 702.116a): the groups the
	// replacements left, the one being split and the ones placed so far
	// (token_attack_choices.go). Targets, Shares and Count are then the
	// group being split's, and Template is unused.
	settled *settledCreation
}

// settledCreation is a creation whose replacements have settled and whose
// groups are still being sent to what they attack.
type settledCreation struct {
	// Groups are the settled groups; Index is the one being split and
	// Chosen the groups placed so far, in order.
	Groups []TokenGroup
	Index  int
	Chosen []TokenGroup
	// Attacking and Together are the creation's own (TokenCreation).
	Attacking uuid.UUID
	Together  bool
}

// attackSplitKey is the split's number branch, an on-disk identity: never
// renamed, never reused.
const attackSplitKey = "token/attack-split"

// settledSplitKey is the number branch of a split whose creation has
// already settled (myriad). A separate key, so a binary from before it,
// which would ask the question as a plain split and make the wrong
// tokens, refuses the file (ErrUnknownEffectKey).
const settledSplitKey = "token/settled-split"

func init() {
	answer := func(g *Game, r NumberAnswer) error {
		if r.split == nil {
			return fmt.Errorf("%w: attack split with no state", ErrInvalidParam)
		}
		s := r.split.clone()
		left := s.Count - sumInts(s.Shares)
		s.Shares = append(s.Shares, min(max(r.Number, 0), left))
		return g.advanceAttackSplitLocked(s)
	}
	RegisterNumberThen(attackSplitKey, answer)
	RegisterNumberThen(settledSplitKey, answer)
}

// StartAttackSplitForEffect makes the tokens `s` describes, asking the
// controller how many attack each thing when there is more than one
// (CR 508.4). With nothing or one thing to attack nothing is asked. The
// tail runs with every token made, also when there is nothing to attack
// (CR 508.4a: they are made, not attacking).
//
// Caller must hold g.mu.
func (g *Game) StartAttackSplitForEffect(s AttackSplit) error {
	if s.Count <= 0 {
		return nil
	}
	return g.advanceAttackSplitLocked(s.clone())
}

// advanceAttackSplitLocked asks the next question, or makes the tokens
// when there is none.
//
// Caller must hold g.mu.
func (g *Game) advanceAttackSplitLocked(s AttackSplit) error {
	left := s.Count - sumInts(s.Shares)
	i := len(s.Shares)
	if len(s.Targets) <= 1 || left <= 0 || i >= len(s.Targets)-1 {
		if i < max(len(s.Targets), 1) {
			s.Shares = append(s.Shares, left)
		}
		if s.settled != nil {
			return g.placeSettledGroupLocked(s)
		}
		return g.createAttackSplitLocked(s)
	}
	goal, noun, key := 0, "tokens", attackSplitKey
	if i == 0 {
		goal = left
	}
	if s.settled != nil {
		// Each opponent's copies are asked about with the number left as
		// the goal, a bot sending them all at the first thing.
		goal, noun, key = left, "copies", settledSplitKey
	}
	return g.QueueChooseNumberForEffect(ChooseNumber{
		Chooser:  s.Controller,
		Source:   s.Source,
		Resource: PayResourceNone,
		Max:      left,
		Goal:     goal,
		Question: s.Label + " — how many of the " + strconv.Itoa(left) + " " + noun + " attack " + g.attackTargetNameLocked(s.Targets[i]) + "?",
		Branch:   NumberBranch{Then: NumberThen{key: key}},
		split:    &s,
	})
}

// createAttackSplitLocked is the creation: shares[i] of the tokens
// attacking targets[i], or none attacking when there is nothing to attack.
//
// Caller must hold g.mu.
func (g *Game) createAttackSplitLocked(s AttackSplit) error {
	var groups []TokenGroup
	for i, n := range s.Shares {
		if n <= 0 {
			continue
		}
		grp := TokenGroup{Template: s.Template, Count: n, Entry: TokenEntryOptions{Tapped: s.Tapped}}
		if i < len(s.Targets) {
			grp.Attacking = s.Targets[i]
		}
		groups = append(groups, grp)
	}
	return g.CreateTokensKeyedForEffect(TokenCreation{
		Controller: s.Controller,
		Groups:     groups,
		Source:     s.Source,
	}, s.Tail)
}

// attackTargetNameLocked is a player's or a permanent's name for a
// question.
//
// Caller must hold g.mu.
func (g *Game) attackTargetNameLocked(id uuid.UUID) string {
	if p := g.playerByIDLocked(id); p != nil && p.Name != "" {
		return p.Name
	}
	if c := findBattlefieldCard(g, id); c != nil && c.Name != "" {
		return c.Name
	}
	return "that player"
}

func sumInts(xs []int) int {
	n := 0
	for _, x := range xs {
		n += x
	}
	return n
}

// clone is a deep copy.
func (s AttackSplit) clone() AttackSplit {
	out := s
	out.Template = cloneCard(s.Template)
	out.Targets = slices.Clone(s.Targets)
	out.Shares = slices.Clone(s.Shares)
	out.Tail = copyKeyedTail(s.Tail)
	out.settled = s.settled.clone()
	return out
}

// clone is a deep copy, nil for nil.
func (c *settledCreation) clone() *settledCreation {
	if c == nil {
		return nil
	}
	out := *c
	out.Groups = cloneTokenGroups(c.Groups)
	out.Chosen = cloneTokenGroups(c.Chosen)
	return &out
}

// cloneTokenGroups gives each group its own template.
func cloneTokenGroups(groups []TokenGroup) []TokenGroup {
	if groups == nil {
		return nil
	}
	out := make([]TokenGroup, len(groups))
	for i, grp := range groups {
		out[i] = grp
		out[i].Template = cloneCard(grp.Template)
	}
	return out
}

// attackSplitSnapshot is an AttackSplit on disk: the template through the
// card mirror.
type attackSplitSnapshot struct {
	Controller uuid.UUID          `json:"controller"`
	Source     uuid.UUID          `json:"source"`
	Template   cardSnapshot       `json:"template"`
	Count      int                `json:"count"`
	Tapped     bool               `json:"tapped,omitempty"`
	Targets    []uuid.UUID        `json:"targets,omitempty"`
	Shares     []int              `json:"shares,omitempty"`
	Label      string             `json:"label,omitempty"`
	Tail       *keyedTailSnapshot `json:"tail,omitempty"`
	// Settled is a myriad split's creation (#4294): the groups the
	// replacements left, the one being split and the ones placed.
	Settled *settledCreationSnapshot `json:"settled,omitempty"`
}

// settledCreationSnapshot is a settledCreation on disk: each group's
// template through the card mirror.
type settledCreationSnapshot struct {
	Groups    []tokenGroupSnapshot `json:"groups"`
	Index     int                  `json:"index,omitempty"`
	Chosen    []tokenGroupSnapshot `json:"chosen,omitempty"`
	Attacking *uuid.UUID           `json:"attacking,omitempty"`
	Together  bool                 `json:"together,omitempty"`
}

func snapshotTokenGroups(groups []TokenGroup, cen *ContinuationCensus) []tokenGroupSnapshot {
	var out []tokenGroupSnapshot
	for _, grp := range groups {
		out = append(out, tokenGroupSnapshot{
			Template: snapshotCard(grp.Template, cen), Count: grp.Count, Entry: snapshotTokenEntry(grp.Entry),
			Attacking: uuidPtrOrNil(grp.Attacking), AttackOpponent: uuidPtrOrNil(grp.AttackOpponent),
			CreatedWith: grp.CreatedWith, Blocking: uuidPtrOrNil(grp.Blocking),
		})
	}
	return out
}

func restoreTokenGroups(snaps []tokenGroupSnapshot) []TokenGroup {
	var out []TokenGroup
	for i := range snaps {
		grp := &snaps[i]
		out = append(out, TokenGroup{
			Template: restoreCard(&grp.Template), Count: grp.Count, Entry: grp.Entry.options(),
			Attacking: uuidOrNil(grp.Attacking), AttackOpponent: uuidOrNil(grp.AttackOpponent),
			CreatedWith: grp.CreatedWith, Blocking: uuidOrNil(grp.Blocking),
		})
	}
	return out
}

func snapshotAttackSplit(s *AttackSplit, cen *ContinuationCensus) *attackSplitSnapshot {
	if s == nil {
		return nil
	}
	out := &attackSplitSnapshot{
		Controller: s.Controller, Source: s.Source, Template: snapshotCard(s.Template, cen),
		Count: s.Count, Tapped: s.Tapped, Targets: copyUUIDs(s.Targets), Shares: copyInts(s.Shares),
		Label: s.Label, Tail: snapshotKeyedTail(s.Tail),
	}
	if c := s.settled; c != nil {
		out.Settled = &settledCreationSnapshot{
			Groups: snapshotTokenGroups(c.Groups, cen), Index: c.Index, Chosen: snapshotTokenGroups(c.Chosen, cen),
			Attacking: uuidPtrOrNil(c.Attacking), Together: c.Together,
		}
	}
	return out
}

func restoreAttackSplit(s *attackSplitSnapshot) *AttackSplit {
	if s == nil {
		return nil
	}
	out := &AttackSplit{
		Controller: s.Controller, Source: s.Source, Template: restoreCard(&s.Template),
		Count: s.Count, Tapped: s.Tapped, Targets: copyUUIDs(s.Targets), Shares: copyInts(s.Shares),
		Label: s.Label, Tail: restoreKeyedTail(s.Tail),
	}
	if c := s.Settled; c != nil {
		out.settled = &settledCreation{
			Groups: restoreTokenGroups(c.Groups), Index: c.Index, Chosen: restoreTokenGroups(c.Chosen),
			Attacking: uuidOrNil(c.Attacking), Together: c.Together,
		}
	}
	return out
}

// cloneAttackSplit is a deep copy, nil for nil.
func cloneAttackSplit(s *AttackSplit) *AttackSplit {
	if s == nil {
		return nil
	}
	out := s.clone()
	return &out
}
