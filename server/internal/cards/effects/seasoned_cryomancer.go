package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seasoned Cryomancer — Creature — Human Wizard {1}{U}{U}, 2/2:
//
//	"When this creature enters, draw two cards, then discard two cards.
//	 When you discard one or more nonland cards this way, tap up to that
//	 many target creatures and put a stun counter on each of them.
//	 {3}{U}{U}, Exile this card from your graveyard: Draw two cards."
//
// The loot is the draw-then-discard machinery, so the discard opens only
// once both draws have landed and the drawn cards are legal discards.
// The second sentence is a reflexive trigger (CR 603.12) created when
// the discard has actually happened and at least one discarded card was
// nonland; it counts them, and picks its "up to that many" targets as it
// goes on the stack. The graveyard ability is instant speed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3936db3-d3b6-4a92-9773-6f3c52dc419d",
		Name:         "Seasoned Cryomancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Seasoned Cryomancer — draw two cards, then discard two cards", seasonedCryomancerETB),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{U}{U}, Exile this card from your graveyard: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}{U}{U}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}

var (
	// The reflexive trigger (Hapatra's frost effect, shared from fra-creature-c): tap up to N target creatures and stun each,
	// N the number of nonland cards discarded, carried in Params.Amount.
	seasonedCryomancerStunBody = game.ReflexiveBody("seasoned-cryomancer/tap-and-stun", simpleBody(rfCreatureCHapatraFrostEffect),
		func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
			return TargetCreature("up to "+strconv.Itoa(p.Amount)+" target creatures").WithCount(0, p.Amount)
		})

	// The loot's discard half, run once both draws have landed.
	seasonedCryomancerDiscard = game.RegisterDrawThen("seasoned-cryomancer/discard-two", func(g *game.Game, d game.DrawThen) error {
		parent := &game.StackItem{Controller: d.Player, SourceCardID: d.Source, Label: "Seasoned Cryomancer"}
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        d.N,
			Question: "Seasoned Cryomancer — discard two cards",
			Then: func(g *game.Game, _ uuid.UUID, discarded []uuid.UUID) error {
				nonland := 0
				for _, id := range discarded {
					if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
						nonland++
					}
				}
				if nonland == 0 {
					return nil
				}
				t := WhenYouDo("Seasoned Cryomancer — tap up to "+strconv.Itoa(nonland)+" target creatures and put a stun counter on each", seasonedCryomancerStunBody)
				t.Params = game.EffectParams{Amount: nonland}
				return t.Apply(NewContext(g, parent))
			},
		})
		return nil
	})
)

func seasonedCryomancerETB(g *game.Game, item *game.StackItem) error {
	return g.DrawNThenForEffect(item.Controller, 2, game.DrawThen{
		Ref: seasonedCryomancerDiscard, Player: item.Controller, Source: item.SourceCardID, N: 2,
	})
}
