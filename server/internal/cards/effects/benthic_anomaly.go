package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Benthic Anomaly — Creature — Eldrazi Serpent {6}{U}, 7/8 (Modern Horizons
// 3 Commander, #5948):
//
//	"Devoid (This card has no color.)
//	 When you cast this spell, for each opponent, choose a creature that
//	 player controls. Create a token that's a copy of one of those
//	 creatures, except its power is equal to the total power of those
//	 creatures, its toughness is equal to the total toughness of those
//	 creatures, and it's a colorless Eldrazi creature."
//
// The cast trigger is Ulamog's FromStack shape and resolves before the
// spell. Nothing is targeted: the controller makes every choice as it
// resolves (CR 608.2c), one creature per opponent in turn order, through
// Highcliff Felidar's ChoosePermanents fan-out; an opponent with no
// creature is skipped. When more than one creature was chosen, the
// controller then picks which of them the token copies. The totals are
// read from the chosen creatures as the trigger resolves, their power and
// toughness with every effect applied. The copy takes the original's
// supertypes, abilities and (here) a generic-only copy of its mana cost,
// so it is colourless with the same mana value (gapB45ColorlessEldraziCopy).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c12e2563-4ad4-4637-b9ae-72f23cca31d8",
		Name:            "Benthic Anomaly",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{{
			FromStack: true,
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: Self,
			Key:       "Benthic Anomaly — for each opponent, choose a creature; create a token copy of one of them",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Benthic Anomaly — for each opponent, choose a creature; create a token copy of one of them")
				item.Controller, item.Owner = ev.Actor, ev.Actor
				return item
			},
			Effect: benthicAnomalyChoose,
		}},
	})
}

func benthicAnomalyChoose(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ChoosePermanents{
		Question: "Benthic Anomaly — choose a creature that player controls",
		Of:       opponentsInTurnOrder(g, ctx.Controller()),
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			ids := permanentsControlledByMatching(g, of, func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsCreature() })
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 1, 1
		},
		Keyed: benthicAnomalyChosen,
	}.Apply(ctx)
}

// benthicAnomalyChosen reads the totals and, with one creature chosen,
// copies it; with several, asks which to copy. Keyed (#4290).
var benthicAnomalyChosen = RegisterPermanentsPick("prompt-run/benthic-anomaly-chosen", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		var ids []uuid.UUID
		power, toughness := 0, 0
		for _, id := range picked.Cards() {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || !onBattlefield(ctx.Game, id) {
				continue
			}
			ids = append(ids, id)
			power += c.PowerForComparison()
			toughness += c.CurrentToughness()
		}
		switch len(ids) {
		case 0:
			return nil
		case 1:
			return benthicAnomalyCopy(ctx, ids[0], power, toughness)
		}
		return ChoosePermanents{
			Question: "Benthic Anomaly — choose one of those creatures to copy",
			Candidates: func(*game.Game, uuid.UUID) ([]uuid.UUID, int, int) {
				return ids, 1, 1
			},
			Keyed: benthicAnomalyCopyChosen,
			Nums:  []int{power, toughness},
		}.Apply(ctx)
	},
})

// benthicAnomalyCopyChosen makes the token copy of the one creature the
// controller chose, at the totals the first choice recorded. Keyed (#4290).
var benthicAnomalyCopyChosen = RegisterPermanentsPick("prompt-run/benthic-anomaly-copy", PermanentsDef{
	ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
		cards := picked.Cards()
		if len(cards) == 0 || len(r.Nums) < 2 {
			return nil
		}
		return benthicAnomalyCopy(ctx, cards[0], r.Nums[0], r.Nums[1])
	},
})

func benthicAnomalyCopy(ctx *Context, of uuid.UUID, power, toughness int) error {
	return CreateTokenCopy{
		Controller: ctx.Controller(),
		Copy:       of,
		N:          1,
		Except:     func(tok *game.Card) { gapB45ColorlessEldraziCopy(tok, power, toughness) },
	}.Apply(ctx)
}
