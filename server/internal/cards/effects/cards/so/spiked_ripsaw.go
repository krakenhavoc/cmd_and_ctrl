package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spiked Ripsaw — {2}{G} Artifact — Equipment (Innistrad: Crimson Vow,
// S82):
//
//	"Equipped creature gets +3/+3.
//	 Whenever equipped creature attacks, you may sacrifice a Forest. If
//	 you do, that creature gains trample until end of turn.
//	 Equip {3}"
//
// Saw's attack trigger: it fires for the equipped creature's attack and
// asks "you may sacrifice" over the lands with the Forest land type
// (read through the card's text, Magical Hack's word), then the
// registered "if you do" gives that attacker trample until end of turn.
// The attacker comes off the trigger's event, so a creature the
// Equipment has left since still gets it (it was the one that attacked).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "660ff9fa-1e9b-4697-a3f5-b12238703161",
		Name:         "Spiked Ripsaw",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Spiked Ripsaw — you may sacrifice a Forest. If you do, that creature gains trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					forest := g.TextReaderForEffect(item).TextSubtype("Forest")
					return effects.MaySacrificeOne(g, item, "Spiked Ripsaw — sacrifice a Forest to give the creature trample until end of turn?",
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() && c.HasSubtype(forest) },
						false, spikedRipsawTrample)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

// spikedRipsawTrample is "if you do, that creature gains trample until end
// of turn": the creature that attacked, off the trigger's event.
var spikedRipsawTrample = effects.RegisterSacrificedOne("prompt-run/spiked-ripsaw-trample", func(ctx *effects.Context, _ uuid.UUID) error {
	attacker := ctx.Trigger().Event.CardID
	if !effects.OnBattlefield(ctx.Game, attacker) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: attacker, Keywords: []string{"trample"}, Label: "Spiked Ripsaw — trample"}.Apply(ctx)
})
