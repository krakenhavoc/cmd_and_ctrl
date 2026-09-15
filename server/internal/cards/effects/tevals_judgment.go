package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// b16TevalsJudgmentLabel is the stack label the per-turn tally keys
// on — every resolution of this trigger is one mode used up.
const b16TevalsJudgmentLabel = "Teval's Judgment — cards left your graveyard"

// Teval's Judgment — Enchantment {2}{B} (EDHREC rank 1773):
//
//	"Whenever one or more cards leave your graveyard, choose one that
//	 hasn't been chosen this turn —
//	 • Draw a card.
//	 • Create a Treasure token.
//	 • Create a 2/2 black Zombie Druid creature token."
//
// The graveyard-recursion payoff. The condition is
// b16CardLeftYourGraveyard — a move out of a graveyard for a card
// the controller owns, whether an EventZoneMove or a flashback's
// EventCast — and "one or more" is the per-label dedup
// (b12TriggerPendingOrOnStack): a Bojuka Bog on your graveyard emits
// one move per card and fires this once. Each mode is one primitive.
//
// Sandbox simplification, declared: the mode is not chosen — it is
// the first mode in printed order that has not been used this turn,
// Gala Greeters' posture exactly. The modal machinery is cast-time
// only, so a resolution-time pick for a trigger has no prompt;
// "hasn't been chosen this turn" is a count of this trigger's
// resolutions since the turn's upkeep began (b15ResolvedThisTurn).
// The fourth batch and later do nothing, as printed. Weaker than
// printed — the controller cannot take the Zombie before the card —
// never stronger.
func init() {
	Register(Spec{
		OracleID:     "71fc2393-f55c-4b06-897c-fb7d4199b5f5",
		Name:         "Teval's Judgment",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The mode isn't chosen: each turn the first trigger draws a card, the second makes a Treasure, the third makes the Zombie Druid."},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventZoneMove, game.EventCast}, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b16CardLeftYourGraveyard(ev, source, g) && !b12TriggerPendingOrOnStack(g, source, b16TevalsJudgmentLabel)
			}, b16TevalsJudgmentLabel, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				// The resolve event for THIS resolution is already
				// logged, so the count includes it: 1 → first mode.
				switch b15ResolvedThisTurn(g, item.SourceCardID, b16TevalsJudgmentLabel) {
				case 1:
					return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
				case 2:
					return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 1}.Apply(ctx)
				case 3:
					return CreateToken{Controller: item.Controller, Template: b16BlackZombieDruidToken(), N: 1}.Apply(ctx)
				}
				return nil
			}),
		},
	})
}
