package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// copy_effects.go — the shared shape behind every "you may have this
// creature enter as a copy of X" card (CR 706.2).
//
// The whole class is one replacement effect with a picker inside it:
// the permanent's own entry is replaced, the controller chooses what
// to copy as it enters, and the copiable values land before the ETB
// event — so the permanent never exists on the battlefield as its
// own printed self and every ETB trigger sees the copy. The engine
// half lives in server/internal/game/copy_choice.go; this file is
// the catalog-facing constructor, and each card file is the "except"
// clause plus a candidate filter.
//
// The four cards in the catalog exercise all three kinds of "except"
// clause between them, which is why they shipped together:
//
//	Clone                 no except clause at all — the baseline.
//	Phyrexian Metamorph   adds a card type ("it's an artifact in
//	                      addition to its other types").
//	Sakashima the Impostor overrides the name and adds a supertype.
//	Spark Double          REMOVES a supertype and changes how the
//	                      permanent enters (an extra counter).

// EntersAsCopyOf builds the replacement effect for "you may have
// this permanent enter as a copy of <candidates>, except <except>".
//
// `candidates` decides what may be copied: Clone takes any creature
// on the battlefield, Spark Double only a creature or planeswalker
// its controller controls. It is evaluated against the live board
// both when the prompt is built and when the answer arrives.
//
// `except` may be nil. When set it receives the copiable values on
// their way onto the permanent — already a private copy, so editing
// them can never touch the card being copied — plus the entry event,
// for a clause that changes how the permanent enters rather than
// what it copies.
func EntersAsCopyOf(
	name string,
	candidates func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID,
	except func(ev *game.ReplacementEvent, v *game.PrintedValues, g *game.Game, source *game.Card),
) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:         []game.EventKind{game.EventZoneMove},
		SelfReplacement: true,
		Label:           name,
		PromptQuestion:  name + " — enter as a copy of…?",
		AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID
		},
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			// The player who cast it chooses. ev.Actor is the
			// announce-time controller the resolution path stamps;
			// the card's own controller / owner is the fallback for
			// an entry driven by something else.
			if ev != nil && ev.Actor != uuid.Nil {
				return ev.Actor
			}
			if src == nil {
				return uuid.Nil
			}
			if src.Controller != uuid.Nil {
				return src.Controller
			}
			return src.Owner
		},
		CopySelector: &game.CopySelector{
			Candidates: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) []uuid.UUID {
				controller := uuid.Nil
				self := uuid.Nil
				if ev != nil {
					controller = ev.Actor
					self = ev.CardID
				}
				if controller == uuid.Nil && src != nil {
					controller = src.Controller
				}
				return candidates(g, controller, self)
			},
			Except: except,
		},
	}
}

// copyCandidates walks the battlefield and returns the permanents
// that pass `ok`, skipping the entering permanent itself — a Clone
// is not on the battlefield yet when the prompt is built, but an
// effect that re-copies an existing permanent would be, and copying
// yourself is never a legal choice.
func copyCandidates(
	g *game.Game,
	self uuid.UUID,
	ok func(c game.Card) bool,
) []uuid.UUID {
	if g == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == self || !ok(c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}
