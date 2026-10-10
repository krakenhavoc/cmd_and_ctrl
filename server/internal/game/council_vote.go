package game

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

// council_vote.go — voting (CR 701.38, ADR 0146): will of the council,
// council's dilemma, and the plain "each player votes" of Coercive
// Portal's and Emissary Green's kind.
//
// # Not the sandbox vote
//
// politics.go's Vote is the S10 table tool behind "Call a vote…": any
// player opens one, anybody may re-cast or end it, and nothing in the
// rules reads it. A rules vote is the other way round on every count.
// A resolving spell or ability opens it, every player votes exactly
// once (plus any extra votes, CR 701.38d) in turn order from a named
// player (CR 701.38a), nobody may change a ballot, and the effect reads
// the tally. Sharing the struct would mean a sandbox "end vote" button
// that could cut a Council's Judgment short, so the two are kept apart
// and the client draws both with the same dock request
// (lib/choiceDock.ts voteRequest).
//
// # One option pick per ballot
//
// Each ballot is an ordinary PendingChoiceOptionPick addressed to the
// voter, so the choice gate, the enumerator, the departure table, the
// wire, the client's inline prompt and the bot's move list all answer
// it unchanged. What makes it a ballot is PendingChoice.CouncilVote,
// the vote so far: ResolveOptionPick hands an answer to castBallotLocked
// instead of to a frame, and a dropped prompt (its voter left the game,
// CR 800.4a) casts nothing and moves on. The vote is plain data and its
// continuation is a registered key (RegisterVoteThen), so a table
// waiting on a ballot is a restore point.
//
// # Public as it is cast
//
// Every ballot emits EventVoteCast, which is the public log's "Alice
// voted for dominion", and the open prompt carries the ballots so far
// on the wire for every seat (protocol.CouncilVoteView), so the table
// sees the tally move as each player votes. CR 701.38 has no secret
// vote; secret council is a different ability word and is not this.

// EventVoteCast records one ballot: Actor voted for the option whose
// printed label is Label, in the vote Source called. Amount is the
// option's index in the vote's list; Target is the option's subject,
// the permanent or player voted for, when it names one. Emitted as the
// ballot is cast, so the table sees each vote before the next voter is
// asked.
const EventVoteCast EventKind = "vote_cast"

// VoteDeclineLabel is the extra option on the prompt for a vote a
// player MAY cast ("While voting, you may vote an additional time",
// Ballot Broker): choosing it casts nothing.
const VoteDeclineLabel = "Don't vote again"

// ExtraVote is a permanent's printed static about voting (CR 701.38d).
type ExtraVote int

const (
	// ExtraVoteNone is the zero value: the permanent says nothing.
	ExtraVoteNone ExtraVote = iota
	// ExtraVoteYouGet is "While voting, you get an additional vote"
	// (Brago's Representative): one more ballot its controller must
	// cast.
	ExtraVoteYouGet
	// ExtraVoteYouMay is "While voting, you may vote an additional
	// time" (Ballot Broker, Tivit, The Valeyard): one more ballot its
	// controller may cast or decline.
	ExtraVoteYouMay
)

// CatalogExtraVote reports a battlefield permanent's voting static,
// keyed by CatalogAbilityKey. Installed by the catalog (carddef.go);
// nil in a test binary without it, which gives nobody an extra vote.
var CatalogExtraVote func(key string) ExtraVote

// VoteBallot is one vote: Voter voted for Options[Option].
type VoteBallot struct {
	Voter  uuid.UUID `json:"voter"`
	Option int       `json:"option"`
}

// CouncilVote is a vote in progress. It rides the open ballot prompt
// (PendingChoice.CouncilVote) and moves to the next voter's prompt
// when the ballot is cast, so there is exactly one copy of it at a
// time, and none once the vote has finished.
type CouncilVote struct {
	// Source is the card whose ability called the vote, and
	// Controller that ability's controller — the "you" of "starting
	// with you" and of the effect the result decides.
	Source     uuid.UUID `json:"source"`
	Controller uuid.UUID `json:"controller"`
	// Question is the prompt's header, written the way the card is.
	Question string `json:"question"`
	// Options are the choices in printed order (CR 701.38b): a word
	// is a Label alone, a permanent or card carries its ID in Cards,
	// a player carries Player. Never pruned: a ballot is an index
	// into this list.
	Options []ChoiceOption `json:"options"`
	// ForController and ForOpponents are the card's hints to a bot,
	// one number per option: how much the controller, and an opponent
	// of the controller, wants that option to win. Higher is better;
	// empty is no hint. Nothing else reads them.
	ForController []int `json:"forController,omitempty"`
	ForOpponents  []int `json:"forOpponents,omitempty"`
	// Ballots are the votes cast so far, in the order they were cast.
	Ballots []VoteBallot `json:"ballots,omitempty"`
	// Voters are the players still to vote after the current one, in
	// turn order.
	Voters []uuid.UUID `json:"voters,omitempty"`
	// Voter is the player voting now; VotesLeft is how many ballots
	// they still must cast (this one included) and MayLeft how many
	// more they may cast after those (CR 701.38d: all of a player's
	// votes are cast at the moment they would vote).
	Voter     uuid.UUID `json:"voter"`
	VotesLeft int       `json:"votesLeft,omitempty"`
	MayLeft   int       `json:"mayLeft,omitempty"`
	// Then is the registered continuation run with the result, and
	// Carry the IDs the card put on the vote for it.
	Then  string      `json:"then"`
	Carry []uuid.UUID `json:"carry,omitempty"`
}

