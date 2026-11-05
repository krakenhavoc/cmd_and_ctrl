package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kai Winn, Faithless Leader — Legendary Creature — Bajoran Cleric
// {1}{B}, 3/1 (Star Trek, #4973 batch 12):
//
//	"As Kai Winn enters, choose a nonland card name.
//	 Whenever an opponent casts a spell with the chosen name or
//	 activates an ability of a source with the chosen name, that player
//	 loses 3 life."
//
// The name is Phyrexian Revoker's as-enters prompt (CR 614.12), stored
// on this permanent (ChosenName), so two Kai Winns name two cards and a
// copy names its own (CR 707.2). Until it is answered the name is empty
// and nothing triggers.
//
// ONE printed trigger, so one registered trigger that watches both the
// cast and the two activation kinds: "an ability" includes a mana
// ability (CR 605), which uses no stack but is still activated.
//
// THE SPELL is matched by the name it has on the stack
// (game.SpellOrLandHasName): a split card has only the half cast, a
// modal double-faced card only the face cast.
//
// THE SOURCE is matched by the name it has where the ability is
// activated from. On the battlefield that is its effective name, after
// copy effects (game.PermanentHasName). Off it (a cycling card in hand,
// an unearth card in a graveyard) it is the card's own name. A source
// that left the battlefield to pay for the ability (a sacrifice cost) is
// read from its last-known information as well (CR 608.2h), so a
// sacrificed copy of the named card still counts, and a token that has
// ceased to exist is found there too.
//
// "That player" is the caster or activator, read off the triggering
// event at resolution. Life loss, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16a5ed42-ee86-4dcb-b8ea-239ac2b13d67",
		Name:         "Kai Winn, Faithless Leader",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCardNameAsEnters("Kai Winn — choose a nonland card name"),
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventCast, game.EventActivateAbility, game.EventManaAbilityActivated},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
						return false
					}
					named := strings.TrimSpace(source.ChosenName)
					if named == "" {
						return false
					}
					if ev.Kind == game.EventCast {
						spell, ok := g.LookupCardForEffect(ev.CardID)
						return ok && game.SpellOrLandHasName(spell, named)
					}
					return kaiWinnSourceNamed(g, ev, named)
				},
				"Kai Winn, Faithless Leader — that player loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, triggeringActor(item), -3)
				}),
		},
	})
}

// kaiWinnSourceNamed reports whether the object whose ability ev
// activated has the name `named`, wherever it was activated from.
func kaiWinnSourceNamed(g *game.Game, ev game.Event, named string) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	if ok && onBattlefield(g, ev.CardID) {
		return game.PermanentHasName(c, named)
	}
	if ok && game.SpellOrLandHasName(c, named) {
		return true
	}
	if lki, left := g.LastKnownPermanentForEffect(ev.CardID); left {
		return strings.EqualFold(strings.TrimSpace(lki.Characteristic.Name), named)
	}
	return false
}
