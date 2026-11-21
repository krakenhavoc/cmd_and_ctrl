package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Magnanimous Magistrate — Creature — Human Advisor {5}{W}, 3/4 (Jumpstart 2022):
//
//	"This creature enters with five reprieve counters on it.
//	 Whenever another nontoken creature you control dies, if its mana value was
//	 1 or greater, you may remove that many reprieve counters from this
//	 creature. If you do, return that card to the battlefield under its
//	 owner's control."
//
// The counters are a self-replacement on entry (Archfiend of the Dross). The
// dies trigger is Liesa's condition narrowed to a mana value of 1 or more,
// read off the card (a dead permanent keeps its printed cost), and names the
// dead card as it goes on the stack (EventCardTriggerItem). On resolution the
// "may" is asked only when the Magistrate is still the same object on the
// battlefield with enough counters; a yes removes them (a cost-like
// instruction, CR 608.2c) and returns the card only if it is still the same
// card in a graveyard (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecd3ddd7-2fa8-466b-9319-ae51c744a05f",
		Name:         "Magnanimous Magistrate",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters("reprieve", 5, "Magnanimous Magistrate enters with five reprieve counters on it"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.B19AnotherNontokenCreatureYouControlDied(ev, source, g) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.ManaValue() >= 1
			},
			Key:   "Magnanimous Magistrate — remove reprieve counters to return that card",
			Build: effects.EventCardTriggerItem("Magnanimous Magistrate — remove reprieve counters to return that card"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				id, ok := effects.DeadCardInGraveyard(g, item)
				if !ok || effects.SourceIsNewObject(g, item) || !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				dead, _ := g.LookupCardForEffect(id)
				mv := dead.ManaValue()
				if effects.CounterCountOn(g, item.SourceCardID, "reprieve") < mv {
					return nil
				}
				return effects.MayChoice{
					Question: "Magnanimous Magistrate — remove reprieve counters to return that card?",
					Then:     magistrateAnswer,
					Carry:    []uuid.UUID{id},
					Objects:  []game.ObjectRef{item.Params.Object},
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// magistrateAnswer is the "yes": the counters come off first, then the card
// returns if it is still the object that died and is still in a graveyard.
// An on-disk key: never renamed, never reused.
var magistrateAnswer = effects.MayYes("confirm/magnanimous-magistrate-reprieve", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	g := ctx.Game
	if len(r.Carry) == 0 || len(r.Objects) == 0 {
		return nil
	}
	self := ctx.Source()
	if !effects.OnBattlefield(g, self) {
		return nil
	}
	dead, ok := g.LookupCardForEffect(r.Carry[0])
	if !ok || dead.ObjectEpoch != r.Objects[0].Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(dead.InstanceID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	mv := dead.ManaValue()
	if mv < 1 || effects.CounterCountOn(g, self, "reprieve") < mv {
		return nil
	}
	if err := (effects.AddCounter{Target: self, Kind: "reprieve", N: -mv}).Apply(ctx); err != nil {
		return err
	}
	return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(dead.InstanceID, dead.Owner, false, game.KeyedTail{})
})
