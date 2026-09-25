package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hideaway.go — the catalog's words for hideaway (CR 702.75, ADR 0091).
//
// A hideaway card declares two things, and they are linked abilities
// (CR 607.2a):
//
//	Triggered: []game.TriggeredAbility{Hideaway("Windbrisk Heights", 4)},
//	…and, in its second ability, PlayHiddenCard{Source: …}
//
// The first is CR 702.75a whole: look at the top N, exile one face
// down (a FaceDownHidden exile its controller may look at), the rest on
// the bottom in a random order. The second is "you may play the exiled
// card without paying its mana cost": a free-play grant over the card
// that permanent hid.
//
// The link is an OBJECT reference, {instance, epoch}, so the second
// ability needs the permanent AS IT WAS when the ability went on the
// stack — a Windbrisk Heights destroyed in response still lets its
// ability play the card it hid (the exile is what CR 607.2a reads, not
// the permanent). HiddenRefOfActivation reads it off an activated
// ability's item; a triggered ability captures game.ObjectRefOf(source)
// in its Build.

// Hideaway is "Hideaway N" (CR 702.75a) — the permanent's own ETB
// trigger. `name` labels the stack item and the prompt.
//
// Build stamps the permanent OBJECT the trigger belongs to onto
// Params (ADR 0041 P9, #1497, tier 4), so the card it hides is linked
// to this incarnation and no other, and a hideaway permanent that
// left the battlefield before the trigger resolved still hides a card
// (CR 603.10 — the ability exists independently of its source) that
// nothing can ever play. Effect reads the object and the printed N
// and name back off the item rather than a captured value, so a
// hideaway trigger waiting on the stack is a restore point.
func Hideaway(name string, n int) game.TriggeredAbility {
	label := name + " — hideaway"
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Key:       label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Object = game.ObjectRef(game.ObjectRefOf(*source))
			item.Params.Amount = n
			item.Params.Name = name
			return item
		},
		Effect: hideawayEffect,
	}
}

// hideawayEffect is Hideaway's resolution: the look-and-exile, over
// the permanent object and printed N and name Build stamped onto
// Params.
func hideawayEffect(g *game.Game, item *game.StackItem) error {
	ref := game.PermissionCardRef(item.Params.Object)
	return hideawayLookAndExile(g, item.Controller, ref, item.Params.Amount, item.Params.Name)
}

// hideawayLookAndExile is CR 702.75a's two sentences. "Look at", not
// reveal: only the controller becomes a knower of the N. "Exile one of
// them" is not optional, so the prompt's floor is one, and a library
// that shows exactly one card needs no prompt at all. "The rest" goes
// to the bottom in a RANDOM order through the game's keyed RNG — after
// the exile, in the continuation, so it is the cards that are still in
// the library that go, and a commander paused on CR 903.9 on its way
// to exile is not among them.
//
// Captures only scalars and the object reference, so an undo replays
// it against the restored game.
func hideawayLookAndExile(g *game.Game, controller uuid.UUID, source game.PermissionCardRef, n int, name string) error {
	looked := g.LookAtTopOfLibraryForEffect(controller, n)
	if len(looked) == 0 {
		return nil
	}
	finish := func(g *game.Game, picked []uuid.UUID) error {
		var hidErr error
		if len(picked) > 0 {
			_, hidErr = g.ExileHiddenForEffect(source, controller, picked[0])
		}
		rest := make([]uuid.UUID, 0, len(looked))
		for _, id := range cardsStillInALibrary(g, looked) {
			if len(picked) == 0 || id != picked[0] {
				rest = append(rest, id)
			}
		}
		if err := g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneLibrary, rest); err != nil && hidErr == nil {
			hidErr = err
		}
		return hidErr
	}
	if len(looked) == 1 {
		return finish(g, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  controller,
		Source:   source.ID,
		Question: name + " — exile one of them face down",
		Cards:    looked,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneLibrary,
		Then:     finish,
	})
	return nil
}

