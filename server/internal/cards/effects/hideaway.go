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
// Build captures the permanent OBJECT the trigger belongs to, so the
// card it hides is linked to this incarnation and no other, and a
// hideaway permanent that left the battlefield before the trigger
// resolved still hides a card (CR 603.10 — the ability exists
// independently of its source) that nothing can ever play.
func Hideaway(name string, n int) game.TriggeredAbility {
	label := name + " — hideaway"
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventETB},
		AppliesTo: Self,
		Key:       label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
			ref := game.ObjectRefOf(*source)
			return game.NewTriggeredItem(source, label, func(g *game.Game, item *game.StackItem) error {
				return hideawayLookAndExile(g, item.Controller, ref, n, name)
			})
		},
	}
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
// one Source's hideaway exiled). It grants Player a free play of it.
//
// A GRANT, not an inline play — ADR 0066's posture on every "you may
// cast it" a resolution offers (cascade, Malcolm): the announce path
// has no frame for a cast collected from inside a resolution
// (CR 608.2g), so the player is handed a permission and plays the card
// with an ordinary action once the ability has resolved. Its shape:
//
//   - Cost "{0}" — "without paying its mana cost". Empty would mean
//     the printed cost.
//   - NOT CastOnly: hideaway says PLAY, and a land may be what was hid.
//     A land then takes the land-play branch, with its own timing and
//     the turn's land drop (CR 305.2b / 305.3) — no timing statement
//     reaches a land play, so a land hidden by a land activated during
//     combat waits for a main phase. Weaker than printed, never
//     stronger.
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
	for _, id := range g.HiddenCardsForEffect(source) {
		g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player: player,
			Zone:   game.ZoneExile,
			Cost:   "{0}",
			Timing: game.TimingFlash,
			Source: source.ID,
			Label:  label,
		})
		// "The exiled card" — one. A second hideaway on the same
		// permanent (Evercoat Ursine) prints "one of them" and is its
		// own clause.
		return
	}
}
