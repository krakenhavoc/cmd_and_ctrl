package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dirgur Focusmage // Braingeyser — a preparation card (CR 722,
// ADR 0090), Secrets of Strixhaven Commander (#4834):
//
//	Dirgur Focusmage {2}{U}, Creature — Djinn Monk 1/4:
//	"Instant and sorcery spells you cast cost {1} less to cast.
//	 Whenever you cast an instant or sorcery spell with mana value 5 or
//	 greater from your hand, this creature becomes prepared."
//	Braingeyser {X}{U}{U}, Sorcery:
//	"Target player draws X cards."
//
// The reduction is Archmage of Runes'. The trigger reads the spell on
// the stack, so its mana value counts the X it was cast with (CR
// 202.3e) and not the reduction, which changes the cost paid and never
// the mana value. A cast from anywhere but the hand does not trigger
// it; that includes the copy of Braingeyser cast from exile. The player
// the copy targets is checked again as it resolves (CR 608.2b).
//
// No simplification.
const dirgurFocusmageOracleID = "655ee0db-1913-464f-bd06-072c6bf446bd"

func init() {
	Register(Spec{
		OracleID:     dirgurFocusmageOracleID,
		Name:         "Dirgur Focusmage",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Instant and sorcery spells you cast cost {1} less to cast.",
				YourSpell(), InstantOrSorcerySpell()),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastFromYourHand(And(Or(Instant(), Sorcery()), ManaValueGE(5))),
				"Dirgur Focusmage — becomes prepared", socB4ThisBecomesPrepared),
		},
	})
	Register(Spec{
		OracleID:     dirgurFocusmageOracleID + "#1",
		Name:         "Braingeyser",
		Completeness: CompletenessFull,
		XMatters:     true,
		Targets:      TargetPlayer("target player"),
		OnResolve:    targetPlayerDrawsX(),
	})
}
