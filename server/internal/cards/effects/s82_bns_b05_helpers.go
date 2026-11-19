package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b05_helpers.go — card-side helpers for the Special Guests batch
// (S82, BNS batch 5, #5929): Champion.

// champion is the keyword "Champion a <subtype>" (CR 702.72a): "When this
// permanent enters, sacrifice it unless you exile another <subtype> you
// control" and "When this permanent leaves the battlefield, return the
// exiled card to the battlefield under its owner's control."
//
// The two abilities are linked (CR 702.72b) and ride one catalog row, the
// way Oblivion Ring's do: the enters trigger exiles the chosen permanent
// until the champion leaves the battlefield (ExileUntil, CR 610.3), so the
// return is the engine's record of that exile, not a second trigger that
// has to find the card again. The choice is made as the trigger resolves
// (a choose_cards prompt naming one permanent or none; none sacrifices the
// champion, as does having no other <subtype> to name). A champion that has
// already left the battlefield by then exiles nothing (a permanent exile
// nobody would choose) and is not sacrificed again (CR 400.7).
//
// `after`, when set, is the keyed continuation of an ability that triggers
// on the championing ("When a Faerie is championed with this creature",
// Mistbind Clique's tap): it runs only when a card was really exiled by
// this champion ability (CR 702.72c), told the exiled card as its IDs.
//
// Each card registers its own pick key (and `after` key): on-disk
// identities, never renamed or reused.
type champion struct {
	subtype  string
	question string
	label    string
	pick     game.ChooseCardsThen
	after    game.TailRef
}

// newChampion registers the pick under pickKey. Call it once, from a
// package-level var. `after` is the zero TailRef for a plain Champion.
func newChampion(cardName, subtype, pickKey string, after game.TailRef) *champion {
	c := &champion{
		subtype:  subtype,
		question: cardName + " — exile another " + subtype + " you control (or sacrifice " + cardName + ")",
		label:    cardName + " — the championed card returns when " + cardName + " leaves the battlefield",
		after:    after,
	}
	c.pick = askedPick(pickKey, c.picked)
	return c
}

// trigger is the enters trigger, labelled with the card's text.
func (c *champion) trigger(cardName string) game.TriggeredAbility {
	return WhenThisEnters(cardName+" — champion a "+c.subtype+": sacrifice it unless you exile another "+c.subtype+" you control", c.enters)
}

func (c *champion) enters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var candidates []uuid.UUID
	for _, p := range g.BattlefieldCardsForEffect() {
		if p.Controller == ctx.Controller() && p.InstanceID != item.SourceCardID && p.HasSubtypeWord(c.subtype) {
			candidates = append(candidates, p.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return SacrificePermanent{Target: item.SourceCardID}.Apply(ctx)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     item.SourceCardID,
		Question:   c.question,
		Cards:      candidates,
		Min:        0,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    c.pick,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// picked exiles the permanent named until the champion leaves, or
// sacrifices the champion when none was named or the one named is gone.
func (c *champion) picked(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 1 && onBattlefield(ctx.Game, r.Picked[0]) {
		e := ExileUntil{Target: r.Picked[0], ThisLeaves: true, Label: c.label}
		if c.after.Key() != "" {
			e.ThenTail = ctx.ItemTail(c.after)
		}
		return e.Apply(ctx)
	}
	return SacrificePermanent{Target: ctx.Source()}.Apply(ctx)
}
