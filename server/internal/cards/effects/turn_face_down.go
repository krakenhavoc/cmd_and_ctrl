package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// turn_face_down.go — #1209, ADR 0082's 2026-09-23 amendment: the one
// direction #1194 did not build.
//
// morph.go is about the keywords that put a card onto the battlefield
// already face down. This file is about the four printed cards that
// take a permanent already sitting there face up and turn it over:
// Ixidron, Backslide, Cyber Conversion and Master of the Veil.
//
// Nothing in here is a rule. The 2/2 body, the silenced text, the
// controller-only look, the counters and Auras that ride through, the
// CR 708.9 reveal on the way out and CR 708.7's answer to "can it come
// back up" are all the engine's, behind game.TurnFaceDownForEffect.
// What a card file says is only WHICH permanents and WHO did it.

// TurnFaceDown turns the named battlefield permanents face down
// (CR 708.2a). `Source` is the object doing it — the resolving spell,
// the permanent whose trigger or "as this enters" clause it is — and
// goes on the event so the log can say who.
//
// A permanent that is not on the battlefield any more is skipped, and
// so is one the rules refuse: CR 708.2b (a face-down permanent can't
// be turned face down) and CR 712.16 (nor can a double-faced one).
// Both are "nothing happens" in the rules, so neither is an error
// here — a spell whose only target turned out to be one of them has
// still resolved.
type TurnFaceDown struct {
	Source  uuid.UUID
	Targets []uuid.UUID
	// Listed is the body the card LISTS for what it turns over
	// (CR 708.2) — CybermanBody() for Cyber Conversion's "It's a 2/2
	// Cyberman artifact creature". nil is CR 708.2a's default
	// nameless 2/2, which is Ixidron, Backslide and Master of the
	// Veil. #1270.
	Listed *game.FaceDownListing
}

func (t TurnFaceDown) Apply(ctx *Context) error {
	targets := ctx.withoutNewSourceObject(t.Targets) // #1432
	if len(targets) == 0 {
		return nil
	}
	ctx.Game.TurnFaceDownListedForEffect(t.Source, t.Listed, targets...)
	return nil
}

// CybermanBody is the face-down body the Doctor Who Cyberman cards
// list — "It's a 2/2 Cyberman artifact creature" (Cyber Conversion,
// Missy) and "They're 2/2 Cyberman artifact creatures" (Cybership,
// The Cyber-Controller). CR 708.2: a listing REPLACES the default
// body, so this is the whole object — no name, no text, no colour.
//
// A fresh value per call, because the engine copies it anyway and a
// shared package-level pointer would be one careless write away from
// turning every Cyberman into something else.
func CybermanBody() *game.FaceDownListing {
	return &game.FaceDownListing{
		Types:     []string{"Artifact", "Creature"},
		Subtypes:  []string{"Cyberman"},
		Power:     2,
		Toughness: 2,
	}
}

// ForestLandBody is Yedora, Grave Gardener's "It's a Forest land. (It
// has no other types or abilities.)" — not a creature, no P/T. The
// "{T}: Add {G}" it has is CR 305.6's, which the engine derives from
// the Forest subtype; the listing does not have to say it.
func ForestLandBody() *game.FaceDownListing {
	return &game.FaceDownListing{Types: []string{"Land"}, Subtypes: []string{"Forest"}}
}

// turnSingleTargetFaceDown is the body of every "turn target creature
// … face down" card: read the one target, turn it over. Shared by
// Backslide as an OnResolve and by Master of the Veil as a trigger
// Effect, which is why it takes the item rather than the Context.
func turnSingleTargetFaceDown(g *game.Game, item *game.StackItem) error {
	return turnSingleTargetFaceDownAs(nil)(g, item)
}

// turnSingleTargetFaceDownAs is turnSingleTargetFaceDown with a listed
// body — Cyber Conversion's. nil is the CR 708.2a default.
func turnSingleTargetFaceDownAs(listed func() *game.FaceDownListing) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		if item == nil || len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
			return nil
		}
		t := TurnFaceDown{
			Source:  item.SourceCardID,
			Targets: []uuid.UUID{item.Targets[0].ID},
		}
		if listed != nil {
			t.Listed = listed()
		}
		return t.Apply(NewContext(g, item))
	}
}

// otherNontokenCreaturesOnBattlefield is Ixidron's "all other
// nontoken creatures" (CR 708.2a), collected against the board as it
// stands BEFORE anything turns over.
//
// "Nontoken" is Ixidron's own word and not a rule: nothing in CR 708
// or CR 111 stops a token being turned face down, and Cyber
// Conversion may hit one. The exclusion is on this card because the
// card prints it.
func otherNontokenCreaturesOnBattlefield(g *game.Game, self uuid.UUID) []uuid.UUID {
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == self || c.IsToken() || !c.IsCreature() {
			continue
		}
		ids = append(ids, c.InstanceID)
	}
	return ids
}

// faceDownCreaturesOnBattlefield counts the CR 708.2 objects that are
// creatures — Ixidron's own size, and the only reason a card needs to
// count them.
//
// It asks FaceDownIsPermanent rather than Card.FaceDown, because a
// card face down in EXILE (a foretold card, a Necropotence exile) is
// not a permanent and has no 2/2 body to count; and it asks
// IsCreature as well, because "face-down creatures" is what Ixidron
// says and an effect could in principle have made one of them
// something else.
func faceDownCreaturesOnBattlefield(g *game.Game) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.FaceDownIsPermanent() && c.IsCreature() {
			n++
		}
	}
	return n
}
