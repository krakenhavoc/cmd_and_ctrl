package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oscorp Industries — Land (Marvel's Spider-Man):
//
//	"This land enters tapped.
//	 When this land enters from a graveyard, you lose 2 life.
//	 {T}: Add {U}, {B}, or {R}.
//	 Mayhem (You may play this card from your graveyard if you
//	 discarded it this turn. Timing rules still apply.)"
//
// Mayhem without a cost is CR 702.187c, "you may play this card from
// your graveyard if you discarded it this turn": the graveyard opening
// every mayhem card declares (MayhemCastableZone) and no alternative
// cost, because a land is played, not cast, and has no mana cost to
// replace. The enters trigger reads the ETB's EnteredFrom (CR 603.6a),
// so a land that came out of a graveyard any other way loses the life
// too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f432eb6a-f1bf-4ce7-b915-1488dccb4bb9",
		Name:            "Oscorp Industries",
		Completeness:    effects.CompletenessFull,
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
		CastableZonesIf: []game.ConditionalCastZone{effects.MayhemCastableZone()},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && effects.EnteredFromAGraveyard(ev, source, lki, g)
			}, "Oscorp Industries — you lose 2 life", func(g *game.Game, item *game.StackItem) error {
				return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -2)
			}),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U|B|R}",
			Label:    "Add {U}, {B}, or {R}",
		}},
	})
}
