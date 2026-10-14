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
// # Public as it is cast, unless it is secret
//
// Every ballot emits EventVoteCast, which is the public log's "Alice
// voted for dominion", and the open prompt carries the ballots so far
// on the wire for every seat (protocol.CouncilVoteView), so the table
// sees the tally move as each player votes.
//
// Secret council (ADR 0180) is the same vote with CouncilVote.Secret
// set: a ballot emits EventSecretVoteCast, which names the voter and
// not the option, the wire carries a seat's own ballots only (the
// protocol filter puts them back per viewer), and when the last ballot
// is in EventVotesRevealed shows them all at once.

// EventVoteCast records one ballot: Actor voted for the option whose
// printed label is Label, in the vote Source called. Amount is the
// option's index in the vote's list; Target is the option's subject,
// the permanent or player voted for, when it names one. Emitted as the
// ballot is cast, so the table sees each vote before the next voter is
// asked.
const EventVoteCast EventKind = "vote_cast"

// EventVotesFinished marks the end of a vote (CR 701.38), "whenever
// players finish voting" (ADR 0146, amendment 2026-10-13). Actor is the
// controller of the ability that called the vote, Source the card it
// came from, Amount how many ballots were cast, and Ballots every one
// of them in the order cast. Emitted once per vote, after the last
// ballot and before the vote's own result is applied, and not at all
// for a vote that never began because nothing could be voted for.
//
// The ballots ride the event, not the vote: a trigger's stack item
// keeps its event (ctx.Trigger().Event), the event is plain data, and
// the event log and the stack are both in the snapshot, so a trigger
// still waiting on the stack reads the right ballots after a restore.
const EventVotesFinished EventKind = "votes_finished"

// EventSecretVoteCast records that Actor cast a ballot in the secret
// vote Source called (ADR 0180), and nothing about what they voted
// for: no option, no subject, no index. The ballot itself stays on the
// vote until EventVotesRevealed.
const EventSecretVoteCast EventKind = "secret_vote_cast"

// EventVotesRevealed is a secret vote's reveal (ADR 0180): "then those
// votes are revealed". Actor is the vote's controller, Source the card,
// Ballots every vote in the order cast, plus one ballot with Option -1
// for each vote a voter cast for nothing (Label is that option's
// words), and VoteOptions the options' labels, so the log can name each
// vote without the vote. Emitted once, after the last ballot and before
// EventVotesFinished.
const EventVotesRevealed EventKind = "votes_revealed"

// VoteDeclineLabel is the extra option on the prompt for a vote a
// player MAY cast ("While voting, you may vote an additional time",
// Ballot Broker): choosing it casts nothing.
const VoteDeclineLabel = "Don't vote again"

