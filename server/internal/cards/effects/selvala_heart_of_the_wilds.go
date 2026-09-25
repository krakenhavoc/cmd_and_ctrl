package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Selvala, Heart of the Wilds — Legendary Creature — Elf Scout
// {1}{G}{G}, 2/3:
//
//	"Whenever another creature enters, its controller may draw a card
//	 if its power is greater than each other creature's power.
//	 {G}, {T}: Add X mana in any combination of colors, where X is
//	 the greatest power among creatures you control."
//
// The green ramp legend that turns one fatty into a whole turn. Two
// clauses, and each has a trap in it.
//
// The TRIGGER is political: it watches EVERY creature that enters,
// not only the ones its controller controls, and the "may" and the
// draw both belong to the ENTERING creature's controller. That is
// why it is not WheneverAnotherCreatureEntersUnderYourControl and why
// the optional prompt carries a Chooser override, exactly as Edric,
// Spymaster of Trest does — the trigger is still Selvala's ability
// and sits on the stack under Selvala's controller, but the question
// and the card go to the other seat.
//
// "Greater than EACH OTHER creature's power" is strictly greater than
// every other creature on the battlefield, any controller, ties
// losing. It is checked when the trigger would fire and again as it
// resolves (CR 603.4), so a bigger creature that arrives in response
// takes the draw away.
//
// Declared weaker than printed in one edge: the printed ruling uses
// last-known information for a creature that has left the
// battlefield before the ability resolves, and this does not — a
// creature killed in response takes its own draw with it. The
// re-check reads the live battlefield, which is the same read the
// prompt was offered on.
//
// The MANA ABILITY is "in any combination of colors", which is X
// independent any-colour picks (AnyCombinationOfColors) and NOT
// "X mana of any one color" (OneColorOfAmount, Nykthos). With a 5/5
// out this is five separate picks and can pay {W}{U}{B}{R}{G};
// written the other way it would be locked to one colour, which is a
// different and much worse card. X is the greatest power among
// creatures the ACTIVATOR controls — Selvala herself included, since
// the text says "creatures you control" and not "other" — and a
// negative or zero X produces the empty string, which adds nothing
// while the ability still costs {G} and taps her.
//
// Summoning sickness applies: the cost has {T} on a creature
// (CR 302.6), and the engine enforces it inside ActivateManaAbility.
func init() {
	Register(Spec{
		OracleID:     "1d725121-e50c-42f0-9128-56802f07c89e",
		Name:         "Selvala, Heart of the Wilds",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := selvalaBiggestCreatureEntered(ev, source, g)
				return ok
			},
			Key: selvalaDrawLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				entered, ok := selvalaBiggestCreatureEntered(ev, source, g)
				if !ok {
					return nil
				}
				item := game.NewTriggeredItem(source, selvalaDrawLabel, nil)
				item.Params.Player = entered.Controller
				item.Params.Object.ID = entered.InstanceID
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				c, ok := g.LookupCardForEffect(item.Params.Object.ID)
				if !ok || !selvalaHasTheGreatestPower(g, c) {
					return nil
				}
				return DrawCards{Player: item.Params.Player, N: 1}.Apply(NewContext(g, item))
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Selvala, Heart of the Wilds — draw a card?",
				Chooser: func(ev game.Event, source *game.Card, g *game.Game) uuid.UUID {
					entered, ok := selvalaBiggestCreatureEntered(ev, source, g)
					if !ok {
						// Nil falls back to the source's controller,
						// and Build declines the trigger anyway.
						return uuid.Nil
					}
					return entered.Controller
				},
			},
		}},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true, Mana: "{G}"},
			ProducedFunc: ProducedAnyCombinationOfColors(selvalaGreatestPower),
			Label:        "{G}, {T}: Add X mana in any combination of colors, where X is the greatest power among creatures you control",
		}},
	})
}

// selvalaDrawLabel is the stack label, shared by the Build and the
// resolution so the "once per" reads cannot disagree.
const selvalaDrawLabel = "Selvala, Heart of the Wilds — its controller may draw a card"

// selvalaBiggestCreatureEntered answers "another creature entered and
// its power is greater than each other creature's power", returning
// the entering creature so the caller can read its controller.
//
// Caller holds g.mu.
func selvalaBiggestCreatureEntered(ev game.Event, source *game.Card, g *game.Game) (game.Card, bool) {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return game.Card{}, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !c.IsCreature() {
		return game.Card{}, false
	}
	return c, selvalaHasTheGreatestPower(g, c)
}

// selvalaHasTheGreatestPower is the "greater than each other
// creature's power" comparison: strictly greater than every OTHER
// creature on the battlefield, whoever controls it. A tie is not
// greater, so two 5/5s entering back to back draw nothing.
//
// Layered power (CurrentPower), so an anthem on somebody else's board
// counts.
func selvalaHasTheGreatestPower(g *game.Game, c game.Card) bool {
	if !c.IsCreature() {
		return false
	}
	power := c.CurrentPower()
	for _, other := range g.BattlefieldCardsForEffect() {
		if other.InstanceID == c.InstanceID || !other.IsCreature() {
			continue
		}
		if other.CurrentPower() >= power {
			return false
		}
	}
	return true
}

// selvalaGreatestPower is the mana ability's X, in the shape
// ProducedAnyCombinationOfColors wants.
func selvalaGreatestPower(g *game.Game, controller, _ uuid.UUID) int {
	return b42GreatestPowerControlledBy(g, controller)
}
