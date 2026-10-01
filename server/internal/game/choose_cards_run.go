package game

import "github.com/google/uuid"

// choose_cards_run.go — "each player chooses …" over cards only that
// player may see: one choose_cards prompt per seat, asked as ONE
// printed instruction (#1743, ADR 0013 §5ai).
//
// The prompted run (prompt_run.go) was built for a sacrifice (#1019)
// and generalised for a discard (#1027): a run is one instruction in
// flight, however many prompts it takes, with one continuation that
// waits for the last of them. The discard leg already IS a
// choose_cards prompt with a run link. What no caller could do was
// hand the run an arbitrary choose_cards question per seat — each over
// that seat's own look, with its own bounds and its own set rule —
// because the link (ChooseCardsPrompt.promptRun) is unexported, on
// purpose: a catalog card that set it by hand would own a piece of
// the run's bookkeeping.
//
// Explore the Vastlands is the card: "each player looks at the top
// five cards of their library and may reveal a land card and/or an
// instant or sorcery card from among them. Each player puts the cards
// they revealed this way into their hand …". Every seat answers its
// own question, and nothing may move until every seat has.

// ChooseCardsRunThenForEffect queues one choose_cards prompt per entry
// of `legs`, as one run, and runs `then` once — after the last of them
// has been answered — with what each seat chose.
//
// # Each leg is an ordinary choose_cards prompt
//
// Its own chooser, candidates, bounds, Zone re-check and set-level
// Validate, so everything about answering one is choose_cards' own and
// unchanged: checkChooseCardsPicksLocked on submit, the enumerator's
// ChooseCardsPickLegalLocked, the wire projection and its non-chooser
// redaction, the candidate prunes, the departure row. What the run
// adds is only the thing a leg's own Then cannot give: the rest of the
// card, ONCE, after every seat has answered.
//
// # Order (CR 101.4)
//
// The legs go up together, in the order given, and may be answered in
// any order. For a choice made in secret that is APNAP order and not a
// relaxation of it: CR 101.4a lets cards in a hidden zone stay face
// down as they are chosen, so the caller passes the legs APNAP, the
// queue lists them that way, and no seat learns another's answer until
// `then` acts on all of them. A card whose choices are PUBLIC as they
// are made (Selective Obliteration's colour) has to ask one seat at a
// time, and this is not its door.
//
// # What is not asked
//
// A leg with no candidates is not queued — "you may choose" over
// nothing is not a question — and neither is a leg whose chooser has
// left the game (QueueChoiceForEffect's CR 800.4a guard). Neither has
// an entry in the answer, which is the run's "asked and chose nothing"
// versus "never asked" difference (SeatCards). A caller that needs
// every seat accounted for keeps its own list of who was meant to be
// asked; the answer is only who WAS.
//
// A leg's own Then, when set, runs first with that seat's picks, then
// the leg settles. A leg that is withdrawn unanswered — its chooser
// left, or every candidate left the library under it — settles with
// nothing chosen through the run link (defaultDroppedChoiceLocked),
// and its own Then is not run.
//
// `then` receives one entry per seat that was asked, in leg order. It
// runs inline when no leg was queued. Nil `then` is a run nobody waits
// on, as everywhere else.
//
// Reports how many prompts were queued.
//
// Caller must hold g.mu.
func (g *Game) ChooseCardsRunThenForEffect(legs []ChooseCardsPrompt, then func(g *Game, picks PromptedPicks) error) (int, error) {
	asks := make([]uuid.UUID, len(legs))
	for i, leg := range legs {
		asks[i] = leg.Chooser
	}
	var landedThen func(*Game, []SeatCards) error
	if then != nil {
		landedThen = func(g *Game, landed []SeatCards) error {
			return then(g, PromptedPicks(landed))
		}
	}
	// runPromptsLocked calls queue once per ask, in ask order, so the
	// next leg is the one at `next`.
	next := 0
	return g.runPromptsLocked(asks, landedThen, func(seat, run uuid.UUID) bool {
		leg := legs[next]
		next++
		if len(leg.Cards) == 0 {
			return false
		}
		source, own := leg.Source, leg.Then
		leg.promptRun = run
		leg.Then = func(g *Game, picked []uuid.UUID) error {
			if own != nil {
				if err := own(g, picked); err != nil {
					g.emitChoiceEffectErrorLocked(seat, source, err)
				}
			}
			return g.settleRunLegLocked(run, seat, picked)
		}
		return g.QueueChooseCardsForEffect(leg) != uuid.Nil
	})
}
