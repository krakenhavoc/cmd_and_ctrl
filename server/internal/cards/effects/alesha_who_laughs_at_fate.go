package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alesha, Who Laughs at Fate — Legendary Creature — Human Warrior {1}{B}{R},
// 2/2 (Foundations, #4465):
//
//	"First strike
//	 Whenever Alesha attacks, put a +1/+1 counter on it.
//	 Raid — At the beginning of your end step, if you attacked this turn,
//	 return target creature card with mana value less than or equal to
//	 Alesha's power from your graveyard to the battlefield."
//
// The attack trigger is Gau, Feral Youth's counter on itself. The raid
// trigger is Fire Nation Engineer's end-step shape (an intervening "if",
// CR 603.4: checked as it triggers and again as it resolves; any attack
// counts) with Winter Soldier's target: the mana-value bound is data on
// the clause (RelativeToSource, ManaValueNoGreaterThanPower), judged
// against Alesha's power as the target is chosen and again as the ability
// resolves (CR 608.2b), and against her last-known power once she has
// left (CR 608.2h). The card returns under its owner's control, yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "17ce2c17-b69b-4ebf-9c3c-c64e771af101",
		Name:            "Alesha, Who Laughs at Fate",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Alesha, Who Laughs at Fate — put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				}),
			Targeting(On(game.EventBeginEndStep, AllOf(ByYou, tmcB6YouAttackedThisTurn),
				"Alesha, Who Laughs at Fate — Raid: return target creature card with mana value less than or equal to her power from your graveyard to the battlefield",
				Do(If{Cond: tmcB6CondYouAttacked{}, Then: []Applier{
					Instruction(func(ctx *Context) error {
						return returnFirstLegalGraveyardTargetToBattlefield(ctx.Game, ctx.Item)
					}),
				}})),
				RelativeToSource(
					TargetCardInGraveyard("target creature card with mana value less than or equal to Alesha's power from your graveyard",
						Creature(), YouOwn()),
					ManaValueNoGreaterThanPower())),
		},
	})
}
