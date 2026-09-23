package game

import (
	"sort"
	"strings"

	"github.com/google/uuid"
)

// choose_card_name.go — "as this permanent enters, choose a card
// name" (CR 614.12): Pithing Needle, Phyrexian Revoker, Sorcerous
// Spyglass, Meddling Mage, Nevermore. #1210, ADR 0073's amendment of
// 2026-09-22.
//
// The FOURTH as-enters choice, after S26's creature type, #742's
// colour and #1007/#980's player, and it is built the same way for
// the same reasons: queued from the permanent's AsEnters hook rather
// than by pausing the CR 614 entry pipeline (which can pause only for
// a LAND — see creature_type_choice.go for the long form of that
// argument, and note that three of the five cards above are
// artifacts), stored on the permanent, and read for the rest of that
// permanent's life.
//
// # The one way it differs: there is no vocabulary
//
// NamedTribe is validated against CR 205.3m's 345 creature types and
// normalised to a canonical spelling; ChosenColor against five
// letters; ChosenPlayer against the seats at the table. A CARD NAME
// has no such list. CR 201.2 lets a player name any card name at all
// — one in nobody's deck, one from a set this server has never
// imported, one printed after this binary was built. A server that
// refused an unknown name would be refusing a legal choice, and a
// server that tried to keep the list would be wrong the week after
// the next set.
//
// So the answer is FREE TEXT, checked for SHAPE and nothing else:
// trimmed, non-empty, and capped at a length no printed card comes
// near. The wire carries `name_options` as a convenience for the
// client's picker — the names of cards in PUBLIC zones, so the list
// cannot leak a hidden card — and the picker is a suggestion list
// over a text box, never a closed set.
//
// # Matching
//
// CardNameMatches is the one comparison, and every reader goes
// through it: case-insensitive on trimmed strings, asking every FACE
// (CR 201.2b — naming one half of a split card or one face of a modal
// DFC names the card). A card file that compared Card.Name == chosen
// would be wrong about Fire // Ice and about every transforming
// permanent whose back face is up.

// PendingChoiceCardName is the "choose a card name" prompt. Answered
// with a `{card_name: "Sol Ring"}` payload; the chooser is the
// entering permanent's controller.
//
// Like PendingChoiceCreatureType, the legal set is not stored on the
// choice — but for the opposite reason. The creature-type prompt has
// one vocabulary the view can always rebuild; this one has NO legal
// set at all, and what the view sends is a suggestion list it
// rebuilds from the public zones at serialisation time.
const PendingChoiceCardName PendingChoiceKind = "choose_card_name"

// EventCardNameChosen records a CR 614.12 answer in the event log.
// `Label` carries the chosen name and `CardID` the permanent it was
// chosen for, so the log reads "Pithing Needle — Sol Ring".
const EventCardNameChosen EventKind = "card_name_chosen"

// MaxChosenNameLen is the cap on a chosen card name. The longest
// printed Magic card name is well under a hundred characters and the
// longest split name not much more; this is loose enough that no
// legal answer is refused and tight enough that a client cannot park
// a megabyte in game state that every snapshot then carries.
const MaxChosenNameLen = 200

// QueueCardNameChoiceForEffect queues the "choose a card name" prompt
// for `chooser`, on behalf of the permanent `source`. Returns the
// choice ID.
//
// Caller must hold g.mu (an AsEnters hook does).
func (g *Game) QueueCardNameChoiceForEffect(chooser, source uuid.UUID, reason string) uuid.UUID {
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:    PendingChoiceCardName,
		Chooser: chooser,
		Count:   1,
		Source:  source,
		Reason:  reason,
	})
}

