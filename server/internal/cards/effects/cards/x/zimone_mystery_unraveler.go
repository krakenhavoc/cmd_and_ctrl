package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zimone, Mystery Unraveler — Legendary Creature — Human Wizard
// {2}{G}{U}, 3/3:
//
//	"Landfall — Whenever a land you control enters, manifest dread if
//	 this is the first time this ability has resolved this turn.
//	 Otherwise, you may turn a permanent you control face up. (To
//	 manifest dread, look at the top two cards of your library. Put one
//	 onto the battlefield face down as a 2/2 creature and the other
//	 into your graveyard. Turn it face up any time for its mana cost if
//	 it's a creature card.)"
//
// "The first time this ability has resolved this turn" is the engine's
// per-object resolution tally (Game.ResolvedThisTurn, the one
// Sephiroth reads), keyed on this trigger's stack label. The count
// includes the resolution in progress, so the first resolution reads 1
// and every later one reads more. It is per OBJECT (CR 400.7): a Zimone
// that died and came back this turn starts again at zero. A landfall
// trigger that is countered or fizzles never resolved and does not
// count.
//
// The "otherwise" branch is a resolution-time pick of up to one
// face-down permanent the controller controls that an effect could
// turn over (CR 701.40b keeps a manifested noncreature card face
// down, so it is not offered), turned for free (ADR 0082's second
// 2026-10-07 amendment, #2590). With nothing to turn, no prompt is
// asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75bf0f27-9df1-4ab4-97a8-94bb3223bdf9",
		Name:         "Zimone, Mystery Unraveler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall(zimoneLandfallLabel, zimoneLandfall),
		},
	})
}

const zimoneLandfallLabel = "Zimone, Mystery Unraveler — landfall: manifest dread the first time this turn, otherwise you may turn a permanent you control face up"

func zimoneLandfall(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if g.ResolvedThisTurn(item.SourceCardID, zimoneLandfallLabel) <= 1 {
		return effects.ManifestDread{}.Apply(ctx)
	}
	return effects.ChoosePermanents{
		Question: "Zimone — you may turn a permanent you control face up",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := effects.FaceDownPermanentsThatCanTurnUp(g, of)
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 0, 1
		},
		Keyed: zimoneChosen,
	}.Apply(ctx)
}

// zimoneChosen is the chosen permanent turned face up, keyed (#4290).
var zimoneChosen = effects.RegisterPermanentsPick("prompt-run/zimone-mystery-unraveler-turn-face-up", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		return effects.TurnFaceUp{Targets: picked.Cards()}.Apply(ctx)
	},
})
