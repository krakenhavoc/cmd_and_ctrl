package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// text_words.go — "change the text of target spell or permanent by
// replacing all instances of one color word with another" (CR 612.2;
// ADR 0160's 2026-10-15 amendment, #2779). Glamerdye, Sleight of Mind,
// Alter Reality, Magical Hack, Mind Bend, Crystal Spray, Whim of
// Volrath.
//
// The words are chosen as the spell RESOLVES ("You choose the words to
// change on resolution", the cards' ruling), in two ordinary option
// picks: the word to replace, then the word that replaces it, from the
// same class. "It can't change a word to the same word", so the second
// list leaves the first answer out. The engine side, and what "the
// text" means in an engine whose rules are closures, is
// game/text_words.go.
//
// Both picks are keyed option picks (changeTextWordFirst,
// changeTextWordSecond; #4253 sweep), and what is printed after the change
// (Crystal Spray's draw, Balduvian Shaman's grant) is a registered tail they
// hand on, so a table waiting on either is a restore point. A creature-type
// change (ChangeCreatureTypeWord) asks through keyed creature-type prompts
// (the sweep's second pass) for the same reason. The answer comes within the
// resolution that asked, as for a pile split.

// textWordChoice is one word a change can replace.
type textWordChoice struct {
	kind  game.TextWordKind
	word  string // the stored spelling: a WUBRG letter, or a land type
	label string // what the chooser reads
}

// textWordChoices lists every word of the given classes, colours WUBRG
// first, then the basic land types in CR 305.6 order.
func textWordChoices(kinds []game.TextWordKind) []textWordChoice {
	var out []textWordChoice
	for _, k := range kinds {
		switch k {
		case game.TextWordColor:
			for _, l := range []string{"W", "U", "B", "R", "G"} {
				out = append(out, textWordChoice{kind: k, word: l, label: capitalise(game.ColorName(l))})
			}
		case game.TextWordBasicLandType:
			for _, t := range game.BasicLandTypes {
				out = append(out, textWordChoice{kind: k, word: t, label: t})
			}
		}
	}
	return out
}

func capitalise(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-'a'+'A') + s[1:]
}

// ChangeTextWord is "change the text of <Target> by replacing all
// instances of one <kind> with another" (CR 612.2). Kinds is the class
// or classes the card names: a colour word (Sleight of Mind), a basic
// land type (Magical Hack), or either (Mind Bend: "one color word with
// another or one basic land type with another" — ONE change, of either
// class). UntilEndOfTurn is Crystal Spray's and Whim of Volrath's
// duration; otherwise the change lasts indefinitely.
//
// The controller picks the word to replace, then its replacement. With
// no live target there is nothing to change and nothing is asked. A
// creature type is ChangeCreatureTypeWord: its vocabulary is the
// creature-type prompt's, not an option list.
type ChangeTextWord struct {
	Target         uuid.UUID
	Kinds          []game.TextWordKind
	UntilEndOfTurn bool
	// ThenTail is the rest of the card (Crystal Spray's "Draw a card."),
	// a registered tail run once the words are chosen and the change made,
	// or when no change could be asked for. Optional.
	ThenTail game.KeyedTail
}

func (t ChangeTextWord) Apply(ctx *Context) error {
	// #1432, CR 400.7: a source that left and came back is a new object
	// this ability knows nothing about.
	if t.Target == uuid.Nil || ctx.isNewSourceObject(t.Target) {
		return nil
	}
	words := textWordChoices(t.Kinds)
	if len(words) == 0 {
		return t.ThenTail.Run(ctx.Game, nil)
	}
	options := make([]game.ChoiceOption, len(words))
	for i, w := range words {
		options[i] = game.ChoiceOption{Label: w.label}
	}
	return PickOption{
		Question: cardNameFor(ctx.Game, ctx.Source()) + " — choose the word to replace",
		Options:  options,
		ThenKey:  changeTextWordFirst,
		Carry:    append([]uuid.UUID{t.Target}, intsToCarry([]int{textWordKindsMask(t.Kinds), textWordDuration(t.UntilEndOfTurn)})...),
		Tail:     t.ThenTail,
	}.Apply(ctx)
}

// textWordKindsMask writes the classes of words as bits (colour 1, basic land
// type 2), for a prompt that carries only IDs and small numbers. The order the
// card named them in is not kept: textWordChoices lists colours first.
func textWordKindsMask(kinds []game.TextWordKind) int {
	mask := 0
	for _, k := range kinds {
		switch k {
		case game.TextWordColor:
			mask |= 1
		case game.TextWordBasicLandType:
			mask |= 2
		}
	}
	return mask
}

// textWordKindsFromMask is the inverse of textWordKindsMask.
func textWordKindsFromMask(mask int) []game.TextWordKind {
	var kinds []game.TextWordKind
	if mask&1 != 0 {
		kinds = append(kinds, game.TextWordColor)
	}
	if mask&2 != 0 {
		kinds = append(kinds, game.TextWordBasicLandType)
	}
	return kinds
}