// VoteAbstainIndex is what OptionIndex returns, and the wire's
// `offered` carries, for a vote's Abstain option ("votes for up to one
// creature": a vote for nothing). -1 stays the decline option.
const VoteAbstainIndex = -2

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
	// Zone is where the cards the options name are, for an object vote
	// over cards that are not permanents (Custodi Squire's "an
	// artifact, creature, or enchantment card in your graveyard",
	// ADR 0146 amendment 2026-10-13). Empty is the battlefield, which
	// every earlier object vote means. An option whose card has left
	// that zone is not offered (CR 400.7: it is a new object).
	Zone ZoneKind `json:"zone,omitempty"`
	// Secret is secret council (ADR 0180): every player votes secretly,
	// then the votes are revealed. Nothing outside the game shows a
	// ballot before the last one is in (see the file comment).
	Secret bool `json:"secret,omitempty"`
	// OthersOnly is "votes for another player" (Mob Verdict): a ballot
	// does not offer the option naming its voter.
	OthersOnly bool `json:"othersOnly,omitempty"`
	// Abstain, when set, is the label of a "vote for nothing" option on
	// every ballot: "votes for up to one creature" (Vault 11). Choosing
	// it casts that vote as no ballot.
	Abstain string `json:"abstain,omitempty"`
	// SelfBest is a bot hint beside ForController and ForOpponents: the
	// option naming the voter is the one the voter wants (Círdan: the
	// votes you receive draw you cards). Nothing else reads it.
	SelfBest bool `json:"selfBest,omitempty"`
	// Ballots are the votes cast so far, in the order they were cast.
	Ballots []VoteBallot `json:"ballots,omitempty"`
	// Abstained lists, in order, each voter who cast a vote for nothing
	// (Abstain), once per such vote. Only the reveal reads it.
	Abstained []uuid.UUID `json:"abstained,omitempty"`
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
	out.Abstained = copyUUIDs(v.Abstained)
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
	// OptionZone is where the cards the options name are. Empty means
	// the battlefield (permanents voted for); set it to ZoneGraveyard
	// for "votes for a creature card in your graveyard".
	OptionZone ZoneKind
	// Secret, OthersOnly, Abstain and SelfBest are CouncilVote's (ADR
	// 0180): a secret vote, "another player", "up to one", and the bot
	// hint that a voter wants the option naming them.
	Secret     bool
	OthersOnly bool
	Abstain    string
	SelfBest   bool
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
		Zone:          p.OptionZone,
		Secret:        p.Secret,
		OthersOnly:    p.OthersOnly,
		Abstain:       p.Abstain,
		SelfBest:      p.SelfBest,
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
		if v.Abstain != "" {
			options = append(options, ChoiceOption{Label: v.Abstain})
		}
		if v.Optional() {
			options = append(options, ChoiceOption{Label: VoteDeclineLabel})
			question += " (you may vote an additional time)"
		}
		// "You choose how each player votes this turn" (Illusion of
		// Choice): the ballot is the voter's, the prompt is the
		// chooser's. FromPlayer follows the chooser so a departed
		// chooser's prompt is dropped like any ballot's (never
		// reassigned to a stranger), and the header says whose vote it
		// is.
		chooser := g.voteChooserLocked(v.Voter)
		if chooser != v.Voter {
			if p := g.playerByIDLocked(v.Voter); p != nil {
				question += " — you choose " + p.Name + "'s vote"
			}
		}
		queued := g.QueueChoiceForEffect(PendingChoice{
			Kind:        PendingChoiceOptionPick,
			Chooser:     chooser,
			FromPlayer:  chooser,
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

// voteChooserLocked is who answers `voter`'s ballot: the voter, unless
// a ModVoteChooser record stands (Illusion of Choice), in which case
// the player it names — the newest record naming a player still in the
// game. A chooser who has left no longer chooses (their effect has
// nothing to apply it), so the ballot goes back to the voter.
// Caller must hold g.mu.
func (g *Game) voteChooserLocked(voter uuid.UUID) uuid.UUID {
	best, bestAt := uuid.Nil, int64(0)
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModVoteChooser || m.Player == uuid.Nil || !g.playerInGameLocked(m.Player) {
				continue
			}
			if best == uuid.Nil || e.Timestamp >= bestAt {
				best, bestAt = m.Player, e.Timestamp
			}
		}
	}
	if best == uuid.Nil {
		return voter
	}
	return best
}

// ChooseHowEachPlayerVotesThisTurnForEffect is "you choose how each
// player votes this turn" (Illusion of Choice, ADR 0146 amendment
// 2026-10-13): every ballot of every vote until the end of the turn is
// put to `player` instead of the voter, in turn order and with each
// voter's extra votes (CR 701.38d) as ever. A stored ModVoteChooser
// record, swept at cleanup (CR 514.2). Reports whether one was
// written; nothing for a player who is not seated. Caller must hold
// g.mu (write).
func (g *Game) ChooseHowEachPlayerVotesThisTurnForEffect(sourceID, player uuid.UUID, label string) bool {
	if player == uuid.Nil || g.playerByIDLocked(player) == nil {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, uuid.Nil,
		[]Mod{{Kind: ModVoteChooser, Player: player}}, g.UntilEndOfTurnDuration(), label)
}

// openVoteOptionsLocked is the vote's option list as the next ballot
// offers it: every option, less a card that has left the zone the vote
// is over (the battlefield unless CouncilVote.Zone says otherwise), a
// player who has left the game, and, for "another player"
// (OthersOnly), the voter. The labels and subjects are the vote's own,
// so an answer maps back by subject. Caller must hold g.mu.
func (g *Game) openVoteOptionsLocked(v *CouncilVote) []ChoiceOption {
	out := make([]ChoiceOption, 0, len(v.Options))
	for _, o := range v.Options {
		if o.Player != uuid.Nil && !g.playerInGameLocked(o.Player) {
			continue
		}
		if v.OthersOnly && o.Player != uuid.Nil && o.Player == v.Voter {
			continue
		}
		if len(o.Cards) > 0 && !g.voteOptionCardPresentLocked(v, o.Cards[0]) {
			continue
		}
		out = append(out, o)
	}
	return cloneChoiceOptions(out)
}

