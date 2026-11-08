package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancestor Dragon — Creature — Dragon {4}{W}{W}, 5/6 (Foundations,
// #4465):
//
//	"Flying
//	 Whenever one or more creatures you control attack, you gain 1 life
//	 for each attacking creature."
//
// Haunted Hellride's "whenever you attack" (one trigger per attack
// declaration of yours, OncePerBatch over attackDeclaredByYou), whether
// or not the Dragon is among the attackers. The attacking creatures are
// counted as the trigger resolves (AttackingCreatures).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "60b03a44-a7bd-48fa-8c8f-1704aed02cd7",
		Name:            "Ancestor Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Ancestor Dragon — you gain 1 life for each attacking creature",
				func(g *game.Game, item *game.StackItem) error {
					return GainLife{Amount: len(AttackingCreatures(g))}.Apply(NewContext(g, item))
				})),
		},
	})
}
