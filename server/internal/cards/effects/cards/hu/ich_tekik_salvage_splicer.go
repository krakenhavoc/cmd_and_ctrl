package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ich-Tekik, Salvage Splicer — Legendary Creature — Phyrexian Human
// Artificer {4}{G}, 1/1 (Commander Legends):
//
//	"When Ich-Tekik enters, create a 3/3 colorless Phyrexian Golem
//	 artifact creature token.
//	 Whenever an artifact is put into a graveyard from the battlefield,
//	 put a +1/+1 counter on Ich-Tekik and a +1/+1 counter on each Golem
//	 you control.
//	 Partner"
//
// The entry trigger is Blade Splicer's token. The second is Joven and
// Chandler's condition (anyone's artifact, a token too). The Golems are
// read as the trigger resolves, Ich-Tekik's own counter only if it is
// still on the battlefield; a Golem Ich-Tekik himself is (a changeling
// aside) never counted twice, since he gets one counter either way.
//
// Partner is a deck-construction rule (CR 702.124), read off the oracle
// text, so it declares nothing here.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcbada94-54fe-4330-8303-c46162420228",
		Name:         "Ich-Tekik, Salvage Splicer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ich-Tekik, Salvage Splicer — create a 3/3 colorless Phyrexian Golem artifact creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 colorless Phyrexian Golem artifact"), N: 1})),
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B23ArtifactPutIntoGraveyardFromBattlefield(ev, g)
			}, "Ich-Tekik, Salvage Splicer — put a +1/+1 counter on Ich-Tekik and a +1/+1 counter on each Golem you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					golem := ctx.TextSubtype("Golem")
					var ids []uuid.UUID
					if effects.B09SourceStillOnBattlefield(g, item) {
						ids = append(ids, item.SourceCardID)
					}
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.InstanceID != item.SourceCardID && c.HasSubtype(golem) {
							ids = append(ids, c.InstanceID)
						}
					}
					return effects.B13PutCounterOnEach(ctx, ids)
				}),
		},
	})
}
