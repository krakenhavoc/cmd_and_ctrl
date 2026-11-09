package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fear of Missing Out — Enchantment Creature — Nightmare {1}{R}, 2/3
// (Duskmourn, #4466):
//
//	"When this creature enters, discard a card, then draw a card.
//	 Delirium — Whenever this creature attacks for the first time each
//	 turn, if there are four or more card types among cards in your
//	 graveyard, untap target creature. After this phase, there is an
//	 additional combat phase."
//
// The enters trigger is a loot: the discard is a real choice, and the draw
// is its continuation, so it waits for the answer (and happens even with an
// empty hand, the discard having been impossible). The attack trigger is
// Aurelia's "first time each turn" (TimesAttackedThisTurn) with delirium as
// an intervening "if" (CR 603.4: checked on attack and again on resolution,
// by the card types among your graveyard's cards). It untaps the target and
// adds the combat phase after this one.
//
// No simplification.
func init() {
	delirium := func(g *game.Game, source *game.Card) bool {
		return b16CardTypesInGraveyard(g, source.Controller) >= 4
	}
	attack := On(game.EventAttack, AllOf(attacksForTheFirstTimeThisTurn,
		func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			return delirium(g, source)
		}),
		"Fear of Missing Out — untap target creature; after this phase, there is an additional combat phase",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			if src, ok := g.LookupCardForEffect(item.SourceCardID); !ok || !delirium(g, &src) {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					if err := (UntapTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return ExtraCombatAfterThisPhase().Apply(ctx)
		})
	Register(Spec{
		OracleID:     "236ce592-1321-4a7c-b695-fd84496f8924",
		Name:         "Fear of Missing Out",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Fear of Missing Out — discard a card, then draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
						Player:   item.Controller,
						Source:   item.SourceCardID,
						N:        1,
						Question: "Fear of Missing Out — discard a card",
					}, fearOfMissingOutDraw, nil)
				}),
			Targeting(attack, TargetCreature("target creature")),
		},
	})
}

// fearOfMissingOutDraw is the loot's "then draw a card". An on-disk key.
var fearOfMissingOutDraw = game.RegisterDiscardThen("choose-cards/fear-of-missing-out-draw", func(g *game.Game, r game.DiscardsDone) error {
	return g.DrawNForEffect(r.Player, 1)
})
