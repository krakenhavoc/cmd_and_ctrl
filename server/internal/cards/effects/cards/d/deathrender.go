package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deathrender — Artifact — Equipment {4}:
//
//	"Equipped creature gets +2/+2.
//	 Whenever equipped creature dies, you may put a creature card from your hand
//	 onto the battlefield and attach this Equipment to it.
//	 Equip {2}"
//
// The pump is the usual attached static and Equip {2} the usual equip. The
// death trigger is the Equipment's own (EquippedCreatureDied reads the creature
// that was attached as it died, CR 603.10). The put is Howlpack Piper's optional
// pick from hand, so declining is an answer and a hand with no creature card
// asks nothing; the attach rides it as a keyed tail, run once the card has
// landed, with the trigger's own item so a Deathrender that has since left the
// battlefield, or come back as a new object, attaches to nothing (CR 701.3b).
// The creature enters unattached for the replacement window, as any entry does,
// and is attached afterwards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05ac6e5c-b773-45ed-9c74-3cd4b05bc287",
		Name:         "Deathrender",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.EquippedCreatureDied(ev, source)
			}, "Deathrender — you may put a creature card from your hand onto the battlefield and attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PutFromHandOntoBattlefield{
						Match:    effects.Creature(),
						Optional: true,
						Label:    "Deathrender — you may put a creature card from your hand onto the battlefield",
						ThenTail: deathrenderAttach,
						ThenItem: true,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

// deathrenderAttach attaches the Equipment to the creature that arrived. An
// on-disk key.
var deathrenderAttach = effects.PutFromHandTail("deathrender/attach-to-it", func(ctx *effects.Context, res effects.PutFromHandResult) error {
	if res.Entered == uuid.Nil || ctx.Item == nil || !effects.OnBattlefield(ctx.Game, res.Entered) {
		return nil
	}
	return ctx.Game.AttachSourceForEffect(ctx.Item, game.TargetRef{Kind: game.TargetCard, ID: res.Entered})
})
