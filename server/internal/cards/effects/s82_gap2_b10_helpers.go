package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b10_helpers.go — shared helpers for the GAP2 batch 10 cards
// (S82, #5948).

// IsLeavingTheBattlefield reports whether ev is a permanent leaving the
// battlefield, for any destination (CR 603.6c): the ordinary LTB event,
// or a zone move out of the battlefield.
func IsLeavingTheBattlefield(ev game.Event) bool {
	if ev.Kind == game.EventLTB {
		return true
	}
	return ev.Kind == game.EventZoneMove && ev.OldZone == game.ZoneBattlefield && ev.NewZone != game.ZoneBattlefield
}

// DoublesEnteringOrLeaving declares "If a <filter> entering or leaving
// the battlefield causes a triggered ability of a permanent you control to
// trigger, that ability triggers an additional time" (Gandalf the White,
// CR 603.2d). The entering object is judged as it is on the battlefield,
// the leaving one as it last existed there (the harvest's subject).
func DoublesEnteringOrLeaving(filter CardPredicate, rowLabel string) game.TriggerDoubler {
	return game.TriggerDoubler{
		Label:    "entering or leaving",
		RowLabel: rowLabel,
		Applies: func(g *game.Game, q game.TriggerDoublingQuery) bool {
			if q.FromSpell || !q.HasSubject || !q.SourceIsPermanent {
				return false
			}
			if !IsEntering(q.Event) && !IsLeavingTheBattlefield(q.Event) {
				return false
			}
			if q.SourceLKI.Controller != q.DoublerLKI.Controller {
				return false
			}
			return filter == nil || filter(g, q.DoublerLKI.Controller, subjectCard(g, q))
		},
	}
}

// printedMarkOfTheRaniToken is The Rani's Mark of the Rani: a red Aura
// enchantment token with enchant creature and "Enchanted creature gets
// +2/+2 and is goaded." Its slug is an on-disk identity: never renamed.
func printedMarkOfTheRaniToken() TokenTemplate {
	return TokenTemplate{
		Slug: "mark-of-the-rani",
		Card: game.Card{
			Name:     "Mark of the Rani",
			TypeLine: "Token Enchantment — Aura",
			Colors:   []string{"R"},
		},
		Enchant: TargetCreature("creature"),
		Static:  []game.StaticAbility{PumpAttached(2, 2), GoadAttached()},
		Text:    "Enchant creature\nEnchanted creature gets +2/+2 and is goaded.",
	}
}

// MarkOfTheRaniToken is that Mark's card value for CreateTokensForEffect.
func MarkOfTheRaniToken() game.Card { return TokenFromCatalog(printedMarkOfTheRaniToken) }

// CreatureWithChosenCreatureCardName reports whether the creature `c` has
// one of `names` as its name, where that name is a creature card name.
// "Choose a creature card name" (Day of the Moon) is a restriction on the
// CHOICE, and the engine takes a card name as free text with no vocabulary
// to check it against (game/choose_card_name.go; Petrified Hamlet's land
// card name). So it is kept where the name is read: a creature matches only
// when the object it is a copy of, or its own card, is a creature card by
// that name (its copiable values, CR 707.2). An animated land named in the
// prompt reaches nothing, as a name no legal choice could make.
func CreatureWithChosenCreatureCardName(c game.Card, names []string) bool {
	if c.FaceDown || !c.IsCreature() {
		return false
	}
	v := game.CopiableValuesOf(c)
	if !strings.Contains(strings.ToLower(v.TypeLine), "creature") {
		return false
	}
	current := c.Effective().Name
	for _, n := range names {
		n = strings.TrimSpace(n)
		if n != "" && strings.EqualFold(strings.TrimSpace(current), n) && strings.EqualFold(strings.TrimSpace(v.Name), n) {
			return true
		}
	}
	return false
}

