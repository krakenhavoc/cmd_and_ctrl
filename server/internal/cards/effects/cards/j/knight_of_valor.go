package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Valor — Creature — Human Knight {2}{W}, 2/2 (Visions, #5752):
//
//	"Flanking (Whenever a creature without flanking blocks this creature,
//	 the blocking creature gets -1/-1 until end of turn.)
//	 {1}{W}: Each creature without flanking blocking this creature gets
//	 -1/-1 until end of turn. Activate only once each turn."
//
// Flanking is CR 702.25a's trigger (flanking.go). The activated ability
// finds its creatures as it resolves: each creature blocking the Knight
// then that has no flanking, so a creature removed from combat (CR 506.4)
// is out, and a Knight that has left the battlefield is blocked by
// nothing. It can be activated with nothing blocking, and then does
// nothing. The limit is OncePerTurnActivation (CR 602.1b). The ruling:
// "The second ability is not Flanking and is not removed by an effect that
// removes Flanking", so it is its own row, not part of Flanking.
//
// No simplification.
func init() {
	const label = "{1}{W}: Each creature without flanking blocking this creature gets -1/-1 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:        "f2edb543-de33-4eed-892c-98038eec96e2",
		Name:            "Knight of Valor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{effects.RetiredFlankingRow("Knight of Valor")},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Cost:      effects.ManaCost("{1}{W}"),
			Condition: effects.OncePerTurnActivation(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				knight := ctx.Source()
				if !effects.OnBattlefield(g, knight) || ctx.IsNewSourceObject(knight) {
					return nil
				}
				member := ctx.AsGroupMember() // #1463
				for _, c := range g.BattlefieldCardsForEffect() {
					if !c.IsCreature() || !c.IsBlockingAttacker(knight) || game.HasKeyword(&c, game.KeywordFlanking) {
						continue
					}
					if err := (effects.BoostUntilEOT{Target: c.InstanceID, Power: -1, Toughness: -1, Label: "Knight of Valor — -1/-1"}).Apply(member); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}
