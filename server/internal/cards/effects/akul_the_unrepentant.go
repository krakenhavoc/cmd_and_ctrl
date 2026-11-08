package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akul the Unrepentant — Legendary Creature — Scorpion Dragon Rogue
// {B}{B}{R}{R}, 5/5 (Outlaws of Thunder Junction, #4468):
//
//	"Flying, trample
//	 Sacrifice three other creatures: You may put a creature card from
//	 your hand onto the battlefield. Activate only as a sorcery and only
//	 once each turn."
//
// The cost is Elesh Norn, the Argent Etchings' three other creatures
// (SacrificeAnotherN), the effect Quicksilver Amulet's optional put from
// hand (PutFromHandOntoBattlefield), and the two limits are the row's
// SorcerySpeed (CR 602.5d) and Aggressive Mining's once-each-turn
// condition. Declining the put still spends the turn's activation: the
// creatures were a cost.
//
// No simplification.
func init() {
	const label = "Sacrifice three other creatures: You may put a creature card from your hand onto the battlefield. Activate only as a sorcery and only once each turn."
	Register(Spec{
		OracleID:        "d6b56d6d-e2f6-47f5-a9d1-cd0e17f94a0d",
		Name:            "Akul the Unrepentant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Activated: []ActivatedAbility{{
			Label:        label,
			Cost:         SacrificeAnotherN(3, "three other creatures", Creature()),
			SorcerySpeed: true,
			Condition:    OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutFromHandOntoBattlefield{
					Match:    Creature(),
					Optional: true,
					Label:    "Akul the Unrepentant — you may put a creature card from your hand onto the battlefield",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
