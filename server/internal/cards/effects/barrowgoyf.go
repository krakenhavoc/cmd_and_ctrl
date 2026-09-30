package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barrowgoyf — Creature — Lhurgoyf {2}{B}, */1+*:
//
//	"Deathtouch, lifelink
//	 Barrowgoyf's power is equal to the number of card types among
//	 cards in all graveyards and its toughness is equal to that number
//	 plus 1.
//	 Whenever this creature deals combat damage to a player, you may
//	 mill that many cards. If you do, you may put a creature card from
//	 among them into your hand."
//
// Tarmogoyf's characteristic-defining ability, printed word for word,
// so it is the same layer 7a self-only static reading the same
// game.DistinctCardTypesInAllGraveyards — power N, toughness N+1.
// Deathtouch and lifelink ride PrintedKeywords. The printed P/T in
// the corner is "*"/"1+*"; the engine carries numbers and the CDA
// overrides them at every recompute, so what the wire ships is the
// computed pair.
//
// "That many" is the combat damage the event carried, captured in
// Build the way Screaming Nemesis captures its own — the number is a
// fact about the damage that was dealt, not about the board when the
// trigger resolves, so a Barrowgoyf that grew or shrank in response
// still mills what it hit for. A deathtouch-and-lifelink 4/5 that
// connects mills four and hands back a creature, which is the card.
//
// Both "may"s are asked DURING the resolution rather than before the
// ability goes on the stack, which is where the printed card asks
// them: the mill decision is a MayChoice, and the creature it turns
// up is a choose-cards prompt inside the mill's continuation. That
// ordering is not cosmetic — the second question is only worth
// answering once the first one's cards have landed, and the cards it
// offers are the ones that landed.
//
// The size is current at once, including after its own mill: the
// layer cache is invalidated whenever a card enters or leaves any
// graveyard (game/layer_listener.go, #1117's graveyard half), not
// only on battlefield motion.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "74c0164c-130f-4572-9066-626c77f6e2ff",
		Name:            "Barrowgoyf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "lifelink"},
		Static: []game.StaticAbility{{
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := game.DistinctCardTypesInAllGraveyards(g)
				c.Power = n
				c.Toughness = n + 1
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: ThisDealtCombatDamageToAPlayer,
			Key:       "Barrowgoyf — you may mill that many cards",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Barrowgoyf — you may mill that many cards")
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return barrowgoyfMill(item.Params.Amount)(g, item)
			},
		}},
	})
}

// barrowgoyfMill is the combat-damage trigger's resolution: "you may
// mill `n` cards. If you do, you may put a creature card from among
// them into your hand."
//
// `n` is the damage the event carried, bound when the ability was
// built. Zero damage cannot reach here — combatDamageToPlayerBy
// already refuses an amount of zero or less — but the guard stays,
// because a mill of nothing would still open a pointless prompt.
func barrowgoyfMill(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if n <= 0 {
			return nil
		}
		return MayChoice{
			Question: "Barrowgoyf — mill " + strconv.Itoa(n) + " cards?",
			OnYes: func(ctx *Context) error {
				return MillToZone{
					N: n,
					Then: func(ctx *Context, milled []uuid.UUID) error {
						return mayTakeOneFromAmongThem(ctx, milled, Creature(),
							"Barrowgoyf — you may put a creature card from among them into your hand")
					},
				}.Apply(ctx)
			},
		}.Apply(NewContext(g, item))
	}
}
