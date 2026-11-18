package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vilespawn Spider — {G}{U} Creature — Spider 2/3 (Innistrad: Crimson
// Vow, S82):
//
//	"Reach
//	 At the beginning of your upkeep, mill a card.
//	 {2}{G}{U}, {T}, Sacrifice this creature: Create a 1/1 green Insect
//	 creature token for each creature card in your graveyard. Activate
//	 only as a sorcery."
//
// The count is read as the ability resolves, after the Spider itself
// went to the graveyard as a cost (Abzan Monument's sacrifice-to-make
// shape), so the sacrificed Spider counts. A token is never a card, so
// the creature cards counted are real cards (CR 108.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3e969073-76cc-4a2c-9abd-3a4094488fe5",
		Name:            "Vilespawn Spider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Vilespawn Spider — mill a card", Do(MillCards{N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{2}{G}{U}, {T}, Sacrifice this creature: Create a 1/1 green Insect creature token " +
				"for each creature card in your graveyard. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{G}{U}"), TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				n, _ := cmdr2S30CreatureCardsIn(g, item.Controller)
				if n <= 0 {
					return nil
				}
				return CreateToken{Template: TokenCard("1/1 green Insect"), N: n}.Apply(NewContext(g, item))
			},
		}},
	})
}
