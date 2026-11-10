package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skeleton Crew — Creature — Skeleton Pirate {3}{B}, 3/3:
//
//	"Each other creature you control that's a Skeleton or Pirate gets +1/+1.
//	 Whenever one or more creature cards leave your graveyard, create a 2/2
//	 black Skeleton Pirate creature token. (This ability triggers only from
//	 the battlefield.)
//	 {5}{B}: Return this card from your graveyard to the battlefield tapped."
//
// The anthem is a tribal lord over two types (a creature that is either
// gets it, once). The trigger is a batched leave-the-graveyard watcher
// shared with Desecrated Tomb: several creature cards leaving at once make
// one token. The return is an activated ability that works from the
// graveyard (CR 602.2).
func init() {
	skeletonsAndPirates := TribeFilter{Tribes: []string{"Skeleton", "Pirate"}, Others: true, YoursOnly: true}
	Register(Spec{
		OracleID:     "2536750c-3c4f-4e00-b517-680ad4fd2ecc",
		Name:         "Skeleton Crew",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{TribalAnthem(skeletonsAndPirates, 1, 1)},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(leftGraveyardEvents,
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return creatureCardLeftYourGraveyard(ev, source, g)
				},
				"Skeleton Crew — create a 2/2 black Skeleton Pirate creature token",
				func(g *game.Game, item *game.StackItem) error {
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 black Skeleton Pirate"),
						N:          1,
					}.Apply(NewContext(g, item))
				})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{5}{B}: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{5}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisFromGraveyardTapped,
		}},
	})
}