func cloneCouncilVote(v *CouncilVote) *CouncilVote {
	if v == nil {
		return nil
	}
	out := *v
	out.Options = cloneChoiceOptions(v.Options)
	out.ForController = append([]int(nil), v.ForController...)
	out.ForOpponents = append([]int(nil), v.ForOpponents...)
	out.Ballots = append([]VoteBallot(nil), v.Ballots...)
	out.Voters = copyUUIDs(v.Voters)
	out.Carry = copyUUIDs(v.Carry)
	return &out
}

// Optional reports whether the open ballot is one the voter may
// decline: their required votes are cast and an extra one is on offer.
func (v *CouncilVote) Optional() bool {
	return v != nil && v.VotesLeft <= 0 && v.MayLeft > 0
}

// VoteResult is what a vote's continuation is handed once every
// player has voted.
type VoteResult struct {
	// Controller and Source are the vote's.
	Controller, Source uuid.UUID
	// Options are the vote's options, in printed order.
	Options []ChoiceOption
	// Ballots are every vote cast, in the order cast.
	Ballots []VoteBallot
	// Carry is the plain data the card put on the vote.
	Carry []uuid.UUID
}

// Votes is how many votes option i got.
func (r VoteResult) Votes(i int) int {
	n := 0
	for _, b := range r.Ballots {
		if b.Option == i {
			n++
		}
	}
	return n
}

// MoreVotes reports whether option i got more votes than option j —
// will of the council's "if <i> gets more votes".
func (r VoteResult) MoreVotes(i, j int) bool { return r.Votes(i) > r.Votes(j) }

// MostVotes lists the options with the most votes or tied for the
// most, in printed order — "each permanent with the most votes or tied
// for most votes". An option nobody voted for is never among them, so
// a vote in which nobody voted returns nothing.
func (r VoteResult) MostVotes() []int {
	best := 0
	for i := range r.Options {
		if n := r.Votes(i); n > best {
			best = n
		}
	}
	if best == 0 {
		return nil
	}
	var out []int
	for i := range r.Options {
		if r.Votes(i) == best {
			out = append(out, i)
		}
	}
	return out
}

// VotersFor lists one entry per vote for option i, naming the voter,
// in the order the votes were cast — council's dilemma's "for each
// money vote, choose a permanent owned by the voter". A player who
// voted for it twice appears twice.
func (r VoteResult) VotersFor(i int) []uuid.UUID {
	var out []uuid.UUID
	for _, b := range r.Ballots {
		if b.Option == i {
			out = append(out, b.Voter)
		}
	}
	return out
}

// VoteFunc is a registered vote continuation. It runs with g.mu held,
// may queue prompts of its own, and must capture nothing.
type VoteFunc func(g *Game, r VoteResult) error

// VoteThen names a registered continuation. The key is unexported, so
// the only way to hold one is RegisterVoteThen.
type VoteThen struct{ key string }

// Key is the continuation's on-disk key.
func (t VoteThen) Key() string { return t.key }

// voteThens is guarded by effectRegistryMu: the keys share the effect
// key namespace and ledger (testdata/effect_keys.txt lists one as
// "vote <key>").
var voteThens = map[string]VoteFunc{}

// RegisterVoteThen registers a vote's continuation. Call it once, from
// a package-level var. Panics on a bad or duplicate key. The key is an
// on-disk identity: never renamed, never reused.
func RegisterVoteThen(key string, fn VoteFunc) VoteThen {
	effectRegistryMu.Lock()
	defer effectRegistryMu.Unlock()
	checkEffectKey("vote continuation", key)
	if fn == nil {
		panic(fmt.Sprintf("game: vote continuation %q has no function", key))
	}
	voteThens[key] = fn
	return VoteThen{key: key}
}

