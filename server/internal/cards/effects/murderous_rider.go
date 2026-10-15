package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murderous Rider // Swift End — an Adventure card (CR 715):
//
//	Murderous Rider — Creature — Zombie Knight {1}{B}{B}, 2/3
//	  "Lifelink
//	   When this creature dies, put it on the bottom of its owner's
//	   library."
//	Swift End — Instant — Adventure {1}{B}{B}
//	  "Destroy target creature or planeswalker. You lose 2 life. (Then
//	   exile this card. You may cast the creature later from exile.)"
//
// The dies trigger moves the card only while it is still the graveyard
// object that died (CR 400.7), through the tuck route, so a
// commander's owner is offered the command zone (CR 903.9) and a token,
// which ceased to exist (CR 111.7), is not found. Swift End is
// Dreadbore's destroy plus the caster's life loss.
//
// No simplification.
const murderousRiderOracleID = "1080c5b5-6651-4c6a-93e6-099fbe389e26"

func init() {
	Register(Spec{
		OracleID:        murderousRiderOracleID,
		Name:            "Murderous Rider",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Murderous Rider — put it on the bottom of its owner's library",
				func(g *game.Game, item *game.StackItem) error {
					z := g.FindCardZoneForEffect(item.SourceCardID)
					if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
						return nil
					}
					return PutIntoLibrary{Card: item.SourceCardID, ToBottom: true}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     murderousRiderOracleID + "#1",
		Name:         "Swift End",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("target creature or planeswalker",
			Or(Creature(), Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyChosenPermanent(ctx.Game, item); err != nil {
				return err
			}
			return ctx.Game.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -2)
		},
	})
}
