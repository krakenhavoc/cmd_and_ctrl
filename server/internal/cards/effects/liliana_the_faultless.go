package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Liliana the Faultless — Legendary Creature — Human Cleric {W}, 1/1
// (Reality Fracture):
//
//	"Whenever another creature or planeswalker you control enters, you
//	 gain 1 life.
//	 {1}, {T}, Discard a card: Another target creature or planeswalker
//	 you control gains hexproof until end of turn."
//
// The discard is a cost paid at announce (the activator picks the card
// from hand), so a response that kills the target does not give it back.
func init() {
	Register(Spec{
		OracleID:     "22c2e66d-a2a1-46db-a70a-dff5ad5d6c14",
		Name:         "Liliana the Faultless",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, anotherCreatureOrPlaneswalkerEnteredUnderYourControl,
				"Liliana the Faultless — you gain 1 life",
				Do(GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Discard a card: Another target creature or planeswalker you control gains hexproof until end of turn.",
			Cost:    Plus(ManaCost("{1}"), TapCost(), DiscardACard()),
			Targets: Another(TargetPermanent("another target creature or planeswalker you control", Or(Creature(), Planeswalker()), YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