func lookupVoteThen(key string) (VoteFunc, bool) {
	effectRegistryMu.RLock()
	defer effectRegistryMu.RUnlock()
	fn, ok := voteThens[key]
	return fn, ok
}

// KnownVoteThen reports whether this binary has a vote continuation
// registered under key.
func KnownVoteThen(key string) bool { _, ok := lookupVoteThen(key); return ok }

// registeredVoteKeysLocked lists the ledger lines, "vote <key>".
// Caller holds effectRegistryMu.
func registeredVoteKeysLocked() []string {
	out := make([]string, 0, len(voteThens))
	for k := range voteThens {
		out = append(out, "vote "+k)
	}
	sort.Strings(out)
	return out
}

// VotePrompt describes a vote for StartVoteForEffect.
type VotePrompt struct {
	// Starting votes first ("starting with you"); the rest follow in
	// turn order (CR 701.38a). A player who has left the game is
	// skipped, and so is a Starting who has: the next player in turn
	// order begins.
	Starting uuid.UUID
	// Controller is the controller of the ability that called the
	// vote; Source its card.
	Controller, Source uuid.UUID
	// Question is every ballot prompt's header.
	Question string
	// Options are the choices, in printed order. At least one.
	Options []ChoiceOption
	// ForController and ForOpponents are the bot hints (see
	// CouncilVote). Optional.
	ForController, ForOpponents []int
	// Then runs with the result once everyone has voted. Required.
	Then VoteThen
	// Carry is plain data handed to Then.
	Carry []uuid.UUID
}

// StartVoteForEffect runs a vote (CR 701.38a): each player in turn
// order from Starting votes for one of Options, with any extra votes
// their permanents give them (CR 701.38d), and then Then runs with the
// tally. Each ballot is a prompt, so this usually returns with the
// first one queued; with no one able to vote, Then runs at once with
// no ballots.
//
// Caller must hold g.mu.
func (g *Game) StartVoteForEffect(p VotePrompt) error {
	if p.Then.key == "" {
		return fmt.Errorf("vote: no continuation")
	}
	if len(p.Options) == 0 {
		return fmt.Errorf("vote: no options")
	}
	v := &CouncilVote{
		Source:        p.Source,
		Controller:    p.Controller,
		Question:      p.Question,
		Options:       cloneChoiceOptions(p.Options),
		ForController: append([]int(nil), p.ForController...),
		ForOpponents:  append([]int(nil), p.ForOpponents...),
		Voters:        g.votersFromLocked(p.Starting),
		Then:          p.Then.key,
		Carry:         copyUUIDs(p.Carry),
	}
	return g.askNextBallotLocked(v)
}

// votersFromLocked lists the players still in the game in turn order,
// starting with `start` (or, if it has left, the next player after its
// seat). Caller must hold g.mu.
func (g *Game) votersFromLocked(start uuid.UUID) []uuid.UUID {
	n := len(g.Seats)
	first := 0
	for i, s := range g.Seats {
		if s != nil && s.ID == start {
			first = i
			break
		}
	}
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		s := g.Seats[(first+i)%n]
		if s == nil || s.Eliminated {
			continue
		}
		out = append(out, s.ID)
	}
	return out
}

// extraVotesLocked counts the extra votes `voter`'s permanents give
// them (CR 701.38d): `must` they get, `may` they may cast. A permanent
// that has lost its abilities gives none. Caller must hold g.mu.
func (g *Game) extraVotesLocked(voter uuid.UUID) (must, may int) {
	if CatalogExtraVote == nil || g.Battlefield == nil {
		return 0, 0
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != voter {
			continue
		}
		key := catalogAbilityKeyOf(c)
		if key == "" {
			continue
		}
		switch CatalogExtraVote(key) {
		case ExtraVoteYouGet:
			must++
		case ExtraVoteYouMay:
			may++
		}
	}
	return must, may
}

