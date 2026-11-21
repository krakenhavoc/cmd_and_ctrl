package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wildfire Awakener — Creature — Human Wizard {X}{1}{R}{W} 3/2 (March of
// the Machine Commander):
//
//	"Convoke
//	 When this creature enters, create X 1/1 red Elemental creature tokens
//	 with 'Whenever this token becomes tapped, it deals 1 damage to target
//	 player.'"
//
// X is the value announced for the spell (CR 107.3m), read off the
// permanent as the trigger is built. The Elemental carries its own
// trigger (ADR 0083): any tap of it counts, an attack or a convoke
// included (ThisBecameTapped), and the damage is the token's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d22196b5-b111-45ce-ae51-a06ca02533b6",
		Name:         "Wildfire Awakener",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: Self,
			Key:       wildfireAwakenerLabel,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, wildfireAwakenerLabel)
				item.Params.Amount = source.CastX()
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if item.Params.Amount <= 0 {
					return nil
				}
				return CreateToken{
					Controller: item.Controller,
					Template:   wildfireElementalToken(),
					N:          item.Params.Amount,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

const wildfireAwakenerLabel = "Wildfire Awakener — create X 1/1 red Elemental tokens that ping a player when they become tapped"

// wildfireElementalToken is Wildfire Awakener's 1/1 red Elemental with
// "Whenever this token becomes tapped, it deals 1 damage to target player."
func wildfireElementalToken() game.Card { return TokenFromCatalog(printedWildfireElementalToken) }

func printedWildfireElementalToken() TokenTemplate {
	return TokenTemplate{
		Slug: "elemental-tapped-ping",
		Card: game.Card{
			Name:      "Elemental",
			TypeLine:  "Token Creature — Elemental",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"R"},
		},
		Triggered: []game.TriggeredAbility{
			Targeting(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, ThisBecameTapped,
				"Elemental — 1 damage to target player", SourceDealsOneToFirstTarget),
				TargetPlayer("target player")),
		},
		Text: "Whenever this token becomes tapped, it deals 1 damage to target player.",
	}
}
