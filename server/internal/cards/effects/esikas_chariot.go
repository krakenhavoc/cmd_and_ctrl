package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Esika's Chariot — Legendary Artifact — Vehicle, 4/4, for {3}{G}:
//
//	"When Esika's Chariot enters, create two 2/2 green Cat creature
//	 tokens.
//	 Whenever Esika's Chariot attacks, create a token that's a copy
//	 of target token you control.
//	 Crew 4"
//
// The two Cats are what crews it: 2 + 2 = 4, exactly the crew number,
// which is the card's whole design and the reason crew reads
// EFFECTIVE power — a pair of Cats under any anthem still crews, and
// a Cat that has been shrunk below 2 stops the Chariot cold.
//
// The attack trigger targets a TOKEN you control, which includes the
// Cats and, notably, includes the copy it made last turn.
func init() {
	Register(Spec{
		OracleID:     "8e7b079d-9ede-421c-bd2b-f9a5126a8e6f",
		Name:         "Esika's Chariot",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The Cat token is created colorless instead of green, so anything that cares about a creature's color doesn't see it."},
		Activated: []ActivatedAbility{{
			Label:  "Crew 4",
			Cost:   CrewCost(4),
			Effect: CrewEffect("Esika's Chariot"),
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Esika's Chariot — create two 2/2 Cats", Do(CreateToken{
				Template: CatToken(),
				N:        2,
			})),
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclared(ev, source)
				},
				Targets: TargetPermanent("target token you control", YouControl(), IsTokenPredicate()),
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Esika's Chariot — copy target token",
						func(g *game.Game, item *game.StackItem) error {
							if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
								return nil
							}
							return CreateTokenCopy{
								Controller: item.Controller,
								Copy:       item.Targets[0].ID,
								N:          1,
							}.Apply(NewContext(g, item))
						})
				},
			},
		},
	})
}