// askNextBallotLocked queues the next ballot, moving to the next voter
// when the current one has no votes left, and runs the continuation
// once nobody has. Caller must hold g.mu.
func (g *Game) askNextBallotLocked(v *CouncilVote) error {
	for {
		if v.VotesLeft <= 0 && v.MayLeft <= 0 {
			if len(v.Voters) == 0 {
				return g.finishVoteLocked(v)
			}
			v.Voter, v.Voters = v.Voters[0], v.Voters[1:]
			if !g.playerInGameLocked(v.Voter) {
				continue
			}
			must, may := g.extraVotesLocked(v.Voter)
			v.VotesLeft, v.MayLeft = 1+must, may
		}
		if !g.playerInGameLocked(v.Voter) {
			v.VotesLeft, v.MayLeft = 0, 0
			continue
		}
		options := g.openVoteOptionsLocked(v)
		if len(options) == 0 {
			// Nothing left to vote for (every permanent on the list
			// has gone): the player casts no vote.
			v.VotesLeft, v.MayLeft = 0, 0
			continue
		}
		question := v.Question
		if v.Optional() {
			options = append(options, ChoiceOption{Label: VoteDeclineLabel})
			question += " (you may vote an additional time)"
		}
		queued := g.QueueChoiceForEffect(PendingChoice{
			Kind:        PendingChoiceOptionPick,
			Chooser:     v.Voter,
			FromPlayer:  v.Voter,
			Count:       1,
			Source:      v.Source,
			Reason:      question,
			PickOptions: options,
			CouncilVote: v,
		})
		if queued != uuid.Nil {
			return nil
		}
		v.VotesLeft, v.MayLeft = 0, 0
	}
}

// openVoteOptionsLocked is the vote's option list as the next ballot
// offers it: every option, less a permanent that has left the
// battlefield or a player who has left the game. The labels and
// subjects are the vote's own, so an answer maps back by subject.
// Caller must hold g.mu.
func (g *Game) openVoteOptionsLocked(v *CouncilVote) []ChoiceOption {
	out := make([]ChoiceOption, 0, len(v.Options))
	for _, o := range v.Options {
		if o.Player != uuid.Nil && !g.playerInGameLocked(o.Player) {
			continue
		}
		if len(o.Cards) > 0 {
			if _, ok := g.battlefieldCardLocked(o.Cards[0]); !ok {
				continue
			}
		}
		out = append(out, o)
	}
	return cloneChoiceOptions(out)
}

// OptionIndex maps an answered option back to its index in the
// vote's own list, by subject and label: the offered list may have
// been shorter (a permanent gone) or pruned while open (a seat gone).
// -1 for the decline option.
func (v *CouncilVote) OptionIndex(o ChoiceOption) int {
	for i, opt := range v.Options {
		if opt.Label == o.Label && opt.subject() == o.subject() {
			return i
		}
	}
	return -1
}

// castBallotLocked records the voter's answer (nil: they cast none,
// because their prompt was dropped) and asks the next ballot. The
// prompt has already been taken out of the queue. Caller must hold
// g.mu.
func (g *Game) castBallotLocked(v *CouncilVote, voter uuid.UUID, chosen *ChoiceOption) error {
	if v == nil {
		return nil
	}
	required := v.VotesLeft > 0
	switch {
	case chosen == nil:
		// CR 800.4a: a player who has left casts nothing more.
		v.VotesLeft, v.MayLeft = 0, 0
	default:
		i := v.OptionIndex(*chosen)
		if i < 0 {
			// The decline option: no more optional votes.
			v.MayLeft = 0
			break
		}
		v.Ballots = append(v.Ballots, VoteBallot{Voter: voter, Option: i})
		g.EmitEvent(Event{
			Kind:   EventVoteCast,
			Actor:  voter,
			Source: v.Source,
			Target: v.Options[i].subject(),
			Amount: i,
			Label:  v.Options[i].Label,
		})
		if required {
			v.VotesLeft--
		} else {
			v.MayLeft--
		}
	}
	return g.askNextBallotLocked(v)
}

// RunVoteThenForEffect runs a vote's continuation directly, for a card
// whose vote had nothing to vote for (no permanent qualified): nobody
// voted, and the rest of the card still runs. Caller must hold g.mu.
func (g *Game) RunVoteThenForEffect(then VoteThen, r VoteResult) error {
	if then.key == "" {
		return nil
	}
	return g.finishVoteLocked(&CouncilVote{
		Controller: r.Controller, Source: r.Source, Options: r.Options,
		Ballots: r.Ballots, Then: then.key, Carry: r.Carry,
	})
}

// finishVoteLocked runs the vote's continuation with its result.
// Caller must hold g.mu.
func (g *Game) finishVoteLocked(v *CouncilVote) error {
	fn, ok := lookupVoteThen(v.Then)
	if !ok {
		return fmt.Errorf("%w: vote continuation %q", ErrUnknownEffectKey, v.Then)
	}
	return fn(g, VoteResult{
		Controller: v.Controller,
		Source:     v.Source,
		Options:    cloneChoiceOptions(v.Options),
		Ballots:    append([]VoteBallot(nil), v.Ballots...),
		Carry:      copyUUIDs(v.Carry),
	})
}
