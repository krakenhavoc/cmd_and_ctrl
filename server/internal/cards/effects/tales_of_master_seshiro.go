package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tales of Master Seshiro // Seshiro's Living Legacy — a transforming Saga
// (Kamigawa: Neon Dynasty, #5493, S82).
//
// Front face, Enchantment — Saga {4}{G}:
//
//	"I, II — Put a +1/+1 counter on target creature or Vehicle you control.
//	 It gains vigilance until end of turn.
//	 III — Exile this Saga, then return it to the battlefield transformed
//	       under your control."
//
// Back face, Enchantment Creature — Snake Warrior, 5/5: "Vigilance, haste".
//
// A Vehicle that isn't a creature is a legal target and takes the counter
// and the keyword. If the target is gone as the chapter resolves, neither
// half happens (CR 608.2b).
//
// No simplification.
const talesOfMasterSeshiroOracleID = "096fbb51-be73-4edc-a051-770a12c09005"

func init() {
	const name = "Tales of Master Seshiro"
	body := func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		target := FirstLegalBattlefieldTarget(ctx)
		if target == [16]byte{} {
			return nil
		}
		if err := (AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: target, Keywords: []string{"vigilance"}, Label: name + " — vigilance"}.Apply(ctx)
	}
	chapter := func(n int) game.TriggeredAbility {
		return ChapterTriggerTargeting(n, SagaChapterLabel(name, n, "a +1/+1 counter on a creature or Vehicle, and vigilance"),
			TargetPermanent("target creature or Vehicle you control", Or(Creature(), OfSubtype("Vehicle")), YouControl()), body)
	}
	Register(Spec{
		OracleID:     talesOfMasterSeshiroOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			chapter(1), chapter(2),
			ChapterTrigger(3, SagaChapterLabel(name, 3, "exile it, then return it transformed"),
				ChapterExileAndReturnTransformed),
		},
	})
	Register(Spec{
		OracleID:        talesOfMasterSeshiroOracleID + "#1",
		Name:            "Seshiro's Living Legacy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "haste"},
	})
}
