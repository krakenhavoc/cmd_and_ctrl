package game

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// secret_choice_round.go — hidden or simultaneous choices, then a reveal
// (#5948 Q3, ADR 0180's amendment of 2026-11-24).
//
// ADR 0180 built the secret vote, and Prisoner's Dilemma turned it into a
// secret CHOICE (CouncilVote.Choice): the ballots ride the open prompt in
// a field no view serialises, FilterViewFor hands each seat back its own,
// one EventVotesRevealed shows them all at once, and a seat may not undo
// the commit that revealed them. Every voter answered the same question
// from the same list.
//
// The cards of Q3 ask different questions of different players, and some
// of them want a number:
//
//   - Itazura, Lingering Wick: "Each opponent secretly chooses a number 0
//     or greater. Then those numbers are revealed."
//   - Call to the Void: "Each player secretly chooses a creature they
//     control and a creature they don't control. Then those choices are
//     revealed."
//   - The Toymaker's Trap: "secretly choose a number between 1 and 5 that
//     hasn't been chosen. … an opponent guesses which number you chose,
//     then you reveal the number you chose."
//   - Gollum, Scheming Guide: "choose land or nonland. An opponent guesses
//     whether the top card of your library is the chosen kind."
//
// So a secret choice may now be a list of ASKS (CouncilVote.Asks), each
// one question to one player: which of the vote's options it offers
// (VoteAsk.Allowed, or VoteAsk.Relation for "a creature they control" and
// "a creature they don't control", read live as the ask is put), or a
// number instead of an option (VoteAsk.Number). Everything else is the
// secret vote's: the hiding, the per-seat filter, the reveal, the undo
// refusal, the departure rule, the registered continuation and the
// restore point. An option ask is the option_pick prompt; a number ask is
// the pay_amount prompt with nothing paid (choose_number.go's), carrying
// the same CouncilVote. A number answer is a ballot with Option
// VoteNumberIndex and the number in VoteBallot.Number; every ballot of an
// asks round records which ask it answered (VoteBallot.Ask).
//
// Asks are put one at a time, in the order the card lists them, as every
// multi-player choice in the engine is (ADR 0180 decision 1). Nothing
// about the order leaks an answer: whose turn it is to choose is public.

// VoteNumberIndex is a ballot's Option for an answer that is a number,
// not one of the vote's options (VoteAsk.Number). The number is in
// VoteBallot.Number.
const VoteNumberIndex = -3

// Relations a VoteAsk can put between its voter and a permanent option.
// On-disk values: never renamed.
const (
	// VoteAskControls offers only the permanents the voter controls: "a
	// creature they control".
	VoteAskControls = "controls"
	// VoteAskDoesNotControl offers only the permanents the voter doesn't
	// control: "a creature they don't control".
	VoteAskDoesNotControl = "doesNotControl"
)

// VoteAsk is one question of a secret choice round: Voter answers it.
type VoteAsk struct {
	Voter uuid.UUID `json:"voter"`
	// Question is this ask's header; empty is the round's Question.
	Question string `json:"question,omitempty"`
	// Allowed are the indices into the round's Options this ask offers,
	// in that order. Empty offers every option.
	Allowed []int `json:"allowed,omitempty"`
	// Relation narrows a permanent option list to the ones the voter
	// controls (VoteAskControls) or doesn't (VoteAskDoesNotControl),
	// read as the ask is put. Empty narrows nothing.
	Relation string `json:"relation,omitempty"`
	// Number, when set, makes the answer a number rather than an option.
	Number *VoteNumber `json:"number,omitempty"`
}

// VoteNumber is the range of a number ask: Min up to Max, or no ceiling
// at all with NoMax ("a number 0 or greater"). Goal is the number a bot
// answers (and a person's stepper opens on); SelfDamage tells a bot the
// number may come back to it as damage.
type VoteNumber struct {
	Min        int  `json:"min,omitempty"`
	Max        int  `json:"max,omitempty"`
	NoMax      bool `json:"noMax,omitempty"`
	Goal       int  `json:"goal,omitempty"`
	SelfDamage bool `json:"selfDamage,omitempty"`
}

func cloneVoteAsk(a VoteAsk) VoteAsk {
	a.Allowed = append([]int(nil), a.Allowed...)
	if a.Number != nil {
		n := *a.Number
		a.Number = &n
	}
	return a
}

