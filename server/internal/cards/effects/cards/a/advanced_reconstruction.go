package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Advanced Reconstruction — Enchantment — Class {3}{R}:
//
//	"(Gain the next level as a sorcery to add its ability.)
//	 At the beginning of your first main phase, mill a card, then exile a
//	 card from your graveyard at random. You may play the exiled card this
//	 turn.
//	 {1}{R}: Level 2
//	 Whenever one or more cards leave your graveyard, this Class deals 2
//	 damage to each opponent.
//	 {1}{R}: Level 3
//	 Spells you cast from anywhere other than your hand cost {2} less to
//	 cast."
//
// Level 1 is always there (CR 716.2b). The mill rides a keyed tail so the
// random exile waits for the milled card to land (a milled commander's
// command-zone offer included); the card is drawn from the whole graveyard
// through the engine's random stream, as Tersa Lightshatter's is, and gets
// a play (not cast-only) permission for the turn.
//
// Level 2 is Tormod's "one or more cards leave your graveyard" batch
// trigger (a flashback cast from it counts), gated at Level(2), so the
// level-1 exile itself triggers it once the Class is level 2. Level 3 is
// Fortune Teller's Talent's gated cost modifier.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f00bc101-8082-4459-b987-f0c4b7a12344",
		Name:         "Advanced Reconstruction",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Advanced Reconstruction — mill a card, then exile a card from your graveyard at random; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 1, ThenTail: ctx.ItemTail(advancedReconstructionExileRandom)}.Apply(ctx)
				}),
			effects.AtLevel(2, effects.OncePerBatch(effects.OnAny([]game.EventKind{game.EventZoneMove, game.EventCast},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.CardLeftYourGraveyard(ev, source, g)
				}, "Advanced Reconstruction — 2 damage to each opponent (cards left your graveyard)",
				effects.Do(effects.DamageEachOpponent{N: 2})))),
		},
		Activated: []effects.ActivatedAbility{
			effects.LevelUp(2, effects.ManaCost("{1}{R}")),
			effects.LevelUp(3, effects.ManaCost("{1}{R}")),
		},
		CostModifiers: []game.CostModifier{
			effects.GatedCostModifier(effects.Level(3), effects.CostsLess(2,
				"Advanced Reconstruction — spells you cast from anywhere other than your hand cost {2} less",
				effects.YourSpell(), effects.CastFromOutsideYourHand())),
		},
	})
}

// advancedReconstructionExileRandom is the clause after the mill: a card
// at random from the whole graveyard, exiled with a play permission for
// the turn. An on-disk key.
var advancedReconstructionExileRandom = effects.CardTail("advanced-reconstruction/exile-random", func(ctx *effects.Context, _ game.TailOutcome) error {
	g, controller := ctx.Game, ctx.Controller()
	pick := g.ChooseAtRandomForEffect(effects.RandomDraw(ctx), effects.GraveyardIDs(g, controller, nil), 1)
	if len(pick) == 0 {
		return nil
	}
	return g.ExileCardWithPermissionForEffect(pick[0], game.CastPermission{
		Player: controller,
		Label:  "Advanced Reconstruction — play it this turn",
	})
})
