package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armored Kincaller — Creature — Dinosaur {2}{G}, 3/3 (The Lost Caverns of
// Ixalan, #4470):
//
//	"When this creature enters, you may reveal a Dinosaur card from your
//	 hand. If you do or if you control another Dinosaur, you gain 3 life."
//
// A choose_cards prompt over the Dinosaur cards in the controller's hand
// with a floor of zero (a keyed pick, so a table waiting on it is a restore
// point). Revealing one shows it to the table and gains the life; declining
// still gains it when another Dinosaur is on the battlefield under the
// controller's control. The other Dinosaur is looked for as the trigger
// resolves, through the layered types (a changeling counts). With no
// Dinosaur card in hand and no other Dinosaur nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6aad6906-96ec-4994-bfc3-9a76f893bc8c",
		Name:         "Armored Kincaller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Armored Kincaller — you may reveal a Dinosaur card from your hand; gain 3 life if you do or if you control another Dinosaur",
				armoredKincallerEnters),
		},
	})
}

// armoredKincallerDinosaursInHand are the controller's Dinosaur cards in
// hand.
func armoredKincallerDinosaursInHand(ctx *Context) []uuid.UUID {
	p := ctx.Game.PlayerByIDForEffect(ctx.Controller())
	if p == nil || p.Hand == nil {
		return nil
	}
	var out []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.HasSubtypeWord("Dinosaur") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func armoredKincallerControlsAnotherDinosaur(ctx *Context) bool {
	for _, c := range MatchingBattlefield(ctx, And(OfCreatureType("Dinosaur"), YouControl())) {
		if c.InstanceID != ctx.Source() {
			return true
		}
	}
	return false
}

func armoredKincallerEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	dinos := armoredKincallerDinosaursInHand(ctx)
	if len(dinos) == 0 {
		if armoredKincallerControlsAnotherDinosaur(ctx) {
			return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		}
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Armored Kincaller — you may reveal a Dinosaur card from your hand",
		Cards:    dinos,
		Min:      0,
		Max:      1,
		ThenKey:  armoredKincallerReveal,
		LiveZone: game.ZoneHand,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// armoredKincallerReveal is the answer: a revealed card gains the life; a
// declined reveal gains it only when another Dinosaur is controlled.
var armoredKincallerReveal = askedPick("choose-cards/armored-kincaller-reveal", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) > 0 {
		ctx.Game.RevealForEffect(game.RevealSpec{
			Player: r.Chooser,
			Source: ctx.Source(),
			Reason: "Armored Kincaller — reveal a Dinosaur card from your hand",
			Cards:  r.Picked,
		})
	} else if !armoredKincallerControlsAnotherDinosaur(ctx) {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
})
