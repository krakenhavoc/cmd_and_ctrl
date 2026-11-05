package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fire Navy Trebuchet — Artifact Creature — Wall {2}{B}, 0/4:
//
//	"Defender, reach
//	 Whenever you attack, create a 2/1 colorless Construct artifact
//	 creature token with flying named Ballistic Boulder that's tapped and
//	 attacking. Sacrifice that token at the beginning of the next end
//	 step."
//
// "Whenever you attack" is one trigger per attack declaration (Adeline's
// OncePerBatch over the declaration), so it fires once however many
// creatures attack, and the Trebuchet itself, a defender, never has to.
// The Boulder enters tapped and attacking a player, planeswalker or
// battle its controller chooses, not necessarily what the others attack
// (the ruling); it was never declared as an attacker, so "whenever a
// creature attacks" does not see it (CR 508.4). Mobilize's delayed
// trigger (CR 603.7) sacrifices exactly that token at the next end step.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:        "1f018a8c-6ceb-4d03-94a7-8713467c9cf5",
		Name:            "Fire Navy Trebuchet",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, "Fire Navy Trebuchet — create a tapped and attacking Ballistic Boulder",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokensAttackingYourChoice{
						Template: TokenCard(ballisticBoulderToken),
						N:        1,
						Tapped:   true,
						Label:    "Fire Navy Trebuchet",
						Tail:     game.KeyedTail{Ref: fireNavyTrebuchetSacrificeAtEndStep},
					}.Apply(NewContext(g, item))
				})),
		},
	})
}

// ballisticBoulderToken is the set's printed Ballistic Boulder (TTLA 13).
const ballisticBoulderToken = "2/1 colorless Construct artifact with flying named Ballistic Boulder"

// fireNavyTrebuchetSacrificeAtEndStep is "Sacrifice that token at the
// beginning of the next end step", told the token that landed.
var fireNavyTrebuchetSacrificeAtEndStep = AttackingTokensThen("attacking-tokens/fire-navy-trebuchet-sacrifice-the-boulder",
	func(ctx *Context, created []uuid.UUID, _ game.KeyedTail) error {
		if len(created) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label: "Fire Navy Trebuchet — sacrifice Ballistic Boulder",
			Cards: created,
			Body:  sacrificeListedCardsBody,
		}.Apply(ctx)
	})
