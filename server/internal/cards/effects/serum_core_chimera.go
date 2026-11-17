package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Serum-Core Chimera — Creature — Phyrexian Chimera {2}{U}{R}, 2/4 (Phyrexia: All Will Be One):
//
//	"Flying
//	 Whenever you cast a noncreature spell, put an oil counter on this creature.
//	 Remove three oil counters from this creature: Draw a card. Then you may discard a
//	 nonland card. When you discard a card this way, this creature deals 3 damage to
//	 target creature or planeswalker. Activate only as a sorcery."
//
// The discard is optional and only a nonland card is a legal pick (a set rule over the
// chosen card), so a hand of lands, or an empty one, discards nothing. The reflexive
// trigger (CR 603.12) goes on the stack once a card has really been discarded and picks
// its target then; the damage is dealt by the Chimera as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ed54a59b-b24c-4681-b862-fa634fe759fd",
		Name:            "Serum-Core Chimera",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{s82OneB09OilOnNoncreatureCast("Serum-Core Chimera")},
		Activated: []ActivatedAbility{{
			Label:        "Remove three oil counters from this creature: Draw a card. Then you may discard a nonland card. When you discard a card this way, this creature deals 3 damage to target creature or planeswalker. Activate only as a sorcery.",
			Purpose:      game.Purpose{Draws: 1, DamageToCreature: 3},
			Cost:         RemoveCountersFromThis(s82OneB09Oil, 3),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
					Ref: serumCoreChimeraDiscard, Player: item.Controller, Source: item.SourceCardID, N: 1,
				})
			},
		}},
	})
}

var (
	// "Then you may discard a nonland card": up to one, and the set rule refuses a land.
	serumCoreChimeraDiscard = game.RegisterDrawThen("serum-core-chimera/may-discard", func(g *game.Game, d game.DrawThen) error {
		return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
			Player:   d.Player,
			Source:   d.Source,
			N:        1,
			UpTo:     true,
			Question: "Serum-Core Chimera — you may discard a nonland card",
		}, serumCoreChimeraDiscarded, nil)
	})

	serumCoreChimeraDiscarded = game.RegisterDiscardThenRule("choose-cards/serum-core-chimera-discarded",
		func(picked []game.Card) bool { return len(picked) == 1 && !picked[0].IsLand() },
		func(g *game.Game, r game.DiscardsDone) error {
			if len(r.Discarded.Cards()) == 0 {
				return nil
			}
			parent := &game.StackItem{Controller: r.Player, SourceCardID: r.Source, Label: "Serum-Core Chimera"}
			t := WhenYouDo("Serum-Core Chimera — 3 damage to target creature or planeswalker", serumCoreChimeraDamageBody)
			t.Params = game.EffectParams{Amount: 3}
			return t.Apply(NewContext(g, parent))
		})

	serumCoreChimeraDamageBody = game.ReflexiveBody("serum-core-chimera/damage", breechesBlastEffect,
		constTargets(func() *game.TargetSpec {
			return TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))
		}))
)
