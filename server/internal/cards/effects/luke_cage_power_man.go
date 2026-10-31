package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "c95d1558-09a4-4dbc-b08a-fe21b47bc5d2",
		Name:         "Luke Cage, Power Man",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.CardID == source.InstanceID && attackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
			}, lukeCageLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
					return nil
				}
				if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: lukeCageLabel}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"indestructible"}, Label: lukeCageLabel}.Apply(ctx)
			}),
		},
	})
}

const lukeCageLabel = "Luke Cage, Power Man — Unbreakable Skin: +2/+0 and indestructible until end of turn"