// voteOptionCardPresentLocked reports whether the card an option names
// is still where the vote is over: on the battlefield for an ordinary
// object vote, or in a zone of the vote's Zone kind (a graveyard, for
// Custodi Squire). Caller must hold g.mu.
func (g *Game) voteOptionCardPresentLocked(v *CouncilVote, id uuid.UUID) bool {
	if v.Zone == "" || v.Zone == ZoneBattlefield {
		_, ok := g.battlefieldCardLocked(id)
		return ok
	}
	z := g.findCardZoneLocked(id)
	return z != nil && z.Kind == v.Zone
}

// OptionIndex maps an answered option back to its index in the
// vote's own list, by subject and label: the offered list may have
// been shorter (a permanent gone) or pruned while open (a seat gone).
// -1 for the decline option, VoteAbstainIndex for the Abstain option.
func (v *CouncilVote) OptionIndex(o ChoiceOption) int {
	if v.Abstain != "" && o.Label == v.Abstain && o.subject() == uuid.Nil {
		return VoteAbstainIndex
	}
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
	if chosen != nil && !g.playerInGameLocked(voter) {
		// The voter left while someone else was choosing for them
		// (Illusion of Choice): CR 800.4a, they cast nothing.
		chosen = nil
	}
	switch {
	case chosen == nil:
		// CR 800.4a: a player who has left casts nothing more.
		v.VotesLeft, v.MayLeft = 0, 0
	default:
		i := v.OptionIndex(*chosen)
		if i == VoteAbstainIndex {
			// "Up to one": a vote for nothing. It is that vote, cast.
			v.Abstained = append(v.Abstained, voter)
			g.emitBallotCastLocked(v, voter, i)
			if required {
				v.VotesLeft--
			} else {
				v.MayLeft--
			}
			break
		}
		if i < 0 {
			// The decline option: no more optional votes.
			v.MayLeft = 0
			break
		}
		v.Ballots = append(v.Ballots, VoteBallot{Voter: voter, Option: i})
		g.emitBallotCastLocked(v, voter, i)
		if required {
			v.VotesLeft--
		} else {
			v.MayLeft--
		}
	}
	return g.askNextBallotLocked(v)
}

// emitBallotCastLocked announces one ballot: in an open vote the
// option, as it is cast (ADR 0146 decision 5), and in a secret vote only
// that the voter has voted (ADR 0180). i is the option's index, or
// VoteAbstainIndex for a vote for nothing. Caller must hold g.mu.
func (g *Game) emitBallotCastLocked(v *CouncilVote, voter uuid.UUID, i int) {
	if v.Secret {
		g.EmitEvent(Event{Kind: EventSecretVoteCast, Actor: voter, Source: v.Source})
		return
	}
	ev := Event{Kind: EventVoteCast, Actor: voter, Source: v.Source, Amount: i, Label: v.Abstain}
	if i >= 0 {
		ev.Target, ev.Label = v.Options[i].subject(), v.Options[i].Label
	}
	g.EmitEvent(ev)
}

// RunVoteThenForEffect runs a vote's continuation directly, for a card
// whose vote had nothing to vote for (no permanent qualified): nobody
// voted, and the rest of the card still runs. No vote began, so nobody
// "finishes voting" and EventVotesFinished is not emitted. Caller must
// hold g.mu.
func (g *Game) RunVoteThenForEffect(then VoteThen, r VoteResult) error {
	if then.key == "" {
		return nil
	}
	return g.runVoteThenLocked(&CouncilVote{
		Controller: r.Controller, Source: r.Source, Options: r.Options,
		Ballots: r.Ballots, Then: then.key, Carry: r.Carry,
	})
}

