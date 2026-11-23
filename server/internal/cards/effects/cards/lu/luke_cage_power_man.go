package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Luke Cage, Power Man — Legendary Creature — Human Hero {3}{W}, 2/5
// (Marvel Super Heroes, #4455):
//
//	"Unbreakable Skin — Whenever Luke Cage attacks alone, he gets +2/+0
//	 and gains indestructible until end of turn."
//
// Bilbo's Ring's "alone": game.AttackedAlone is read when the attack
// trigger is harvested, after every attacker of the declaration is
// stamped, and the attacker must be Luke Cage himself. If he is gone when
// the trigger resolves he gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c95d1558-09a4-4dbc-b08a-fe21b47bc5d2",
		Name:         "Luke Cage, Power Man",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && effects.AttackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
			}, lukeCageLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: lukeCageLabel}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"}, Label: lukeCageLabel}.Apply(ctx)
			}),
		},
	})
}

const lukeCageLabel = "Luke Cage, Power Man — Unbreakable Skin: +2/+0 and indestructible until end of turn"
