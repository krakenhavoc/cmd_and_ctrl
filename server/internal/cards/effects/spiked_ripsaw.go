package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "660ff9fa-1e9b-4697-a3f5-b12238703161",
		Name:         "Spiked Ripsaw",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return source.IsAttachedTo(ev.CardID)
			}, "Spiked Ripsaw — you may sacrifice a Forest. If you do, that creature gains trample until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					forest := g.TextReaderForEffect(item).TextSubtype("Forest")
					return maySacrificeOne(g, item, "Spiked Ripsaw — sacrifice a Forest to give the creature trample until end of turn?",
						func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() && c.HasSubtype(forest) },
						false, spikedRipsawTrample)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

// spikedRipsawTrample is "if you do, that creature gains trample until end
// of turn": the creature that attacked, off the trigger's event.
var spikedRipsawTrample = RegisterSacrificedOne("prompt-run/spiked-ripsaw-trample", func(ctx *Context, _ uuid.UUID) error {
	attacker := ctx.Trigger().Event.CardID
	if !onBattlefield(ctx.Game, attacker) {
		return nil
	}
	return GrantKeywordUntilEOT{Target: attacker, Keywords: []string{"trample"}, Label: "Spiked Ripsaw — trample"}.Apply(ctx)
})
