package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Floral Evoker — Creature — Snake Druid {2}{G}, 2/3 (Tarkir: Dragonstorm
// Commander):
//
//	"Landfall — Whenever a land you control enters, put a +1/+1 counter on
//	 this creature.
//	 {G}, Discard a creature card: Return target land card from your graveyard
//	 to the battlefield tapped."
//
// Landfall is the shared trigger (every land entering under your control,
// played or put there, CR 305.4). The ability is Fauna Shaman's discard cost
// (a creature card from hand, paid at announce, CR 602.2b) with Glimpse the
// Core's tapped return of a land card from your graveyard; the target is
// chosen as the ability is activated and checked again on resolution
// (CR 608.2b), so a land that left the graveyard fizzles it, though the
// discard stays paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4da17e9e-df3e-4c46-a0cf-baafbe67186f",
		Name:         "Floral Evoker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Landfall("Floral Evoker — put a +1/+1 counter on this creature", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{G}, Discard a creature card: Return target land card from your graveyard to the battlefield tapped.",
			Cost:    Plus(ManaCost("{G}"), DiscardCardsMatching(1, "a creature card", MatchCreature)),
			Targets: TargetCardInGraveyard("target land card from your graveyard", Land(), YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Tapped: true}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