// ResolveCardNameChoice processes a resolve_choice action for a
// PendingChoiceCardName entry: the chooser names a card and it is
// stamped onto the source permanent's ChosenName.
//
// The name is trimmed and length-checked and NOT validated against
// any vocabulary — see the note at the top of this file on why there
// is none to validate against.
//
// The permanent is located live rather than trusted from the queue:
// the prompt is asynchronous and the permanent can have left in the
// meantime (a Pithing Needle destroyed in response to its own entry
// is a legal, if odd, board). A missing source is NOT an error — the
// choice is still made and simply has nowhere to land, which is what
// CR 608.2 does with an effect whose object has gone.
//
// Caller must NOT hold g.mu.
func (g *Game) ResolveCardNameChoice(choiceID, chooserID uuid.UUID, name string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	idx := -1
	for i, c := range g.PendingChoices {
		if c != nil && c.ID == choiceID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return ErrPendingChoiceNotFound
	}
	choice := g.PendingChoices[idx]
	if choice.Kind != PendingChoiceCardName {
		return ErrInvalidParam
	}
	if choice.Chooser != chooserID {
		return ErrNotTheChooser
	}
	chosen := strings.TrimSpace(name)
	if chosen == "" || len(chosen) > MaxChosenNameLen {
		return ErrInvalidParam
	}
	g.dequeueChoiceLocked(idx)

	if i := findCardOnBattlefield(g, choice.Source); i >= 0 {
		g.Battlefield.Cards[i].ChosenName = chosen
		// The layer version is NOT bumped, and that is the one place
		// this differs from NamedTribe. A named tribe is an AppliesTo
		// input to its permanent's static abilities, so the layer
		// engine's cached resolution has to be invalidated when it
		// lands. A chosen NAME is not: the one reader is the
		// activation gate, which is asked fresh at every announce and
		// caches nothing — the same argument choose_player.go makes
		// for protection.
	}
	g.EmitEvent(Event{
		Kind:   EventCardNameChosen,
		Actor:  chooserID,
		CardID: choice.Source,
		Label:  chosen,
	})
	g.runStateChecksLocked()
	return nil
}

// ChosenNameOf returns the card name chosen for the permanent
// `sourceID` currently on the battlefield, or "" when none has been
// chosen (or the permanent is gone).
//
// The accessor exists for the reason NamedTribeOf, ChosenColorOf and
// ChosenPlayerOf do: a caller outside this package reads the answer
// without reaching into the battlefield slice itself.
//
// Caller must hold either lock.
func (g *Game) ChosenNameOf(sourceID uuid.UUID) string {
	if i := findCardOnBattlefield(g, sourceID); i >= 0 {
		return g.Battlefield.Cards[i].ChosenName
	}
	return ""
}

// CardNameMatches reports whether `c` has the name `named` for the
// purposes of a "cards named X" / "sources with the chosen name"
// clause (CR 201.2).
//
// Case-insensitive on trimmed strings, because the name arrives as
// free text a human typed and "sol ring" is not a different card
// from "Sol Ring".
//
// Asks every FACE as well as the card's own name (CR 201.2b): naming
// "Fire" names Fire // Ice, and naming "Brutal Cathar" names the
// permanent whose Moonrage Brute face is currently up. Card.Name is
// the ACTIVE face's name (the flat printed fields are the
// materialisation of Faces[ActiveFace]), so without the face walk a
// transformed permanent would stop matching the name it was named by.
//
// A face-down object has no name (CR 708.2a) and matches nothing; the
// face list of such a card is the face-down one, which carries no
// printed name to match.
func CardNameMatches(c Card, named string) bool {
	want := strings.TrimSpace(named)
	if want == "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(c.Name), want) {
		return true
	}
	for _, f := range c.Faces {
		if strings.EqualFold(strings.TrimSpace(f.Name), want) {
			return true
		}
	}
	return false
}

// PublicCardNamesLocked is the suggestion list the "choose a card
// name" picker is offered: every distinct card name visible in a
// PUBLIC zone right now, sorted, so the client can filter it instead
// of making the player type a name they are looking at.
//
// PUBLIC ONLY — the battlefield, every graveyard and the stack. Not a
// hand, not a library, not a face-down exile. The list goes on the
// wire to every viewer the prompt is projected to, and a suggestion
// list is not worth a hidden-information leak; the free-text entry is
// the general answer and covers every name this list does not.
//
// The names are the ACTIVE face's, which is what a player reading the
// board sees. A transformed permanent therefore suggests its back
// face; CardNameMatches asks both, so naming either works.
//
// Caller must hold g.mu (read or write).
func (g *Game) PublicCardNamesLocked() []string {
	seen := map[string]bool{}
	add := func(c Card) {
		n := strings.TrimSpace(c.Name)
		if n == "" || c.FaceDown {
			return
		}
		seen[n] = true
	}
	if g.Battlefield != nil {
		for _, c := range g.Battlefield.Cards {
			add(c)
		}
	}
	if g.Stack != nil {
		for _, c := range g.Stack.Cards {
			add(c)
		}
	}
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			add(c)
		}
	}
	out := make([]string, 0, len(seen))
	for n := range seen {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
