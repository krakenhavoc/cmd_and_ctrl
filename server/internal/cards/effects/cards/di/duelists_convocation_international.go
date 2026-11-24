package cardsdi

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Duelists' Convocation International — Enchantment {2}{U} (Mystery
// Booster 2):
//
//	"When Duelists' Convocation International enters the battlefield,
//	 choose a 10-digit number at random and write it down.
//	 Whenever you play a land or cast a spell, if it shares a mana value
//	 with one or more uncrossed digits in the chosen number, cross out one
//	 of those digits and draw a card. Then if all ten digits are crossed
//	 out, you win the game."
//
// The number is drawn from the game's keyed random stream
// (Game.ChooseNumberAtRandomForEffect: every 10-digit number equally
// likely, so the first digit is never 0) and written down as a note on
// this enchantment (Card.Notes, #6097), which every seat sees on the card:
// "Number 3141592653". Each digit crossed out is noted too ("Crossed out
// the 3 in place 1"), so the list is per object: a Convocation that leaves
// and returns writes a new number (CR 400.7).
//
// The second ability watches the controller's land plays and casts (The
// Endstone's two conditions). Its "if" is an intervening if (CR 603.4),
// checked as it triggers and again as it resolves: a land's mana value is
// read off the permanent (0 for a land with no mana cost, CR 202.3a), and
// a spell's on the stack with its X (CR 202.3e), or as it last existed
// there when it has left (CR 608.2h). A digit equal to that mana value is
// crossed out (which of two equal digits makes no difference; the
// leftmost is noted), the controller draws, and then the game is won if
// all ten are crossed out. A mana value of 10 or more matches no digit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "27a7380d-4c18-4754-b5fb-1302d939c9d9",
		Name:         "Duelists' Convocation International",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Duelists' Convocation International — choose a 10-digit number at random and write it down",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.OnBattlefield(g, ctx.Source()) {
						return nil
					}
					digits := g.ChooseNumberAtRandomForEffect(game.RandomDraw{Player: ctx.Controller(), Source: ctx.Source()}, 10)
					var b strings.Builder
					for _, d := range digits {
						b.WriteString(strconv.Itoa(d))
					}
					g.AddNoteForEffect(ctx.Source(), dciNumberNote+b.String())
					return nil
				}),
			effects.InterveningIf(effects.CondBoard{
				Name: "it shares a mana value with an uncrossed digit",
				Test: func(ctx *effects.Context) bool {
					_, ok := dciUncrossedPlace(ctx.Game, ctx.Source(), dciManaValueOf(ctx.Game, ctx.Trigger().Event))
					return ok
				},
			}, effects.OnAny([]game.EventKind{game.EventZoneMove, game.EventCast}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B31YouPlayedALandOrCastASpell(ev, source, g) {
					return false
				}
				_, ok := dciUncrossedPlace(g, source.InstanceID, dciManaValueOf(g, ev))
				return ok
			}, "Duelists' Convocation International — cross out a digit and draw a card; then if all ten are crossed out, you win the game",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					place, ok := dciUncrossedPlace(g, ctx.Source(), dciManaValueOf(g, ctx.Trigger().Event))
					if !ok {
						return nil
					}
					digits, _ := dciState(g, ctx.Source())
					g.AddNoteForEffect(ctx.Source(), dciCrossedNote(digits[place], place))
					if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					if _, crossed := dciState(g, ctx.Source()); len(crossed) == 10 {
						return effects.WinTheGame{}.Apply(ctx)
					}
					return nil
				})),
		},
	})
}

// The notes this card writes. On-disk words: never reworded.
const (
	dciNumberNote  = "Number "
	dciCrossedHead = "Crossed out the "
)

// dciCrossedNote is the note for crossing out `digit` at 0-based `place`.
func dciCrossedNote(digit, place int) string {
	return dciCrossedHead + strconv.Itoa(digit) + " in place " + strconv.Itoa(place+1)
}

// dciState reads the number written down for the Convocation `id` and the
// 0-based places crossed out; nil digits when none was written.
func dciState(g *game.Game, id uuid.UUID) (digits []int, crossed map[int]bool) {
	crossed = map[int]bool{}
	for _, n := range g.NotesForEffect(id) {
		switch {
		case strings.HasPrefix(n, dciNumberNote):
			digits = nil
			for _, r := range strings.TrimPrefix(n, dciNumberNote) {
				if r >= '0' && r <= '9' {
					digits = append(digits, int(r-'0'))
				}
			}
		case strings.HasPrefix(n, dciCrossedHead):
			i := strings.LastIndex(n, " ")
			if p, err := strconv.Atoi(n[i+1:]); err == nil && p >= 1 {
				crossed[p-1] = true
			}
		}
	}
	return digits, crossed
}

// dciUncrossedPlace is the leftmost uncrossed place holding the digit
// `mv`, and whether there is one.
func dciUncrossedPlace(g *game.Game, id uuid.UUID, mv int) (int, bool) {
	if mv < 0 || mv > 9 {
		return 0, false
	}
	digits, crossed := dciState(g, id)
	for i, d := range digits {
		if d == mv && !crossed[i] {
			return i, true
		}
	}
	return 0, false
}

// dciManaValueOf is the mana value of what the event played or cast, or -1
// when it can't be read.
func dciManaValueOf(g *game.Game, ev game.Event) int {
	switch ev.Kind {
	case game.EventCast:
		if mv, ok := g.LastKnownSpellManaValueForEffect(ev.CardID); ok {
			return mv
		}
	case game.EventZoneMove:
		if mv, ok := g.PermanentManaValueForEffect(ev.CardID); ok {
			return mv
		}
		if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
			return info.ManaValue
		}
	}
	return -1
}
