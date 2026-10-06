package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Necrobloom — Legendary Creature — Plant {1}{W}{B}{G}, 2/7
// (EDHREC rank 3703):
//
//	"Landfall — Whenever a land you control enters, create a 0/1
//	 green Plant creature token. If you control seven or more lands
//	 with different names, create a 2/2 black Zombie creature token
//	 instead.
//	 Land cards in your graveyard have dredge 2. (You may return a
//	 land card from your graveyard to your hand and mill two cards
//	 instead of drawing a card.)"
//
// The lands commander. Landfall is Maja's condition
// (b33LandYouControlEntered — a land entered under the controller's
// control, played or fetched or tokened), and the body reads the
// land count at resolution, so the land that entered is among the
// seven: a Plant, or a Zombie once seven differently-named lands are
// in play (Field of the Dead's count, b04LandNamesControlled — two
// copies of the same land count once, as printed).
//
// Dredge 2 for lands (#2127, dredge.go): a "may" replacement on the
// Necrobloom itself. Say yes and the draw is replaced; then pick WHICH
// land card in your graveyard to return, mill two cards and return it.
// Offered only with a land card in the graveyard and two cards in the
// library (CR 702.52b), once per draw, alongside any dredge card of
// your own in the graveyard.
//
// The land is picked after the "yes", which gives the same options as
// a dredge per land card: any one land, or none.
func init() {
	Register(Spec{
		OracleID:     "b981af39-4ee6-4fbc-9a89-618dcad9dfbf",
		Name:         "The Necrobloom",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{necrobloomDredge()},
		Triggered: []game.TriggeredAbility{
			Landfall("The Necrobloom — create a 0/1 green Plant, or a 2/2 black Zombie with seven differently-named lands", b35NecrobloomLandfall),
		},
	})
}

// necrobloomDredge is "land cards in your graveyard have dredge 2".
func necrobloomDredge() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches:  []game.EventKind{game.EventDrawCard},
		Optional: true,
		// The bot prices a dredge that names no card as returning a
		// LAND (aiseat/heuristic/dredge.go, #2390), which is this
		// grant's printed class.
		Dredge: 2,
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventDraw || ev.DrawCount <= 0 || src == nil || ev.DrawPlayer != src.Controller {
				return false
			}
			p := g.PlayerByIDForEffect(ev.DrawPlayer)
			return p != nil && p.Library != nil && p.Library.Size() >= 2 &&
				len(gyCardIDs(g, ev.DrawPlayer, isLandCard)) > 0
		},
		DrawInstead: game.RegisterDrawInstead("necrobloom-dredge-2", func(g *game.Game, drawer, source uuid.UUID, done func(*game.Game) error) error {
			lands := gyCardIDs(g, drawer, isLandCard)
			if len(lands) == 0 {
				return done(g)
			}
			g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  drawer,
				Source:   source,
				Question: "The Necrobloom — dredge 2: return which land card from your graveyard to your hand?",
				Cards:    lands,
				Min:      1,
				Max:      1,
				Zone:     game.ZoneGraveyard,
				Then: func(g *game.Game, picked []uuid.UUID) error {
					if len(picked) == 0 {
						return done(g)
					}
					return dredgeBody(g, drawer, picked[0], 2, done)
				},
			})
			return nil
		}),
		Controller: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) uuid.UUID {
			return ev.DrawPlayer
		},
		PromptQuestion: "The Necrobloom — dredge 2: mill two cards and return a land card from your graveyard to your hand instead of drawing?",
		Label:          "The Necrobloom: land cards in your graveyard have dredge 2",
	}
}