// finishVoteLocked ends a vote that began: it announces that the
// players finished voting, then runs the vote's continuation with its
// result. The announcement comes first so the log reads in the order it
// happened (the last ballot, the end of the vote, what the ability does
// with it); a "finish voting" trigger goes on the stack after the
// ability that called the vote has finished resolving either way.
// Caller must hold g.mu.
func (g *Game) finishVoteLocked(v *CouncilVote) error {
	if v.Secret {
		g.EmitEvent(revealEventOf(v))
	}
	g.EmitEvent(Event{
		Kind:    EventVotesFinished,
		Actor:   v.Controller,
		Source:  v.Source,
		Amount:  len(v.Ballots),
		Ballots: append([]VoteBallot(nil), v.Ballots...),
	})
	return g.runVoteThenLocked(v)
}

// revealEventOf is a secret vote's EventVotesRevealed (ADR 0180): every
// ballot in the order cast, then one Option -1 ballot per vote for
// nothing, with the options' labels beside them.
func revealEventOf(v *CouncilVote) Event {
	ballots := append([]VoteBallot(nil), v.Ballots...)
	for _, voter := range v.Abstained {
		ballots = append(ballots, VoteBallot{Voter: voter, Option: -1})
	}
	labels := make([]string, len(v.Options))
	for i, o := range v.Options {
		labels[i] = o.Label
	}
	return Event{
		Kind:        EventVotesRevealed,
		Actor:       v.Controller,
		Source:      v.Source,
		Amount:      len(v.Ballots),
		Ballots:     ballots,
		VoteOptions: labels,
		Label:       v.Abstain,
	}
}

// RevealedSecretVoteAfter reports whether a secret vote was revealed
// by an event later than seq (ADR 0180 decision 4): the room refuses a
// seat's undo of the commit that revealed one, which would let the last
// voter vote again knowing every ballot. Takes g.mu (read): the caller
// must not hold it.
func (g *Game) RevealedSecretVoteAfter(seq uint64) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for i := len(g.Events) - 1; i >= 0; i-- {
		ev := &g.Events[i]
		if ev.Seq <= seq {
			return false
		}
		if ev.Kind == EventVotesRevealed {
			return true
		}
	}
	return false
}

// LastEventSeq is the Seq of the newest event in the log, or 0 for an
// empty log. Takes g.mu (read): the caller must not hold it.
func (g *Game) LastEventSeq() uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if n := len(g.Events); n > 0 {
		return g.Events[n-1].Seq
	}
	return 0
}

// VotedFor lists the distinct options voter voted for in ballots, in
// the order they first voted for each. A player who voted twice (CR
// 701.38d) names every option either vote went to, and a player who
// cast no ballot names none.
func VotedFor(ballots []VoteBallot, voter uuid.UUID) []int {
	var out []int
	seen := map[int]bool{}
	for _, b := range ballots {
		if b.Voter == voter && !seen[b.Option] {
			seen[b.Option] = true
			out = append(out, b.Option)
		}
	}
	return out
}

// VotedForAChoiceVotedForBy reports whether voter voted for at least
// one option that other also voted for — "an opponent who voted for a
// choice you voted for". Both are read against the same ballots, so a
// player with several votes shares a choice when any of theirs match
// any of the other's, and a player who cast nothing shares nothing.
func VotedForAChoiceVotedForBy(ballots []VoteBallot, voter, other uuid.UUID) bool {
	theirs := VotedFor(ballots, other)
	for _, o := range VotedFor(ballots, voter) {
		for _, t := range theirs {
			if o == t {
				return true
			}
		}
	}
	return false
}

// VotedForAChoiceNotVotedForBy reports whether voter voted for at least
// one option that other did not vote for — "an opponent who voted for a
// choice you didn't vote for". A voter with a choice outside all of
// other's counts even if they also shared one, and when other cast no
// ballot at all every choice voter voted for qualifies. A voter who
// cast nothing never does.
func VotedForAChoiceNotVotedForBy(ballots []VoteBallot, voter, other uuid.UUID) bool {
	theirs := VotedFor(ballots, other)
outer:
	for _, o := range VotedFor(ballots, voter) {
		for _, t := range theirs {
			if o == t {
				continue outer
			}
		}
		return true
	}
	return false
}

// runVoteThenLocked runs the vote's continuation with its result.
// Caller must hold g.mu.
func (g *Game) runVoteThenLocked(v *CouncilVote) error {
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
