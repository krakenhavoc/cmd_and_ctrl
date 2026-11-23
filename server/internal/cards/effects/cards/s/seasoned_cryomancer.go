package cardss

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "c3936db3-d3b6-4a92-9773-6f3c52dc419d",
		Name:         "Seasoned Cryomancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Seasoned Cryomancer — draw two cards, then discard two cards", seasonedCryomancerETB),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}{U}, Exile this card from your graveyard: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{U}{U}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.Do(effects.DrawCards{N: 2}),
		}},
	})
}

var (
	// The reflexive trigger (Hapatra's frost effect, shared from fra-creature-c): tap up to N target creatures and stun each,
	// N the number of nonland cards discarded, carried in Params.Amount.
	seasonedCryomancerStunBody = game.ReflexiveBody("seasoned-cryomancer/tap-and-stun", effects.SimpleBody(effects.RfCreatureCHapatraFrostEffect),
		func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
			return effects.TargetCreature("up to "+strconv.Itoa(p.Amount)+" target creatures").WithCount(0, p.Amount)
		})

	// The loot's discard half, run once both draws have landed.
	seasonedCryomancerDiscard = game.RegisterDrawThen("seasoned-cryomancer/discard-two", func(g *game.Game, d game.DrawThen) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        d.N,
			Question: "Seasoned Cryomancer — discard two cards",
		}, seasonedCryomancerDiscarded, nil)
	})

	// The loot's discard answered, keyed (#4253): the reflexive trigger for
	// the nonland cards that really left the hand.
	seasonedCryomancerDiscarded = game.RegisterDiscardThen("choose-cards/seasoned-cryomancer-discarded", func(g *game.Game, r game.DiscardsDone) error {
		nonland := 0
		for _, id := range r.Discarded.By(r.Player) {
			if c, ok := g.LookupCardForEffect(id); ok && !c.IsLand() {
				nonland++
			}
		}
		if nonland == 0 {
			return nil
		}
		parent := &game.StackItem{Controller: r.Player, SourceCardID: r.Source, Label: "Seasoned Cryomancer"}
		t := effects.WhenYouDo("Seasoned Cryomancer — tap up to "+strconv.Itoa(nonland)+" target creatures and put a stun counter on each", seasonedCryomancerStunBody)
		t.Params = game.EffectParams{Amount: nonland}
		return t.Apply(effects.NewContext(g, parent))
	})
)

func seasonedCryomancerETB(g *game.Game, item *game.StackItem) error {
	return g.DrawNThenForEffect(item.Controller, 2, game.DrawThen{
		Ref: seasonedCryomancerDiscard, Player: item.Controller, Source: item.SourceCardID, N: 2,
	})
}
