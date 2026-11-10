package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aurora Awakener — Creature — Giant Druid {6}{G}, 7/7 (Lorwyn Eclipsed,
// #4458):
//
//	"Trample
//	 Vivid — When this creature enters, reveal cards from the top of your
//	 library until you reveal X permanent cards, where X is the number of
//	 colors among permanents you control. Put any number of those permanent
//	 cards onto the battlefield, then put the rest of the revealed cards on
//	 the bottom of your library in a random order."
//
// Vivid is an ability word with no rules meaning (CR 207.2c). X is counted
// once, as the trigger resolves (the ruling), through Explosive Prodigy's
// colorsAmongPermanentsYouControl. Cards are revealed one at a time off the top
// until X of them are permanent cards (a creature, artifact, enchantment,
// land, planeswalker or battle card; never a token), or the library runs
// out, in which case it reveals what it has (Mass Polymorph's reveal). Every
// revealed card is shown to the table. You then choose any number of the
// permanent cards, none included, and they enter together under your
// control (CR 603.6a); everything else revealed, the cards you did not put
// and the nonpermanent cards, goes to the bottom in a random order
// (PutRestOnBottomInRandomOrderTail). An Aura put onto the battlefield this
// way is attached as it enters, to anything its enchant ability allows,
// not to another card entering with it (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cf9d13e2-bdcc-4c66-95ad-654ef795dbd3",
		Name:            "Aurora Awakener",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aurora Awakener — reveal until X permanent cards, put any number onto the battlefield, the rest on the bottom (vivid)",
				auroraAwakenerReveal),
		},
	})
}

// stdB66IsPermanentCard is the reveal's count and the put's filter: a
// card that is a permanent card (CR 110.4), and never a token (CR 108.2).
func stdB66IsPermanentCard(c game.Card) bool {
	return !c.IsToken() && (c.IsCreature() || c.IsArtifact() || c.IsEnchantment() || c.IsLand() || c.IsPlaneswalker() || c.IsBattle())
}

func auroraAwakenerReveal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	want := len(colorsAmongPermanentsYouControl(g, player))
	p := ctx.PlayerByID(player)
	var run []uuid.UUID
	hit := map[uuid.UUID]bool{}
	if p != nil && p.Library != nil && want > 0 {
		for i := len(p.Library.Cards) - 1; i >= 0 && len(hit) < want; i-- {
			c := p.Library.Cards[i]
			run = append(run, c.InstanceID)
			if stdB66IsPermanentCard(c) {
				hit[c.InstanceID] = true
			}
		}
	}
	if len(run) == 0 {
		return nil
	}
	g.RevealForEffect(game.RevealSpec{
		Player: player,
		Source: ctx.Source(),
		Reason: "Aurora Awakener — reveal until X permanent cards",
		Cards:  run,
	})
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit[c.InstanceID] },
		Max:      0,
		Optional: true,
		Label:    "Aurora Awakener — put any number of those permanent cards onto the battlefield",
		ThenTail: PutRestOnBottomInRandomOrderTail,
	}.Apply(ctx)
}
