package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Captain America, Skybound — Legendary Creature — Human Soldier Hero
// {3}{W}{W}, 3/5 (Marvel Super Heroes Commander, #4455):
//
//	"Flying, vigilance
//	 Battalion — Whenever Captain America and at least two other
//	 creatures attack, put a +1/+1 counter on each attacking creature.
//	 They gain indestructible until end of turn."
//
// "Them" is the creatures that got a counter: the set is read once as
// the trigger resolves (CR 611.2c) and both halves use it, so a creature
// that attacks later does not join.
func init() {
	Register(Spec{
		OracleID:        "e9566303-15cf-4444-a044-fc19540f21b6",
		Name:            "Captain America, Skybound",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclared(ev, source) && attackersYouControlOtherThan(g, source.Controller, source.InstanceID) >= 2
			}, "Captain America — a +1/+1 counter on each attacking creature; they gain indestructible (battalion)",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ids := b13AttackingCreaturesYouControl(g, item.Controller)
					if err := (GrantKeywordUntilEOT{
						Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
							return slices.Contains(ids, c.InstanceID)
						},
						Keywords: []string{"indestructible"},
						Label:    "Captain America, Skybound — indestructible",
					}).Apply(ctx); err != nil {
						return err
					}
					return b13PutCounterOnEach(ctx, ids)
				}),
		},
	})
}
