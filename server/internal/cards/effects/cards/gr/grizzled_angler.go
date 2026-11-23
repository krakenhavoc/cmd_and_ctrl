package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grizzled Angler // Grisly Anglerfish — {2}{U} Creature — Human 2/3 //
// Creature — Eldrazi Fish 4/5 (Innistrad Remastered, S82 GAP, #5948):
//
//	Front: "{T}: Mill two cards. Then if there is a colorless creature card
//	        in your graveyard, transform this creature."
//	Back:  "{6}: Creatures your opponents control attack this turn if able."
//
// The mill is a single two-card mill; the check is made afterwards against
// the whole graveyard, not only the cards just milled (a colorless creature
// card already there counts). The back face's ability is Bident of Thassa's
// attack requirement over creatures your opponents control, a live scope
// that also reaches creatures they cast later this turn (CR 508.1d).
//
// No simplification.
func init() {
	const oracle = "d37effd3-d194-4609-9994-6d70038a68f9"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Grizzled Angler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Mill two cards. Then if there is a colorless creature card in your graveyard, transform this creature.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.MillToZone{N: 2, ThenTail: ctx.ItemTail(grizzledAnglerMilled)}.Apply(ctx)
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Grisly Anglerfish",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}: Creatures your opponents control attack this turn if able.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{6}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.OpponentsCreaturesAttackIfAble{
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Grisly Anglerfish — creatures your opponents control attack this turn if able",
				}.Apply(ctx)
			},
		}},
	})
}

// grizzledAnglerMilled is "then if there is a colorless creature card in
// your graveyard, transform this creature". An on-disk key: never renamed,
// never reused.
var grizzledAnglerMilled = effects.CardTail("grizzled-angler/after-mill", func(ctx *effects.Context, _ game.TailOutcome) error {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() && c.IsColorless() {
			return effects.TransformThis{}.Apply(ctx)
		}
	}
	return nil
})
