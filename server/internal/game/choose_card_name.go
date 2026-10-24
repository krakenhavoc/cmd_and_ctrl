package game

import (
	"fmt"
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

// CardNamePrompt is the as-enters card-name prompt with its two
// options spelled out (#1997): which of the source's names the answer
// fills, and whether a basic land card's name is refused.
type CardNamePrompt struct {
	// Chooser names the card. Source is the permanent the name is
	// chosen for. Reason is the prompt's header.
	Chooser, Source uuid.UUID
	Reason          string
	// SecondName stores the answer as the source's ChosenSecondName
	// instead of its ChosenName: the name another player chooses for it
	// (Null Chamber's opponent).
	SecondName bool
	// NotBasicLandName is "a card name other than a basic land card
	// name" (CR 201.4a).
	NotBasicLandName bool
}

// QueueCardNamePromptForEffect queues a CardNamePrompt and returns the
// choice ID, or uuid.Nil when its chooser has left the game. Plain data,
// like QueueCardNameChoiceForEffect: a table waiting on it is a restore
// point.
//
// Caller must hold g.mu.
func (g *Game) QueueCardNamePromptForEffect(p CardNamePrompt) uuid.UUID {
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:             PendingChoiceCardName,
		Chooser:          p.Chooser,
		Count:            1,
		Source:           p.Source,
		Reason:           p.Reason,
		SecondName:       p.SecondName,
		NotBasicLandName: p.NotBasicLandName,
	})
}

// basicLandCardNames are the names of the basic land cards in the Oracle
// card reference (CR 205.4c names the ten from before Eighth Edition;
// Wastes and Snow-Covered Wastes are the two since, and Barry's Land the
// one un-set basic). Lower case, for IsBasicLandCardName.
var basicLandCardNames = map[string]bool{
	"plains": true, "island": true, "swamp": true, "mountain": true, "forest": true, "wastes": true,
	"snow-covered plains": true, "snow-covered island": true, "snow-covered swamp": true,
	"snow-covered mountain": true, "snow-covered forest": true, "snow-covered wastes": true,
	"barry's land": true,
}

// IsBasicLandCardName reports whether `name` is the name of a basic land
// card, compared as every card name is (case-insensitive, trimmed): the
// names "a card name other than a basic land card name" rules out.
func IsBasicLandCardName(name string) bool {
	return basicLandCardNames[strings.ToLower(strings.TrimSpace(name))]
}

