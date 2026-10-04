package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// In the Darkness Bind Them — Enchantment — Saga {2}{U}{B}{R}:
//
//	"(As this Saga enters and after your draw step, add a lore
//	 counter. Sacrifice after IV.)
//	 I, II, III — Create a 3/3 black Wraith creature token with menace.
//	     The Ring tempts you.
//	 IV — For each opponent, gain control of up to one target creature
//	     that player controls until end of turn. Untap those creatures.
//	     They gain haste until end of turn. The Ring tempts you."
//
// Chapters I to III tempt once the token has entered, so it may be
// chosen as the Ring-bearer (2023-06-16 ruling). Chapter IV is Molten
// Primordial's clauses, one "up to one" per opponent: with no target
// chosen it still tempts, and with every chosen target gone it does
// nothing (CR 608.2b). A stolen creature may be chosen as the
// Ring-bearer.
//
// No simplification.
func init() {
	steal := ChapterTrigger(4, "In the Darkness Bind Them — gain control of a creature of each opponent, then the Ring tempts you", inTheDarknessSteal)
	steal.TargetsFrom = moltenPrimordialClauses
	Register(Spec{
		OracleID:     "c8689a83-cefa-4ef3-8ec1-58b4c64e332b",
		Name:         "In the Darkness Bind Them",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "In the Darkness Bind Them — create a Wraith, then the Ring tempts you", inTheDarknessWraith),
			ChapterTrigger(2, "In the Darkness Bind Them — create a Wraith, then the Ring tempts you", inTheDarknessWraith),
			ChapterTrigger(3, "In the Darkness Bind Them — create a Wraith, then the Ring tempts you", inTheDarknessWraith),
			steal,
		},
	})
}

// inTheDarknessWraith is chapters I to III.
func inTheDarknessWraith(g *game.Game, item *game.StackItem) error {
	tempt := ringTemptsYouNext(item)
	return g.CreateTokensThenForEffect(game.TokenCreation{
		Controller: item.Controller,
		Source:     item.SourceCardID,
		Groups:     []game.TokenGroup{{Template: TokenCard("3/3 black Wraith with menace"), Count: 1}},
	}, func(g *game.Game, _ []uuid.UUID) error { return tempt(g) })
}

// inTheDarknessSteal is chapter IV: each still-legal pick is taken
// until end of turn, untapped and given haste, then the tempt.
func inTheDarknessSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, id := range legalTargetIDs(ctx) {
		if err := (GainControl{Target: id, Duration: DurationUntilEndOfTurn(ctx), Label: "In the Darkness Bind Them — gain control"}).Apply(ctx); err != nil {
			return err
		}
		if err := (UntapTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "In the Darkness Bind Them — haste"}).Apply(ctx); err != nil {
			return err
		}
	}
	return TheRingTemptsYou{}.Apply(ctx)
}