func cloneVoteAsks(in []VoteAsk) []VoteAsk {
	if in == nil {
		return nil
	}
	out := make([]VoteAsk, len(in))
	for i, a := range in {
		out[i] = cloneVoteAsk(a)
	}
	return out
}

// SecretChoicePrompt describes a secret choice round for
// StartSecretChoicesForEffect.
type SecretChoicePrompt struct {
	// Controller is the controller of the ability ("you"); Source its
	// card.
	Controller, Source uuid.UUID
	// Question is the default header of every ask.
	Question string
	// Options are the choices the option asks pick from, in printed
	// order. May be empty when every ask is a number.
	Options []ChoiceOption
	// ForController and ForOpponents are the bot hints of CouncilVote,
	// one per option. They ride the wire to every seat, so they must be
	// built from public information only.
	ForController, ForOpponents []int
	// Asks are the questions, in the order they are put. At least one.
	Asks []VoteAsk
	// Then runs with the result once every ask is answered (or skipped).
	// Required.
	Then VoteThen
	// Carry is plain data handed to Then.
	Carry []uuid.UUID
}

// StartSecretChoicesForEffect runs a secret choice round: each ask is put
// to its voter in order, every answer stays hidden from every other seat,
// and once the last one is in they are revealed together
// (EventVotesRevealed) and Then runs. An ask whose voter has left, or
// that has no option to offer, is skipped and records nothing. It is not a
// vote (CR 701.38): no extra vote, no Illusion of Choice, and no "whenever
// players finish voting".
//
// Caller must hold g.mu.
func (g *Game) StartSecretChoicesForEffect(p SecretChoicePrompt) error {
	if p.Then.key == "" {
		return fmt.Errorf("secret choice: no continuation")
	}
	if len(p.Asks) == 0 {
		return fmt.Errorf("secret choice: no asks")
	}
	for _, a := range p.Asks {
		if a.Number == nil && len(p.Options) == 0 {
			return fmt.Errorf("secret choice: an option ask with no options")
		}
	}
	v := &CouncilVote{
		Source:        p.Source,
		Controller:    p.Controller,
		Question:      p.Question,
		Options:       cloneChoiceOptions(p.Options),
		ForController: append([]int(nil), p.ForController...),
		ForOpponents:  append([]int(nil), p.ForOpponents...),
		Secret:        true,
		Choice:        true,
		Asks:          cloneVoteAsks(p.Asks),
		AskNo:         -1,
		Then:          p.Then.key,
		Carry:         copyUUIDs(p.Carry),
	}
	return g.askNextSecretLocked(v)
}

// askNextSecretLocked puts the next ask of a round, skipping one with
// nobody to answer it or nothing to offer, and finishes the round once
// none is left. Caller must hold g.mu.
func (g *Game) askNextSecretLocked(v *CouncilVote) error {
	for {
		if len(v.Asks) == 0 {
			v.Asking = nil
			v.VotesLeft, v.MayLeft = 0, 0
			return g.finishVoteLocked(v)
		}
		a := cloneVoteAsk(v.Asks[0])
		v.Asks = v.Asks[1:]
		v.AskNo++
		v.Asking = &a
		v.Voter = a.Voter
		v.VotesLeft, v.MayLeft = 1, 0
		if !g.playerInGameLocked(a.Voter) {
			continue
		}
		question := a.Question
		if question == "" {
			question = v.Question
		}
		if n := a.Number; n != nil {
			if g.queueSecretNumberLocked(v, n, question) {
				return nil
			}
			continue
		}
		options := g.openAskOptionsLocked(v, &a)
		if len(options) == 0 {
			// "A creature they control" from a player who controls none:
			// nothing is chosen (CR 608.2d), and the round goes on.
			continue
		}
		queued := g.QueueChoiceForEffect(PendingChoice{
			Kind:        PendingChoiceOptionPick,
			Chooser:     a.Voter,
			FromPlayer:  a.Voter,
			Count:       1,
			Source:      v.Source,
			Reason:      question,
			PickOptions: options,
			CouncilVote: v,
		})
		if queued != uuid.Nil {
			return nil
		}
	}
}

