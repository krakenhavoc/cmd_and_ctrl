package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iname as One — Legendary Creature — Spirit {8}{B}{B}{G}{G}, 8/8 (Saviors of
// Kamigawa, #5601):
//
//	"When Iname as One enters, if you cast it from your hand, you may search
//	 your library for a Spirit permanent card, put it onto the battlefield,
//	 then shuffle.
//	 When Iname as One dies, you may exile it. If you do, return target Spirit
//	 permanent card from your graveyard to the battlefield."
//
// The enters trigger is Angel of the Dire Hour's cast-from-hand check, made
// as it triggers and again as it resolves (CR 603.4); the search is optional
// and finds a permanent card (not a token) with the Spirit subtype. The dies
// trigger targets a Spirit permanent card in your graveyard as it is put on the
// stack, asks "you may exile it" as it resolves, and returns the target only if
// Iname was exiled; with Iname itself as the target it is gone by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "865e569b-6999-4a03-ad02-4bfd595d5820",
		Name:         "Iname as One",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.B06SelfETB(ev, source, lki, g) && effects.B30CastFromHand(g, source.InstanceID)
			}, "Iname as One — you may search your library for a Spirit permanent card, put it onto the battlefield, then shuffle", effects.OldB587InameSearch),
			effects.Targeting(effects.Optional(
				effects.WhenThisDies("Iname as One — you may exile it; if you do, return target Spirit permanent card from your graveyard to the battlefield", effects.OldB587InameDies),
				"Iname as One — exile it?"),
				effects.TargetCardInGraveyard("target Spirit permanent card from your graveyard", effects.YouOwn(), effects.OldB587SpiritPermanentCardTarget())),
		},
	})
}
