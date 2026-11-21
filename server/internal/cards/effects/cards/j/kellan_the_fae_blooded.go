package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kellan, the Fae-Blooded // Birthright Boon — Legendary Creature — Human
// Faerie {2}{R}, 2/2 // Sorcery — Adventure {1}{W} (Wilds of Eldraine,
// #4471):
//
//	Kellan, the Fae-Blooded — "Double strike
//	 Other creatures you control get +1/+0 for each Aura and Equipment
//	 attached to Kellan."
//	Birthright Boon — "Search your library for an Aura or Equipment card,
//	 reveal it, put it into your hand, then shuffle. (Then exile this
//	 card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID
// and "#1"). The anthem is a layer 7c static (CR 613.4c) over the other
// creatures Kellan's controller controls, counting at every recompute the
// Auras and Equipment attached to Kellan himself (Thran Power Suit's count,
// read off the source rather than the host), so an attach or unattach
// moves the bonus in the same pass and Kellan alone gets nothing.
// Birthright Boon is Open the Armory's tutor; the Adventure lifecycle is
// the engine's (CR 715.3d).
//
// No simplification.
const KellanTheFaeBloodedOracleID = "10c48d19-d91f-4674-b186-477acaf08e5a"

func init() {
	effects.Register(effects.Spec{
		OracleID:        KellanTheFaeBloodedOracleID,
		Name:            "Kellan, the Fae-Blooded",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static: []game.StaticAbility{{
			Label:     "Other creatures you control get +1/+0 for each Aura and Equipment attached to Kellan.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.OtherCreaturesYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if g == nil || source == nil {
					return
				}
				for i := range g.Battlefield.Cards {
					att := &g.Battlefield.Cards[i]
					if att.IsAttachedTo(source.InstanceID) && (att.HasSubtype("Aura") || att.HasSubtype("Equipment")) {
						c.Power++
					}
				}
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     KellanTheFaeBloodedOracleID + "#1",
		Name:         "Birthright Boon",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: func(c game.Card) bool { return c.HasSubtype("Aura") || c.HasSubtype("Equipment") },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Birthright Boon — an Aura or Equipment card",
			}.Apply(ctx)
		},
	})
}
