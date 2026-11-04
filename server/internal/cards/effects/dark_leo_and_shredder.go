package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Leo & Shredder — Legendary Creature — Mutant Ninja Turtle Human
// {W}{B}, 1/3 (TMT, #4457):
//
//	"Sneak {W}{B}
//	 Attacking Ninjas you control have deathtouch.
//	 Whenever Dark Leo & Shredder deal combat damage to a player, create
//	 a 1/1 black Ninja creature token. Then if you control five or more
//	 Ninjas, that player loses half their life, rounded up."
//
// Sneak is Sneak(cost) (CR 702.190). The deathtouch grant is Ohran
// Frostfang's attacking-creatures static narrowed to Ninjas, the subtype
// read through the source's text the way Blight Mound reads Pests. The
// trigger makes the token first, then counts Ninjas (the new token is
// one), and the damaged player loses half their life as it is then,
// rounded up (CR 107.1a).
//
// No simplification.
func init() {
	deathtouch := AttackingCreaturesHave("Attacking Ninjas you control have deathtouch.", nil, "deathtouch")
	base := deathtouch.AppliesTo
	deathtouch.AppliesTo = func(target *game.Card, g *game.Game, source *game.Card) bool {
		return base(target, g, source) && target.HasSubtype(source.TextSubtype("Ninja"))
	}
	Register(Spec{
		OracleID:         "c3ea6af5-48b9-45d2-9576-46cacb2db5b0",
		Name:             "Dark Leo & Shredder",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{W}{B}")},
		Static:           []game.StaticAbility{deathtouch},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Dark Leo & Shredder — create a 1/1 black Ninja; then with five or more Ninjas, that player loses half their life",
				darkLeoAndShredderNinja),
		},
	})
}

// darkLeoAndShredderNinja is the combat-damage trigger's body.
func darkLeoAndShredderNinja(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CreateToken{Template: TokenCard("1/1 black Ninja"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if tmtB11NinjasYouControl(ctx, item.Controller) < 5 {
		return nil
	}
	return triggeringTargetLosesHalfTheirLife(g, item)
}
