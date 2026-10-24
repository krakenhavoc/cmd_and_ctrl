package game

import (
	"strings"

	"github.com/google/uuid"
)

// sticker_prompt.go — "you may put a sticker on <an object you own>" (CR
// 123.3, ADR 0208 §4). Three keyed questions, each a restore point:
//
//  1. which object (a choose_cards pick of one or none: "you may"), with
//     the sticker kinds the instruction allows riding as the prompt's word;
//  2. which sticker (an option pick over the stickers the player has
//     access to, can afford and has not used, CR 123.3 and 123.3c, plus
//     "Don't put a sticker"), the object and the stickers riding as Carry;
//  3. for a name sticker, where in the name it goes (CR 123.6b): an option
//     per position, asked of the object's controller.
//
// The ticket cost is paid as the sticker is put on (placeStickerLocked),
// by the object's owner, who is the player (CR 123.3b, 123.3c). Every key
// is an on-disk identity: never renamed, never reused.

const (
	stickerObjectKey   = "sticker/choose-object"
	stickerPickKey     = "sticker/choose-sticker"
	stickerPositionKey = "sticker/name-position"
)

var (
	stickerObjectThen   ChooseCardsThen
	stickerPickThen     OptionPickThen
	stickerPositionThen OptionPickThen
)

func init() {
	stickerObjectThen = RegisterChooseCardsThen(stickerObjectKey, stickerObjectChosen)
	stickerPickThen = RegisterOptionPickThen(stickerPickKey, stickerPicked)
	stickerPositionThen = RegisterOptionPickThen(stickerPositionKey, stickerPositioned)
}

// StickerRequest is one "you may put a sticker on …" instruction.
type StickerRequest struct {
	// Player puts the sticker; Source is the card whose effect it is.
	Player, Source uuid.UUID
	// Objects are the objects the instruction names ("a nonland
	// permanent you own"). Any that is not the player's, or not in a
	// public zone, is dropped (CR 123.3b, 123.5).
	Objects []uuid.UUID
	// Kinds are the kinds of sticker the instruction allows ("a name
	// sticker", "a power and toughness sticker"); empty is any sticker.
	Kinds []StickerKind
	// Question heads the first prompt; empty uses a default.
	Question string
}

// DeclineStickerLabel is the option that puts no sticker.
const DeclineStickerLabel = "Don't put a sticker"

// PutStickerForEffect asks the player to put a sticker on one of the
// objects (CR 123.3). It asks nothing, and does nothing, when no object
// qualifies or the player has no sticker of an allowed kind they could
// put on (no sheets, every one used, or none affordable).
//
// Caller must hold g.mu in write mode.
func (g *Game) PutStickerForEffect(r StickerRequest) {
	g.RecomputeLayersIfStaleLocked()
	var objects []uuid.UUID
	for _, id := range r.Objects {
		if host := g.stickerHostLocked(id); host != nil && host.Owner == r.Player {
			objects = append(objects, id)
		}
	}
	if len(objects) == 0 || len(g.AvailableStickersLocked(r.Player, r.Kinds)) == 0 {
		return
	}
	question := r.Question
	if question == "" {
		question = "You may put a sticker on a permanent you own. Choose it, or none."
	}
	g.QueueChooseCardsForEffect(ChooseCardsPrompt{
		Chooser:   r.Player,
		Source:    r.Source,
		Question:  question,
		Cards:     objects,
		Min:       0,
		Max:       1,
		ThenKey:   stickerObjectThen,
		CarryWord: stickerKindsWord(r.Kinds),
	})
}

// stickerKindsWord writes the allowed kinds as the prompt's word, and
// parseStickerKinds reads them back.
func stickerKindsWord(kinds []StickerKind) string {
	parts := make([]string, 0, len(kinds))
	for _, k := range kinds {
		parts = append(parts, string(k))
	}
	return strings.Join(parts, ",")
}

