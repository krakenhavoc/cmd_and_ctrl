package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Furnace Layer — Plane — New Phyrexia (#4870):
//
//	"When you planeswalk to Furnace Layer and at the beginning of your upkeep,
//	 select target player at random. That player discards a card. If that
//	 player discards a land card this way, they lose 3 life.
//	 Whenever chaos ensues, you may destroy target nonland permanent."
//
// The player is a target chosen at random as the ability goes on the stack
// (ChosenAtRandom, CR 603.3d), and may be the planar controller (the ruling).
// Their discard is their own choice; a keyed answer reads what was discarded,
// and a land card costs them 3 life. With an empty hand nothing happens, and
// no life is lost. The chaos ability's "you may" is asked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1799d361-85e7-4c04-aa04-6f83e6499573",
		Name:         "Furnace Layer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenYouPlaneswalkToThisAndAtYourUpkeep("Furnace Layer — a player selected at random discards a card; a land card costs them 3 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   t.ID,
						Source:   item.SourceCardID,
						N:        1,
						Question: "Furnace Layer — discard a card (a land card costs you 3 life)",
					}, furnaceLayerDiscarded, nil)
				}),
				ChosenAtRandom(TargetPlayer("target player selected at random"))),
			Targeting(Optional(WheneverChaosEnsues("Furnace Layer — you may destroy target nonland permanent", destroyFirstLegalTarget), "Furnace Layer — destroy the target nonland permanent?"),
				TargetPermanent("target nonland permanent", Not(Land()))),
		},
	})
}

// furnaceLayerDiscarded is the discard's answer: 3 life lost for each land
// card the player discarded. An on-disk key: never renamed, never reused.
var furnaceLayerDiscarded = game.RegisterDiscardThen("choose-cards/furnace-layer-discarded", func(g *game.Game, r game.DiscardsDone) error {
	for _, id := range r.Discarded.By(r.Player) {
		c, ok := g.LookupCardForEffect(id)
		if !ok || !c.IsLand() {
			continue
		}
		return g.ChangePlayerLifeForEffect(r.Source, r.Player, -3)
	}
	return nil
})
