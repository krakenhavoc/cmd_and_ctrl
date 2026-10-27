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
// all in one entry (game.ManifestFromHandsKeyedForEffect).
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
// (game.ChooseCardsRunThenKeyedForEffect): nothing moves until every player
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
	// that entered, in APNAP order. It is a registered continuation
	// (RegisterManifestThen, #4290), so a table waiting on a player's pick
	// is a restore point; its Context is rebuilt from the resolving item
	// as data.
	Then ManifestThen
}

// ManifestThen is a registered "for each card manifested this way" rest of
// the card. Hold one only from RegisterManifestThen; the zero value is no
// rest.
type ManifestThen struct{ run game.PromptRunThen }

// RegisterManifestThen registers the rest of a card after
// EachManifestsFromHand under key (ledgered "run <key>": never renamed,
// never reused). `body` (nil for none) is handed a Context rebuilt from the
// resolving item and the permanents that entered. Call it once, from a
// package-level var.
func RegisterManifestThen(key string, body func(ctx *Context, manifested []uuid.UUID) error) ManifestThen {
	// The entry's own continuation is a registered tail too (#4253 slice
	// 6), named after the run's key, so a table waiting on a manifested
	// card's entry question is a restore point.
	var tail game.TailRef
	if body != nil {
		tail = CardTail("manifest-from-hand/"+key, func(ctx *Context, o game.TailOutcome) error {
			return body(ctx, o.IDs)
		})
	}
	return ManifestThen{run: RunThen(key, func(ctx *Context, r game.RunDone) error {
		return manifestPicked(ctx, r, tail)
	})}
}

// manifestNoThen is the run of a manifest with no rest of the card.
var manifestNoThen = RegisterManifestThen("prompt-run/manifest-from-hand", nil)

// manifestPicked is the run's continuation: manifest every player's cards
// in one entry and run the card's rest. The carry is the players in APNAP
// order, then, for each player who manifests their whole hand without a
// choice, a list led by the player and followed by that hand.
func manifestPicked(ctx *Context, r game.RunDone, tail game.TailRef) error {
	lists := unpackIDLists(r.Carry)
	if len(lists) == 0 {
		// Nobody to manifest: the rest of the card still runs, on nothing.
		lists = [][]uuid.UUID{nil}
	}
	auto := map[uuid.UUID][]uuid.UUID{}
	for _, l := range lists[1:] {
		if len(l) > 0 {
			auto[l[0]] = l[1:]
		}
	}
	picks := r.Picks()
	var all []uuid.UUID
	for _, id := range lists[0] {
		if hand, ok := auto[id]; ok {
			all = append(all, hand...)
			continue
		}
		all = append(all, picks.By(id)...)
	}
	var then game.KeyedTail
	if tail.Key() != "" {
		then = ctx.ItemTail(tail)
	}
	if len(all) == 0 {
		return then.Run(ctx.Game, nil)
	}
	return ctx.Game.ManifestFromHandsKeyedForEffect(all, then)
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
	autos := [][]uuid.UUID{players}
	var legs []game.ChooseCardsPrompt
	for _, id := range players {
		hand := handCardIDs(g, id)
		switch {
		case len(hand) == 0 || n <= 0:
		case len(hand) <= n:
			autos = append(autos, append([]uuid.UUID{id}, hand...))
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
	then := e.Then.run
	if !then.Valid() {
		then = manifestNoThen.run
	}
	// Keyed (#4290): the players and the hands manifested without a choice
	// ride the run as its carry, the resolving item as its asking item.
	_, err := g.ChooseCardsRunThenKeyedForEffect(legs, then, source, ctx.RunData(packIDLists(autos...)))
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
