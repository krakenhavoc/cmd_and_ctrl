package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winota, Joiner of Forces — Legendary Creature — Human Warrior {2}{R}{W},
// 4/4 (Final Fantasy, #4462):
//
//	"Whenever a non-Human creature you control attacks, look at the top six
//	 cards of your library. You may put a Human creature card from among
//	 them onto the battlefield tapped and attacking. It gains indestructible
//	 until end of turn. Put the rest of the cards on the bottom of your
//	 library in a random order."
//
// One trigger per non-Human attacker (CR 508.3a; a changeling is a Human, so
// it never triggers it), each resolving on its own look at six. The look is
// private. The pick is a choose_cards prompt over the Human creature cards
// among the six, up to one; the rest go to the bottom in a random order
// first, so the put finds its card still in the library. The put is the
// shared tapped-and-attacking entry: its controller chooses what it attacks
// (CR 508.4), which need not be what the non-Human attacks, and it was never
// declared as an attacker, so "whenever attacks" triggers (this card's
// included) don't see it, and a rule that stops it attacking doesn't stop it
// entering attacking. The new object gains indestructible until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "02598a35-1b11-4abe-89cd-fc44c8c36c15",
		Name:         "Winota, Joiner of Forces",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, winotaNonHumanAttacked,
				"Winota, Joiner of Forces — look at the top six cards; you may put a Human creature card onto the battlefield tapped and attacking",
				winotaLook),
		},
	})
}

// winotaNonHumanAttacked is "a non-Human creature you control attacks".
func winotaNonHumanAttacked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && c.Controller == source.Controller && !c.HasSubtype("Human")
}

// winotaLook looks at six and asks which Human creature card, if any.
func winotaLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	looked := g.LookAtTopOfLibraryForEffect(player, 6)
	if len(looked) == 0 {
		return nil
	}
	var humans []uuid.UUID
	for _, id := range looked {
		if c, ok := g.LookupCardForEffect(id); ok && c.IsCreature() && !IsToken(c) && c.HasSubtype("Human") {
			humans = append(humans, id)
		}
	}
	if len(humans) == 0 {
		return g.PutOnBottomInRandomOrderForEffect(player, game.ZoneLibrary, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Winota, Joiner of Forces — you may put a Human creature card from among them onto the battlefield tapped and attacking (the rest go to the bottom in a random order)",
		Cards:    humans,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneLibrary,
		ThenKey:  winotaPicked,
		Carry:    looked,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// winotaPicked puts the rest on the bottom in a random order, then the pick
// onto the battlefield tapped and attacking. An on-disk key: never renamed,
// never reused.
var winotaPicked = askedPick("choose-cards/winota-joiner-of-forces-human", func(ctx *Context, r game.CardsChosen) error {
	var picked uuid.UUID
	if len(r.Picked) > 0 {
		picked = r.Picked[0]
	}
	var rest []uuid.UUID
	for _, id := range r.Carry {
		if id != picked {
			rest = append(rest, id)
		}
	}
	if len(rest) > 0 {
		if err := ctx.Game.PutOnBottomInRandomOrderForEffect(ctx.Controller(), game.ZoneLibrary, rest); err != nil {
			return err
		}
	}
	if picked == uuid.Nil {
		return nil
	}
	return finB11PutAttackingThen(ctx, picked, game.ZoneLibrary, "Winota, Joiner of Forces", winotaEntered, winotaAttackPicked)
})

// winotaAttackPicked is the answer to "what does it attack?". An on-disk
// key: never renamed, never reused.
var winotaAttackPicked = OptionPickThen("option-pick/winota-joiner-of-forces-attack", func(ctx *Context, r game.OptionPicked) error {
	card, target, err := finB11PickedAttack(r)
	if err != nil {
		return nil
	}
	return finB11PutAttacking(ctx.Game, ctx.Controller(), card, game.ZoneLibrary, target, ctx.Tail(winotaEntered))
})

// winotaEntered gives the new object indestructible until end of turn. An
// on-disk key: never renamed, never reused.
var winotaEntered = EnteredTail("winota/entered-indestructible", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   entered,
		Keywords: []string{"indestructible"},
		Label:    "Winota, Joiner of Forces — indestructible until end of turn",
	}.Apply(ctx)
})
