package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bortuk Bonerattle — Legendary Creature — Troll Shaman {4}{B}{G}, 4/4:
//
//	"Domain — When Bortuk Bonerattle enters, if you cast it, choose target
//	 creature card in your graveyard. Return that card to the battlefield
//	 if its mana value is less than or equal to the number of basic land
//	 types among lands you control. Otherwise, put it into your hand."
//
// Cresting Mosasaurus's "if you cast it" (the permanent came from the stack
// and its caster controls it, ADR 0104) on an enters trigger that targets a
// creature card in the controller's graveyard (Altanak's graveyard target).
// Both the mana value and the domain count are read as the trigger
// resolves, so a land that arrived or left in response counts; the card is
// then returned to the battlefield or put into the hand. A target that has
// left the graveyard is not returned (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d965e8c2-34b1-42eb-9647-59951d9fc06b",
		Name:         "Bortuk Bonerattle",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, crestingMosasaurusYouCastIt),
					"Bortuk Bonerattle — return target creature card from your graveyard to the battlefield or your hand",
					bortukBonerattleReturn),
				TargetCardInGraveyard("target creature card in your graveyard", Creature(), YouOwn())),
		},
	})
}

func bortukBonerattleReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		dest := game.ZoneHand
		if c.ManaValue() <= DomainCount(g, ctx.Controller()) {
			dest = game.ZoneBattlefield
		}
		return ReturnFromGraveyard{Target: t.ID, Dest: dest}.Apply(ctx)
	}
	return nil
}
