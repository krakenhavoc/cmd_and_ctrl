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
// The pick holds a closure continuation, so a table waiting on it is not
// a restore point (the census counts it). The answer comes within the
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
	if then == nil {
		then = func(*game.Game) error { return nil }
	}
	name := cardNameFor(g, source)
	options := make([]game.ChoiceOption, len(words))
	for i, w := range words {
		options[i] = game.ChoiceOption{Label: w.label}
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

// ChangeCreatureTypeWord is ChangeTextWord for a creature type
// (Artificial Evolution): both words are asked through the
// choose-a-creature-type prompt. Except names types the NEW word may not
// be ("The new creature type can't be Wall"); the replaced word is
// always excluded too, since a word can't be changed to itself.
type ChangeCreatureTypeWord struct {
	Target         uuid.UUID
	Except         []string
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
