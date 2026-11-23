package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of New Benalia — Creature — Human Soldier {1}{W}, 2/2
// (Dominaria United, #5487, seam #5658):
//
//	"Enlist (As this creature attacks, you may tap a nonattacking creature
//	 you control without summoning sickness. When you do, add its power to
//	 this creature's until end of turn.)
//	 Whenever this creature enlists a creature, scry 2.
//	 Discard a card: This creature gains indestructible until end of turn.
//	 Tap it."
//
// Enlist is the engine's keyword (CR 702.154, game/enlist.go): the creature
// to tap is chosen with the attack declaration, tapped as the declaration
// locks in, and the linked trigger adds its power as it resolves.
// "Enlists a creature" is the cost being paid (CR 702.154c): the scry
// triggers on the same event as the linked +X/+0 (WheneverThisEnlists) and
// goes on the stack with it and the attack triggers, in the order the
// player chooses. The activated ability is Iron-Shield Elf's: the discard
// is a cost with no {T}, so a tapped or summoning-sick Guardian can
// activate it, and "tap it" does nothing to one already tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "33747a34-79cc-446f-8718-b499bd497687",
		Name:            "Guardian of New Benalia",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordEnlist},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisEnlists("Guardian of New Benalia — scry 2", effects.Do(effects.Scry{N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: This creature gains indestructible until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.DiscardACard(),
			Effect:  guardianOfNewBenaliaShield,
		}},
	})
}

// guardianOfNewBenaliaShield is "this creature gains indestructible until
// end of turn. Tap it." A Guardian that has left does nothing.
func guardianOfNewBenaliaShield(g *game.Game, item *game.StackItem) error {
	if !effects.B09SourceStillOnBattlefield(g, item) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	grant := effects.GrantKeywordUntilEOT{
		Target:   item.SourceCardID,
		Keywords: []string{"indestructible"},
		Label:    "Guardian of New Benalia — indestructible",
	}
	if err := grant.Apply(ctx); err != nil {
		return err
	}
	return effects.TapTarget{Target: item.SourceCardID}.Apply(ctx)
}
