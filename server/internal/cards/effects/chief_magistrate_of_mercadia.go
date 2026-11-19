package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chief Magistrate of Mercadia — Legendary Creature — Human Noble
// {4}{R}{W}, 3/3 (Mystery Booster Commander Edition):
//
//	"When Chief Magistrate of Mercadia enters, you become the monarch.
//	 At the beginning of your upkeep, create a 2/1 red Goblin creature
//	 token with haste. Then if you're the monarch, for each creature token
//	 you control, create a token that's a copy of it."
//
// Ocelot Pride's shape: the Goblin is made first, and "then if you're the
// monarch" is a clause of the effect read after it (not an intervening
// if). The set copied is every creature token the controller controls,
// taken after the Goblin exists (it is copied too) and before any copy is
// made, so a copy is not itself copied; the copies are created together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b4544d8-0579-4621-8660-4adb3b2a90be",
		Name:         "Chief Magistrate of Mercadia",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Chief Magistrate of Mercadia — you become the monarch", Do(BecomeTheMonarch{})),
			AtYourUpkeep("Chief Magistrate of Mercadia — create a 2/1 Goblin with haste; if you're the monarch, copy each creature token you control",
				chiefMagistrateUpkeep),
		},
	})
}

func chiefMagistrateUpkeep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if err := (CreateToken{Controller: item.Controller, Template: TokenCard("2/1 red Goblin with haste"), N: 1}).Apply(ctx); err != nil {
		return err
	}
	if g.Monarch != item.Controller {
		return nil
	}
	var tokens []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && IsToken(c) && c.IsCreature() {
			tokens = append(tokens, c.InstanceID)
		}
	}
	return CreateTokenCopiesTogether{Controller: item.Controller, Copies: tokens}.Apply(ctx)
}
