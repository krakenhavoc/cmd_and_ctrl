package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cemetery Illuminator — Creature — Spirit {1}{U}{U}, 2/3 (Innistrad:
// Crimson Vow):
//
//	"Flying
//	 Whenever this creature enters or attacks, exile a card from a
//	 graveyard.
//	 You may look at the top card of your library any time.
//	 Once each turn, you may cast a spell from the top of your library if
//	 it shares a card type with a card exiled with this creature."
//
// The exile is linked to the one enters-or-attacks trigger (CR 607.2a).
// The cast permission is one gated top-of-library permission per card type
// a spell can have, each open while this creature has an exiled card of that
// type; they share one once-each-turn tally, so a spell of two shared types
// or casts under two permissions still use the turn's one cast. A land
// card cannot be cast, so no land permission exists. The spell is cast for
// its normal cost.
//
// No simplification.
func init() {
	gated := make([]game.CastPermissionGate, 0, len(cemeteryIlluminatorTypes))
	for _, t := range cemeteryIlluminatorTypes {
		gated = append(gated, game.CastPermissionGate{
			Condition: cemeteryIlluminatorOpen(t),
			Permission: game.CastPermission{
				Zone:             game.ZoneLibrary,
				TopOfLibraryOnly: true,
				CastOnly:         true,
				Filter:           game.PermissionFilter{CardType: t},
				FollowUp:         cemeteryIlluminatorTally,
				Label:            "Once each turn, cast a " + t + " spell from the top of your library if it shares a card type with a card exiled with this creature (Cemetery Illuminator)",
			},
		})
	}
	effects.Register(effects.Spec{
		OracleID:          "14fe4531-96cd-415b-aa4a-b41fd271d902",
		Name:              "Cemetery Illuminator",
		Completeness:      effects.CompletenessFull,
		PrintedKeywords:   []string{"flying"},
		LibraryTopVisible: game.LibraryTopOwner,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks(cemeteryIlluminatorExileLabel, func(g *game.Game, item *game.StackItem) error {
				effects.VowB16QueueExile(g, item, cemeteryIlluminatorExileLabel, uuid.Nil, cemeteryIlluminatorExiled)
				return nil
			}),
		},
		GatedCastPermissions: gated,
	})
}

const (
	cemeteryIlluminatorExileLabel = "Cemetery Illuminator — exile a card from a graveyard"
	cemeteryIlluminatorDoneLabel  = "Cemetery Illuminator — cast from the top of the library (done)"
)

// cemeteryIlluminatorTypes are the card types a castable spell can have.
var cemeteryIlluminatorTypes = []string{"artifact", "creature", "enchantment", "instant", "sorcery", "planeswalker", "battle", "kindred"}

var cemeteryIlluminatorExiled = effects.VowB16LinkedExileThen("choose-cards/cemetery-illuminator-exile")

var cemeteryIlluminatorTally = game.RegisterCastFollowUp("cemetery-illuminator-once-each-turn",
	func(g *game.Game, f game.CastFollowUp) error {
		g.NoteDoneThisTurn(f.Source, cemeteryIlluminatorDoneLabel)
		return nil
	})

// cemeteryIlluminatorOpen is the gate for one card type: the turn's cast is
// unused and a card of that type is exiled with this creature.
func cemeteryIlluminatorOpen(cardType string) func(g *game.Game, controller, source uuid.UUID) bool {
	return func(g *game.Game, _, source uuid.UUID) bool {
		if g.ResolvedThisTurn(source, cemeteryIlluminatorDoneLabel) != 0 {
			return false
		}
		self, ok := g.LookupCardForEffect(source)
		if !ok {
			return false
		}
		for _, c := range effects.VowB16ExiledWithCard(g, &self, cemeteryIlluminatorExileLabel) {
			if c.HasCardType(cardType) {
				return true
			}
		}
		return false
	}
}
