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
// A change with nothing printed after it (every card but Crystal Spray and
// Balduvian Shaman) asks both picks as keyed option picks (changeTextWordFirst,
// changeTextWordSecond; #4253 sweep), so a table waiting on either is a
// restore point. A change with a Then, and a creature-type change
// (ChangeCreatureTypeWord, which needs a keyed creature-type prompt), still
// hold a closure continuation, so a table waiting on them is not (the census
// counts it). The answer comes within the resolution that asked, as for a
// pile split.

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
	// Then is the rest of the card (Crystal Spray's "Draw a card."),
	// run once the words are chosen and the change made, or when no
	// change could be asked for. Optional.
	Then func(g *game.Game) error
}

func (t ChangeTextWord) Apply(ctx *Context) error {
	// #1432, CR 400.7: a source that left and came back is a new object
	// this ability knows nothing about.
	if t.Target == uuid.Nil || ctx.isNewSourceObject(t.Target) {
		return nil
	}
	words := textWordChoices(t.Kinds)
	if len(words) == 0 {
		return nil
	}
	g := ctx.Game
	chooser, source, target, eot, then := ctx.Controller(), ctx.Source(), t.Target, t.UntilEndOfTurn, t.Then
	name := cardNameFor(g, source)
	options := make([]game.ChoiceOption, len(words))
	for i, w := range words {
		options[i] = game.ChoiceOption{Label: w.label}
	}
	if then == nil {
		// Nothing printed after the change: the two picks are keyed (#4253
		// sweep), so a table waiting on either is a restore point.
		return PickOption{
			Question: name + " — choose the word to replace",
			Options:  options,
			ThenKey:  changeTextWordFirst,
			Carry:    append([]uuid.UUID{target}, intsToCarry([]int{textWordKindsMask(t.Kinds), textWordDuration(eot)})...),
		}.Apply(ctx)
	}
	queued := g.QueueOptionPickForEffect(game.OptionPickPrompt{
		Chooser:  chooser,
		Source:   source,
		Question: name + " — choose the word to replace",
		Options:  options,
		Then: func(g *game.Game, i int) error {
			if i < 0 || i >= len(words) {
				return then(g)
			}
			from := words[i]
			var to []textWordChoice
			for _, w := range textWordChoices([]game.TextWordKind{from.kind}) {
				if w.word != from.word {
					to = append(to, w)
				}
			}
			toOptions := make([]game.ChoiceOption, len(to))
			for j, w := range to {
				toOptions[j] = game.ChoiceOption{Label: w.label}
			}
			asked := g.QueueOptionPickForEffect(game.OptionPickPrompt{
				Chooser:  chooser,
				Source:   source,
				Question: fmt.Sprintf("%s — replace %s with", name, from.label),
				Options:  toOptions,
				Then: func(g *game.Game, j int) error {
					if j >= 0 && j < len(to) {
						changeTextWordNow(g, source, target, from.kind, from.word, to[j].word, from.label, to[j].label, name, eot)
					}
					return then(g)
				},
			})
			if asked == uuid.Nil {
				return then(g)
			}
			return nil
		},
	})
	if queued == uuid.Nil {
		return then(g)
	}
	return nil
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
// prompt as IDs. Registered in init, because the body asks the second pick,
// which is registered beside it.
var changeTextWordFirst game.OptionPickThen

// changeTextWordSecond is the word that replaces it, chosen: the same data and
// the index of the word replaced.
var changeTextWordSecond game.OptionPickThen

func init() {
	changeTextWordFirst = OptionPickThen("option-pick/change-text-word-replace", func(ctx *Context, r game.OptionPicked) error {
		if len(r.Carry) < 3 {
			return nil
		}
		nums := carryToInts(r.Carry[1:3])
		words := textWordChoices(textWordKindsFromMask(nums[0]))
		if r.Index < 0 || r.Index >= len(words) {
			return nil
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
		}.Apply(ctx)
	})
	changeTextWordSecond = OptionPickThen("option-pick/change-text-word-with", func(ctx *Context, r game.OptionPicked) error {
		if len(r.Carry) < 4 {
			return nil
		}
		nums := carryToInts(r.Carry[1:4])
		words := textWordChoices(textWordKindsFromMask(nums[0]))
		if nums[2] < 0 || nums[2] >= len(words) {
			return nil
		}
		from := words[nums[2]]
		to := textWordReplacements(from)
		if r.Index < 0 || r.Index >= len(to) {
			return nil
		}
		g, source := ctx.Game, ctx.Source()
		changeTextWordNow(g, source, r.Carry[0], from.kind, from.word, to[r.Index].word, from.label, to[r.Index].label, cardNameFor(g, source), nums[1] != 0)
		return nil
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
	chooser, source, target, eot := ctx.Controller(), ctx.Source(), t.Target, t.UntilEndOfTurn
	except := append([]string(nil), t.Except...)
	name := cardNameFor(g, source)
	if fixed := t.To; fixed != "" {
		g.QueueCreatureTypeChoiceThenForEffect(game.CreatureTypePrompt{
			Chooser:  chooser,
			Source:   source,
			Question: fmt.Sprintf("%s — choose the creature type to replace with %s", name, fixed),
			Except:   []string{fixed},
			Then: func(g *game.Game, from string) error {
				if from == "" || from == fixed {
					return nil
				}
				changeTextWordNow(g, source, target, game.TextWordCreatureType, from, fixed, from, fixed, name, eot)
				return nil
			},
		})
		return nil
	}
	g.QueueCreatureTypeChoiceThenForEffect(game.CreatureTypePrompt{
		Chooser:  chooser,
		Source:   source,
		Question: name + " — choose the creature type to replace",
		Then: func(g *game.Game, from string) error {
			if from == "" {
				return nil
			}
			g.QueueCreatureTypeChoiceThenForEffect(game.CreatureTypePrompt{
				Chooser:  chooser,
				Source:   source,
				Question: fmt.Sprintf("%s — replace %s with", name, from),
				Except:   append(append([]string(nil), except...), from),
				Then: func(g *game.Game, to string) error {
					if to == "" {
						return nil
					}
					changeTextWordNow(g, source, target, game.TextWordCreatureType, from, to, from, to, name, eot)
					return nil
				},
			})
			return nil
		},
	})
	return nil
}

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
func changeTextOfTarget(kinds []game.TextWordKind, untilEndOfTurn bool, then func(item *game.StackItem) func(*game.Game) error) func(*game.StackItem, *Context) error {
	return func(item *game.StackItem, ctx *Context) error {
		var rest func(*game.Game) error
		if then != nil {
			rest = then(item)
		}
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			return ChangeTextWord{Target: t.ID, Kinds: kinds, UntilEndOfTurn: untilEndOfTurn, Then: rest}.Apply(ctx)
		}
		if rest != nil {
			return rest(ctx.Game)
		}
		return nil
	}
}

var (
	textColorWords  = []game.TextWordKind{game.TextWordColor}
	textLandWords   = []game.TextWordKind{game.TextWordBasicLandType}
	textEitherWords = []game.TextWordKind{game.TextWordColor, game.TextWordBasicLandType}
)
