package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Danitha, Spear of Agony — Legendary Creature — Human Knight {2}{B}, 2/2:
//
//	"First strike
//	 Whenever you cast a spell that targets an opponent or a creature
//	 an opponent controls, put a +1/+1 counter on Danitha."
//
// Read at the cast (CR 601.2i), from the targets the spell announced:
// a player other than the caster, or a creature on the battlefield
// that someone other than the caster controls. A spell with two
// qualifying targets still triggers once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b5c7f11-75f2-4bad-86cd-f0232861f569",
		Name:            "Danitha, Spear of Agony",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller {
					return false
				}
				return rfCreatureACastSpellTargetsAny(g, ev.CardID, func(t game.TargetRef) bool {
					if t.Kind == game.TargetPlayer {
						return t.ID != source.Controller
					}
					c, ok := rfCreatureATargetIsCreatureOnBattlefield(g, t)
					return ok && c.Controller != source.Controller
				})
			}, "Danitha, Spear of Agony — put a +1/+1 counter on Danitha",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
