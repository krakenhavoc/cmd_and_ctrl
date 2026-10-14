package effects

import (
	"fmt"
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_from_hand.go — "manifest N cards from your hand" (CR 701.40a),
// asked of several players at once (Kozilek, the Broken Reality) or of
// the controller alone (Scroll of Fate). ADR 0189 decision 3.
//
//	EachManifestsFromHand{Players: …, N: 2, Then: …}
//	  "up to two target players each manifest two cards from their hands.
//	   For each card manifested this way, you draw a card."
//	EachManifestsFromHand{N: 1}
//	  "Manifest a card from your hand."

// EachManifestsFromHand has each of `Players` choose `N` cards from
// their own hand and manifest them: the chosen cards enter the
// battlefield face down as 2/2 creatures under their owner's control,
// all in one entry (game.ManifestFromHandsThenForEffect).
//
// # What each player is asked
//
// Exactly N cards, as one choose_cards prompt tagged
// game.ChooseManifest ("Choose two cards from your hand to manifest").
// A player whose hand holds N cards or fewer is not asked: they
// manifest all of them (the 2024-06-07 Kozilek ruling, "if a targeted
// player has only one card in hand … they manifest that card"). A
// player with an empty hand is skipped, and so is one who has left
// the game.
//
// # Secrecy (CR 101.4)
//
// The questions go up together, in APNAP order, as ONE run
// (game.ChooseCardsRunThenForEffect): nothing moves until every player
// has answered, so no seat learns another's picks first. Asking each
// player in turn would make the picks public one by one and leak them
// to the later choosers. Nobody but the chooser sees the prompt or the
// cards, and the manifested cards are face down for everyone else
// (CR 708.5).
//
// # After the answers
//
// Every pick is manifested in one entry, the players' picks in APNAP
// order, and Then runs once with the permanents that actually
// entered. A card a CR 614 replacement kept off the battlefield is not
// among them, which is what lets "for each card manifested this way,
// draw a card" count only the cards that arrived. A table where no
// hand has a card manifests nothing, and Then sees an empty list.
type EachManifestsFromHand struct {
	// Players are the players who manifest. Nil means the resolving
	// item's controller. They are asked in APNAP order whatever order
	// they are listed in, and a player listed twice is asked once.
	Players func(*Context) []uuid.UUID

	// N is how many cards each player manifests.
	N int

	// Then, when set, runs once after the entry, told the permanents
	// that entered, in APNAP order. It runs with g.mu held, may queue
	// further prompts, and must not capture a *Card or a pointer into a
	// zone (the context it is handed is rebuilt for the resolving
	// item).
	Then func(ctx *Context, manifested []uuid.UUID) error
}

func (e EachManifestsFromHand) Apply(ctx *Context) error {
	g := ctx.Game
	want := []uuid.UUID{ctx.Controller()}
	if e.Players != nil {
		want = e.Players(ctx)
	}
	n := e.N
	source := ctx.Source()

	// Frozen before the run starts and never written after: the run's
	// continuation shares them with every undo snapshot.
	var players []uuid.UUID
	for _, id := range apnapPlayers(g) {
		if slices.Contains(want, id) {
			players = append(players, id)
		}
	}
	auto := make(map[uuid.UUID][]uuid.UUID, len(players))
	var legs []game.ChooseCardsPrompt
	for _, id := range players {
		hand := handCardIDs(g, id)
		switch {
		case len(hand) == 0 || n <= 0:
		case len(hand) <= n:
			auto[id] = hand
		default:
			legs = append(legs, game.ChooseCardsPrompt{
				Chooser:     id,
				FromPlayer:  id,
				Source:      source,
				Question:    fmt.Sprintf("Choose %s from your hand to manifest", cardCount(n)),
				Cards:       hand,
				Min:         n,
				Max:         n,
				Zone:        game.ZoneHand,
				Destination: game.ChooseManifest,
			})
		}
	}
	resolving, then := ctx.Item, e.Then
	_, err := g.ChooseCardsRunThenForEffect(legs, func(g *game.Game, picks game.PromptedPicks) error {
		var all []uuid.UUID
		for _, id := range players {
			if hand, ok := auto[id]; ok {
				all = append(all, hand...)
				continue
			}
			all = append(all, picks.By(id)...)
		}
		finish := func(g *game.Game, manifested []uuid.UUID) error {
			if then == nil {
				return nil
			}
			return then(NewContext(g, resolving), manifested)
		}
		if len(all) == 0 {
			return finish(g, nil)
		}
		return g.ManifestFromHandsThenForEffect(all, finish)
	})
	return err
}

// handCardIDs lists the cards in a player's hand, in hand order.
func handCardIDs(g *game.Game, playerID uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(playerID)
	if p == nil || p.Hand == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(p.Hand.Cards))
	for _, c := range p.Hand.Cards {
		out = append(out, c.InstanceID)
	}
	return out
}

// cardCount spells a small card count the way a prompt reads: "a
// card", "two cards".
func cardCount(n int) string {
	words := []string{"zero", "a", "two", "three", "four", "five", "six", "seven"}
	if n == 1 {
		return "a card"
	}
	if n > 0 && n < len(words) {
		return words[n] + " cards"
	}
	return fmt.Sprintf("%d cards", n)
}
