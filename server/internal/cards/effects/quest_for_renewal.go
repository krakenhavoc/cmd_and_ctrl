package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quest for Renewal — Enchantment {1}{G} (EDHREC rank 3383):
//
//	"Whenever a creature you control becomes tapped, you may put a
//	 quest counter on this enchantment.
//	 As long as there are four or more quest counters on this
//	 enchantment, untap all creatures you control during each other
//	 player's untap step."
//
// The budget Seedborn Muse. The counter trigger is Beastmaster
// Ascension's shape with Magda's condition: it reads two event kinds,
// because the engine taps an attacker without an EventTapCard — see
// b11DwarfYouControlBecameTapped for why an EventAttack whose
// creature is now tapped is "became tapped", and why a vigilance
// attacker is not. Tapping for mana, convoking and attacking all
// count, as printed, and each is its own yes/no prompt.
//
// Sandbox simplification, declared (weaker than printed): the untap
// happens at the beginning of each other player's UPKEEP, as a
// triggered ability on the stack, rather than during their untap
// step. The engine's untap step is a turn-based action with no hook a
// card can add to (Tangle's gap from the other side), and the upkeep
// is the first moment after it that anything can happen. The four-
// counter condition is checked when the trigger fires and again as
// it resolves. The difference is that the untap can be responded to
// — a tap effect in response leaves the creatures tapped — where the
// printed untap cannot; never stronger.
func init() {
	Register(Spec{
		OracleID:     "cba4ab80-09e8-4868-a082-a9a3bade9571",
		Name:         "Quest for Renewal",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With four or more quest counters, your creatures untap at the beginning of each other player's upkeep (as a trigger that can be responded to) rather than during their untap step."},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventTapCard, game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b32CreatureYouControlBecameTapped(ev, source, g)
				},
				OptionalPrompt: &game.TriggerOptionalPrompt{
					Question: "Quest for Renewal — put a quest counter on it?",
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, "Quest for Renewal — put a quest counter",
						func(g *game.Game, item *game.StackItem) error {
							if !b09SourceStillOnBattlefield(g, item) {
								return nil
							}
							return AddCounter{Target: item.SourceCardID, Kind: "quest", N: 1}.Apply(NewContext(g, item))
						})
				},
			},
			{
				Watches: []game.EventKind{game.EventBeginUpkeep},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return b32AnotherPlayersUpkeepBeganWithQuestCounters(ev, source, 4)
				},
				Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					return game.NewTriggeredItem(source, b32QuestForRenewalUntapLabel, b32UntapAllCreaturesIfQuestCounters(4))
				},
			},
		},
	})
}