// textWordDuration is the change's duration as a small number for a prompt's
// Carry: 1 for until end of turn, 0 for indefinitely.
func textWordDuration(untilEndOfTurn bool) int {
	if untilEndOfTurn {
		return 1
	}
	return 0
}

// textWordReplacements is the words `from` can become: every word of its own
// class but itself ("it can't change a word to the same word").
func textWordReplacements(from textWordChoice) []textWordChoice {
	var to []textWordChoice
	for _, w := range textWordChoices([]game.TextWordKind{from.kind}) {
		if w.word != from.word {
			to = append(to, w)
		}
	}
	return to
}

// changeTextWordFirst is the word to replace, chosen, keyed (#4253 sweep): the
// target, the classes of words (textWordKindsMask) and the duration ride the
// prompt as IDs, and the rest of the card as its tail (r.Tail), run when no
// word is chosen. Registered in init, because the body asks the second pick,
// which is registered beside it.
var changeTextWordFirst game.OptionPickThen

// changeTextWordSecond is the word that replaces it, chosen: the same data and
// the index of the word replaced. The tail runs after the change, or in its
// place when none is chosen.
var changeTextWordSecond game.OptionPickThen

func init() {
	changeTextWordFirst = OptionPickThen("option-pick/change-text-word-replace", func(ctx *Context, r game.OptionPicked) error {
		if len(r.Carry) < 3 {
			return r.Tail.Run(ctx.Game, nil)
		}
		nums := carryToInts(r.Carry[1:3])
		words := textWordChoices(textWordKindsFromMask(nums[0]))
		if r.Index < 0 || r.Index >= len(words) {
			return r.Tail.Run(ctx.Game, nil)
		}
		from := words[r.Index]
		to := textWordReplacements(from)
		toOptions := make([]game.ChoiceOption, len(to))
		for j, w := range to {
			toOptions[j] = game.ChoiceOption{Label: w.label}
		}
		return PickOption{
			Question: fmt.Sprintf("%s — replace %s with", cardNameFor(ctx.Game, ctx.Source()), from.label),
			Options:  toOptions,
			ThenKey:  changeTextWordSecond,
			Carry:    append(append([]uuid.UUID(nil), r.Carry...), intsToCarry([]int{r.Index})...),
			Tail:     r.Tail,
		}.Apply(ctx)
	})
	changeTextWordSecond = OptionPickThen("option-pick/change-text-word-with", func(ctx *Context, r game.OptionPicked) error {
		if len(r.Carry) < 4 {
			return r.Tail.Run(ctx.Game, nil)
		}
		nums := carryToInts(r.Carry[1:4])
		words := textWordChoices(textWordKindsFromMask(nums[0]))
		if nums[2] < 0 || nums[2] >= len(words) {
			return r.Tail.Run(ctx.Game, nil)
		}
		from := words[nums[2]]
		to := textWordReplacements(from)
		if r.Index >= 0 && r.Index < len(to) {
			g, source := ctx.Game, ctx.Source()
			changeTextWordNow(g, source, r.Carry[0], from.kind, from.word, to[r.Index].word, from.label, to[r.Index].label, cardNameFor(g, source), nums[1] != 0)
		}
		return r.Tail.Run(ctx.Game, nil)
	})
}

// ChangeCreatureTypeWord is ChangeTextWord for a creature type
// (Artificial Evolution): both words are asked through the
// choose-a-creature-type prompt. Except names types the NEW word may not
// be ("The new creature type can't be Wall"); the replaced word is
// always excluded too, since a word can't be changed to itself.
//
// To fixes the new word instead ("replacing all instances of one
// creature type with Vampire", New Blood, #3868): only the replaced word
// is asked, and To itself is not offered for it.
type ChangeCreatureTypeWord struct {
	Target         uuid.UUID
	Except         []string
	To             string
	UntilEndOfTurn bool
}

func (t ChangeCreatureTypeWord) Apply(ctx *Context) error {
	if t.Target == uuid.Nil || ctx.isNewSourceObject(t.Target) {
		return nil
	}
	g := ctx.Game
	name := cardNameFor(g, ctx.Source())
	then := withCarry(withN(ctx.Tail(changeCreatureTypeWordReplace), textWordDuration(t.UntilEndOfTurn)), t.Target)
	if fixed := t.To; fixed != "" {
		then.Ref = changeCreatureTypeWordToFixed
		then.Words = []string{fixed}
		g.QueueCreatureTypeChoiceThenForEffect(game.CreatureTypePrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: fmt.Sprintf("%s — choose the creature type to replace with %s", name, fixed),
			Except:   []string{fixed},
			ThenTail: then,
		})
		return nil
	}
	then.Words = append([]string(nil), t.Except...)
	g.QueueCreatureTypeChoiceThenForEffect(game.CreatureTypePrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: name + " — choose the creature type to replace",
		ThenTail: then,
	})
	return nil
}