func parseStickerKinds(word string) []StickerKind {
	if word == "" {
		return nil
	}
	var out []StickerKind
	for _, p := range strings.Split(word, ",") {
		switch k := StickerKind(strings.TrimSpace(p)); k {
		case StickerName, StickerAbility, StickerPT, StickerArt:
			out = append(out, k)
		}
	}
	return out
}

// stickerObjectChosen is question 1's answer: ask which sticker.
func stickerObjectChosen(g *Game, r CardsChosen) error {
	if !r.Answered || len(r.Picked) == 0 {
		return nil
	}
	obj := r.Picked[0]
	host := g.stickerHostLocked(obj)
	if host == nil || host.Owner != r.Chooser {
		return nil
	}
	choices := g.AvailableStickersLocked(r.Chooser, parseStickerKinds(r.Word))
	if len(choices) == 0 {
		return nil
	}
	options := make([]ChoiceOption, 0, len(choices)+1)
	options = append(options, ChoiceOption{Label: DeclineStickerLabel})
	carry := make([]uuid.UUID, 0, len(choices)+1)
	carry = append(carry, obj)
	for _, s := range choices {
		options = append(options, ChoiceOption{Label: s.Label()})
		carry = append(carry, s.ID)
	}
	g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  r.Chooser,
		Source:   r.Source,
		Question: "Which sticker do you put on " + EffectiveOf(host).Name + "?",
		Options:  options,
		ThenKey:  stickerPickThen,
		Carry:    carry,
	})
	return nil
}

// stickerPicked is question 2's answer: put the sticker on, or, for a
// name sticker, ask where in the name it goes.
func stickerPicked(g *Game, r OptionPicked) error {
	if r.Option == nil || r.Index <= 0 || len(r.Carry) < 2 {
		return nil
	}
	obj, ids := r.Carry[0], r.Carry[1:]
	i := r.Index - 1
	if i >= len(ids) {
		return nil
	}
	choice, ok := g.stickerChoiceByIDLocked(r.Chooser, ids[i])
	if !ok {
		return nil
	}
	if choice.Print.Kind != StickerName {
		g.placeStickerLocked(r.Chooser, obj, choice.ID, 0, r.Source)
		return nil
	}
	host := g.stickerHostLocked(obj)
	if host == nil {
		return nil
	}
	g.RecomputeLayersIfStaleLocked()
	name := EffectiveOf(host).Name
	words := NameWordCount(name)
	options := make([]ChoiceOption, 0, words+1)
	for k := 0; k <= words; k++ {
		options = append(options, ChoiceOption{Label: InsertNameWord(name, choice.Print.Text, k)})
	}
	// CR 123.6b: the object's controller chooses the position. A card
	// off the battlefield is controlled by its owner.
	chooser := host.Controller
	if chooser == uuid.Nil {
		chooser = host.Owner
	}
	id := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  chooser,
		Source:   r.Source,
		Question: "Where in the name does “" + choice.Print.Text + "” go?",
		Options:  options,
		ThenKey:  stickerPositionThen,
		Carry:    []uuid.UUID{obj, choice.ID, r.Chooser},
	})
	if id == uuid.Nil {
		// The controller has left the game: the word goes at the end.
		g.placeStickerLocked(r.Chooser, obj, choice.ID, words, r.Source)
	}
	return nil
}

// stickerPositioned is question 3's answer: put the name sticker on after
// that many words. An unanswered prompt puts it at the end.
func stickerPositioned(g *Game, r OptionPicked) error {
	if len(r.Carry) < 3 {
		return nil
	}
	obj, id, owner := r.Carry[0], r.Carry[1], r.Carry[2]
	after := r.Index
	if r.Option == nil || after < 0 {
		after = 0
		if host := g.stickerHostLocked(obj); host != nil {
			after = NameWordCount(EffectiveOf(host).Name)
		}
	}
	g.placeStickerLocked(owner, obj, id, after, r.Source)
	return nil
}
