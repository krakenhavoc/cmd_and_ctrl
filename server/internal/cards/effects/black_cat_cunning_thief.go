package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Black Cat, Cunning Thief — Legendary Creature — Human Rogue Villain
// {3}{B}{B}, 2/3 (Marvel's Spider-Man):
//
//	"When Black Cat enters, look at the top nine cards of target opponent's
//	 library, exile two of them face down, then put the rest on the bottom
//	 of their library in a random order. You may play the exiled cards for
//	 as long as they remain exiled. Mana of any type can be spent to cast
//	 spells this way."
//
// Vivien, Champion of the Wilds' look-and-exile with Gonti's grant
// (gontiExileTopForPlay's permission): the looker (Black Cat's controller)
// is the only knower of the nine cards (CR 701.20, "look at" is not
// "reveal"), picks two, and they are exiled face down, playable by that
// player alone for as long as each stays in exile, lands included, spending
// mana of any type. The rest go to the bottom of the opponent's library in
// a random order. A library with two or fewer cards exiles them all with no
// choice to make. The pick is keyed, so a table waiting on it restores.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5cdb413-2b72-4dc6-b93d-2532ac2ae568",
		Name:         "Black Cat, Cunning Thief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Black Cat — look at the top nine cards of target opponent's library, exile two face down, put the rest on the bottom in a random order", blackCatLook),
				TargetPlayer("target opponent", Opponent()),
			),
		},
	})
}

func blackCatLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
			break
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	controller, source := item.Controller, item.SourceCardID
	looked := g.LookAtTopOfPlayersLibraryForEffect(controller, opp, 9)
	if len(looked) == 0 {
		return nil
	}
	if len(looked) <= 2 {
		return blackCatFinish(g, controller, source, looked, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    controller,
		FromPlayer: opp,
		Source:     source,
		Question:   "Black Cat — exile two of them face down",
		Cards:      looked,
		Min:        2,
		Max:        2,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    blackCatPicked,
		Carry:      looked,
	})
	return nil
}

// blackCatPicked is the two exiled cards chosen: the carry is the nine looked
// at. An on-disk key: never renamed, never reused.
var blackCatPicked = game.RegisterChooseCardsThen("choose-cards/spm-black-cat-exile", func(g *game.Game, r game.CardsChosen) error {
	return blackCatFinish(g, r.Chooser, r.Source, r.Carry, r.Picked)
})

// blackCatFinish exiles the picks face down with the any-type play grant, then
// puts the rest still in the library on the bottom in a random order.
func blackCatFinish(g *game.Game, controller, source uuid.UUID, looked, picked []uuid.UUID) error {
	var firstErr error
	if len(picked) > 0 {
		firstErr = g.ExileCardsFaceDownWithPermissionForEffect(picked, controller, game.CastPermission{
			AnyColor:   true,
			AnyType:    true,
			Duration:   game.WhileInZoneDuration(),
			Source:     source,
			SourceName: "Black Cat, Cunning Thief",
			Label:      "You may play the exiled cards for as long as they remain exiled; mana of any type can be spent to cast spells this way",
		})
	}
	took := make(map[uuid.UUID]bool, len(picked))
	for _, id := range picked {
		took[id] = true
	}
	var rest []uuid.UUID
	for _, id := range cardsStillInALibrary(g, looked) {
		if !took[id] {
			rest = append(rest, id)
		}
	}
	if err := g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneLibrary, rest); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}
