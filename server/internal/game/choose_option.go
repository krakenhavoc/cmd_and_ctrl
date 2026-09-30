package game

import "github.com/google/uuid"

// choose_option.go — "As this permanent enters, choose <A> or <B>"
// (CR 614.12) where A and B are NAMED OPTIONS printed on the card:
// the anchor words of the Siege cycles. #1572, ADR 0071's amendment of
// 2026-09-27.
//
//	Frostcliff Siege   "As this enchantment enters, choose Jeskai or Temur.
//	                    • Jeskai — Whenever one or more creatures you control …
//	                    • Temur — Creatures you control get +1/+0 …"
//
// The FIFTH as-enters choice, after S26's creature type, #742's
// colour, #980's player and #1210's card name, and built the same way
// for the same reasons: queued from the permanent's AsEnters hook
// rather than by pausing the CR 614 entry pipeline (see
// creature_type_choice.go for the long form of that argument), stored
// on the permanent (Card.ChosenOption), and read for the rest of that
// permanent's life.
//
// # No new prompt kind
//
// The question — here is a closed list of words, pick exactly one — is
// the question PendingChoiceOptionPick already asks, so this is one of
// those with one option per word, addressed to the entering
// permanent's controller. The choice gate, `internal/legal`, the wire
// projection and the client modal all answer it unchanged, exactly as
// they answer ChoosePlayerAsEnters' seat list. None of the options
// names a seat or a card, so nothing prunes the list while it is open
// and the INDEX the chooser sends is a stable answer (the #994 reason
// the seat form needed ThenSeat does not arise).
//
// # Who reads it
//
// The ADR 0071 designation gate, and nothing else needs to. The ability
// printed after an anchor word carries ActiveWhen: ChosenOptionIs(word),
// and the four ability accessors drop it while that word is not the
// chosen one — so the other mode is not a disabled ability, it is not
// in the permanent's ability list at all. The Fate Reforged rulings say
// the same thing: each Siege has ONE of its two listed abilities,
// depending on the choice made as it enters.
//
// # The window
//
// Between entering and answering the permanent has NEITHER ability —
// DesignationChosenOption never matches an empty answer. Nothing can
// act in that window (an open PendingChoice stops priority), and the
// direction is the safe one: a Siege with neither mode is weaker than
// printed for no observable time, never stronger.
//
// # Bots
//
// The options are offered in PRINTED order and the enumerator marks the
// first always-legal (legal.choiceMoves), so a bot with nothing better
// to say takes the first anchor word. A legality-and-termination
// policy, not a strength one — the same posture docs/bot.md records for
// the seat form.

// EventOptionChosen records a CR 614.12 "choose <A> or <B>" answer in
// the event log, the way EventColorChosen, EventCreatureTypeChosen,
// EventPlayerChosen and EventCardNameChosen record theirs. `CardID` is
// the permanent it was chosen for and `Label` the chosen word, so the
// log reads "Ian chose Temur for Frostcliff Siege".
//
// It is also a layer-invalidating event (layer_listener.go): the answer
// switches a gated static on, so the cached resolution must be dropped
// the moment it lands.
const EventOptionChosen EventKind = "option_chosen"

// QueueChooseOptionAsEntersForEffect queues "As this permanent enters,
// choose <options[0]> or <options[1]> …" (CR 614.12) and returns the
// choice ID, or uuid.Nil when nothing was queued — no options, or a
// chooser who has already left the game (QueueChoiceForEffect's CR
// 800.4a guard). The answer lands on `source`'s Card.ChosenOption.
//
// `options` are the card's own words in printed order, and are copied:
// the continuation reads the chosen word back out of this copy by the
// index the chooser sent, never out of the caller's slice.
//
// A prompt DROPPED unanswered (#1006 — its chooser left, which takes
// the permanent with them) runs the continuation with NoChoiceIndex,
// which stores nothing: the permanent keeps neither ability. Never a
// default pick on the chooser's behalf.
//
// Caller must hold g.mu (an AsEnters hook does).
func (g *Game) QueueChooseOptionAsEntersForEffect(chooser, source uuid.UUID, question string, options []string) uuid.UUID {
	if len(options) == 0 {
		return uuid.Nil
	}
	words := append([]string(nil), options...)
	choices := make([]ChoiceOption, len(words))
	for i, w := range words {
		choices[i] = ChoiceOption{Label: w}
	}
	permanent, who := source, chooser
	return g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  chooser,
		Source:   source,
		Question: question,
		Options:  choices,
		Then: func(g *Game, index int) error {
			if index < 0 || index >= len(words) {
				return nil
			}
			return g.setChosenOptionLocked(permanent, who, words[index])
		},
	})
}

// setChosenOptionLocked stamps the CR 614.12 answer onto the permanent
// and announces it.
//
// The permanent is located LIVE rather than trusted from the queue: the
// prompt is asynchronous, and a Siege destroyed in response to its own
// entry is a legal if odd board. A missing source is not an error — the
// choice was made and simply has nowhere to land (CR 608.2) — and the
// event is still emitted, because the choice was still made out loud —
// by `chooser`, which is why the actor is not read off the permanent.
//
// The layer bump rides the event (layer_listener.go's designation case)
// rather than being written here: the answer switches a gated static
// on, which is exactly what a Class level or a solved Case does.
//
// Caller must hold g.mu.
func (g *Game) setChosenOptionLocked(source, chooser uuid.UUID, option string) error {
	if i := findCardOnBattlefield(g, source); i >= 0 {
		g.Battlefield.Cards[i].ChosenOption = option
	}
	g.EmitEvent(Event{
		Kind:   EventOptionChosen,
		Actor:  chooser,
		CardID: source,
		Label:  option,
	})
	return nil
}

// ChosenOptionOf returns the named option chosen for the permanent
// `sourceID` currently on the battlefield, or "" when none has been
// chosen (or the permanent is gone).
//
// The accessor exists for the reason ChosenPlayerOf and ChosenColorOf
// do: a caller outside this package reads the answer without reaching
// into the battlefield slice. The designation gate reads the field
// directly, because it is already holding the card.
//
// Caller must hold either lock.
func (g *Game) ChosenOptionOf(sourceID uuid.UUID) string {
	if i := findCardOnBattlefield(g, sourceID); i >= 0 {
		return g.Battlefield.Cards[i].ChosenOption
	}
	return ""
}
