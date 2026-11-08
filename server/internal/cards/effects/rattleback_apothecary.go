package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rattleback Apothecary — Creature — Gorgon Warlock {2}{B}, 3/2 (Outlaws
// of Thunder Junction, #4468):
//
//	"Deathtouch
//	 Whenever you commit a crime, target creature you control gains your
//	 choice of menace or lifelink until end of turn. (Targeting opponents,
//	 anything they control, and/or cards in their graveyards is a crime.)"
//
// The crime is Overzealous Muscle's annotation (YouCommittedCrime, CR
// 700.13), once per announcement and with no once-each-turn limit. The
// trigger's own target is a creature you control, so it is no crime and
// can't feed itself. The keyword is chosen as the trigger resolves
// (Viper, Cruel Conspirator's PickOption shape, the target riding the
// prompt as data) and granted until end of turn; a target that has gone
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5538d4cf-98eb-4e20-a9c1-42c7cdcad194",
		Name:            "Rattleback Apothecary",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return YouCommittedCrime(ev, source)
			}, "Rattleback Apothecary — target creature you control gains your choice of menace or lifelink until end of turn (crime)",
				rattlebackChoosesAKeyword),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

// rattlebackKeywords are the two keywords offered, in printed order.
var rattlebackKeywords = []string{"menace", "lifelink"}

// rattlebackChoosesAKeyword asks for the keyword as the trigger resolves.
func rattlebackChoosesAKeyword(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	options := make([]game.ChoiceOption, len(rattlebackKeywords))
	for i, kw := range rattlebackKeywords {
		options[i] = game.ChoiceOption{Label: "Gains " + kw + " until end of turn"}
	}
	return PickOption{
		Question: "Rattleback Apothecary — the creature gains your choice of menace or lifelink until end of turn",
		Options:  options,
		ThenKey:  rattlebackKeywordChosen,
		Carry:    []uuid.UUID{id},
	}.Apply(ctx)
}

// rattlebackKeywordChosen grants the chosen keyword to the carried target.
// An on-disk key: never renamed, never reused.
var rattlebackKeywordChosen = OptionPickThen("option-pick/rattleback-apothecary-keyword", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index >= len(rattlebackKeywords) || len(r.Carry) == 0 {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   r.Carry[0],
		Keywords: []string{rattlebackKeywords[r.Index]},
		Label:    "Rattleback Apothecary — gains " + rattlebackKeywords[r.Index] + " until end of turn",
	}.Apply(ctx)
})
