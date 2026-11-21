package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dredge the Mire — Sorcery {3}{B} (Starter Commander Decks):
//
//	"Each opponent chooses a creature card in their graveyard. Put those
//	 cards onto the battlefield under your control."
//
// Each opponent is asked, over their own graveyard, as one run of keyed
// choose_cards prompts (Kishla Trawlers' shape), so nobody sees another's
// pick first. An opponent with no creature card in their graveyard, or one
// who has left the game, is not asked. When the last has answered, the
// chosen cards are put onto the battlefield under the caster's control (a
// card a replacement keeps off the battlefield stays where it is).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10e286b0-7245-4c30-a7a9-6d4a9c49b0a7",
		Name:         "Dredge the Mire",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var legs []game.ChooseCardsPrompt
			var asked []uuid.UUID
			for _, opp := range ctx.Opponents() {
				p := ctx.Game.PlayerByIDForEffect(opp)
				if p == nil || p.Eliminated || p.Graveyard == nil {
					continue
				}
				var cards []uuid.UUID
				for _, c := range p.Graveyard.Cards {
					if c.HasCardType("creature") {
						cards = append(cards, c.InstanceID)
					}
				}
				if len(cards) == 0 {
					continue
				}
				asked = append(asked, opp)
				legs = append(legs, game.ChooseCardsPrompt{
					Chooser:    opp,
					FromPlayer: opp,
					Source:     ctx.Source(),
					Question:   "Dredge the Mire — choose a creature card in your graveyard",
					Cards:      cards,
					Min:        1,
					Max:        1,
					Zone:       game.ZoneGraveyard,
				})
			}
			if len(legs) == 0 {
				return nil
			}
			_, err := ctx.Game.ChooseCardsRunThenKeyedForEffect(legs, dredgeTheMireChosen, ctx.Source(), ctx.RunData(asked))
			return err
		},
	})
}

// dredgeTheMireChosen puts every opponent's chosen card onto the battlefield
// under the caster's control. An on-disk key.
var dredgeTheMireChosen = effects.RunThen("prompt-run/dredge-the-mire-chosen", func(ctx *effects.Context, r game.RunDone) error {
	for _, opp := range r.Carry {
		for _, id := range r.Picks().By(opp) {
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				continue
			}
			if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	return nil
})