// QueueCardNameChoiceThenForEffect queues the AT-RESOLUTION form: the
// answer is handed to `then` and stored nowhere, so the permanent's
// ChosenName is untouched. `then` runs with g.mu held and may queue
// further prompts.
//
// Caller must hold g.mu.
func (g *Game) QueueCardNameChoiceThenForEffect(chooser, source uuid.UUID, reason string, then func(g *Game, name string) error) uuid.UUID {
	return g.QueueChoiceForEffect(PendingChoice{
		Kind:    PendingChoiceCardName,
		Chooser: chooser,
		Count:   1,
		Source:  source,
		Reason:  reason,
		// The ceiling on resume-frame routes only falls
		// (closure_fields.txt), so this rides the colour prompt's
		// frame: both are "hand the single string answer to a
		// closure", and the census already counts it.
		chooseValueResume: &chooseValueFrame{then: then},
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
	if choice.NotBasicLandName && IsBasicLandCardName(chosen) {
		// CR 201.4a: not a name this prompt may be answered with. The
		// prompt stays open for another answer, and the refusal says why.
		return fmt.Errorf("%w: a basic land card's name can't be chosen", ErrInvalidParam)
	}
	g.dequeueChoiceLocked(idx)

	if frame := choice.chooseValueResume; frame != nil {
		g.EmitEvent(Event{
			Kind:   EventCardNameChosen,
			Actor:  chooserID,
			CardID: choice.Source,
			Label:  chosen,
		})
		if frame.then != nil {
			if err := frame.then(g, chosen); err != nil {
				g.EmitEvent(Event{
					Kind:     EventEffectError,
					Actor:    chooserID,
					Source:   choice.Source,
					ErrorMsg: err.Error(),
				})
			}
		}
		g.runStateChecksLocked()
		return nil
	}

	if i := findCardOnBattlefield(g, choice.Source); i >= 0 {
		if choice.SecondName {
			g.Battlefield.Cards[i].ChosenSecondName = chosen
		} else {
			g.Battlefield.Cards[i].ChosenName = chosen
		}
		// The chosen name is an AppliesTo input to a static ability
		// since Alpine Moon (ADR 0109 §2): "Lands your opponents
		// control with the chosen name lose all land types and
		// abilities". The layer engine caches its resolution until
		// something invalidates it, and nothing else in this path
		// emits an event the layer listener watches, so the bump is
		// explicit, as it is for a named tribe. The activation gate
		// (Pithing Needle) is asked fresh at every announce and needs
		// none.
		g.layerVersion.Add(1)
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

// ChosenNamesOf returns every card name chosen for the permanent
// `sourceID` on the battlefield: its ChosenName and then its
// ChosenSecondName, each only once it has been answered. Nil for a
// permanent with none, or one that is gone. Null Chamber's two names
// (#1997).
//
// Caller must hold either lock.
func (g *Game) ChosenNamesOf(sourceID uuid.UUID) []string {
	i := findCardOnBattlefield(g, sourceID)
	if i < 0 {
		return nil
	}
	return ChosenNamesOn(&g.Battlefield.Cards[i])
}

// ChosenNamesOn is ChosenNamesOf for a permanent already in hand: its
// ChosenName and then its ChosenSecondName, each only once answered.
func ChosenNamesOn(c *Card) []string {
	if c == nil {
		return nil
	}
	var out []string
	for _, n := range []string{c.ChosenName, c.ChosenSecondName} {
		if strings.TrimSpace(n) != "" {
			out = append(out, n)
		}
	}
	return out
}

// SpellOrLandHasName reports whether the card `c` being cast or played
// has the name `named`: "spells with the chosen name can't be cast" and
// "lands with the chosen name can't be played" (Null Chamber, #1997).
//
// The card is the one the cast or play gate is handed, with the face
// being cast or played already materialised, so Card.Name is the name it
// has on the stack or will have on the battlefield. Only that face
// counts: a split card on the stack has only the half being cast (CR
// 709.3b), and a modal double-faced card only the face cast or played
// (CR 712.11c, CR 712.12). A fused split spell has both halves' names (CR 709.4d),
// which the engine spells "Fire // Ice", so each half of a name with
// " // " in it is asked too. Compared as every card name is
// (case-insensitive, trimmed).
func SpellOrLandHasName(c Card, named string) bool {
	want := strings.TrimSpace(named)
	if want == "" || c.FaceDown {
		return false
	}
	for _, part := range strings.Split(c.Name, " // ") {
		if strings.EqualFold(strings.TrimSpace(part), want) {
			return true
		}
	}
	return false
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

// PermanentHasName reports whether the permanent `c` has the name
// `named` right now: its effective name, which is the name of the face
// that is up (CR 712.8d-f) after any copy effect, compared the way
// CardNameMatches compares (case-insensitive on trimmed strings). A
// face-down permanent has no name (CR 708.2a).
//
// It is the reader for a chosen name with characteristics, where
// CardNameMatches' every-face walk would be too wide. Alpine Moon
// names "a nonbasic land card name" (CR 201.4a), and the face a name
// may be chosen from is judged by that face's own characteristics (CR
// 201.4d). Naming the instant front face of a modal double-faced land
// is not a legal choice, so it must not reach the land face that is up;
// naming the land face itself is, and does. Inside a layer pass this
// reads the name the pass has computed so far, so a copy effect (layer
// 1) is already in it.
func PermanentHasName(c Card, named string) bool {
	want := strings.TrimSpace(named)
	if want == "" || c.FaceDown {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(c.Effective().Name), want)
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
