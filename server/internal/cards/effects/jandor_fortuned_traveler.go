package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jandor, Fortuned Traveler — {1}{R}{G}{W} Legendary Creature — Human
// Citizen 2/4 (Mystery Booster Commander, S82 GAP batch 3, #5948):
//
//	"Each Beast, Camel, Horse, Elephant, and Wolf creature you control is
//	 a Mount in addition to its other types and has saddle 2. (Tap any
//	 number of other creatures you control with total power 2 or more:
//	 That Mount becomes saddled until end of turn. Saddle only as a
//	 sorcery.)
//	 Whenever a saddled creature you control attacks, it gets +2/+2 until
//	 end of turn. Draw a card. Create a Food token."
//
// The first sentence is two layered effects over the same set of
// creatures (yours, with one of the five types, Jandor not excepted since
// he is a Human): a layer-4 subtype add (Mount) and a layer-6 grant of a
// Saddle 2 bundle, so a creature that stops being one of the five loses
// both and a Saddle 2 the creature already printed is a second, separate
// ability. The trigger is one per saddled creature of yours that
// attacks, read when it is declared; the +2/+2 is for the attacker (if it
// has left, nothing), and the draw and the Food are the controller's.
//
// No simplification.
func init() {
	const grant = "jandor-fortuned-traveler/saddle-2"
	mountTribes := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		if !target.IsCreature() || target.Controller != source.Controller {
			return false
		}
		for _, t := range []string{"Beast", "Camel", "Horse", "Elephant", "Wolf"} {
			if target.HasSubtype(t) {
				return true
			}
		}
		return false
	}
	mount := game.StaticAbility{
		Layer:     game.Layer4Type,
		AppliesTo: mountTribes,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			if eotHasType(c.Subtypes, "Mount") {
				return
			}
			c.Subtypes = append(append([]string(nil), c.Subtypes...), "Mount")
		},
		Label: "Each Beast, Camel, Horse, Elephant, and Wolf creature you control is a Mount in addition to its other types.",
	}
	Register(Spec{
		OracleID:     "8c05de43-6f9e-46c1-bc01-551a7fc570d0",
		Name:         "Jandor, Fortuned Traveler",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       grant,
			Activated: []ActivatedAbility{Saddle(2)},
			Text:      "Saddle 2",
		}},
		Static: []game.StaticAbility{mount, GrantAbilities(mountTribes, grant)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && isSaddledSource(&c)
			}, "Jandor, Fortuned Traveler — the saddled creature gets +2/+2, draw a card, create a Food token",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if item.Trigger != nil && onBattlefield(g, item.Trigger.Event.CardID) {
						if err := (BoostUntilEOT{Target: item.Trigger.Event.CardID, Power: 2, Toughness: 2,
							Label: "Jandor, Fortuned Traveler — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
					}
					if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					return CreateToken{Template: FoodToken(), N: 1}.Apply(ctx)
				}),
		},
	})
}
