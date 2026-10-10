package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clan Crafter — Legendary Enchantment — Background {1}{U} (EDHREC rank
// 11577):
//
//	"Commander creatures you own have "{2}, Sacrifice an artifact: Put
//	 a +1/+1 counter on this creature and draw a card.""
//
// An ADR 0093 grant of an activated ability to each commander creature
// you own, under anyone's control (choose_a_background.go): whoever
// controls the creature activates it, pays with their own artifacts
// and draws. The cost is Maximus, Knight Apparent's shape, {N} plus a
// sacrificed artifact, and any artifact will do, the creature itself
// included if it is one.
//
// No simplification.
const clanCrafterGrant = "clan-crafter/counter-and-draw"

func init() {
	Register(Spec{
		OracleID:     "6ebd4b3f-246e-475f-8cec-838eaaf597e2",
		Name:         "Clan Crafter",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: clanCrafterGrant,
			Activated: []ActivatedAbility{{
				Label:   "{2}, Sacrifice an artifact: Put a +1/+1 counter on this creature and draw a card.",
				Cost:    Plus(ManaCost("{2}"), game.AbilityCost{SacrificeOther: sacrificeSpec("an artifact", Artifact())}),
				Purpose: game.Purpose{Draws: 1, Answers: game.AnswerPump},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return DrawCards{N: 1}.Apply(ctx)
				},
			}},
			Text: "{2}, Sacrifice an artifact: Put a +1/+1 counter on this creature and draw a card.",
		}},
		Static: []game.StaticAbility{grantToCommanderCreaturesYouOwn(clanCrafterGrant)},
	})
}
