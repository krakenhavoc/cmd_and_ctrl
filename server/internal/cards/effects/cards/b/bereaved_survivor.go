package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bereaved Survivor // Dauntless Avenger — Creature — Human Peasant {2}{W},
// 2/1 // Creature — Human Soldier 3/2 (Midnight Hunt):
//
//	Bereaved Survivor
//	  "When another creature you control dies, transform this creature."
//	Dauntless Avenger
//	  "Whenever this creature attacks, return target creature card with
//	   mana value 2 or less from your graveyard to the battlefield tapped
//	   and attacking."
//
// The front trigger fires once per other creature that dies; the engine
// ignores a second transform once the Survivor has already turned over
// (ADR 0152 §3), so two creatures dying together turn it over once. The
// Avenger's target is chosen as the trigger goes on the stack and
// re-checked as it resolves (CR 608.2b). The creature enters tapped and
// attacking (CR 508.4): the controller picks what it attacks, and it was
// never declared as an attacker, so no "whenever attacks" trigger sees it.
//
// No simplification.
func init() {
	const oracle = "bfa60dd1-fcc8-4126-ab7a-4695c74f5bce"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Bereaved Survivor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B39AnotherCreatureYouControlDied(ev, source, g)
			}, "Bereaved Survivor — transform this creature", effects.Do(effects.TransformThis{})),
		},
	})
	avenge := effects.WheneverThisAttacks("Dauntless Avenger — return target creature card with mana value 2 or less from your graveyard to the battlefield tapped and attacking",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, id := range effects.LegalTargetIDs(ctx) {
				return effects.PutAttackingYourChoice(ctx, id, game.ZoneGraveyard, "Dauntless Avenger")
			}
			return nil
		})
	avenge.Targets = effects.TargetCardInGraveyard("target creature card with mana value 2 or less in your graveyard",
		effects.YouOwn(), effects.Creature(), effects.ManaValueLE(2))
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Dauntless Avenger",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{avenge},
	})
}