// queueSecretNumberLocked asks the current voter for a number, through
// the pay_amount prompt with nothing paid. Reports whether it was asked.
// Caller must hold g.mu.
func (g *Game) queueSecretNumberLocked(v *CouncilVote, n *VoteNumber, question string) bool {
	floor := max(n.Min, 0)
	ceiling := n.Max
	if n.NoMax {
		ceiling = PayAmountNoMaxCeiling
	}
	if ceiling < floor {
		return false
	}
	goal := n.Goal
	if goal < floor || goal > ceiling {
		goal = 0
	}
	id := g.QueueChoiceForEffect(PendingChoice{
		Kind:       PendingChoicePayAmount,
		Chooser:    v.Voter,
		FromPlayer: v.Voter,
		Count:      1,
		Source:     v.Source,
		Reason:     question,
		PayAmount: &PayAmountPrompt{
			Min: floor, Max: ceiling, Goal: goal, Unit: PayAmountOther,
			Resource: PayResourceNone, NoMax: n.NoMax, SelfDamage: n.SelfDamage,
		},
		CouncilVote: v,
	})
	return id != uuid.Nil
}

// openAskOptionsLocked is the option list the ask offers now: its
// Allowed indices (all of them when empty), narrowed by its Relation, less
// every option openVoteOptionsLocked would drop (a permanent gone, a
// player gone). Caller must hold g.mu.
func (g *Game) openAskOptionsLocked(v *CouncilVote, a *VoteAsk) []ChoiceOption {
	allowed := map[int]bool{}
	for _, i := range a.Allowed {
		allowed[i] = true
	}
	open := g.openVoteOptionsLocked(v)
	out := make([]ChoiceOption, 0, len(open))
	for _, o := range open {
		i := v.OptionIndex(o)
		if i < 0 || (len(a.Allowed) > 0 && !allowed[i]) {
			continue
		}
		if a.Relation != "" && len(o.Cards) > 0 {
			c, ok := g.battlefieldCardLocked(o.Cards[0])
			if !ok {
				continue
			}
			mine := c.Controller == a.Voter
			if (a.Relation == VoteAskControls) != mine {
				continue
			}
		}
		out = append(out, o)
	}
	return out
}

// castSecretAnswerLocked records the current ask's answer — an option, a
// number, or nothing (the prompt was dropped: its voter left) — and puts
// the next ask. The prompt has already been taken out of the queue.
// Caller must hold g.mu.
func (g *Game) castSecretAnswerLocked(v *CouncilVote, chosen *ChoiceOption, number *int) error {
	if v == nil {
		return nil
	}
	voter := v.Voter
	if !g.playerInGameLocked(voter) {
		chosen, number = nil, nil
	}
	switch {
	case number != nil:
		v.Ballots = append(v.Ballots, VoteBallot{Voter: voter, Option: VoteNumberIndex, Number: *number, Ask: v.AskNo})
		g.EmitEvent(Event{Kind: EventSecretVoteCast, Actor: voter, Source: v.Source, Label: SecretChoiceLabel})
	case chosen != nil:
		if i := v.OptionIndex(*chosen); i >= 0 {
			v.Ballots = append(v.Ballots, VoteBallot{Voter: voter, Option: i, Ask: v.AskNo})
			g.EmitEvent(Event{Kind: EventSecretVoteCast, Actor: voter, Source: v.Source, Label: SecretChoiceLabel})
		}
	}
	return g.askNextSecretLocked(v)
}

// SecretChoiceLabel is the Label of a secret CHOICE's EventSecretVoteCast
// and EventVotesRevealed (not a vote's): the log says "chose in secret"
// and "Choices revealed" rather than speaking of votes.
const SecretChoiceLabel = "secret_choice"

// AnswerTo is the ballot that answered ask number `ask` (0-based, in the
// order the card listed its asks), and whether one did: an ask that was
// skipped, or whose voter left, has none.
func (r VoteResult) AnswerTo(ask int) (VoteBallot, bool) {
	for _, b := range r.Ballots {
		if b.Ask == ask {
			return b, true
		}
	}
	return VoteBallot{}, false
}

// revealLabelsOf is a reveal's VoteOptions and Ballots: the options'
// labels, and after them one label per number answer, which that ballot
// points at, so the log names a number the way it names an option.
func revealLabelsOf(v *CouncilVote, ballots []VoteBallot) ([]string, []VoteBallot) {
	labels := make([]string, len(v.Options))
	for i, o := range v.Options {
		labels[i] = o.Label
	}
	out := append([]VoteBallot(nil), ballots...)
	for i := range out {
		if out[i].Option == VoteNumberIndex {
			labels = append(labels, strconv.Itoa(out[i].Number))
			out[i].Option = len(labels) - 1
		}
	}
	return labels, out
}
