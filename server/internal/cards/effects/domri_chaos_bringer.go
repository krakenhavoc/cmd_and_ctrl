package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Domri, Chaos Bringer — Legendary Planeswalker — Domri {2}{R}{G},
// loyalty 5:
//
//	"+1: Add {R} or {G}. If that mana is spent on a creature spell, it
//	     gains riot. (It enters with your choice of a +1/+1 counter or
//	     haste.)
//	 −3: Look at the top four cards of your library. You may reveal up
//	     to two creature cards from among them and put them into your
//	     hand. Put the rest on the bottom of your library in a random
//	     order.
//	 −8: You get an emblem with "At the beginning of each end step,
//	     create a 4/4 red and green Beast creature token with trample.""
//
// The +1 is a loyalty ability, so it is not a mana ability (CR 605.1a):
// it uses the stack and adds the mana as it resolves. The mana carries a
// spend rider (ADR 0109 §11): spent on a creature spell, it gives the
// SPELL riot, indefinitely, and the spell carries it onto the permanent
// (CR 400.7a), where the entry look-ahead finds it and riot asks its
// question (ADR 0109 §10, CR 702.136a). Both mana of a doubled +1 are
// one production, so one riot.
//
// The −3 is the ordinary look-and-take dig with a ceiling of two, and
// the emblem's trigger fires at every end step, anyone's.
func init() {
	Register(Spec{
		OracleID:        "2a5408ed-8b47-4896-97e4-aa102a4b85c9",
		Name:            "Domri, Chaos Bringer",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"With strict mana off, the game doesn't see which mana you spent, so the creature spell its +1 mana pays for never gets its choice of a +1/+1 counter or haste."},
		StartingLoyalty: 5,
		Emblem: &EmblemSpec{
			Label: "Domri, Chaos Bringer emblem",
			Text:  "At the beginning of each end step, create a 4/4 red and green Beast creature token with trample.",
			Triggered: []game.TriggeredAbility{
				On(game.EventBeginEndStep, AnyPlayer, "Domri, Chaos Bringer emblem — create a 4/4 Beast",
					Do(CreateToken{Template: TokenCard("4/4 red and green Beast with trample"), N: 1})),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Add {R} or {G}. If that mana is spent on a creature spell, it gains riot.",
				Cost:  LoyaltyCost(1),
				Effect: Do(AddMana{Produced: "{R|G}", Riders: []game.ManaSpendRider{
					SpentSpellGains([]string{game.KeywordRiot}, false, ManaRestrictCast, ManaRestrictType("Creature")),
				}}),
			},
			{
				Label: "−3: Look at the top four cards of your library. You may reveal up to two creature cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.",
				Cost:  LoyaltyCost(-3),
				Effect: LookAtTopThenMayTakeToHand(4, Creature(), 2,
					"Domri, Chaos Bringer — reveal up to two creature cards and put them into your hand"),
			},
			{
				Label: "−8: You get an emblem with \"At the beginning of each end step, create a 4/4 red and green Beast creature token with trample.\"",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
