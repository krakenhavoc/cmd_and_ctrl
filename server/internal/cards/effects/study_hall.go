package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Study Hall — Land (Secrets of Strixhaven Commander, #4834):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color. When you spend this mana to cast
//	 your commander, scry X, where X is the number of times it's been cast
//	 from the command zone this game."
//
// Opal Palace's shape with Path of Ancestry's scry. Both mana abilities
// are real; the {1}, {T} one is unrestricted in colour, and carries a
// spend trigger (WhenManaSpent) that fires only when the mana pays for a
// spell that is a commander its caster owns (opalPalaceYourCommander).
//
// X is the commander tax tally of CR 903.8, read as the trigger resolves
// off the spell the mana paid for (Payload[0]) rather than captured: it
// includes the cast being paid for when that cast came from the command
// zone, as the printed "number of times it's been cast from the command
// zone this game" does. A commander cast from anywhere else still scries
// for its earlier command-zone casts; with none, X is 0 and the scry does
// nothing. Mana that never entered the pool was not this card's mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "eb735501-19e7-4910-aa6a-6667fff6f4e5",
		Name:           "Study Hall",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color",
				SpendRiders: []game.ManaSpendRider{WhenManaSpent("Study Hall", game.ManaSpendTrigger{
					Label:     "Study Hall — scry X",
					Condition: opalPalaceYourCommander,
					Effect:    studyHallScry,
				}, game.ManaRestrictCast)},
			},
		},
	})
}

// studyHallScry is the rider's effect: the player who spent the mana
// scries X, X being the number of times the commander it paid for has
// been cast from the command zone (CR 903.8).
func studyHallScry(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cards := ctx.PayloadCards()
	if len(cards) == 0 {
		return nil
	}
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil {
		return nil
	}
	n := p.CommanderCasts[cards[0]]
	if n <= 0 {
		return nil
	}
	return Scry{Player: item.Controller, N: n}.Apply(ctx)
}