// The creature-type change's answers, keyed (#4253 sweep, second pass). Each
// runs against the stand-in of the asking effect's controller and source
// (ctx.Tail); the target is the tail's one carried ID and the duration its N
// (textWordDuration). On-disk keys.
var (
	// changeCreatureTypeWordToFixed is the replaced word (the tail's Word)
	// chosen when the new word is fixed (its first Word).
	changeCreatureTypeWordToFixed = CardTail("then/change-creature-type-word-to-fixed", func(ctx *Context, o game.TailOutcome) error {
		from := o.Tail.Word
		if from == "" || len(o.Tail.Carry) == 0 || len(o.Tail.Words) == 0 || from == o.Tail.Words[0] {
			return nil
		}
		to := o.Tail.Words[0]
		changeTextWordNow(ctx.Game, ctx.Source(), o.Tail.Carry[0], game.TextWordCreatureType, from, to, from, to,
			cardNameFor(ctx.Game, ctx.Source()), o.Tail.N != 0)
		return nil
	})

	// changeCreatureTypeWordReplace is the replaced word (the tail's Word)
	// chosen; it asks the new one, which may be none of the tail's Words
	// (the card's exceptions) nor the replaced word itself.
	changeCreatureTypeWordReplace = CardTail("then/change-creature-type-word-replace", func(ctx *Context, o game.TailOutcome) error {
		from := o.Tail.Word
		if from == "" || len(o.Tail.Carry) == 0 {
			return nil
		}
		then := o.Tail
		then.Ref = changeCreatureTypeWordWith
		then.Word = ""
		then.Words = []string{from}
		ctx.Game.QueueCreatureTypeChoiceThenForEffect(game.CreatureTypePrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: fmt.Sprintf("%s — replace %s with", cardNameFor(ctx.Game, ctx.Source()), from),
			Except:   append(append([]string(nil), o.Tail.Words...), from),
			ThenTail: then,
		})
		return nil
	})

	// changeCreatureTypeWordWith is the new word (the tail's Word) chosen for
	// the replaced one (its first Word).
	changeCreatureTypeWordWith = CardTail("then/change-creature-type-word-with", func(ctx *Context, o game.TailOutcome) error {
		to := o.Tail.Word
		if to == "" || len(o.Tail.Carry) == 0 || len(o.Tail.Words) == 0 {
			return nil
		}
		from := o.Tail.Words[0]
		changeTextWordNow(ctx.Game, ctx.Source(), o.Tail.Carry[0], game.TextWordCreatureType, from, to, from, to,
			cardNameFor(ctx.Game, ctx.Source()), o.Tail.N != 0)
		return nil
	})
)

// changeTextWordNow registers the change once both words are known.
func changeTextWordNow(g *game.Game, source, target uuid.UUID, kind game.TextWordKind, from, to, fromLabel, toLabel, name string, eot bool) {
	d := game.IndefiniteDuration()
	if eot {
		d = g.UntilEndOfTurnDuration()
	}
	label := fmt.Sprintf("%s — %s becomes %s", name, fromLabel, toLabel)
	g.ChangeTextWordForEffect(source, target, game.TextWordChange{Kind: kind, From: from, To: to}, d, label)
}

// cardNameFor is the name of the card `id`, wherever it is, or "This
// spell" when it cannot be found.
func cardNameFor(g *game.Game, id uuid.UUID) string {
	if c, ok := g.LookupCardForEffect(id); ok && c.Name != "" {
		return c.Name
	}
	return "This spell"
}

// changeTextOfTarget is the OnResolve of every "change the text of
// target spell or permanent" card: the change applied to the first
// target still legal (CR 608.2b already countered a spell whose only
// target left).
//
// `then` is what is printed after the change, a registered tail run against
// the asking item (ctx.ItemTail); the zero TailRef when nothing is.
func changeTextOfTarget(kinds []game.TextWordKind, untilEndOfTurn bool, then game.TailRef) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		var rest game.KeyedTail
		if then.Key() != "" {
			rest = ctx.ItemTail(then)
		}
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return ChangeTextWord{Target: t.ID, Kinds: kinds, UntilEndOfTurn: untilEndOfTurn, ThenTail: rest}.Apply(ctx)
		}
		return rest.Run(ctx.Game, nil)
	}
}

// nothingAfterTheChange is changeTextOfTarget's `then` for a card with
// nothing printed after the change: the zero TailRef.
var nothingAfterTheChange game.TailRef

var (
	textColorWords  = []game.TextWordKind{game.TextWordColor}
	textLandWords   = []game.TextWordKind{game.TextWordBasicLandType}
	textEitherWords = []game.TextWordKind{game.TextWordColor, game.TextWordBasicLandType}
)
