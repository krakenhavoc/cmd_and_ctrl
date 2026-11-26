package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b313_helpers.go — shared pieces of OLD batch 313, Battle for
// Zendikar (S82, #5547): rally Allies whose effect is not a team grant, and
// Munda's look-and-order.

// S82OldB313Rally is "Rally — Whenever this creature or another Ally you
// control enters, <effect>": one trigger per entering Ally (bfzB1RallyWhen).
func S82OldB313Rally(label string, effect Effect) game.TriggeredAbility {
	return bfzB1Rally(label, effect)
}

// S82OldB313MundaLook is Munda's "you may look at the top four cards of your
// library. If you do, reveal any number of Ally cards from among them, then
// put those cards on top of your library in any order and the rest on the
// bottom in any order." The "may" is the trigger's (Optional); this is the
// "if you do".
func S82OldB313MundaLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player, source := ctx.Controller(), ctx.Source()
	looked := g.LookAtTopOfLibraryForEffect(player, 4)
	if len(looked) == 0 {
		return nil
	}
	var allies []uuid.UUID
	ally := ctx.TextSubtype("Ally")
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsToken() && c.HasSubtype(ally) {
			allies = append(allies, id)
		}
	}
	if len(allies) == 0 {
		return s82OldB313MundaPlace(g, source, player, looked, nil)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   source,
		Question: "Munda, Ambush Leader — reveal any number of Ally cards from among them",
		Cards:    allies,
		Min:      0,
		Max:      len(allies),
		LiveZone: game.ZoneLibrary,
		ThenKey:  s82OldB313MundaPick,
		Carry:    looked,
	})
	return nil
}

var s82OldB313MundaPick = game.RegisterChooseCardsThen("choose-cards/s82-old-b313-munda-reveal-allies",
	func(g *game.Game, r game.CardsChosen) error {
		var picked []uuid.UUID
		if r.Answered {
			picked = r.Picked
		}
		if len(picked) > 0 {
			g.RevealForEffect(game.RevealSpec{
				Player: r.Chooser,
				Source: r.Source,
				Cards:  picked,
				Reason: "Munda, Ambush Leader — reveal any number of Ally cards from among them",
			})
		}
		return s82OldB313MundaPlace(g, r.Source, r.Chooser, r.Carry, picked)
	})

// s82OldB313MundaPlace puts the rest of the look on the bottom in the order
// the player chooses, then the revealed cards on top in the order they
// choose. The two lanes do not touch, so the order they are asked in does
// not matter.
func s82OldB313MundaPlace(g *game.Game, source, player uuid.UUID, looked, revealed []uuid.UUID) error {
	isRevealed := map[uuid.UUID]bool{}
	for _, id := range revealed {
		isRevealed[id] = true
	}
	var rest []uuid.UUID
	for _, id := range looked {
		if !isRevealed[id] {
			rest = append(rest, id)
		}
	}
	tail := game.KeyedTail{Ref: s82OldB313MundaTop, Player: player, Source: source, Carry: revealed}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   player,
		Source:    source,
		Cards:     rest,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceBottom,
		Reason:    "Munda, Ambush Leader — put the rest on the bottom of your library in any order",
		ThenTail:  tail,
	})
}

var s82OldB313MundaTop = game.RegisterTail("then/s82-old-b313-munda-top", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	return g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
		Chooser:   o.Tail.Player,
		Source:    o.Tail.Source,
		Cards:     o.Tail.Carry,
		From:      game.ZoneLibrary,
		Placement: game.LibraryPlaceTop,
		Reason:    "Munda, Ambush Leader — put those cards on top of your library in any order",
	})
})
