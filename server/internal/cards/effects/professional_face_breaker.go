package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Professional Face-Breaker — Creature — Human Warrior {2}{R}, 2/3
// (EDHREC rank 209):
//
//	"Menace
//	 Whenever one or more creatures you control deal combat damage
//	 to a player, create a Treasure token.
//	 Sacrifice a Treasure: Exile the top card of your library. You
//	 may play that card this turn."
//
// Red's card advantage engine: attack, get a Treasure, crack it for
// a card. Three abilities, all real:
//
//   - Menace rides PrintedKeywords; the combat engine enforces it.
//   - "ONE OR MORE creatures … deal combat damage" is one trigger
//     per combat damage step, not one per creature. The engine
//     emits one EventDealDamage per creature, so the AppliesTo
//     declines any event that arrives while a Face-Breaker trigger
//     is already waiting on PendingTriggers — see
//     OncePerBatch for why that is exactly "once per
//     batch". Without it the card would be STRONGER than printed
//     (three attackers, three Treasures), which is the #259
//     direction and not shippable. First-strike and regular damage
//     are two batches and two triggers, as in paper.
//   - The impulse draw is Ragavan's ExileTopWithPermission with the
//     "play" grant (lands included — it says play, not cast), paid
//     for by sacrificing any permanent with the Treasure subtype,
//     the same SacrificeOther cost shape Goblin Bombardment's
//     creature clause uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04152e7a-969c-4858-841b-0a569a9fc1bf",
		Name:            "Professional Face-Breaker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Professional Face-Breaker — create a Treasure", Do(CreateToken{
				Template: TreasureToken(),
				N:        1,
			}))),
		},
		Activated: []ActivatedAbility{{
			Label: "Sacrifice a Treasure: Exile the top card of your library. You may play that card this turn.",
			Cost:  game.AbilityCost{SacrificeOther: sacrificeSpec("a Treasure", isTreasure)},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ExileTopWithPermission{
					From:    item.Controller,
					GrantTo: item.Controller,
					N:       1,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
