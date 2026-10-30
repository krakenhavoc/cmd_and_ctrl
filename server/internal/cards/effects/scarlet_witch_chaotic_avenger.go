package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scarlet Witch, Chaotic Avenger — Legendary Creature — Mutant Warlock
// Hero {2}{U}{R}, 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever Scarlet Witch deals combat damage to a player, look at the
//	 top two cards of your library, then exile them face down. Then you
//	 may cast a Hero or noncreature spell from among cards exiled with
//	 Scarlet Witch without paying its mana cost."
//
// The exile is Kayla's Music Box's: each card lands FaceDownHidden (its
// controller may look, CR 406.3, 601.3f) and carries Card.ExiledWith
// naming Scarlet Witch as the object she is now (CR 607.2a, 400.7), so a
// Scarlet Witch that left and came back has exiled nothing.
//
// "From among cards exiled with Scarlet Witch" is EVERY card she has
// exiled and that is still there, not just these two: the cast is one
// permission over all of them, so an earlier hit's Hero or noncreature
// card that was passed up is on offer again. It is ONE cast ("a spell"):
// the permission is pass-closed and spent whole by the first cast, and an
// uncast card stays in exile for the next hit.
//
// "A Hero or noncreature spell" describes the spell the cast makes, so it
// is the permission's filter, judged on the face being cast (#4509):
// PermissionFilter.NoncreatureOrCreatureType "Hero" admits an instant, an
// Aura or a land's Adventure as well as a Hero creature, and refuses a
// creature that is not a Hero. A land is played, never cast, so it is
// never offered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8c72d54e-97a6-408e-ae77-18d3c96b3bd8",
		Name:            "Scarlet Witch, Chaotic Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer(scarletWitchLabel, scarletWitchExileAndCast),
		},
	})
}

const (
	scarletWitchName  = "Scarlet Witch, Chaotic Avenger"
	scarletWitchLabel = scarletWitchName + " — look at the top two cards, exile them face down, then cast a Hero or noncreature spell from among cards exiled with her free"
)

// scarletWitchExileAndCast exiles the top two cards face down, linked to
// Scarlet Witch, then offers the free cast over everything exiled with her.
func scarletWitchExileAndCast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you := ctx.Controller()
	if you == uuid.Nil {
		return nil
	}
	source := HiddenRefOfActivation(item)
	var firstErr error
	for _, id := range g.LookAtTopOfLibraryForEffect(you, 2) {
		if _, err := g.ExileHiddenLinkedForEffect(source, you, id); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	perm := game.CastPermission{
		Player:      you,
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Filter:      game.PermissionFilter{NoncreatureOrCreatureType: "Hero"},
		Source:      ctx.Source(),
		SourceName:  scarletWitchName,
		Label:       scarletWitchName + " — cast a Hero or noncreature spell from among cards exiled with her without paying its mana cost",
	}
	var exiledWith []uuid.UUID
	for _, c := range g.ExiledWithForEffect(source) {
		exiledWith = append(exiledWith, c.InstanceID)
	}
	candidates := freeCastCandidates(g, &perm, exiledWith)
	if len(candidates) == 0 {
		return firstErr
	}
	if g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
		Chooser:      you,
		Source:       ctx.Source(),
		Reason:       scarletWitchName,
		Cards:        candidates,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: declineLeaveInExile,
	}) {
		return firstErr
	}
	// No resolution open (nothing in the catalog reaches this): the
	// permission alone is the window.
	var cards []game.Card
	for _, id := range candidates {
		if c, ok := g.LookupCardForEffect(id); ok {
			cards = append(cards, c)
		}
	}
	g.GrantCastPermissionToCardsForEffect(perm, cards)
	return firstErr
}
