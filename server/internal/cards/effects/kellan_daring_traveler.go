package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kellan, Daring Traveler // Journey On — Legendary Creature — Human
// Faerie Scout {1}{W}, 2/3 // Sorcery — Adventure {G} (The Lost Caverns of
// Ixalan, #4470):
//
//	Kellan — "Whenever Kellan attacks, reveal the top card of your
//	 library. If it's a creature card with mana value 3 or less, put it
//	 into your hand. Otherwise, you may put it into your graveyard."
//	Journey On — "Create X Map tokens, where X is one plus the number of
//	 opponents who control an artifact. (Then exile this card. You may
//	 cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"), as Bellowing Bruiser's do.
//
// The attack trigger reveals the top card to every player. A creature
// card with mana value 3 or less goes to your hand, with no choice.
// Anything else is Bison Whistle's last sentence: you may put it into
// your graveyard (a plain move, not a mill), and declining leaves it on
// top. A card that left the library before the answer is a new object
// and stays where it is (CR 608.2b).
//
// Journey On counts, as it resolves, the opponents still in the game who
// control an artifact, and makes one Map more than that (the catalog's
// Map token, CR 111.10s).
//
// No simplification.
const kellanDaringTravelerOracleID = "7af69a44-e958-42e5-9d65-75e53a601e96"

func init() {
	Register(Spec{
		OracleID:     kellanDaringTravelerOracleID,
		Name:         "Kellan, Daring Traveler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kellan, Daring Traveler — reveal the top card of your library", kellanRevealsTheTop),
		},
	})
	Register(Spec{
		OracleID:     kellanDaringTravelerOracleID + "#1",
		Name:         "Journey On",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 1
			for _, opp := range ctx.Opponents() {
				if cloudsculptControlsAnArtifact(ctx.Game, opp) {
					n++
				}
			}
			return CreateToken{Template: MapToken(), N: n}.Apply(ctx)
		},
	})
}

// kellanRevealsTheTop is the attack trigger's resolution.
func kellanRevealsTheTop(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	revealed := g.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 1, "Kellan, Daring Traveler — reveal the top card")
	if len(revealed) == 0 {
		return nil
	}
	top, ok := g.LookupCardForEffect(revealed[0])
	if !ok {
		return nil
	}
	if top.IsCreature() && top.ManaValue() <= 3 {
		return g.BounceToHandForEffect(top.InstanceID)
	}
	return MayChoice{
		Player:   item.Controller,
		Question: "Kellan, Daring Traveler — put " + top.Name + " into your graveyard?",
		YesLabel: "Put it into your graveyard",
		NoLabel:  "Leave it on top",
		Then:     kellanToGraveyard,
		Carry:    []uuid.UUID{top.InstanceID},
	}.Apply(ctx)
}

// kellanToGraveyard is the "yes": the revealed card goes to the graveyard
// if it is still in the library. Keyed: never renamed, never reused.
var kellanToGraveyard = MayYes("confirm/kellan-daring-traveler-to-graveyard", putCarriedLibraryCardIntoGraveyard)
