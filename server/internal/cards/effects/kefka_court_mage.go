package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kefkaCourtMageOracle is shared by both faces: the front registers under
// it, the back under kefkaCourtMageOracle + "#1".
const kefkaCourtMageOracle = "6d3ca5cb-781d-4608-b7f1-38c020536376"

// Kefka, Court Mage // Kefka, Ruler of Ruin (Final Fantasy, #4462) — a
// transforming legendary creature.
//
// Front face, Legendary Creature — Human Wizard {2}{U}{B}{R}, 4/5:
//
//	"Whenever Kefka enters or attacks, each player discards a card. Then
//	 you draw a card for each card type among cards discarded this way.
//	 {8}: Each opponent sacrifices a permanent of their choice. Transform
//	 Kefka. Activate only as a sorcery."
//
// Back face, Legendary Creature — Avatar Wizard, 5/7:
//
//	"Flying
//	 Whenever an opponent loses life during your turn, you draw that many
//	 cards."
//
// The discard is Syphon Mind's fan-out (EachPlayerDiscardsThenKeyedForEffect):
// every player is asked in turn order and a player with no hand is not, and
// the draw is the run's continuation, told every card that really left a
// hand. The card types are counted among those cards as they sit in the
// graveyard now (the rulings: Legendary and Saga are not card types), each
// type once however many cards had it. The {8} is a sorcery-speed edict on
// the opponents (Smallpox's EachPlayerSacrifices) whose continuation turns
// Kefka over once every opponent's permanent has left, as printed; if Kefka
// has changed faces or left since the ability went on the stack, the
// instruction is ignored (ADR 0152 §3). The back face is Exquisite Blood's
// life-loss read (a drain and damage both lose life) on the controller's
// own turn, drawing as many cards as the life lost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     kefkaCourtMageOracle,
		Name:         "Kefka, Court Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Kefka, Court Mage — each player discards a card, then you draw a card for each card type among cards discarded this way",
				func(g *game.Game, item *game.StackItem) error {
					return g.EachPlayerDiscardsThenKeyedForEffect(uuid.Nil,
						game.DiscardPrompt{Source: item.SourceCardID, N: 1, Question: "Kefka, Court Mage — discard a card"},
						kefkaDraw, []uuid.UUID{item.Controller})
				}),
		},
		Activated: []ActivatedAbility{{
			Label:        "{8}: Each opponent sacrifices a permanent of their choice. Transform Kefka. Activate only as a sorcery.",
			Cost:         ManaCost("{8}"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return EachPlayerSacrifices{ExceptController: true, Label: "a permanent", Keyed: kefkaTransform}.Apply(NewContext(g, item))
			},
		}},
	})
	Register(Spec{
		OracleID:        kefkaCourtMageOracle + "#1",
		Name:            "Kefka, Ruler of Ruin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !isActivePlayer(g, source.Controller) {
					return false
				}
				_, ok := b04OpponentLostLife(ev, source.Controller, g)
				return ok
			},
			Key: "Kefka, Ruler of Ruin — you draw that many cards",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				amount, _ := b04OpponentLostLife(ev, source.Controller, g)
				item := game.NewTriggeredItem(source, "Kefka, Ruler of Ruin — you draw that many cards")
				item.Params.Amount = amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}

// kefkaDraw is "you draw a card for each card type among cards discarded
// this way", run once after the last discard has landed. The controller is
// the carry. An on-disk key.
var kefkaDraw = game.RegisterPromptRunThen("prompt-run/kefka-court-mage-draw", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) == 0 {
		return nil
	}
	types := map[string]bool{}
	for _, id := range r.Discards().Cards() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		_, ts, _ := game.ParseTypeLine(c.TypeLine)
		for _, t := range ts {
			types[t] = true
		}
	}
	if len(types) == 0 {
		return nil
	}
	return g.DrawNForEffect(r.Carry[0], len(types))
})

// kefkaTransform is "Transform Kefka", run once every opponent's
// sacrifice has finished. An on-disk key.
var kefkaTransform = RegisterEdictThen("prompt-run/kefka-court-mage-transform", func(ctx *Context, _ game.PromptedSacrifices) error {
	return TransformThis{}.Apply(ctx)
})
