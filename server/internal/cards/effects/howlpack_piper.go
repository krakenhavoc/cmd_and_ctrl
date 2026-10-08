package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Howlpack Piper // Wildsong Howler — {3}{G} Creature — Human Werewolf
// 2/2 // Creature — Werewolf 4/4 (#2586, ADR 0132):
//
//	Front: "This spell can't be countered.
//	        {1}{G}, {T}: You may put a creature card from your hand onto
//	        the battlefield. If it's a Wolf or Werewolf, untap this
//	        creature. Activate only as a sorcery.
//	        Daybound"
//	Back:  "Whenever this creature enters or transforms into Wildsong
//	        Howler, look at the top six cards of your library. You may
//	        reveal a creature card from among them and put it into your
//	        hand. Put the rest on the bottom of your library in a random
//	        order.
//	        Nightbound"
//
// The front's put is Quicksilver Amulet's: a put, not a cast, so nothing
// counters it and no cast trigger fires. "If it's a Wolf or Werewolf"
// reads the permanent that arrived, after the layers, and the untap
// needs the Piper still on the battlefield. The back face's trigger is
// the enter-or-turn-over shape Brutal Cathar uses (EventTransform, Amount
// 1 for the back); a card cast at night enters as Wildsong Howler and
// triggers on entering.
//
// No simplification.
func init() {
	const oracle = "cede233b-4e27-4738-8099-c8e46862ba96"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Howlpack Piper",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"daybound"},
		Activated: []ActivatedAbility{{
			Label:        "{1}{G}, {T}: You may put a creature card from your hand onto the battlefield. If it's a Wolf or Werewolf, untap this creature. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}{G}"), TapCost()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match:    Creature(),
					Optional: true,
					Label:    "Howlpack Piper — you may put a creature card from your hand onto the battlefield",
					Then:     howlpackPiperUntap,
				}.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Wildsong Howler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventTransform},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				if ev.CardID != source.InstanceID {
					return false
				}
				return ev.Kind == game.EventETB || (ev.Kind == game.EventTransform && ev.Amount == 1)
			},
			Key: "Wildsong Howler — look at the top six cards of your library",
			Effect: LookAtTopThenMayTakeToHand(6, Creature(), 1,
				"Wildsong Howler — you may reveal a creature card from among them and put it into your hand"),
		}},
	})
}

// howlpackPiperUntap is "If it's a Wolf or Werewolf, untap this
// creature": the permanent that arrived is read through the layers, and
// the Piper must still be in play to untap.
func howlpackPiperUntap(g *game.Game, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	c, ok := g.LookupCardForEffect(res.Entered)
	if !ok || !(c.HasSubtype("Wolf") || c.HasSubtype("Werewolf")) {
		return nil
	}
	return UntapTarget{Target: res.Source}.Apply(NewContext(g, &game.StackItem{SourceCardID: res.Source, Controller: res.Player}))
}
