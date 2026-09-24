package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Don Andres, the Renegade — Legendary Creature — Vampire Pirate
// {1}{U}{B}{R}, 4/3 (Edea steal-and-sac deck, #1565):
//
//	"Each creature you control but don't own gets +2/+2, has menace
//	 and deathtouch, and is a Pirate in addition to its other types.
//	 Whenever you cast a noncreature spell you don't own, create two
//	 tapped Treasure tokens."
//
// The static is one ability in three layers (CR 613.1): the Pirate
// type in layer 4, menace and deathtouch in layer 6, +2/+2 in layer
// 7c. All three share one AppliesTo, "a creature you control whose
// owner is someone else", read off the post-layer-2 controller, so a
// creature stolen by Edea, Act of Treason or an exchange gets the
// bonus the moment it changes hands and loses it the moment it goes
// back. Agent of Treachery reads the same board fact the same way
// (permanentsYouControlButDontOwn).
//
// The cast trigger reads the spell's OWNER against its caster: a
// noncreature spell cast from an opponent's library or graveyard
// (Gonti, Hostage Taker, Outrageous Robbery, a stolen card cast from
// exile) is a spell you don't own. A creature spell you don't own is
// not a noncreature spell, so it makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "060ef981-db05-436b-b1c1-f55071375344",
		Name:         "Don Andres, the Renegade",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			{
				Layer:     game.Layer4Type,
				AppliesTo: donAndresStolenCreature,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !eotHasType(c.Subtypes, "Pirate") {
						c.Subtypes = append(c.Subtypes, "Pirate")
					}
				},
			},
			{
				Layer:     game.Layer6Ability,
				AppliesTo: donAndresStolenCreature,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					for _, kw := range []string{"menace", "deathtouch"} {
						if !eotHasAbility(c.Abilities, kw) {
							c.Abilities = append(c.Abilities, kw)
						}
					}
				},
			},
			{
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: donAndresStolenCreature,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCast(And(Noncreature(), notOwnedByCaster())),
				"Don Andres, the Renegade — create two tapped Treasure tokens",
				Do(CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: 2})),
		},
	})
}

// donAndresStolenCreature is "each creature you control but don't
// own".
func donAndresStolenCreature(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller && target.Owner != source.Controller
}

// notOwnedByCaster is "a spell you don't own": the card's owner is not
// the player the predicate is asked about.
func notOwnedByCaster() CardPredicate {
	return func(_ *game.Game, caster uuid.UUID, c game.Card) bool { return c.Owner != caster }
}
