package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Titan of Littjara — Creature — Illusion {4}{U}{U}, 6/6 (Marvel Super
// Heroes Commander, #4455):
//
//	"As this creature enters, choose a creature type.
//	 This creature is the chosen type in addition to its other types.
//	 Whenever this creature enters or attacks, you may draw a card for each
//	 other creature you control that shares a creature type with it. If you
//	 do, discard a card."
//
// The first two sentences are Adaptive Automaton's (ChooseCreatureTypeAs-
// Enters and IsAlsoTheChosenType, layer 4), so the Titan shares the chosen
// type with every creature of it. The trigger is one ability with two
// conditions (WhenThisEntersOrAttacks). The count is read on resolution
// through game.SharesCreatureType, so a changeling shares with everything
// (CR 702.73a), and the "may" is one yes/no that draws that many cards;
// the discard is linked to it (CR 607.2) and happens only on the yes
// branch, after the draw. With no creature to count there is nothing to
// draw, so no question is asked. Should the Titan have left, its
// last-known creature types are compared (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f62250c6-0832-46c8-a800-806589504f5f",
		Name:         "Titan of Littjara",
		Completeness: CompletenessFull,
		AsEnters:     ChooseCreatureTypeAsEnters("Titan of Littjara"),
		Static:       []game.StaticAbility{IsAlsoTheChosenType()},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Titan of Littjara — you may draw a card for each other creature you control that shares a creature type with it; if you do, discard a card",
				titanOfLittjaraMayDraw),
		},
	})
}

// titanOfLittjaraSharers is the number of OTHER creatures the Titan's
// controller controls that share a creature type with the Titan.
func titanOfLittjaraSharers(ctx *Context) int {
	g := ctx.Game
	id := ctx.Source()
	live, onField := g.LookupCardForEffect(id)
	onField = onField && onBattlefield(g, id)
	var lkiTypes []string
	if !onField {
		info, ok := g.LastKnownPermanentForEffect(id)
		if !ok {
			return 0
		}
		lkiTypes = info.Characteristic.Subtypes
	}
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == id || c.Controller != ctx.Controller() || !c.IsCreature() {
			continue
		}
		other := c
		if onField {
			if game.SharesCreatureType(&live, &other) {
				n++
			}
			continue
		}
		if titanOfLittjaraSharesWithTypes(lkiTypes, &other) {
			n++
		}
	}
	return n
}

// titanOfLittjaraSharesWithTypes is SharesCreatureType against a list of
// last-known subtypes.
func titanOfLittjaraSharesWithTypes(types []string, other *game.Card) bool {
	if game.HasAllCreatureTypes(other) {
		for _, t := range types {
			if game.IsCreatureType(t) {
				return true
			}
		}
		return false
	}
	for _, t := range types {
		if !game.IsCreatureType(t) {
			continue
		}
		for _, o := range other.Effective().Subtypes {
			if o == t {
				return true
			}
		}
	}
	return false
}

func titanOfLittjaraMayDraw(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := titanOfLittjaraSharers(ctx)
	if n == 0 {
		return nil
	}
	return MayChoice{
		Question: "Titan of Littjara — draw cards for each creature that shares a type, then discard a card?",
		Draws:    n,
		YesLabel: "Draw",
		NoLabel:  "Decline",
		Then:     titanOfLittjaraAnswer,
	}.Apply(ctx)
}

// titanOfLittjaraAnswer is the "yes": draw one card per sharing creature,
// then discard a card. Keyed (#4291).
var titanOfLittjaraAnswer = MayYes("confirm/titan-of-littjara-loot", func(ctx *Context, _ game.ConfirmAnswer) error {
	return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), titanOfLittjaraSharers(ctx), 1, "")
})
