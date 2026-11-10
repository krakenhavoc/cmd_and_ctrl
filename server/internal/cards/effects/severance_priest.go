package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Severance Priest — Creature — Djinn Cleric {W}{B}{G}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Deathtouch
//	 When this creature enters, target opponent reveals their hand. You
//	 may choose a nonland card from it. If you do, exile that card.
//	 When this creature leaves the battlefield, the exiled card's owner
//	 creates an X/X white Spirit creature token, where X is the mana
//	 value of the exiled card."
//
// The enters trigger is Null Summoner's reveal-and-exile with "you may"
// (ChooseFromRevealedHand, Exile + Linked + Optional), the exile linked to
// the Priest as the object it is (CR 607.2a), so a Priest that left and
// came back is a new object with no claim on the card (CR 400.7). The
// leaves trigger reads the cards exiled with the Priest as it last
// existed (CR 608.2h): their owner, which is not the Priest's controller,
// creates the token, X the card's mana value in exile (an {X} in the cost
// is 0, CR 202.3e). With nothing exiled, whether the Priest's controller
// declined or the Priest left before the enters trigger resolved, the
// leaves trigger does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3a38bf69-511f-4851-870f-8a65641d832e",
		Name:            "Severance Priest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Severance Priest — target opponent reveals their hand, you may choose a nonland card and exile it",
					severancePriestExile),
				TargetPlayer("target opponent", Opponent())),
			On(game.EventLTB, Self, "Severance Priest — the exiled card's owner creates an X/X white Spirit creature token",
				severancePriestSpirit),
		},
	})
}

func severancePriestExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ChooseFromRevealedHand{
		Player:   TargetedPlayer(ctx),
		Filter:   Nonland(),
		Label:    "nonland card",
		Exile:    true,
		Linked:   true,
		Optional: true,
	}.Apply(ctx)
}

// severancePriestSpirit makes, for each card exiled with the Priest, an
// X/X white Spirit for that card's owner, X its mana value.
func severancePriestSpirit(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		x := c.ManaValueWithX(0)
		if err := (CreateToken{
			Controller: c.Owner,
			Template: game.Card{
				Name: "Spirit", TypeLine: "Token Creature — Spirit",
				Power: x, Toughness: x, Colors: []string{"W"},
			},
			N: 1,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