// NoteACreatureTypeThenTail asks `chooser` for a creature type that hasn't
// been noted for `source` (Card.Notes, #6097) and hands it to `then`, a
// registered tail told the type as its Word ("" when nobody chose). The
// noted types are offered as nothing: the prompt's Except.
func NoteACreatureTypeThenTail(g *game.Game, chooser, source uuid.UUID, question string, then game.KeyedTail) bool {
	return g.QueueCreatureTypeChoiceThenForEffect(game.CreatureTypePrompt{
		Chooser:  chooser,
		Source:   source,
		Question: question,
		ThenTail: then,
		Except:   g.NotesForEffect(source),
	}) != uuid.Nil
}

// volosJournalNoteTrigger is Volo's Journal's "Whenever you cast a creature
// spell, note one of its creature types that hasn't been noted for this
// artifact."
func volosJournalNoteTrigger() game.TriggeredAbility {
	return On(game.EventCast, YouCastACreatureSpell,
		"Volo's Journal — note one of its creature types that hasn't been noted for this artifact",
		volosJournalNote)
}

// volosJournalNote reads the spell's creature types as it is now, or as it
// last was on the stack if it has gone (CR 608.2h). A spell with every
// creature type (changeling, CR 702.73a) asks over the whole vocabulary;
// otherwise the types not yet noted are offered, and a lone one is noted
// without a question. None left notes nothing.
func volosJournalNote(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	journal := ctx.Source()
	if !OnBattlefield(g, journal) {
		return nil
	}
	tc := ctx.Trigger()
	var subtypes []string
	all := false
	if spell, ok := g.LookupCardForEffect(tc.Event.CardID); ok {
		ch := spell.Effective()
		subtypes, all = ch.Subtypes, ch.AllCreatureTypes
	} else if tc.Object != nil {
		subtypes = tc.Object.Subtypes
	}
	if all {
		NoteACreatureTypeThenTail(g, ctx.Controller(), journal,
			"Volo's Journal — note one of its creature types that hasn't been noted", ctx.Tail(volosJournalNoted))
		return nil
	}
	noted := g.NotesForEffect(journal)
	var options []game.ChoiceOption
	for _, s := range subtypes {
		t, ok := game.CanonicalCreatureType(s)
		if !ok || game.NotedIn(noted, t) {
			continue
		}
		dup := false
		for _, o := range options {
			dup = dup || o.Label == t
		}
		if !dup {
			options = append(options, game.ChoiceOption{Label: t})
		}
	}
	switch len(options) {
	case 0:
		return nil
	case 1:
		g.AddNoteForEffect(journal, options[0].Label)
		return nil
	}
	return PickOption{
		Question: "Volo's Journal — note one of its creature types that hasn't been noted",
		Options:  options,
		ThenKey:  volosJournalPicked,
	}.Apply(ctx)
}

// volosJournalNoted and volosJournalPicked note the answer. On-disk keys:
// never renamed, never reused.
var volosJournalNoted = CreatureTypeChosen("creature-type/volos-journal-note", func(ctx *Context, t string, _ game.TailOutcome) error {
	ctx.Game.AddNoteForEffect(ctx.Source(), t)
	return nil
})

var volosJournalPicked = OptionPickThen("option-pick/volos-journal-note", func(ctx *Context, r game.OptionPicked) error {
	if r.Option == nil {
		return nil
	}
	ctx.Game.AddNoteForEffect(ctx.Source(), r.Option.Label)
	return nil
})

// printedVolosJournalToken is Volo's Journal: a legendary colorless Book
// artifact token with hexproof and the note trigger. Its slug is an
// on-disk identity: never renamed.
func printedVolosJournalToken() TokenTemplate {
	return TokenTemplate{
		Slug: "volos-journal",
		Card: game.Card{
			Name:     "Volo's Journal",
			TypeLine: "Token Legendary Artifact — Book",
			Keywords: []string{"hexproof"},
		},
		Triggered: []game.TriggeredAbility{volosJournalNoteTrigger()},
		Text:      "Hexproof\nWhenever you cast a creature spell, note one of its creature types that hasn't been noted for this artifact.",
	}
}

// VolosJournalToken is that Journal's card value for CreateToken.
func VolosJournalToken() game.Card { return TokenFromCatalog(printedVolosJournalToken) }
