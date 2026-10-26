package game

import (
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// sticker_prompt.go — "you may put a sticker on <an object you own>" (CR
// 123.3, ADR 0208 §4). Three keyed questions, each a restore point:
//
//  1. which object (a choose_cards pick of one or none: "you may"), with
//     the sticker kinds the instruction allows, and any limit on the
//     sticker's ticket cost, riding as the prompt's word;
//  2. which sticker (an option pick over the stickers the player has
//     access to, can afford and has not used, CR 123.3 and 123.3c, plus
//     "Don't put a sticker"), the object and the stickers riding as Carry;
//  3. for a name sticker, where in the name it goes (CR 123.6b): an option
//     per position, asked of the object's controller.
//
// The ticket cost is paid as the sticker is put on (placeStickerLocked),
// by the object's owner, who is the player (CR 123.3b, 123.3c). An
// instruction that puts it on "without paying that sticker's ticket cost"
// (Pin Collection) is a Free limit: the questions two and three carry a
// marker first, so the stickers are offered whatever the owner has and the
// cost is not paid. Every key is an on-disk identity: never renamed, never
// reused.

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

// stickerFreeMarker leads the Carry of the second and third question when
// the sticker goes on without paying its ticket cost. No object or sticker
// has this ID, so a prompt asked before this existed (no marker) reads as it
// always did.
var stickerFreeMarker = uuid.NewSHA1(stickerIDSpace, []byte("free"))

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
	// Limit narrows the stickers further: "with ticket cost X or less",
	// "without paying that sticker's ticket cost".
	Limit StickerLimit
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
	if len(objects) == 0 || len(g.availableStickersLimitedLocked(r.Player, r.Kinds, r.Limit)) == 0 {
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
		CarryWord: stickerRequestWord(r.Kinds, r.Limit),
	})
}

// stickerRequestWord writes the allowed kinds and the limit as the
// prompt's word: "name,art;max=3;free", and stickerRequestFromWord reads
// them back. The kinds alone ("name") is the word it has always been.
func stickerRequestWord(kinds []StickerKind, lim StickerLimit) string {
	parts := make([]string, 0, len(kinds))
	for _, k := range kinds {
		parts = append(parts, string(k))
	}
	word := strings.Join(parts, ",")
	if lim.HasMax {
		word += ";max=" + strconv.Itoa(lim.MaxCost)
	}
	if lim.Free {
		word += ";free"
	}
	return word
}

func stickerRequestFromWord(word string) ([]StickerKind, StickerLimit) {
	var lim StickerLimit
	kindsPart, rest, _ := strings.Cut(word, ";")
	for _, f := range strings.Split(rest, ";") {
		switch {
		case f == "free":
			lim.Free = true
		case strings.HasPrefix(f, "max="):
			if n, err := strconv.Atoi(strings.TrimPrefix(f, "max=")); err == nil {
				lim.HasMax, lim.MaxCost = true, n
			}
		}
	}
	if kindsPart == "" {
		return nil, lim
	}
	var out []StickerKind
	for _, p := range strings.Split(kindsPart, ",") {
		switch k := StickerKind(strings.TrimSpace(p)); k {
		case StickerName, StickerAbility, StickerPT, StickerArt:
			out = append(out, k)
		}
	}
	return out, lim
}

// stickerCarryLimit reads the free marker off the front of a carry and
// returns the limit it stands for and the rest.
func stickerCarryLimit(carry []uuid.UUID) (StickerLimit, []uuid.UUID) {
	if len(carry) > 0 && carry[0] == stickerFreeMarker {
		return StickerLimit{Free: true}, carry[1:]
	}
	return StickerLimit{}, carry
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
	kinds, lim := stickerRequestFromWord(r.Word)
	choices := g.availableStickersLimitedLocked(r.Chooser, kinds, lim)
	if len(choices) == 0 {
		return nil
	}
	options := make([]ChoiceOption, 0, len(choices)+1)
	options = append(options, ChoiceOption{Label: DeclineStickerLabel})
	carry := make([]uuid.UUID, 0, len(choices)+2)
	if lim.Free {
		carry = append(carry, stickerFreeMarker)
	}
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
	lim, carry := stickerCarryLimit(r.Carry)
	if r.Option == nil || r.Index <= 0 || len(carry) < 2 {
		return nil
	}
	obj, ids := carry[0], carry[1:]
	i := r.Index - 1
	if i >= len(ids) {
		return nil
	}
	choice, ok := g.stickerChoiceByIDLocked(r.Chooser, ids[i], lim)
	if !ok {
		return nil
	}
	if choice.Print.Kind != StickerName {
		g.placeStickerUnderLocked(r.Chooser, obj, choice.ID, 0, r.Source, lim)
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
	carryOn := []uuid.UUID{obj, choice.ID, r.Chooser}
	if lim.Free {
		carryOn = append([]uuid.UUID{stickerFreeMarker}, carryOn...)
	}
	id := g.QueueOptionPickForEffect(OptionPickPrompt{
		Chooser:  chooser,
		Source:   r.Source,
		Question: "Where in the name does “" + choice.Print.Text + "” go?",
		Options:  options,
		ThenKey:  stickerPositionThen,
		Carry:    carryOn,
	})
	if id == uuid.Nil {
		// The controller has left the game: the word goes at the end.
		g.placeStickerUnderLocked(r.Chooser, obj, choice.ID, words, r.Source, lim)
	}
	return nil
}

// stickerPositioned is question 3's answer: put the name sticker on after
// that many words. An unanswered prompt puts it at the end.
func stickerPositioned(g *Game, r OptionPicked) error {
	lim, carry := stickerCarryLimit(r.Carry)
	if len(carry) < 3 {
		return nil
	}
	obj, id, owner := carry[0], carry[1], carry[2]
	after := r.Index
	if r.Option == nil || after < 0 {
		after = 0
		if host := g.stickerHostLocked(obj); host != nil {
			after = NameWordCount(EffectiveOf(host).Name)
		}
	}
	g.placeStickerUnderLocked(owner, obj, id, after, r.Source, lim)
	return nil
}