// HiddenRefOfActivation is the hideaway permanent an ACTIVATED linked
// ability belongs to, as the object it was when the ability was
// activated: the item's source and the CR 400.7 epoch the activation
// stamped (StackItem.SourceEpoch).
func HiddenRefOfActivation(item *game.StackItem) game.PermissionCardRef {
	return game.PermissionCardRef{ID: item.SourceCardID, Epoch: item.SourceEpoch}
}

// PlayHiddenCard is hideaway's payoff: "you may play the exiled card
// without paying its mana cost" (CR 607.2a — "the exiled card" is the
// one Source's hideaway exiled). Two shapes, by what was hidden:
//
// A LAND is played NOW, as part of the resolution (CR 608.2g, ADR 0091's
// 2026-09-23 amendment): a "you may" prompt, then
// PlayLandDuringResolutionForEffect — a land play that spends a land
// drop (CR 305.2a) and needs no main phase and no empty stack. With no
// land drop left, or on another player's turn, the instruction is
// ignored (CR 305.2b / 305.3): nobody is asked and the land stays
// hidden.
//
// A SPELL is a GRANT, not an inline cast — ADR 0066's posture on every
// "you may cast it" a resolution offers (cascade, Malcolm): the
// announce path has no frame for a cast collected from inside a
// resolution, so the player is handed a permission and casts the card
// with an ordinary action once the ability has resolved. Its shape:
//
//   - Cost "{0}" — "without paying its mana cost". Empty would mean
//     the printed cost.
//   - NOT CastOnly: hideaway says PLAY. (A land never reaches the grant
//     — see above — but a permission that said "cast" would be the
//     wrong statement of the card.)
//   - TimingFlash — "as part of the resolution" ignores the card's own
//     timing (CR 608.2g), so a sorcery hidden by an instant-speed
//     activation is castable now rather than never.
//   - Until end of turn, the zero Duration — the Malcolm window. What
//     that adds over printed is a response window before the cast and
//     the choice of when in the turn to take it; declared in ADR 0091.
//
// Nothing hidden (already played, or this is a hideaway permanent that
// was bounced and replayed), no grant. The card stays face down while
// it waits: CR 406.3a turns it face up only as it is played.
type PlayHiddenCard struct {
	Source game.PermissionCardRef
	// Player may play it. Zero means the resolving item's controller.
	Player uuid.UUID
	Label  string
}

func (p PlayHiddenCard) Apply(ctx *Context) error {
	player := p.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	grantHiddenPlay(ctx.Game, player, p.Source, p.Label)
	return nil
}

// grantHiddenPlay is PlayHiddenCard's body with nothing but scalars in
// reach, for a continuation that must not hold a stack item across a
// pause (an undo replays it against the restored game).
func grantHiddenPlay(g *game.Game, player uuid.UUID, source game.PermissionCardRef, label string) {
	hidden := g.HiddenCardsForEffect(source)
	if len(hidden) == 0 {
		return
	}
	// "The exiled card" — one. A second hideaway on the same permanent
	// (Evercoat Ursine) prints "one of them" and is its own clause.
	id := hidden[0]
	if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
		// A hidden LAND is played now, as part of this resolution
		// (CR 608.2g, ADR 0091's 2026-09-23 amendment): a land play
		// that spends a land drop (CR 305.2a) and needs no main phase
		// and no empty stack. A player with no land drop left, or on
		// someone else's turn, is not asked — CR 305.2b and 305.3 say
		// the instruction is ignored — and the land stays hidden for a
		// later activation.
		if !g.CanPlayLandDuringResolutionForEffect(player, id) {
			return
		}
		_ = g.QueueMayCastForEffect(player, source.ID, id,
			label+" — play "+c.Name+"?",
			func(g *game.Game) error {
				_, err := g.PlayLandDuringResolutionForEffect(player, id)
				return err
			}, nil)
		return
	}
	g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
		Player: player,
		Zone:   game.ZoneExile,
		Cost:   "{0}",
		Timing: game.TimingFlash,
		Source: source.ID,
		Label:  label,
	})
}
