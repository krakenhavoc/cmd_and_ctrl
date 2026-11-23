package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imotekh the Stormlord — Legendary Artifact Creature — Necron {2}{B}{B},
// 3/3 (Warhammer 40,000 Commander, S82 GAP batch 188, #5948):
//
//	"Phaeron — Whenever one or more artifact cards leave your graveyard,
//	 create two 2/2 black Necron Warrior artifact creature tokens.
//	 Grand Strategist — At the beginning of combat on your turn, another
//	 target artifact creature you control gets +2/+2 and gains menace
//	 until end of turn."
//
// The first is a batched trigger (CR 603.2c): a unearth, a reanimation or
// a mass return that takes several artifact cards at once makes one pair of
// Warriors. A card is judged as it was in the graveyard (CR 603.10a); tokens
// never count. The second targets as the trigger goes on the stack and
// applies both parts to the chosen creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "425fea06-76ec-4cd9-bd3b-750504fd7fd2",
		Name:         "Imotekh the Stormlord",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.CardThatLeftYourGraveyard(ev, source, g)
				return ok && effects.WasCardInAGraveyard(c, game.Card.IsArtifact)
			}, "Imotekh the Stormlord — create two 2/2 black Necron Warrior artifact creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Necron Warrior artifact"), N: 2}))),
			effects.Targeting(effects.AtBeginningOfYourCombat("Imotekh the Stormlord — another target artifact creature you control gets +2/+2 and gains menace until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Imotekh the Stormlord — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"menace"}, Label: "Imotekh the Stormlord — menace"}.Apply(ctx)
				}),
				effects.Another(effects.TargetCreature("another target artifact creature you control", effects.Artifact(), effects.YouControl()))),
		},
	})
}
