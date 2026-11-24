package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Eighth Doctor — Legendary Creature — Time Lord Doctor {4}{W}{U}, 4/4
// (Doctor Who, #5948 GAP2 b07):
//
//	"When The Eighth Doctor enters, mill three cards.
//	 Once during each of your turns, you may play a historic land or cast
//	 a historic permanent spell from your graveyard. If you do, it gains
//	 "If this permanent would leave the battlefield, exile it instead of
//	 putting it anywhere else.""
//
// The second ability is Serra Paragon's STANDING graveyard permission:
// a land (played, CR 305.1) or a permanent spell, each historic
// (legendary, artifact or Saga, CR 700.6), judged on the face being cast;
// ONE use each turn however it is spent (PerType's single
// PermissionTypePermanent, kept per Eighth Doctor object), during your
// turn and with the card's own timing, so a land needs a main phase, an
// empty stack and the turn's land drop (CR 305.2). The rider is the
// permission's GainsExileInsteadOfLeaving (the #6091 seam): the permanent
// the cast or land play produces has "If this permanent would leave the
// battlefield, exile it instead of putting it anywhere else" (CR 614.1a)
// from the moment it is on the battlefield for as long as it is that
// object (CR 400.7). A card whose own text opens the graveyard (unearth,
// flashback) is cast by that text, uses nothing here and gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0956bc16-ff04-4e96-8059-ef62afa2405f",
		Name:         "The Eighth Doctor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("The Eighth Doctor — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
		CastPermissions: []game.CastPermission{{
			Zone:                       game.ZoneGraveyard,
			Filter:                     game.PermissionFilter{PermanentOnly: true, HistoricOnly: true},
			PerType:                    []string{game.PermissionTypePermanent},
			Timing:                     game.TimingYourTurnOnly,
			GainsExileInsteadOfLeaving: true,
			Label:                      "Once during each of your turns, play a historic land or cast a historic permanent spell from your graveyard; it gains \"If this permanent would leave the battlefield, exile it instead\" (The Eighth Doctor)",
		}},
	})
}
